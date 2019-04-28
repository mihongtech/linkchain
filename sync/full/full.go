package full

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/mihongtech/linkchain/common/util/event"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/node"
	p2p_node "github.com/mihongtech/linkchain/p2p/discover"
	"github.com/mihongtech/linkchain/p2p/message"
	p2p_peer "github.com/mihongtech/linkchain/p2p/peer"
	"github.com/mihongtech/linkchain/p2p/peer_error"
	"github.com/mihongtech/linkchain/protobuf"
	"github.com/mihongtech/linkchain/sync/full/downloader"
	"github.com/mihongtech/linkchain/sync/full/fetcher"
	"github.com/mihongtech/linkchain/txpool"
)

// errIncompatibleConfig is returned if the requested protocols and configs are
// not compatible (low protocol version restrictions and high requirements).
var errIncompatibleConfig = errors.New("incompatible configuration")

const (
	txChanSize = 4096
)

type ProtocolManager struct {
	networkId uint64
	maxPeers  int
	peers     *peerSet

	downloader *downloader.Downloader
	fetcher    *fetcher.Fetcher

	SubProtocols  []p2p_peer.Protocol
	eventMux      *event.TypeMux
	eventTx       *event.Feed
	scope         event.SubscriptionScope
	txCh          chan node.TxEvent
	txSub         event.Subscription
	minedBlockSub *event.TypeMuxSubscription

	nodeAPI   *node.PublicNodeAPI
	txPoolAPI *txpool.TxPool

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan *peer
	txsyncCh    chan *txsync
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg sync.WaitGroup
}

// NewProtocolManager returns a new linkchain sub protocol manager. The Linkchain sub protocol manages peers capable
// with the linkchain network.
func NewProtocolManager(config interface{}, nodeSvc *node.PublicNodeAPI, txPoolSvc *txpool.TxPool, networkId uint64, mux *event.TypeMux, tx *event.Feed) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		networkId:   networkId,
		maxPeers:    64,
		eventTx:     tx,
		eventMux:    mux,
		peers:       newPeerSet(),
		newPeerCh:   make(chan *peer),
		noMorePeers: make(chan struct{}),
		nodeAPI:     nodeSvc,
		txPoolAPI:   txPoolSvc,
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
	}

	// Initiate a sub-protocol for every implemented version we can handle
	manager.SubProtocols = make([]p2p_peer.Protocol, 0, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		// Compatible; initialise the sub-protocol
		version := version // Closure for the run
		manager.SubProtocols = append(manager.SubProtocols, p2p_peer.Protocol{
			Name:    ProtocolName,
			Version: version,
			Length:  ProtocolLengths[i],
			Run: func(p *p2p_peer.Peer, rw message.MsgReadWriter) error {
				peer := manager.newPeer(int(version), p, rw)
				select {
				case manager.newPeerCh <- peer:
					manager.wg.Add(1)
					defer manager.wg.Done()
					return manager.handle(peer)
				case <-manager.quitSync:
					return peer_error.DiscQuitting
				}
			},
			NodeInfo: func() interface{} {
				return manager.NodeInfo()
			},
			PeerInfo: func(id p2p_node.NodeID) interface{} {
				if p := manager.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
					return p.Info()
				}
				return nil
			},
		})
	}
	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}

	manager.downloader = downloader.New(manager.eventMux, manager.nodeAPI, manager.removePeer)

	heighter := func() uint64 {
		return uint64(manager.nodeAPI.GetBestBlock().GetHeight())
	}
	validator := func(block *meta.Block) error {
		return manager.nodeAPI.CheckBlock(block)
	}
	manager.fetcher = fetcher.New(manager.nodeAPI.GetBlockByID, validator, manager.BroadcastBlock, heighter, manager.nodeAPI.ProcessBlock, manager.removePeer)

	return manager, nil
}

func (pm *ProtocolManager) Start() bool {
	// broadcast transactions
	pm.txCh = make(chan node.TxEvent, txChanSize)
	pm.txSub = pm.scope.Track(pm.eventTx.Subscribe(pm.txCh))
	go pm.txBroadcastLoop()
	//
	//	 broadcast mined blocks
	pm.minedBlockSub = pm.eventMux.Subscribe(node.NewMinedBlockEvent{})
	go pm.minedBroadcastLoop()
	//
	//	 start sync handlers
	go pm.syncer()
	go pm.txsyncLoop()
	return true
}

