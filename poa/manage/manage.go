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
			AccountManager:     &AccountManage{},
			TransactionManager: &TransactionManage{},
			ChainManager:       &ChainManage{},
			NewBlockEvent:      new(event.TypeMux)}
	})
	return m
}

type Manage struct {
	BlockManager       manager.BlockManager
	AccountManager     manager.AccountManager
	TransactionManager manager.TransactionManager
	ChainManager       manager.ChainManager
	NewBlockEvent      *event.TypeMux
}

func (m *Manage) Init(i interface{}) bool {
	log.Info("Manage init...")
	m.AccountManager.Init(i)
	m.BlockManager.Init(i)
	m.ChainManager.Init(i)
	m.TransactionManager.Init(i)
	return true
}

func (m *Manage) Start() bool {
	log.Info("Manage start...")
	m.ChainManager.Start()
	return true
}

func (m *Manage) Stop() {
	log.Info("Manage stop...")
	m.ChainManager.Stop()
}
