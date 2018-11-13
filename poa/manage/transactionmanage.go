package manage

import (
	"errors"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/tx"
	poameta "github.com/linkchain/poa/meta"
)

type TransactionManage struct {
	txpool []poameta.Transaction
}

/** interface: common.IService **/
func (m *TransactionManage) Init(i interface{}) bool {
	log.Info("BlockManage init...")
	m.txpool = make([]poameta.Transaction, 0)
	return true
}

func (m *TransactionManage) Start() bool {
	log.Info("BlockManage start...")
	return true
}

func (m *TransactionManage) Stop() {
	log.Info("BlockManage stop...")
}

func (m *TransactionManage) AddTransaction(tx tx.ITx) error {
	newTx := *tx.(*poameta.Transaction)
	m.txpool = append(m.txpool, newTx)
	return nil
}

func (m *TransactionManage) GetAllTransaction() []tx.ITx {
	txs := make([]tx.ITx, 0)
	for _, tx := range m.txpool {
		txs = append(txs, &tx)
	}
	return txs
}

func (m *TransactionManage) RemoveTransaction(txid meta.TxID) error {
	deleteIndex := make([]int, 0)
	for index, tx := range m.txpool {
		txHash := tx.GetTxID()
		if txHash.IsEqual(&txid) {
			deleteIndex = append(deleteIndex, index)
		}
	}
	for _, index := range deleteIndex {
		m.txpool = append(m.txpool[:index], m.txpool[index+1:]...)
	}
	return nil
}

func (m *TransactionManage) CheckTx(tx tx.ITx) bool {
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

func (m *TransactionManage) ProcessTx(tx tx.ITx) error {
	log.Info("POA ProcessTx ...")
	//1.checkTx
	if !m.CheckTx(tx) {
		log.Error("POA checkTransaction failed")
		return errors.New("POA checkTransaction failed")
	}
	//2.push Tx into storage
	m.AddTransaction(tx)
	log.Info("POA Add Tranasaction Pool  ...", "txid", tx.GetTxID())
	log.Info("POA Add Tranasaction Pool  ...", "tx", tx)
	return nil
}
