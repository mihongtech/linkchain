package chain

import (
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
)

type IChain interface {
	//maintain blockchain
	AddNewBlock(block.IBlock)

	Rollback(block.IBlock)

	RollbackAtHeight(int)

	//get blockchain info
	GetLastBlock() block.IBlock

	GetHeight() uint32

	GetBlockByID(id meta.BlockID) block.IBlock

	GetBlockByHeight(int) block.IBlock
}

type IChainGraph interface {
}
