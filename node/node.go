package node

import (
	"encoding/json"
	"errors"
	"github.com/linkchain/interpreter"
	"os"
	"sync"

	"github.com/linkchain/app/context"
	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/consensus"
	"github.com/linkchain/consensus/poa"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/genesis"
	"github.com/linkchain/storage"
)

var (
	globalConfig config.LinkChainConfig
)

type Node struct {
	//transaction
	txPool *TxPool

	//account state
	accountMtx sync.RWMutex

	//engine
	engine         consensus.Engine
	validatorAPI   interpreter.Validator
	interpreterAPI interpreter.Interpreter

	//block
	blockMtx            sync.RWMutex
	mapBlockIndexByHash map[math.Hash]meta.Block

	//chain
	chainMtx   sync.RWMutex
	blockchain *BlockChain
	db         lcdb.Database

	//event
	newBlockEvent   *event.TypeMux
	newAccountEvent *event.TypeMux
	newTxEvent      *event.Feed

	// offchain
	offchain interpreter.OffChain

	updateMainState event.Subscription
	updateSideState event.Subscription
	MainChainCh     chan meta.ChainEvent
	SideChainCh     chan meta.ChainSideEvent
}

func NewNode() *Node {
	return &Node{MainChainCh: make(chan meta.ChainEvent, 10), SideChainCh: make(chan meta.ChainSideEvent, 10)}
}

func (n *Node) Setup(i interface{}) bool {
	globalConfig := i.(*context.Context).Config
	log.Info("Manage init...")

	n.newBlockEvent = new(event.TypeMux)
	n.newAccountEvent = new(event.TypeMux)
	n.newTxEvent = new(event.Feed)
	n.mapBlockIndexByHash = make(map[math.Hash]meta.Block)

	n.initAccountManager()

	s := storage.NewStrorage(globalConfig.DataDir)
	if s == nil {
		log.Error("init storage failed")
		return false
	}
	n.db = s.GetDB()

	config, genesisHash, err := n.initGenesis(n.db, globalConfig.GenesisPath)

	n.engine = poa.NewPoa(config, s.GetDB())
	n.validatorAPI = i.(*context.Context).InterpreterAPI
	n.interpreterAPI = i.(*context.Context).InterpreterAPI
	n.offchain = n.interpreterAPI.CreateOffChain(n.db)

	n.blockchain, err = NewBlockChain(s.GetDB(), genesisHash, nil, config, n.interpreterAPI, n.engine)
	if err != nil {
		log.Error("init blockchain failed", "err", err)
		return false
	}

	n.offchain.Setup(i)

	n.txPool = NewTxPool(n.validatorAPI)
	n.txPool.SetUp(i)

	return true
}

func (n *Node) initGenesis(db lcdb.Database, genesisPath string) (*config.ChainConfig, math.Hash, error) {
	if len(genesisPath) == 0 {
		return nil, math.Hash{}, errors.New("genesis file is nil")
	}
	file, err := os.Open(genesisPath)
	if err != nil {
		log.Info("genesis file not found, use default Genesis")
	}
	defer file.Close()

	genesisBlock := new(genesis.Genesis)
	if err == nil {
		if err := json.NewDecoder(file).Decode(genesisBlock); err != nil {
			log.Error("invalid genesis file")
			return nil, math.Hash{}, errors.New("invalid genesis file")
		}
	} else {
		genesisBlock = nil
	}

	config, hash, err := genesis.SetupGenesisBlock(db, genesisBlock)
	if err != nil {
		log.Error("setup genesis failed", "err", err)
		return nil, math.Hash{}, errors.New("setup genesis failed")
	}

	return config, hash, nil
}

func (n *Node) Start() bool {
	log.Info("Node is start...")
	//n.offchain.SetSubscription(n.blockchain.SubscribeChainEvent(n.offchain.MainChainCh), n.blockchain.SubscribeChainSideEvent(n.offchain.SideChainCh))
	n.updateMainState = n.blockchain.SubscribeChainEvent(n.MainChainCh)
	n.updateSideState = n.blockchain.SubscribeChainSideEvent(n.SideChainCh)
	if !n.offchain.Start() {
		return false
	}
	if !n.txPool.Start() {
		return false
	}

	go n.updateState()
	return true
}

func (n *Node) updateState() {
	for {
		select {
		case ev := <-n.MainChainCh:
			n.offchain.UpdateMainChain(ev)
			n.txPool.MainChainCh <- ev
			n.newAccountEvent.Post(AccountEvent{IsUpdate: true})
		case ev := <-n.SideChainCh:
			n.offchain.UpdateSideChain(ev)
		}
	}
}
func (n *Node) Stop() {
	log.Info("Stop node...")
	n.offchain.Stop()
	n.txPool.Stop()
}

//func (n *Node) getBlockEvent() *event.TypeMux {
//	return n.newBlockEvent
//}
//
//func (n *Node) getTxEvent() *event.Feed {
//	return n.newTxEvent
//}
