package interpreter

import (
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/storage/state"
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
