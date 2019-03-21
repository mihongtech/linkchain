package interpreter

import (
	"github.com/linkchain/common/math"
	"github.com/linkchain/core"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/storage/state"
)

type Processor interface {
	ProcessTxState(tx *meta.Transaction, param Params) (error, Result)
	ProcessBlockState(block *meta.Block, stateDb *state.StateDB, chain core.Chain, validator Validator) (error, []Result)

	ExecuteBlockState(block *meta.Block, stateDb *state.StateDB, chain core.Chain, validator Validator) (error, []Result, math.Hash, *meta.Amount)
}
