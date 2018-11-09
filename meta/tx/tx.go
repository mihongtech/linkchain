package tx

import (
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/amount"
	"github.com/linkchain/meta/coin"
)

type TxEvent struct{ Tx ITx }

type ITx interface {
	GetTxID() *meta.TxID

	//tx content

	AddFromCoin(fromcoin coin.IFromCoin)
	AddToCoin(tocoin coin.IToCoin)
	AddSignature(signature math.ISignature)

	GetFromCoins() []coin.IFromCoin
	GetToCoins() []coin.IToCoin
	GetToValue() *amount.Amount
	GetVersion() uint32
	GetType() uint32
	GetNewFromCoins() []coin.IFromCoin

	Verify() error

	//serialize
	serialize.ISerialize
}
