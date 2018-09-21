package manager

import (
	"github.com/linkchain/common"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/tx"
)

type AccountManager interface {
	common.IService

	AccountPoolManager

	NewAccount() account.IAccount
}

type AccountPoolManager interface {
	AddAccount(iAccount account.IAccount) error
	GetAccount(id account.IAccountID) (account.IAccount, error)
	RemoveAccount(id account.IAccountID) error

	GetAccountRelateTXs(txs tx.ITx, isMine bool) ([]account.IAccount, error)
	ConvertAccount(tx tx.ITx, isMine bool) (account.IAccount, account.IAccount)
	UpdateAccountsByTxs(txs []tx.ITx, mineIndex int) error

	CheckTxAccount(tx tx.ITx) error

	GetAllAccounts()
}
