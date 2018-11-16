package sync

import (
	appContext "github.com/linkchain/app/context"
	"github.com/linkchain/node"
	p2p_peer "github.com/linkchain/p2p/peer"
	"github.com/linkchain/sync/full"
)

var ()

type Service struct {
	Engine *full.ProtocolManager
}

func (s *Service) Setup(i interface{}) bool {
	//log.Info("sync service init...");
	nodeService := i.(*appContext.Context).Node.(*node.Node)
	engine, err := full.NewProtocolManager(i,
		nodeService, 0, nodeService.GetBlockEvent(), nodeService.GetTxEvent())
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
