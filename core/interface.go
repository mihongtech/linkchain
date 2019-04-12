package core

import (
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/consensus"
	"github.com/mihongtech/linkchain/core/meta"
)

type Chain interface {
	meta.ChainReader

	// Engine retrieves the ChainReader's consensus engine.
	Engine() consensus.Engine

	// GetHeader returns the hash corresponding to their hash.
	GetHeader(math.Hash, uint64) *meta.BlockHeader

	Config() *config.ChainConfig
}
