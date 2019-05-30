package config

import (
	"math/big"
)

const (
	FirstPubMiner  = "07411e1beff277bf1dd9d810c07a4db0e1e45f5a"
	SecondPubMiner = "0a35c1bd74497c851265774e7e98027b46c27c41"
	ThirdPubMiner  = "56c5636befbe7cc23f5157c9278fca4e09109ffc"

	DefaultBlockVersion       = 0x00000001 //the version of block.
	DefaultDifficulty         = 0xffffffff //the default difficult.
	DefaultNounce             = 0x00000000 //the default nounce of  block.
	DefaultTransactionVersion = 0x00000001 //the version of transaction
	DefaultBlockReward        = 5000000000 //the reward of mining a block

	DefaultNodeDatabaseDir = "nodes"   // Path within the datadir to store the node infos
	DefaultPrivateKeyDir   = "nodekey" // Path within the datadir to the node's private key
	DefaultMaxPeers        = 25
)

var (
	SignMiners         = []string{FirstPubMiner, SecondPubMiner, ThirdPubMiner}
	DefaultPeriod      = 15
	DefaultChainConfig = &ChainConfig{big.NewInt(1337), uint64(DefaultPeriod)}
)
