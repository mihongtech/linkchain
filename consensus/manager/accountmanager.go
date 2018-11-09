package manager

import (
	"github.com/linkchain/common"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/block"
)

type AccountManager interface {
	common.IService

	AccountPoolManager

	//NewAccount() account.IAccount
}

type AccountPoolManager interface {
	AddAccount(iAccount account.IAccount) error
	GetAccount(id meta.IAccountID) (account.IAccount, error)
	GetAllAccounts()
	RemoveAccount(id meta.IAccountID) error

	UpdateAccountsByBlock(block block.IBlock) error
	RevertAccountsByBlock(block block.IBlock) error
}
