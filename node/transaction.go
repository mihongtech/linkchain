package node

import (
	"errors"

	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/storage"
)

func (n *Node) addTransaction(tx *meta.Transaction) error {
	return n.txPool.addTransaction(tx)
}

func (n *Node) getAllTransaction() []meta.Transaction {
	return n.txPool.getAllTransaction()
}

func (n *Node) removeTransaction(txID meta.TxID) error {
	return n.txPool.removeTransaction(txID)
}

func (n *Node) checkTx(tx *meta.Transaction) error {
	err := n.validatorAPI.CheckTx(tx)
	if err != nil {
		return errors.New("CheckTx" + "\ttx:" + tx.GetTxID().String() + "\nerror:" + err.Error())
	}
	return err
}

func (n *Node) processTx(tx *meta.Transaction) error {
	log.Info("ProcessTx ...")
	//1.checkTx
	if err := n.checkTx(tx); err != nil {
		return err
	}
	//2.push Tx into storage
	err := n.addTransaction(tx)
	if err != nil {
		return err
	}
	log.Info("Add Tranasaction Pool  ...", "txid", tx.GetTxID(), "tx", tx)
	return nil
}

func (n *Node) getTxByID(hash meta.TxID) (*meta.Transaction, math.Hash, uint64, uint64) {
	tx, hash, number, index := storage.GetTransaction(n.db, hash)
	if tx == nil {
		return nil, math.Hash{}, 0, 0
	} else {
		return tx, hash, number, index
	}
}
