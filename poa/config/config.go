package config

import (
	"math/big"
)

const (
	BlockVersion       = 0x00000001 //the version of block.
	Difficulty         = 0xffffffff //the default difficult.
	DefaultNounce      = 0x00000000 //the default nounce of  block.
	TransactionVersion = 0x00000001 //the version of transaction
	DefaultPeriod      = 15
)

var (
	DefaultChainConfig = &ChainConfig{big.NewInt(1337), DefaultPeriod}
)
