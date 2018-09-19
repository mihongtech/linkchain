package full

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/p2p/message"
	p2p_peer "github.com/linkchain/p2p/peer"
	"github.com/linkchain/p2p/peer_error"
	"github.com/linkchain/protobuf"
	"gopkg.in/fatih/set.v0"
)

var (
	errClosed            = errors.New("peer set is closed")
	errAlreadyRegistered = errors.New("peer is already registered")
	errNotRegistered     = errors.New("peer is not registered")
)

const (
	maxKnownTxs      = 32768 // Maximum transactions hashes to keep in the known list (prevent DOS)
	maxKnownBlocks   = 1024  // Maximum block hashes to keep in the known list (prevent DOS)
	handshakeTimeout = 5 * time.Second
)

// PeerInfo represents a short summary of the Linkchain sub-protocol metadata known
// about a connected peer.
type PeerInfo struct {
	Version int    `json:"version"` // Linkchain protocol version negotiated
	Head    string `json:"head"`    // SHA3 hash of the peer's best owned block
}

type peer struct {
	id string

	*p2p_peer.Peer
	rw message.MsgReadWriter

	version  int         // Protocol version negotiated
	forkDrop *time.Timer // Timed connection dropper if forks aren't validated in time

	head math.Hash
	lock sync.RWMutex

	knownTxs    set.Interface // Set of transaction hashes known to be known by this peer
	knownBlocks set.Interface // Set of block hashes known to be known by this peer
}

func newPeer(version int, p *p2p_peer.Peer, rw message.MsgReadWriter) *peer {
	id := p.ID()

	return &peer{
		Peer:        p,
		rw:          rw,
		version:     version,
		id:          fmt.Sprintf("%x", id[:8]),
		knownTxs:    set.New(set.ThreadSafe),
		knownBlocks: set.New(set.ThreadSafe),
	}
}

// Info gathers and returns a collection of metadata known about a peer.
func (p *peer) Info() *PeerInfo {
	hash := p.Head()

	return &PeerInfo{
		Version: p.version,
		Head:    hash.GetString(),
	}
}

// Head retrieves a copy of the current head hash and total difficulty of the
// peer.
func (p *peer) Head() (hash meta.DataID) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	// copy(hash[:], p.head[:])
	hash = &math.Hash{}
	hash.SetBytes(p.head[:])

	return hash
}

// SetHead updates the head hash and total difficulty of the peer.
func (p *peer) SetHead(hash meta.DataID) {
	p.lock.Lock()
	defer p.lock.Unlock()

	copy(p.head[:], hash.CloneBytes())
	// p.head.SetBytes(hash.CloneBytes())
}

// MarkBlock marks a block as known for the peer, ensuring that the block will
// never be propagated to this particular peer.
func (p *peer) MarkBlock(hash meta.DataID) {
	// If we reached the memory allowance, drop a previously known block hash
	for p.knownBlocks.Size() >= maxKnownBlocks {
		p.knownBlocks.Pop()
	}
	p.knownBlocks.Add(hash)
}

// MarkTransaction marks a transaction as known for the peer, ensuring that it
// will never be propagated to this particular peer.
func (p *peer) MarkTransaction(hash meta.DataID) {
	// If we reached the memory allowance, drop a previously known transaction hash
	for p.knownTxs.Size() >= maxKnownTxs {
		p.knownTxs.Pop()
	}
	p.knownTxs.Add(hash)
}

// SendTransactions sends transactions to the peer and includes the hashes
// in its transaction hash set for future reference.
func (p *peer) SendTransactions(txs []tx.ITx) error {
	for _, tx := range txs {
		p.knownTxs.Add(tx.GetTxID())
		log.Debug("Send TxMsg", "transaction is", tx)
		message.Send(p.rw, TxMsg, tx.Serialize())
	}
	return nil
}

// SendNewBlock propagates an entire block to a remote peer.
func (p *peer) SendNewBlock(block block.IBlock) error {
	p.knownBlocks.Add(block.GetBlockID())
	log.Debug("Send NewBlockMsg", "block is", block)
	return message.Send(p.rw, NewBlockMsg, block.Serialize())
}

func (p *peer) SendBlock(block block.IBlock) error {
	p.knownBlocks.Add(block.GetBlockID())
	log.Debug("Send BlockMsg", "block is", block)
	return message.Send(p.rw, BlockMsg, block.Serialize())
}

// RequestBodies fetches a batch of blocks' bodies corresponding to the hashes
// specified.
func (p *peer) RequestBlock(hashes []meta.DataID) error {
	p.Log().Debug("Fetching batch of block bodies", "count", len(hashes))
	for _, hash := range hashes {
		data := &getBlockHeadersData{Hash: hash}
		log.Debug("Send GetBlockMsg", "query data is", data)
		message.Send(p.rw, GetBlockMsg, data.Serialize().(*protobuf.GetBlockHeadersData))
	}

	return nil
}

func (p *peer) RequestOneBlock(hash meta.DataID) error {
	p.Log().Debug("Fetching single block", "hash", hash)
	data := &getBlockHeadersData{Hash: hash}
	log.Debug("Send GetBlockMsg", "query data is", data)
	return message.Send(p.rw, GetBlockMsg, data.Serialize().(*protobuf.GetBlockHeadersData))
}

