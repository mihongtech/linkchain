package interpreter

import (
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/consensus"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
)

type BlockValidator interface {
	VerifyBlockState(block *meta.Block, root math.Hash, actualReward *meta.Amount, fee *meta.Amount, headerData []byte) error
	ValidateBlockBody(txValidator TransactionValidator, chain core.Chain, block *meta.Block) error
	ValidateBlockHeader(engine consensus.Engine, chain core.Chain, block *meta.Block) error
}

type TransactionValidator interface {
	CheckTx(tx *meta.Transaction) error
	VerifyTx(tx *meta.Transaction, data Params) error
}

type Validator interface {
	TransactionValidator
	BlockValidator
}
