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
	GetAllAccounts()
	RemoveAccount(id account.IAccountID) error

	GetAccountRelateTXs(txs tx.ITx, isMine bool) ([]account.IAccount, error)
	UpdateAccountsByTxs(txs []tx.ITx, mineIndex int) error
	RevertAccountsByTxs(txs []tx.ITx, mineIndex int) error
	CheckTxAccount(tx tx.ITx) error
}
