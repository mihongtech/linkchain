package node

import (
	appContext "github.com/linkchain/app/context"
	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/storage"
	"sync"
)

var (
	globalConfig config.LinkChainConfig
)

type Node struct {
	//transaction
	txPool  []meta.Transaction
	stateDB *storage.StateDB

	//block
	blockMtx            sync.RWMutex
	mapBlockIndexByHash map[math.Hash]meta.Block

	//chain
	chainMtx       sync.RWMutex
	chains         []meta.Chain     //the chain tree for storing all chains
	mainChainIndex []meta.ChainNode //the mainChain is slice for search block
	mainChain      meta.BlockChain  //the mainChain is linked list for converting chain
	db             lcdb.Database

	//event
	newBlockEvent   *event.TypeMux
	newAccountEvent *event.TypeMux
	newTxEvent      *event.Feed
}

func NewNode() *Node {
	return &Node{}
}

func (n *Node) Setup(i interface{}) bool {
	globalConfig := i.(*appContext.Context).Config

	log.Info("Manage init...")

	n.newBlockEvent = new(event.TypeMux)
	n.newAccountEvent = new(event.TypeMux)
	n.newTxEvent = new(event.Feed)
	n.txPool = make([]meta.Transaction, 0)
	n.stateDB = &storage.StateDB{}
	n.mapBlockIndexByHash = make(map[math.Hash]meta.Block)

	n.initAccountManager()

	s := storage.NewStrorage(globalConfig.DataDir)

	initChainManager(n, s.GetDB(), globalConfig.GenesisPath)

	return true
}

func (n *Node) Start() bool {
	log.Info("Manage start...")

	return true
}

func (n *Node) Stop() {
	log.Info("Manage stop...")
}

//func (n *Node) getBlockEvent() *event.TypeMux {
//	return n.newBlockEvent
//}
//
//func (n *Node) getTxEvent() *event.Feed {
//	return n.newTxEvent
//}
