package sync

import (
	"github.com/mihongtech/linkchain/app/context"
	"github.com/mihongtech/linkchain/node"
	p2p_peer "github.com/mihongtech/linkchain/p2p/peer"
	"github.com/mihongtech/linkchain/sync/full"
)

var ()

type Service struct {
	Engine *full.ProtocolManager
}

func (s *Service) Setup(i interface{}) bool {
	//log.Info("sync service init...");
	nodeAPI := i.(*context.Context).NodeAPI.(*node.PublicNodeAPI)
	engine, err := full.NewProtocolManager(i,
		nodeAPI, 0, nodeAPI.GetBlockEvent(), nodeAPI.GetTxEvent())
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
