package manager

import (
	"github.com/linkchain/common"
	"github.com/linkchain/common/math"
	"github.com/linkchain/meta/block"
)

type ChainReader interface {
	GetBestBlock() block.IBlock

	GetBlockByHash(h math.Hash) (block.IBlock, error)

	GetBlockByHeight(height uint32) (block.IBlock, error)

	GetBlockChainInfo() string

	GetBlockAncestor(block block.IBlock, height uint32) (block.IBlock, error)
}

type ChainWriter interface {
	AddBlock(block block.IBlock)

	UpdateChain() bool
}

type ChainManager interface {
	common.IService
	ChainWriter
	ChainReader
}
