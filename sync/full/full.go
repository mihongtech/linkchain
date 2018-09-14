package full

import (
	"github.com/linkchain/common/util/event"
	p2p_peer "github.com/linkchain/p2p/peer"
)

type Config struct {
	maxPeers int
	peers    *peerSet

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
}

type ProtocolManager struct {
}

func (pm *ProtocolManager) Init(i interface{}) bool {
	return true
}

func (pm *ProtocolManager) Start() bool {

	return true
}

func (pm *ProtocolManager) Stop() {

}
