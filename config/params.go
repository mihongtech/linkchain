package config

import (
	"math/big"
	"time"
)

const (
	FirstPubMiner  = "07411e1beff277bf1dd9d810c07a4db0e1e45f5a"
	SecondPubMiner = "0a35c1bd74497c851265774e7e98027b46c27c41"
	ThirdPubMiner  = "56c5636befbe7cc23f5157c9278fca4e09109ffc"

	DefaultBlockVersion       = 0x00000001 //the version of block.
	DefaultDifficulty         = 0x1f00ffff //the default difficult.
	DefaultNounce             = 0x00000000 //the default nounce of  block.
	DefaultTransactionVersion = 0x00000001 //the version of transaction
	DefaultBlockReward        = 5000000000 //the reward of mining a block

	DefaultNodeDatabaseDir = "nodes"   // Path within the datadir to store the node infos
	DefaultPrivateKeyDir   = "nodekey" // Path within the datadir to the node's private key
	DefaultMaxPeers        = 25

	TxTypeCount = 2
	CoinBaseTx  = 0x00000000 //the coinbase tx for reward to miner
	NormalTx    = 0x00000001 //the normal tx

	NormalAccount = 0x00000000 // the normal account

	TargetTimespan = 10 * time.Second
	MaxTimespan    = 1 * time.Minute
	MinTimespan    = 5 * time.Second
	PowLimitBits   = 0x0100ffff

	BlockSizeLimit = 2 * 1024 * 1024 * 8
)

var (
	SignMiners         = []string{FirstPubMiner, SecondPubMiner, ThirdPubMiner}
	DefaultPeriod      = 15
	DefaultChainConfig = &ChainConfig{big.NewInt(1337), uint64(DefaultPeriod)}

	// PowLimit is the highest proof of work value a Bitcoin block can
	// have for the main network.  It is the value 2^224 - 1.
	PowLimit = new(big.Int).Sub(new(big.Int).Lsh(new(big.Int).SetInt64(0), 224), new(big.Int).SetInt64(0))
)
