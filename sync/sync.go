package sync

import (
	_ "github.com/linkchain/consensus/manager"
	"github.com/linkchain/sync/full"
)

var (
	engine full.ProtocolManager
)

type Service struct {
}

func (s *Service) Init(i interface{}) bool {
	//log.Info("sync service init...");
	engine = full.ProtocolManager{}
	engine.Init(i)
	return true
}

func (s *Service) Start() bool {
	//log.Info("sync service start...");
	engine.Start()
	return true
}

func (s *Service) Stop() {
	//log.Info("sync service stop...");
	engine.Stop()
}
