package full

import (
	"errors"
	"fmt"
	"sync"

	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
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

	eventMux *event.TypeMux
	// txCh          chan core.TxPreEvent
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

	return true
}

func (pm *ProtocolManager) Stop() {

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
