package poamanager

import (
	"sync"

	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/consensus/manager"
)

var m *POAManager
var once sync.Once

func GetManager() *POAManager {
	once.Do(func() {
		m = &POAManager{BlockManager: &POABlockManager{},
			AccountManager:     &POAAccountManager{},
			TransactionManager: &POATxManager{},
			ChainManager:       &POAChainManager{},
			NewBlockEvent:      new(event.TypeMux),
			NewTxEvent:         new(event.Feed)}
	})
	return m
}

type POAManager struct {
	BlockManager       manager.BlockManager
	AccountManager     manager.AccountManager
	TransactionManager manager.TransactionManager
	ChainManager       manager.ChainManager
	NewBlockEvent      *event.TypeMux
	NewTxEvent         *event.Feed
}

func (m *POAManager) Init(i interface{}) bool {
	log.Info("POAManager init...")
	//TODO Account init
	m.AccountManager.Init(i)
	m.BlockManager.Init(i)
	m.ChainManager.Init(i)
	m.TransactionManager.Init(i)

	//TODO Transaction init
	return true
}

func (m *POAManager) Start() bool {
	log.Info("POAManager start...")
	m.ChainManager.Start()
	return true
}

func (m *POAManager) Stop() {
	log.Info("POAManager stop...")
	m.ChainManager.Stop()
}
