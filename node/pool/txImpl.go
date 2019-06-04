package pool

import (
	"errors"
	"sync"

	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/node/bcsi"
)

type TxImpl struct {
	txPool       []meta.Transaction
	validatorAPI bcsi.Validator

	txPollMtx sync.RWMutex

	MainChainCh chan meta.ChainEvent
}

func NewTxPool(validatorApI bcsi.Validator) *TxImpl {
	return &TxImpl{
		txPool:       make([]meta.Transaction, 0),
		validatorAPI: validatorApI,
		MainChainCh:  make(chan meta.ChainEvent, 10),
	}
}

func (t *TxImpl) SetUp(i interface{}) bool {
	return true
}

func (t *TxImpl) Start() bool {
	go t.updateTxLoop()
	return true
}

func (t *TxImpl) Stop() {

}

func (t *TxImpl) updateTxLoop() {
	for {
		select {
		case ev := <-t.MainChainCh:
			t.updateTransaction(ev.Block)
		}
	}
}

func (t *TxImpl) updateTransaction(block *meta.Block) {
	txs := block.GetTxs()
	for i := range txs {
		t.RemoveTransaction(*txs[i].GetTxID())
	}
}

func (t *TxImpl) AddTransaction(tx *meta.Transaction) error {
	t.txPollMtx.Lock()
	defer t.txPollMtx.Unlock()
	newTx := *tx
	t.txPool = append(t.txPool, newTx)
	return nil
}

func (t *TxImpl) GetAllTransaction() []meta.Transaction {
	t.txPollMtx.RLock()
	defer t.txPollMtx.RUnlock()
	txs := make([]meta.Transaction, 0)
	for _, tx := range t.txPool {
		txs = append(txs, tx)
	}
	return txs
}

func (t *TxImpl) RemoveTransaction(txID meta.TxID) error {
	t.txPollMtx.Lock()
	defer t.txPollMtx.Unlock()

	txLen := len(t.txPool)
	for i := 0; i < txLen; i++ {
		txHash := t.txPool[i].GetTxID()
		if txHash.IsEqual(&txID) {
			t.txPool = append(t.txPool[:i], t.txPool[i+1:]...)
			i--
			txLen--
		}
	}

	return nil
}

func (t *TxImpl) CheckTx(tx *meta.Transaction) error {
	err := t.validatorAPI.CheckTx(*tx)
	if err != nil {
		return errors.New("CheckTx" + "\ttx:" + tx.GetTxID().String() + "\nerror:" + err.Error())
	}
	return err
}

func (t *TxImpl) ProcessTx(tx *meta.Transaction) error {
	log.Info("ProcessTx ...")
	//1.checkTx
	if err := t.CheckTx(tx); err != nil {
		return err
	}
	//2.push Tx into storage
	err := t.AddTransaction(tx)
	if err != nil {
		return err
	}
	log.Info("Add Tranasaction Pool  ...", "txid", tx.GetTxID(), "tx", tx)
	return nil
}
