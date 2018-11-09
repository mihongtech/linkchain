package util

import (
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/amount"
	"github.com/linkchain/meta/coin"
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/poa/config"
	poameta "github.com/linkchain/poa/meta"
)

func CreateToCoin(to meta.IAccountID, amount *amount.Amount) coin.IToCoin {
	return poameta.NewToCoin(*to.(*poameta.AccountID), amount)
}

func CreateFromCoin(from meta.IAccountID, ticket ...coin.ITicket) coin.IFromCoin {
	tickets := make([]poameta.Ticket, 0)
	fc := poameta.NewFromCoin(*from.(*poameta.AccountID), tickets)
	for _, c := range ticket {
		fc.AddTicket(c)
	}
	return fc
}

func CreateTempleteTx(version uint32, txtype uint32) tx.ITx {
	return poameta.NewEmptyTransaction(version, txtype)
}

func CreateTransaction(fromCoin coin.IFromCoin, toCoin coin.IToCoin) tx.ITx {
	transaction := CreateTempleteTx(config.DefaultTransactionVersion, config.NormalTx)
	transaction.AddFromCoin(fromCoin)
	transaction.AddToCoin(toCoin)
	return transaction
}

func CreateCoinBaseTx(to meta.IAccountID, amount *amount.Amount) tx.ITx {
	toCoin := poameta.NewToCoin(*to.(*poameta.AccountID), amount)
	transaction := poameta.NewEmptyTransaction(config.DefaultDifficulty, config.CoinBaseTx)
	transaction.AddToCoin(toCoin)
	return transaction
}
