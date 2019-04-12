package interpreter

import (
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/storage/state"
)

type Processor interface {
	ProcessTxState(tx *meta.Transaction, param Params) (error, Result)
	ProcessBlockState(block *meta.Block, stateDb *state.StateDB, chain core.Chain, validator Validator) (error, []Result)

	ExecuteBlockState(block *meta.Block, stateDb *state.StateDB, chain core.Chain, validator Validator) (error, []Result, math.Hash, *meta.Amount)
}
