package manager

import (
	"github.com/linkchain/common"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
)

type BlockManager interface {
	common.IService
	BlockBaseManager
	BlockPoolManager
	BlockValidator

	ProcessBlock(block block.IBlock) error
}

type BlockBaseManager interface {
	CreateBlock() (block.IBlock, error)
	GetGensisBlock() block.IBlock
	RebuildBlock(block block.IBlock) (block.IBlock, error)
}

/**
BlockPoolManager
manager block pool.the block at pool is side chain block or no-parent block
*/
type BlockPoolManager interface {
	GetBlockByID(hash meta.DataID) (block.IBlock, error)
	GetBlockByHeight(height uint32) ([]block.IBlock, error)
	GetBlockAncestor(block block.IBlock, height uint32) (block.IBlock, error)
	AddBlock(block block.IBlock) error
	AddBlocks(block []block.IBlock) error
	RemoveBlock(hash meta.DataID) error
	HasBlock(hash meta.DataID) bool
}

type BlockValidator interface {
	CheckBlock(block block.IBlock) bool
}
