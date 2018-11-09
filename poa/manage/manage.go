package manage

import (
	"sync"

	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/consensus/manager"
)

var m *Manage
var once sync.Once

func GetManager() *Manage {
	once.Do(func() {
		m = &Manage{BlockManager: &BlockManage{},
			AccountManager:     &AccountManage{NewWalletEvent: new(event.TypeMux)},
			TransactionManager: &TransactionManage{},
			ChainManager:       &ChainManage{},
			NewBlockEvent:      new(event.TypeMux),
			NewTxEvent:         new(event.Feed)}
	})
	return m
}

type Manage struct {
	BlockManager       manager.BlockManager
	AccountManager     manager.AccountManager
	TransactionManager manager.TransactionManager
	ChainManager       manager.ChainManager
	NewBlockEvent      *event.TypeMux
	NewTxEvent         *event.Feed
}

func (m *Manage) Init(i interface{}) bool {
	log.Info("Manage init...")
	if !m.AccountManager.Init(i) {
		return false
	}
	if !m.BlockManager.Init(i) {
		return false
	}
	if !m.ChainManager.Init(i) {
		return false
	}
	if !m.TransactionManager.Init(i) {
		return false
	}
	return true
}

func (m *Manage) Start() bool {
	log.Info("Manage start...")
	return m.ChainManager.Start()
}

func (m *Manage) Stop() {
	log.Info("Manage stop...")
	m.ChainManager.Stop()
}
