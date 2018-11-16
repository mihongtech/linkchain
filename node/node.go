package node

import (
	appContext "github.com/linkchain/app/context"
	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/storage"
)

var (
	globalConfig config.LinkChainConfig
)

type Node struct {
	NewBlockEvent *event.TypeMux
	NewTxEvent    *event.Feed
}

func (m *Node) Setup(i interface{}) bool {
	globalConfig := i.(*appContext.Context).Config

	log.Info("Manage init...")

	m.NewBlockEvent = new(event.TypeMux)
	m.NewTxEvent = new(event.Feed)

	initAccountManager()

	s := storage.NewStrorage(globalConfig.DataDir)

	initChainManager(s.GetDB(), globalConfig.GenesisPath)

	return true
}

func (m *Node) Start() bool {
	log.Info("Manage start...")

	return true
}

func (m *Node) Stop() {
	log.Info("Manage stop...")
}

func (m *Node) GetBlockEvent() *event.TypeMux {
	return m.NewBlockEvent
}

func (m *Node) GetTxEvent() *event.Feed {
	return m.NewTxEvent
}
