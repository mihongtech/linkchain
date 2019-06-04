package node

import (
	"encoding/json"
	"errors"
	"github.com/mihongtech/linkchain/node/bcsi"
	"os"
	"sync"

	"github.com/mihongtech/linkchain/common/lcdb"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/util/event"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/node/chain"
	"github.com/mihongtech/linkchain/node/chain/genesis"
	"github.com/mihongtech/linkchain/node/config"
	"github.com/mihongtech/linkchain/node/consensus"
	"github.com/mihongtech/linkchain/node/consensus/poa"
	"github.com/mihongtech/linkchain/node/net"
	"github.com/mihongtech/linkchain/node/net/p2p"
	"github.com/mihongtech/linkchain/node/pool"
	"github.com/mihongtech/linkchain/storage"
)

type Config struct {
	config.BaseConfig
	InterpreterAPI interpreter.Interpreter
}

type Node struct {
	//txPool
	txPool *pool.TxImpl

	//consensus
	engine consensus.Engine

	//BCSI
	bcsiAPI        bcsi.BCSI
	validatorAPI   interpreter.Validator
	interpreterAPI interpreter.Interpreter

	//chain
	chainMtx   sync.RWMutex
	blockchain *chain.ChainImpl
	db         lcdb.Database

	//net p2p
	p2pSvc net.Net

	//event
	newBlockEvent *event.TypeMux
	newTxEvent    *event.Feed

	updateMainState event.Subscription
	updateSideState event.Subscription
	MainChainCh     chan meta.ChainEvent
	SideChainCh     chan meta.ChainSideEvent
}

func NewNode(cfg config.BaseConfig) *Node {
	return &Node{
		p2pSvc:      p2p.NewP2P(cfg),
		MainChainCh: make(chan meta.ChainEvent, 10),
		SideChainCh: make(chan meta.ChainSideEvent, 10)}
}

//Setup is prepared to init node.
func (n *Node) Setup(i interface{}) bool {
	globalConfig := i.(*Config)
	log.Info("Manage init...")

	//Event
	n.newBlockEvent = new(event.TypeMux)
	n.newTxEvent = new(event.Feed)

	//DB
	s := storage.NewStrorage(globalConfig.DataDir)
	if s == nil {
		log.Error("init storage failed")
		return false
	}
	n.db = s.GetDB()

	chainCfg, genesisHash, err := n.initGenesis(n.db, globalConfig.GenesisPath)

	//consensus
	n.engine = poa.NewPoa(chainCfg, s.GetDB())

	//BCSI
	n.validatorAPI = i.(*Config).InterpreterAPI
	n.interpreterAPI = i.(*Config).InterpreterAPI

	//chain
	n.blockchain, err = chain.NewBlockChain(s.GetDB(), genesisHash, nil, chainCfg, n.interpreterAPI, n.engine)
	if err != nil {
		log.Error("init chain failed", "err", err)
		return false
	}

	//tx pool
	n.txPool = pool.NewTxPool(n.validatorAPI)
	n.txPool.SetUp(i)

	//p2p init
	p2pCfg := p2p.NewConfig(n.blockchain, n.txPool, 0, n.newBlockEvent, n.newTxEvent)
	if !n.p2pSvc.Setup(p2pCfg) {
		return false
	}

	return true
}

//initGeneisis() init gensisBlock config form gensis.json
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
	//n.offchain.SetSubscription(n.chain.SubscribeChainEvent(n.offchain.MainChainCh), n.chain.SubscribeChainSideEvent(n.offchain.SideChainCh))
	n.updateMainState = n.blockchain.SubscribeChainEvent(n.MainChainCh)
	n.updateSideState = n.blockchain.SubscribeChainSideEvent(n.SideChainCh)

	if !n.txPool.Start() {
		return false
	}
	if !n.p2pSvc.Start() {
		return false
	}

	go n.updateState()
	return true
}

func (n *Node) updateState() {
	for {
		select {
		case ev := <-n.MainChainCh: //the signal of MainChain update
			n.txPool.MainChainCh <- ev

			//case ev := <-n.SideChainCh: //the signal of SideChain update
		}
	}
}
func (n *Node) Stop() {
	log.Info("Stop node...")
	n.txPool.Stop()
}
