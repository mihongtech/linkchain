package account

import (
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/amount"
	"github.com/linkchain/meta/coin"
)

type IAccount interface {
	GetAmount() *amount.Amount
	GetAccountID() meta.IAccountID

	CheckFromCoin(fromCoin coin.IFromCoin) bool
	GetFromCoinValue(fromCoin coin.IFromCoin) (*amount.Amount, error)
	Contains(ticket coin.ITicket) bool

	//serialize
	serialize.ISerialize
}
