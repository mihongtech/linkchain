package storage

import (
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
