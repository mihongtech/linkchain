package storage

import (
	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
	"github.com/linkchain/meta/chain"
	"github.com/linkchain/meta/tx"
)

type IStroage interface {
	//block
	storeBlock(block block.IBlock)
	loadBlockById(id meta.BlockID) block.IBlock
	loadBlockByHeight(height int) block.IBlock

	//tx
	storeTx(iTx tx.ITx)
	loadTxById(id meta.BlockID) tx.ITx

	//chain info
	storeChain(chain chain.IChain)
	storeChainGraph(graph chain.IChainGraph)
	loadChainGraph() chain.IChainGraph
}

type Storage struct {
	db lcdb.Database
}

func (m *Storage) Init(i interface{}) bool {
	log.Info("Stroage init...")

	//load genesis from storage
	var err error
	m.db, err = lcdb.NewLDBDatabase("data", 1024, 256)
	if err != nil {
		return false
	}

	return true
}

func (m *Storage) Start() bool {
	log.Info("Stroage start...")
	return true
}

func (m *Storage) Stop() {
	log.Info("Stroage stop...")
}

func (m *Storage) GetDB() lcdb.Database {
	return m.db
}
