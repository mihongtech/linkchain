package chain

import (
	"math/big"

	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/node/config"
)

type ChainReader interface {
	GetBestBlock() *meta.Block
	HasBlock(hash meta.BlockID) bool
	GetBlockByID(hash meta.BlockID) (*meta.Block, error)
	GetHeader(hash math.Hash, height uint64) *meta.BlockHeader
	GetBlockByHeight(height uint32) (*meta.Block, error)
	GetChainConfig() *config.ChainConfig
	GetChainID() *big.Int
}

type ChainVerifier interface {
	ProcessBlock(block *meta.Block) error
	CheckBlock(block *meta.Block) error
}

type Chain interface {
	ChainReader
	ChainVerifier
}
