package bcsi

import "github.com/mihongtech/linkchain/core/meta"

//core provide to app for querying information
type Querier interface {

	//core provide
	meta.ChainReader
	GetTx(id meta.TxID) meta.Transaction

	//app provide
	GetBlockState(id meta.BlockID) meta.TreeID
}

//app provide to core for notifying app to update app state
type Processor interface {
	UpdateChain(head *meta.Block) error
	ProcessBlock(block *meta.Block) error
	Commit(id meta.BlockID) error
}

type Validator interface {
	CheckBlock(block *meta.Block) error
	CheckTx(transaction meta.Transaction) error
	FilterTx(txs []meta.Transaction) []meta.Transaction
}

type Configurator interface {
}

type BCSI interface {
	Querier
	Processor
	Validator
	Configurator
}
