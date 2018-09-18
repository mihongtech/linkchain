package manager

import (
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta"
	"github.com/linkchain/common"
)

type TransactionManager interface{
	common.IService
	TransactionValidator
	TransactionPoolManager
	NewTransaction(from account.IAccount,to account.IAccount,amount meta.IAmount) tx.ITx
	SignTransaction(tx tx.ITx) error
	ProcessTx(tx tx.ITx) error
}

type TransactionPoolManager interface{
	AddTransaction(tx tx.ITx) error
	GetAllTransaction() []tx.ITx
	RemoveTransaction(txid meta.DataID) error
}

type TransactionValidator  interface {
	CheckTx(tx tx.ITx) bool
}

