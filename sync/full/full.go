package full

import (
	"errors"
	"fmt"
	_ "math"
	_ "math/big"
	"sync"
	_ "time"

	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta/block"
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/p2p/message"
	"github.com/linkchain/p2p/node"
	p2p_peer "github.com/linkchain/p2p/peer"
	"github.com/linkchain/p2p/peer_error"
)

// errIncompatibleConfig is returned if the requested protocols and configs are
// not compatible (low protocol version restrictions and high requirements).
var errIncompatibleConfig = errors.New("incompatible configuration")

type ProtocolManager struct {
	networkId uint64
	maxPeers  int
	peers     *peerSet

	SubProtocols []p2p_peer.Protocol

	eventMux      *event.TypeMux
	txCh          chan tx.ITx
	txSub         event.Subscription
	minedBlockSub *event.TypeMuxSubscription

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh chan *peer
	// txsyncCh    chan *txsync
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg sync.WaitGroup
}

// NewProtocolManager returns a new ethereum sub protocol manager. The Ethereum sub protocol manages peers capable
// with the ethereum network.
func NewProtocolManager(config interface{}, networkId uint64, mux *event.TypeMux) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		networkId:   networkId,
		maxPeers:    64,
		eventMux:    mux,
		peers:       newPeerSet(),
		newPeerCh:   make(chan *peer),
		noMorePeers: make(chan struct{}),
		// txsyncCh:    make(chan *txsync),
		quitSync: make(chan struct{}),
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
			PeerInfo: func(id node.NodeID) interface{} {
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
	return manager, nil
}

func (pm *ProtocolManager) Start() bool {
	// broadcast transactions
	//	pm.txCh = make(chan core.TxPreEvent, txChanSize)
	//	pm.txSub = pm.txpool.SubscribeTxPreEvent(pm.txCh)
	go pm.txBroadcastLoop()
	//
	//	// broadcast mined blocks
	//	pm.minedBlockSub = pm.eventMux.Subscribe(core.NewMinedBlockEvent{})
	go pm.minedBroadcastLoop()
	//
	//	// start sync handlers
	//	go pm.syncer()
	//	go pm.txsyncLoop()
	return true
}

func (pm *ProtocolManager) Stop() {
	log.Info("Stopping Ethereum protocol")

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

	log.Info("Ethereum protocol stopped")
}

