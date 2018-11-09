package manager

import (
	"github.com/linkchain/common"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
)

type BlockManager interface {
	common.IService
	BlockPoolManager
	BlockValidator

	ProcessBlock(block block.IBlock) error
}

/**
BlockPoolManager
manager block pool.the block at pool is side chain block or no-parent block
*/
type BlockPoolManager interface {
	GetBlockByID(hash meta.BlockID) (block.IBlock, error)
	GetBlockByHeight(height uint32) ([]block.IBlock, error)
	GetBlockAncestor(block block.IBlock, height uint32) (block.IBlock, error)
	AddBlock(block block.IBlock) error
	AddBlocks(block []block.IBlock) error
	RemoveBlock(hash meta.BlockID) error
	HasBlock(hash meta.BlockID) bool
}

type BlockValidator interface {
	CheckBlock(block block.IBlock) bool
}
