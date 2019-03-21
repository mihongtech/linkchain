package interpreter

import (
	"github.com/linkchain/core"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/storage/state"
)

type Params interface {
	GetBlockSigner() meta.AccountID
	GetBlockHeader() *meta.BlockHeader
	GetStateDB() *state.StateDB
	GetChainReader() meta.ChainReader
}

type Result interface {
	GetTxFee() *meta.Amount
	GetReceipt() *core.Receipt
	WriteResult() error
}