// handle is the callback invoked to manage the life cycle of an eth peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		return peer_error.DiscTooManyPeers
	}
	p.Log().Debug("Ethereum peer connected", "name", p.Name())

	// Execute the Ethereum handshake
	var (
	// genesis = pm.blockchain.Genesis()
	// head    = pm.blockchain.CurrentHeader()
	// hash    = head.Hash()
	// number  = head.Number.Uint64()
	// td      = pm.blockchain.GetTd(hash, number)
	)
	//	if err := p.Handshake(pm.networkId, td, hash, genesis.Hash()); err != nil {
	//		p.Log().Debug("Ethereum handshake failed", "err", err)
	//		return err
	//	}

	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		p.Log().Error("Ethereum peer registration failed", "err", err)
		return err
	}
	defer pm.removePeer(p.id)

	// Register the peer in the downloader. If the downloader considers it banned, we disconnect
	//	if err := pm.downloader.RegisterPeer(p.id, p.version, p); err != nil {
	//		return err
	//	}
	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	// pm.syncTransactions(p)

	// main loop. handle incoming messages.

	//	for {
	//		if err := pm.handleMsg(p); err != nil {
	//			p.Log().Debug("Ethereum message handling failed", "err", err)
	//			return err
	//		}
	//	}

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
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	// Block header query, collect the requested headers and reply
	case msg.Code == GetBlockHeadersMsg:
		// do nothing
		return nil

	case msg.Code == BlockHeadersMsg:
		// do nothing
		return nil

	case msg.Code == GetBlockBodiesMsg:
		// Decode the retrieval message
		//		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		//		if _, err := msgStream.List(); err != nil {
		//			return err
		//		}
		//		// Gather blocks until the fetch or network limits is reached
		//		var (
		//			hash   math.Hash
		//			bytes  int
		//			bodies []rlp.RawValue
		//		)
		//		for bytes < softResponseLimit && len(bodies) < downloader.MaxBlockFetch {
		//			// Retrieve the hash of the next block
		//			if err := msgStream.Decode(&hash); err == rlp.EOL {
		//				break
		//			} else if err != nil {
		//				return errResp(ErrDecode, "msg %v: %v", msg, err)
		//			}
		//			// Retrieve the requested block body, stopping if enough was found
		//			if data := pm.blockchain.GetBodyRLP(hash); len(data) != 0 {
		//				bodies = append(bodies, data)
		//				bytes += len(data)
		//			}
		//		}
		//		return p.SendBlockBodiesRLP(bodies)

	case msg.Code == BlockBodiesMsg:
		// A batch of block bodies arrived to one of our previous requests
		//		var request blockBodiesData
		//		if err := msg.Decode(&request); err != nil {
		//			return errResp(ErrDecode, "msg %v: %v", msg, err)
		//		}
		//		// Deliver them all to the downloader for queuing
		//		trasactions := make([][]*types.Transaction, len(request))
		//		uncles := make([][]*types.Header, len(request))
		//
		//		for i, body := range request {
		//			trasactions[i] = body.Transactions
		//			uncles[i] = body.Uncles
		//		}
		//		// Filter out any explicitly requested bodies, deliver the rest to the downloader
		//		filter := len(trasactions) > 0 || len(uncles) > 0
		//		if filter {
		//			trasactions, uncles = pm.fetcher.FilterBodies(p.id, trasactions, uncles, time.Now())
		//		}
		//		if len(trasactions) > 0 || len(uncles) > 0 || !filter {
		//			err := pm.downloader.DeliverBodies(p.id, trasactions, uncles)
		//			if err != nil {
		//				log.Debug("Failed to deliver bodies", "err", err)
		//			}
		//		}

	case msg.Code == NewBlockHashesMsg:
		//		var announces newBlockHashesData
		//		if err := msg.Decode(&announces); err != nil {
		//			return errResp(ErrDecode, "%v: %v", msg, err)
		//		}
		//		// Mark the hashes as present at the remote node
		//		for _, block := range announces {
		//			p.MarkBlock(block.Hash)
		//		}
		//		// Schedule all the unknown hashes for retrieval
		//		unknown := make(newBlockHashesData, 0, len(announces))
		//		for _, block := range announces {
		//			if !pm.blockchain.HasBlock(block.Hash, block.Number) {
		//				unknown = append(unknown, block)
		//			}
		//		}
		//		for _, block := range unknown {
		//			pm.fetcher.Notify(p.id, block.Hash, block.Number, time.Now(), p.RequestOneHeader, p.RequestBodies)
		//		}

	case msg.Code == NewBlockMsg:
		// Retrieve and decode the propagated block
		//		var request newBlockData
		//		if err := msg.Decode(&request); err != nil {
		//			return errResp(ErrDecode, "%v: %v", msg, err)
		//		}
		//		request.Block.ReceivedAt = msg.ReceivedAt
		//		request.Block.ReceivedFrom = p
		//
		//		// Mark the peer as owning the block and schedule it for import
		//		p.MarkBlock(request.Block.Hash())
		//		pm.fetcher.Enqueue(p.id, request.Block)
		//
		//		// Assuming the block is importable by the peer, but possibly not yet done so,
		//		// calculate the head hash and TD that the peer truly must have.
		//		var (
		//			trueHead = request.Block.ParentHash()
		//			trueTD   = new(big.Int).Sub(request.TD, request.Block.Difficulty())
		//		)
		//		// Update the peers total difficulty if better than the previous
		//		if _, td := p.Head(); trueTD.Cmp(td) > 0 {
		//			p.SetHead(trueHead, trueTD)
		//
		//			// Schedule a sync if above ours. Note, this will not fire a sync for a gap of
		//			// a singe block (as the true TD is below the propagated block), however this
		//			// scenario should easily be covered by the fetcher.
		//			currentBlock := pm.blockchain.CurrentBlock()
		//			if trueTD.Cmp(pm.blockchain.GetTd(currentBlock.Hash(), currentBlock.NumberU64())) > 0 {
		//				go pm.synchronise(p)
		//			}
		//		}

	case msg.Code == TxMsg:
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		//		if atomic.LoadUint32(&pm.acceptTxs) == 0 {
		//			break
		//		}
		//		// Transactions can be processed, parse all of them and deliver to the pool
		//		var txs []*types.Transaction
		//		if err := msg.Decode(&txs); err != nil {
		//			return errResp(ErrDecode, "msg %v: %v", msg, err)
		//		}
		//		for i, tx := range txs {
		//			// Validate and mark the remote transaction
		//			if tx == nil {
		//				return errResp(ErrDecode, "transaction %d is nil", i)
		//			}
		//			p.MarkTransaction(tx.Hash())
		//		}
		//		pm.txpool.AddRemotes(txs)

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
	log.Debug("Removing Ethereum peer", "peer", id)

	// Unregister the peer from the downloader and Ethereum peer set
	//	pm.downloader.UnregisterPeer(id)
	//	if err := pm.peers.Unregister(id); err != nil {
	//		log.Error("Peer removal failed", "peer", id, "err", err)
	//	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(peer_error.DiscUselessPeer)
	}
}

