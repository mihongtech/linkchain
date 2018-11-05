package config

import (
	"math/big"
)

type ChainConfig struct {
	ChainId *big.Int `json:"chainId"` // Chain id identifies the current chain and is used for replay protection
	Period  uint64   `json:"period"`  // Number of seconds between blocks to enforce
}

var (
	DefaultChainConfig = &ChainConfig{big.NewInt(1337), uint64(DefaultPeriod)}
)
