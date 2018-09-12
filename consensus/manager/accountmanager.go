package manager

import (
	"github.com/linkchain/meta/account"
	"github.com/linkchain/common"
	"github.com/linkchain/meta/tx"
)

type AccountManager interface{
	common.IService

	AccountPoolManager

	NewAccount() account.IAccount
}

type AccountPoolManager interface{
	AddAccount(iAccount account.IAccount) error
	GetAccount(id account.IAccountID) (account.IAccount,error)
	RemoveAccount(id account.IAccountID) error
	UpdateAccount(iAccount account.IAccount) error
	UpdateAccountByTX(tx tx.ITx) error
	CheckTxFromAccount(tx tx.ITx) error

	GetAllAccounts()
}
