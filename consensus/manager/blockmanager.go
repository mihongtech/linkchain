package manager

import (
	"github.com/linkchain/meta/block"
	"github.com/linkchain/common"
	"github.com/linkchain/meta"
)

type BlockManager interface {
	common.IService
	BlockBaseManager
	BlockPoolManager
	BlockValidator

	ProcessBlock(block block.IBlock) error
}

type BlockBaseManager interface {
	NewBlock() block.IBlock
	GetGensisBlock() block.IBlock
}
/**
	BlockPoolManager
	manager block pool.the block at pool is side chain block or no-parent block
 */
type BlockPoolManager interface{
	GetBlockByID(hash meta.DataID) (block.IBlock,error)
	GetBlockByHeight(height uint32) ([]block.IBlock,error)
	AddBlock(block block.IBlock) error
	AddBlocks(block []block.IBlock) error
	RemoveBlock(hash meta.DataID) error
}

type BlockValidator  interface {
	CheckBlock(block block.IBlock) bool
}