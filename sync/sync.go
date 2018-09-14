package sync

import (
	_ "github.com/linkchain/consensus/manager"
	"github.com/linkchain/sync/full"
)

var ()

type Service struct {
	engine *full.ProtocolManager
}

func (s *Service) Init(i interface{}) bool {
	//log.Info("sync service init...");
	engine, err := full.NewProtocolManager(i, 0, nil)
	if err != nil {
		return false
	}
	s.engine = engine
	return true
}

func (s *Service) Start() bool {
	//log.Info("sync service start...");
	s.engine.Start()
	return true
}

func (s *Service) Stop() {
	//log.Info("sync service stop...");
	s.engine.Stop()
}