func (pm *ProtocolManager) Stop() {
	log.Info("Stopping Linkchain protocol")
	pm.scope.Close()
	pm.txSub.Unsubscribe()         // quits txBroadcastLoop
	pm.minedBlockSub.Unsubscribe() // quits blockBroadcastLoop

	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	pm.noMorePeers <- struct{}{}

	// Quit fetcher, txsyncLoop.
	close(pm.quitSync)

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	pm.peers.Close()

	// Wait for all peer handler goroutines and the loops to come down.
	pm.wg.Wait()

	log.Info("Linkchain protocol stopped")
}

// handle is the callback invoked to manage the life cycle of an full peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		return peer_error.DiscTooManyPeers
	}
	p.Log().Trace("Linkchain peer connected", "name", p.Name())

	// Execute the Linkchain handshake
	var (
		genesis, _ = pm.nodeAPI.GetBlockByHeight(0)
		current    = pm.nodeAPI.GetBestBlock()
		hash       = *current.GetBlockID()
		number     = uint64(current.GetHeight())
	)
	p.Log().Debug("Linkchain handshake data", "genesis", genesis, "number", current.GetHeight(), "current", hash)
	if err := p.Handshake(pm.networkId, number, hash, *genesis.GetBlockID()); err != nil {
		p.Log().Debug("Linkchain handshake failed", "err", err)
		return err
	}

	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		p.Log().Error("Linkchain peer registration failed", "err", err)
		return err
	}
	defer pm.removePeer(p.id)

	// Register the peer in the downloader. If the downloader considers it banned, we disconnect
	if err := pm.downloader.RegisterPeer(p.id, p.version, p); err != nil {
		return err
	}
	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	// pm.syncTransactions(p)

	// main loop. handle incoming messages.

	for {
		if err := pm.handleMsg(p); err != nil {
			p.Log().Debug("Linkchain message handling failed", "err", err)
			return err
		}
	}

	return nil
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	defer msg.Discard()

	// Handle the message depending on its contents
	switch {
	case msg.Code == StatusMsg:
		log.Error("uncontrolled status message")
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	// Block header query, collect the requested headers and reply
	case msg.Code == GetBlockMsg:
		// Decode the complex header query

		var query protobuf.GetBlockHeadersData
		if err := msg.Decode(&query); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		data := &getBlockHeadersData{}
		data.Deserialize(&query)

		var (
			blocks  []*meta.Block
			unknown bool
		)
		for !unknown && len(blocks) < int(data.Amount) && len(blocks) < downloader.MaxBlockFetch {
			// Retrieve the next header satisfying the query
			var block *meta.Block
			var err error
			if data.Hash.IsEmpty() {
				block, err = pm.nodeAPI.GetBlockByHeight(uint32(data.Number))
				log.Debug("get block by height", "number", data.Number, "block", block)
			} else {
				block, err = pm.nodeAPI.GetBlockByID(data.Hash)
				log.Debug("get block by id", "Hash", data.Hash, "block", block)
			}
			if err != nil || block == nil {
				log.Debug("get block msg error", "query data", data, "err", err)
				break
			}
			// number := uint64(block.GetHeight())
			blocks = append(blocks, block)

			// Advance to the next header of the query
			switch {
			case !data.Hash.IsEmpty():
				// Hash based traversal towards the leaf block
				var (
					current = uint64(block.GetHeight())
					next    = current + data.Skip + 1
				)
				if next <= current {
					infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
					p.Log().Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip, "next", next, "attacker", infos)
					unknown = true
				} else {
					if b, e := pm.nodeAPI.GetBlockByHeight(uint32(next)); (b != nil) && (e == nil) {
						log.Debug("get block by height", "number", current, "skip", data.Skip, "next", next)
						data.Hash.SetBytes(b.GetBlockID().CloneBytes())
					} else {
						unknown = true
					}

				}
			case data.Hash.IsEmpty():
				// Number based traversal towards the leaf block
				data.Number += data.Skip + 1
			}
		}
		for i, b := range blocks {
			log.Debug("Receive GetBlockMsg", "query is", data, "index", i, "block", b)
		}

		p.SendBlock(blocks)

		return nil

	case msg.Code == BlockMsg:

		blocks := []*meta.Block{}
		var b protobuf.Blocks
		if err := msg.Decode(&b); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		for _, prob := range b.Block {
			data := &meta.Block{}
			data.Deserialize(prob)
			blocks = append(blocks, data)
		}

		log.Debug("Receive BlockMsg", "len(block) is", len(blocks))
		for i, b := range blocks {
			log.Debug("Receive BlockMsg", "index", i, "block", b)
		}
		filter := len(blocks) == 1
		if filter {
			blocks = pm.fetcher.FilterBlocks(p.id, blocks, time.Now())
		}
		if len(blocks) > 0 || !filter {
			err := pm.downloader.DeliverBlocks(p.id, blocks)
			if err != nil {
				log.Debug("Failed to deliver blocks", "err", err)
			}
		}
		pm.downloader.ImportBlocks(p.id, blocks)

	case msg.Code == NewBlockMsg:
		// Retrieve and decode the propagated block
		var b protobuf.Block
		if err := msg.Decode(&b); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		block := &meta.Block{}
		block.Deserialize(&b)

		// Mark the peer as owning the block and schedule it for import
		p.MarkBlock(*block.GetBlockID())
		pm.fetcher.Enqueue(p.id, block)

		var (
			trueHead = *block.GetPrevBlockID()
		)
		log.Debug("Receive NewBlockMsg", "block is", block)
		p.SetHead(trueHead, uint64(block.GetHeight()))

		go pm.synchronise(p)

	case msg.Code == TxMsg:

		// TODO: add interface
		var t protobuf.Transaction
		if err := msg.Decode(&t); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		transaction := &meta.Transaction{}
		transaction.Deserialize(&t)
		p.MarkTransaction(*transaction.GetTxID())
		log.Debug("Receive TxMsg", "transaction is", transaction)
		if err = pm.txPoolAPI.ProcessTx(transaction); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		pm.BroadcastTx(*transaction.GetTxID(), transaction)
		//		for _, t := range pm.txmanager.getAllTransaction() {
		//			log.Debug("all txs is", "tx", t)
		//		}

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

