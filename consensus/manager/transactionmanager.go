package manager

import (
	"github.com/linkchain/common"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/tx"
)

type TransactionManager interface {
	common.IService
	TransactionValidator
	TransactionPoolManager

	ProcessTx(tx tx.ITx) error
}

type TransactionPoolManager interface {
	AddTransaction(tx tx.ITx) error
	GetAllTransaction() []tx.ITx
	RemoveTransaction(txid meta.TxID) error
}

type TransactionValidator interface {
	CheckTx(tx tx.ITx) bool
}
