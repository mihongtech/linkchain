package core

import (
	"github.com/linkchain/common/math"
	"github.com/linkchain/config"
	"github.com/linkchain/consensus"
	"github.com/linkchain/core/meta"
)

type Chain interface {
	meta.ChainReader

	// Engine retrieves the ChainReader's consensus engine.
	Engine() consensus.Engine

	// GetHeader returns the hash corresponding to their hash.
	GetHeader(math.Hash, uint64) *meta.BlockHeader

	Config() *config.ChainConfig
}
