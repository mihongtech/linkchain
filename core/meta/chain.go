package meta

import (
	"github.com/mihongtech/linkchain/common/math"
)

// ChainReader defines a small collection of methods needed to access the local
// chain during header and/or uncle verification.
type ChainReader interface {
	// Config retrieves the chain's chain configuration.
	//	Config() *config.ChainConfig
	//
	//	// CurrentHeader retrieves the current header from the local chain.
	//	CurrentHeader() *meta.BlockHeader

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetBlockByID(hash BlockID) (*Block, error)

	GetBlockByHeight(height uint32) (*Block, error)
}

type ChainInfo struct {
	ChainId    int
	BestHeight uint32
	BestHash   string
}

type ChainEvent struct {
	Block *Block
	Hash  math.Hash
}

type ChainSideEvent struct {
	Block *Block
}

type ChainHeadEvent struct{ Block *Block }
