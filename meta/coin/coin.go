package coin

import (
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/amount"
)

type ITicket interface {
	SetTxid(id meta.TxID)
	GetTxid() *meta.TxID

	SetIndex(index uint32)
	GetIndex() uint32
	//serialize
	serialize.ISerialize
}

type IFromCoin interface {
	SetId(id meta.IAccountID)
	GetId() meta.IAccountID

	AddTicket(ticket ITicket)
	GetTickets() []ITicket
	//serialize
	serialize.ISerialize
}

type IToCoin interface {
	SetId(id meta.IAccountID)
	GetId() meta.IAccountID

	SetValue(value *amount.Amount)
	GetValue() *amount.Amount
	CheckValue() bool

	//serialize
	serialize.ISerialize
}
