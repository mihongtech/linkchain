package node

import (
	"errors"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/config"
)

var (
	txPool = make([]meta.Transaction, 0)
)

func AddTransaction(tx *meta.Transaction) error {
	newTx := *tx
	txPool = append(txPool, newTx)
	return nil
}


func GetAllTransaction() []meta.Transaction {
	txs := make([]meta.Transaction, 0)
	for _, tx := range txPool {
		txs = append(txs, tx)
	}
	return txs
}

func removeTransaction(txID meta.TxID) error {
	deleteIndex := make([]int, 0)
	for index, tx := range txPool {
		txHash := tx.GetTxID()
		if txHash.IsEqual(&txID) {
			deleteIndex = append(deleteIndex, index)
		}
	}
	for _, index := range deleteIndex {
		txPool = append(txPool[:index], txPool[index+1:]...)
	}
	return nil
}

func checkTx(tx *meta.Transaction) bool {
	log.Info("POA CheckTx ...")
	err := tx.Verify()
	if err != nil {
		log.Error("POA CheckTx", "failed", err)
		return false
	}

	//Check Tx Amount > 0
	for _, coin := range tx.GetToCoins() {
		if !coin.CheckValue() {
			log.Error("POA CheckTx", "failed", "Transaction toCoin-Value need plus 0")
			return false
		}
	}

	if err != nil {
		log.Error("POA CheckTx", "failed", err)
		return false
	}

	return true
}

func processTx(tx *meta.Transaction) error {
	log.Info("POA ProcessTx ...")
	//1.checkTx
	if !checkTx(tx) {
		log.Error("POA checkTransaction failed")
		return errors.New("POA checkTransaction failed")
	}
	//2.push Tx into storage
	AddTransaction(tx)
	log.Info("POA Add Tranasaction Pool  ...", "txid", tx.GetTxID())
	log.Info("POA Add Tranasaction Pool  ...", "tx", tx)
	return nil
}

func CreateToCoin(to meta.AccountID, amount *meta.Amount) *meta.ToCoin {
	return meta.NewToCoin(to, amount)
}

func CreateFromCoin(from meta.AccountID, ticket ...meta.Ticket) *meta.FromCoin {
	tickets := make([]meta.Ticket, 0)
	fc := meta.NewFromCoin(from, tickets)
	for _, c := range ticket {
		fc.AddTicket(&c)
	}
	return fc
}

func CreateTempleteTx(version uint32, txtype uint32) *meta.Transaction {
	return meta.NewEmptyTransaction(version, txtype)
}

func CreateTransaction(fromCoin meta.FromCoin, toCoin meta.ToCoin) *meta.Transaction{
	transaction := CreateTempleteTx(config.DefaultTransactionVersion, config.NormalTx)
	transaction.AddFromCoin(fromCoin)
	transaction.AddToCoin(toCoin)
	return transaction
}

func CreateCoinBaseTx(to meta.AccountID, amount *meta.Amount) *meta.Transaction {
	toCoin := meta.NewToCoin(to, amount)
	transaction := meta.NewEmptyTransaction(config.DefaultDifficulty, config.CoinBaseTx)
	transaction.AddToCoin(*toCoin)
	return transaction
}
