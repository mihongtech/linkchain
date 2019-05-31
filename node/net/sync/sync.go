package sync

import (
	"github.com/mihongtech/linkchain/common/util/event"
	"github.com/mihongtech/linkchain/node/chain"
	"github.com/mihongtech/linkchain/node/net/sync/full"
	"github.com/mihongtech/linkchain/node/pool"

	p2p_peer "github.com/mihongtech/linkchain/node/net/p2p/peer"
)

var ()

type Service struct {
	Engine *full.ProtocolManager
}

type Config struct {
	Chain     chain.Chain
	TxPool    pool.TxPool
	EventMux  *event.TypeMux
	EventTx   *event.Feed
	NetworkId uint64
}

func (s *Service) Setup(i interface{}) bool {
	//log.Info("sync service init...");
	cfg := i.(*Config)
	engine, err := full.NewProtocolManager(cfg.Chain, cfg.TxPool, cfg.NetworkId, cfg.EventMux, cfg.EventTx)
	if err != nil {
		return false
	}
	s.Engine = engine
	return true
}

func (s *Service) Start() bool {
	//log.Info("sync service start...");
	s.Engine.Start()
	return true
}

func (s *Service) Stop() {
	//log.Info("sync service stop...");
	s.Engine.Stop()
}

func (s *Service) Protocols() []p2p_peer.Protocol {
	return s.Engine.SubProtocols
}
