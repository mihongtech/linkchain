package node

import (
	"errors"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
)

func (n *Node) addTransaction(tx *meta.Transaction) error {
	newTx := *tx
	n.txPool = append(n.txPool, newTx)
	return nil
}

func (n *Node) getAllTransaction() []meta.Transaction {
	txs := make([]meta.Transaction, 0)
	for _, tx := range n.txPool {
		txs = append(txs, tx)
	}
	return txs
}

func (n *Node) removeTransaction(txID meta.TxID) error {
	deleteIndex := make([]int, 0)
	for index, tx := range n.txPool {
		txHash := tx.GetTxID()
		if txHash.IsEqual(&txID) {
			deleteIndex = append(deleteIndex, index)
		}
	}
	for _, index := range deleteIndex {
		n.txPool = append(n.txPool[:index], n.txPool[index+1:]...)
	}
	return nil
}

func (n *Node) checkTx(tx *meta.Transaction) bool {
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

func (n *Node) processTx(tx *meta.Transaction) error {
	log.Info("POA ProcessTx ...")
	//1.checkTx
	if !n.checkTx(tx) {
		log.Error("POA checkTransaction failed")
		return errors.New("POA checkTransaction failed")
	}
	//2.push Tx into storage
	n.addTransaction(tx)
	log.Info("POA Add Tranasaction Pool  ...", "txid", tx.GetTxID())
	log.Info("POA Add Tranasaction Pool  ...", "tx", tx)
	return nil
}
