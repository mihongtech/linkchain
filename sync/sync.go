package sync

import (
	"github.com/linkchain/common/util/event"
	"github.com/linkchain/consensus"
	"github.com/linkchain/sync/full"
)

var ()

type Service struct {
	engine *full.ProtocolManager
}

func (s *Service) Init(i interface{}) bool {
	//log.Info("sync service init...");
	engine, err := full.NewProtocolManager(i, i.(*consensus.Service), 0, &(event.TypeMux{}), &(event.Feed{}))
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
