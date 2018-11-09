package state

import (
	"github.com/linkchain/common"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/block"
)

type StateDBer interface {
	common.IService
	GetAccount(id meta.IAccountID) (account.IAccount, bool)
	SetAccount(iAccount account.IAccount) error
	GetAllAccount()

	UpdateAccountsByBlock(block block.IBlock) error
	RollBack(block block.IBlock) error
}
