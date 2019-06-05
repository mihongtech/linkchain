package bcsi

import (
	"github.com/mihongtech/linkchain/core/meta"
)

//app provide to core for querying information
type Querier interface {
	//app provide
	GetBlockState(id meta.BlockID) meta.TreeID
}

//app provide to core for notifying app to update app state
type Processor interface {
	UpdateChain(head *meta.Block) error
	ProcessBlock(block *meta.Block) error
	Commit(id meta.BlockID) error
}

//app provide to core for validating data
type Validator interface {
	CheckBlock(block *meta.Block) error
	CheckTx(transaction meta.Transaction) error
	FilterTx(txs []meta.Transaction) []meta.Transaction
}

//app provide to core for setting core option
type Configurator interface {
}

type BCSI interface {
	Querier
	Processor
	Validator
	Configurator
}