// NodeInfo represents a short summary of the Ethereum sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network uint64 `json:"network"` // Ethereum network ID (1=Frontier, 2=Morden, Ropsten=3, Rinkeby=4)
	//	Difficulty *big.Int            `json:"difficulty"` // Total difficulty of the host's blockchain
	Genesis math.Hash `json:"genesis"` // hash of the host's genesis block
	//	Config     *params.ChainConfig `json:"config"`     // Chain configuration for the fork rules
	Head math.Hash `json:"head"` // hash of the host's best owned block
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (self *ProtocolManager) NodeInfo() *NodeInfo {
	return &NodeInfo{
		Network: self.networkId,
		// TODO: add node info
		//		Difficulty: self.blockchain.GetTd(currentBlock.Hash(), currentBlock.NumberU64()),
		// Genesis: self.blockchain.Genesis().Hash(),
		//		Config:     self.blockchain.Config(),
		// Head: currentBlock.Hash(),
	}
}

func (self *ProtocolManager) txBroadcastLoop() {
	for {
		select {
		case event := <-self.txCh:
			self.BroadcastTx(event.GetTxID(), event)

		// Err() channel will be closed when unsubscribing.
		case <-self.txSub.Err():
			return
		}
	}
}

// Mined broadcast loop
func (self *ProtocolManager) minedBroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range self.minedBlockSub.Chan() {
		switch ev := obj.Data.(type) {
		case block.IBlock:
			self.BroadcastBlock(ev, true)  // First propagate block to peers
			self.BroadcastBlock(ev, false) // Only then announce to the rest
		}
	}
}

// BroadcastBlock will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastBlock(block block.IBlock, propagate bool) {
	hash := block.GetBlockID()
	peers := pm.peers.PeersWithoutBlock(hash)

	// If propagation is requested, send to a subset of the peer
	if propagate {
		//		// Calculate the TD of the block (it's not imported yet, so block.Td is not valid)
		//		var td *big.Int
		//		if parent := pm.blockchain.GetBlock(block.ParentHash(), block.NumberU64()-1); parent != nil {
		//			td = new(big.Int).Add(block.Difficulty(), pm.blockchain.GetTd(block.ParentHash(), block.NumberU64()-1))
		//		} else {
		//			log.Error("Propagating dangling block", "number", block.Number(), "hash", hash)
		//			return
		//		}
		//		// Send the block to a subset of our peers
		//		transfer := peers[:int(core_math.Sqrt(float64(len(peers))))]
		//		for _, peer := range transfer {
		//			peer.SendNewBlock(block, td)
		//		}
		//		log.Trace("Propagated block", "hash", hash, "recipients", len(transfer), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
		//		return
	}
	// Otherwise if the block is indeed in out own chain, announce it
	//if pm.blockchain.HasBlock(hash, block.NumberU64()) {
	// for _, peer := range peers {
	// 	peer.SendNewBlockHashes([]math.Hash{hash}, []uint64{block.NumberU64()})
	// }
	log.Trace("Announced block", "hash", hash, "recipients", len(peers))
	// }
}

// BroadcastTx will propagate a transaction to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTx(hash tx.ITxID, t tx.ITx) {
	// Broadcast transaction to a batch of peers not knowing about it
	peers := pm.peers.PeersWithoutTx(hash)
	//FIXME include this again: peers = peers[:int(math.Sqrt(float64(len(peers))))]
	for _, peer := range peers {
		peer.SendTransactions([]tx.ITx{t})
	}
	log.Trace("Broadcast transaction", "hash", hash, "recipients", len(peers))
}
