package sync

import (
	"github.com/linkchain/config"
	"github.com/linkchain/node"
	p2p_peer "github.com/linkchain/p2p/peer"
	"github.com/linkchain/sync/full"
)

var ()

type Service struct {
	Engine *full.ProtocolManager
}

func (s *Service) Init(i interface{}) bool {
	//log.Info("sync service init...");
	nodeService := i.(*config.LinkChainConfig).NodeService.(*node.Node)
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