func (pm *ProtocolManager) newPeer(pv int, p *p2p_peer.Peer, rw message.MsgReadWriter) *peer {
	return newPeer(pv, p, rw)
}

func (pm *ProtocolManager) removePeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing Linkchain peer", "peer", id)

	// Unregister the peer from the downloader and Linkchain peer set
	pm.downloader.UnregisterPeer(id)
	if err := pm.peers.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(peer_error.DiscUselessPeer)
	}
}

// NodeInfo represents a short summary of the Linkchain sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network uint64       `json:"network"` // Linkchain network ID (1=Frontier, 2=Morden, Ropsten=3, Rinkeby=4)
	Genesis meta.BlockID `json:"genesis"` // hash of the host's genesis block
	Head    meta.BlockID `json:"head"`    // hash of the host's best owned block
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (pm *ProtocolManager) NodeInfo() *NodeInfo {
	genesis, _ := pm.nodeAPI.GetBlockByHeight(0)
	return &NodeInfo{
		Network: pm.networkId,
		Genesis: *genesis.GetBlockID(),
		Head:    *pm.nodeAPI.GetBestBlock().GetBlockID(),
	}
}

func (pm *ProtocolManager) txBroadcastLoop() {
	for {
		select {
		case event := <-pm.txCh:
			pm.BroadcastTx(*event.Tx.GetTxID(), event.Tx)

			// Err() channel will be closed when unsubscribing.
		case <-pm.txSub.Err():
			return
		}
	}
}

// Mined broadcast loop
func (pm *ProtocolManager) minedBroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range pm.minedBlockSub.Chan() {
		switch ev := obj.Data.(type) {
		case node.NewMinedBlockEvent:
			pm.BroadcastBlock(ev.Block, true)  // First propagate block to peers
			pm.BroadcastBlock(ev.Block, false) // Only then announce to the rest
		}
	}
}

// BroadcastBlock will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastBlock(block *meta.Block, propagate bool) {
	hash := *block.GetBlockID()
	peers := pm.peers.PeersWithoutBlock(hash)

	// If propagation is requested, send to a subset of the peer
	if propagate {
		// Send the block to a subset of our peers
		transfer := peers[:int(math.Sqrt(float64(len(peers))))]
		for _, peer := range transfer {
			peer.SendNewBlock(block)
		}
		log.Trace("Propagated block", "hash", hash, "recipients", len(transfer))
		return
	}
	// Otherwise if the block is indeed in out own chain, announce it
	if pm.nodeAPI.HasBlock(hash) {
		for _, peer := range peers {
			peer.SendNewBlock(block)
			// peer.SendNewBlockHashes([]meta.DataID{hash}, []uint64{uint64(block.GetHeight())})
		}
		log.Trace("Announced block", "hash", hash, "recipients", len(peers))
	}
}

// BroadcastTx will propagate a transaction to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTx(hash meta.TxID, t *meta.Transaction) {
	// Broadcast transaction to a batch of peers not knowing about it
	peers := pm.peers.PeersWithoutTx(hash)
	//FIXME include this again: peers = peers[:int(math.Sqrt(float64(len(peers))))]
	for _, peer := range peers {
		peer.SendTransactions([]*meta.Transaction{t})
	}
	log.Trace("Broadcast transaction", "hash", hash, "recipients", len(peers))
}
