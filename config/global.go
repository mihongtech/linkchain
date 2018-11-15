package config

const (
	FirstPubMiner  = "025aa040dddd8f873ac5d02dfd249adc4d2c9d6def472a4405252fa6f6650ee1f0"
	SecondPubMiner = "02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50"
	ThirdPubMiner  = "03de3b38a7f61312003c61ab8bee55ba6c6aa94464dc7e5a91f4ff11bf1c60dc59"
)

var (
	SignMiners    = []string{FirstPubMiner, SecondPubMiner, ThirdPubMiner}
	varA          = 12
	VarB          = 33
	DefaultPeriod = 15
)

const (
	DefaultBlockVersion       = 0x00000001 //the version of block.
	DefaultDifficulty         = 0xffffffff //the default difficult.
	DefaultNounce             = 0x00000000 //the default nounce of  block.
	DefaultTransactionVersion = 0x00000001 //the version of transaction
	DefaultBlockReward        = 50         //the reward of mining a block
	CoinBaseTx                = 0x00000000 //the coinbase tx for reward to miner
	NormalTx                  = 0x00000001 //the normal tx

	NormalAccount = 0x00000001 // the normal account

	DafaultClearTime = 0x00000000 // the default clearTime

)

