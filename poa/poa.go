package poa

import (
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/poa/manage"
)

type Service struct {
}

func (s *Service) Init(i interface{}) bool {
	log.Info("poa consensus service init...")
	return s.GetManager().Init(i)
}

func (s *Service) Start() bool {
	log.Info("poa consensus service start...")
	return s.GetManager().Start()
}

func (s *Service) Stop() {
	log.Info("poa consensus service stop...")
	s.GetManager().Stop()
}

func (s *Service) GetManager() *manage.Manage {
	return manage.GetManager()
}