// Handshake executes the eth protocol handshake, negotiating version number,
// network IDs, difficulties, head and genesis blocks.
func (p *peer) Handshake(network uint64, head meta.DataID, genesis meta.DataID) error {
	// Send out own handshake in a new thread
	errc := make(chan error, 2)
	var status statusData // safe to read after two values have been received from errc

	go func() {
		data := &statusData{
			ProtocolVersion: uint32(p.version),
			NetworkId:       network,
			CurrentBlock:    head,
			GenesisBlock:    genesis,
		}
		log.Debug("Send StatusMsg", "data is", data)
		errc <- message.Send(p.rw, StatusMsg, data.Serialize())
	}()
	go func() {
		errc <- p.readStatus(network, &status, genesis)
	}()
	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-errc:
			if err != nil {
				return err
			}
		case <-timeout.C:
			return peer_error.DiscReadTimeout
		}
	}
	p.SetHead(status.CurrentBlock)
	// copy(p.head[:], status.CurrentBlock.CloneBytes())
	return nil
}

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

func (p *peer) readStatus(network uint64, status *statusData, genesis meta.DataID) (err error) {
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Code != StatusMsg {
		return errResp(ErrNoStatusMsg, "first msg has code %x (!= %x)", msg.Code, StatusMsg)
	}
	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	// Decode the handshake and make sure everything matches
	data := protobuf.StatusData{}
	if err := msg.Decode(&data); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	log.Debug("read status data is", "data", data, "current status is", status)
	status.Deserialize(&data)
	if !status.GenesisBlock.IsEqual(genesis) {
		return errResp(ErrGenesisBlockMismatch, "%x (!= %x)", status.GenesisBlock, genesis)
	}
	if status.NetworkId != network {
		return errResp(ErrNetworkIdMismatch, "%d (!= %d)", status.NetworkId, network)
	}
	if int(status.ProtocolVersion) != p.version {
		return errResp(ErrProtocolVersionMismatch, "%d (!= %d)", status.ProtocolVersion, p.version)
	}
	return nil
}

// String implements fmt.Stringer.
func (p *peer) String() string {
	return fmt.Sprintf("Peer %s [%s]", p.id,
		fmt.Sprintf("eth/%2d", p.version),
	)
}

func (p *peer) RequestBlocksByHash(h meta.DataID, amount int, skip int, reverse bool) error {
	p.Log().Debug("Fetching block by hash", "hash", h)
	data := &getBlockHeadersData{Hash: h, Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse}
	log.Debug("Send GetBlockMsg", "query data is", data)
	return message.Send(p.rw, GetBlockMsg, data.Serialize().(*protobuf.GetBlockHeadersData))
}
func (p *peer) RequestBlocksByNumber(i uint64, amount int, skip int, reverse bool) error {
	p.Log().Debug("Fetching block by number", "number", i)
	data := &getBlockHeadersData{Number: i, Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse}
	log.Debug("Send GetBlockMsg", "query data is", data)
	return message.Send(p.rw, GetBlockMsg, data.Serialize().(*protobuf.GetBlockHeadersData))
}

func (p *peer) SendNewBlockHashes(hashes []meta.DataID, numbers []uint64) error {
	for _, hash := range hashes {
		p.knownBlocks.Add(hash)
	}
	msg := make([]*protobuf.NewBlockHashData, 0, len(hashes))
	for i := 0; i < len(hashes); i++ {
		request := &newBlockHashData{
			hashes[i],
			numbers[i],
		}
		msg = append(msg, request.Serialize().(*protobuf.NewBlockHashData))
	}
	log.Debug("Send NewBlockHashesMsg", "block hash is", msg)
	return message.Send(p.rw, NewBlockHashesMsg, &(protobuf.NewBlockHashesDatas{Data: msg}))
}

// peerSet represents the collection of active peers currently participating in
// the Ethereum sub-protocol.
type peerSet struct {
	peers  map[string]*peer
	lock   sync.RWMutex
	closed bool
}

// newPeerSet creates a new peer set to track the active participants.
func newPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[string]*peer),
	}
}

// Register injects a new peer into the working set, or returns an error if the
// peer is already known.
func (ps *peerSet) Register(p *peer) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.closed {
		return errClosed
	}
	if _, ok := ps.peers[p.id]; ok {
		return errAlreadyRegistered
	}
	ps.peers[p.id] = p
	return nil
}

// Unregister removes a remote peer from the active set, disabling any further
// actions to/from that particular entity.
func (ps *peerSet) Unregister(id string) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if _, ok := ps.peers[id]; !ok {
		return errNotRegistered
	}
	delete(ps.peers, id)
	return nil
}

// Peer retrieves the registered peer with the given id.
func (ps *peerSet) Peer(id string) *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return ps.peers[id]
}

// Len returns if the current number of peers in the set.
func (ps *peerSet) Len() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.peers)
}

// PeersWithoutBlock retrieves a list of peers that do not have a given block in
// their set of known hashes.
func (ps *peerSet) PeersWithoutBlock(hash meta.DataID) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.knownBlocks.Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

// PeersWithoutTx retrieves a list of peers that do not have a given transaction
// in their set of known hashes.
func (ps *peerSet) PeersWithoutTx(hash meta.DataID) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.knownTxs.Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

// BestPeer retrieves the known peer with the currently highest total difficulty.
func (ps *peerSet) BestPeer() *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	var (
		bestPeer *peer
	)
	for _, p := range ps.peers {
		bestPeer = p
	}
	return bestPeer
}

// Close disconnects all peers.
// No new peers can be registered after Close has returned.
func (ps *peerSet) Close() {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, p := range ps.peers {
		p.Disconnect(peer_error.DiscQuitting)
	}
	ps.closed = true
}
