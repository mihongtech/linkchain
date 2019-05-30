package pool

import (
	"github.com/mihongtech/linkchain/interpreter"
	"sync"

	"github.com/mihongtech/linkchain/core/meta"
)

type TxPool struct {
	txPool       []meta.Transaction
	validatorAPI interpreter.Validator

	txPollMtx sync.RWMutex

	MainChainCh chan meta.ChainEvent
}

func NewTxPool(validatorApI interpreter.Validator) *TxPool {
	return &TxPool{
		txPool:       make([]meta.Transaction, 0),
		validatorAPI: validatorApI,
		MainChainCh:  make(chan meta.ChainEvent, 10),
	}
}

func (t *TxPool) SetUp(i interface{}) bool {
	return true
}

func (t *TxPool) Start() bool {
	go t.updateTxLoop()
	return true
}

func (t *TxPool) Stop() {

}

func (t *TxPool) updateTxLoop() {
	for {
		select {
		case ev := <-t.MainChainCh:
			t.updateTransaction(ev.Block)
		}
	}
}

func (t *TxPool) updateTransaction(block *meta.Block) {
	txs := block.GetTxs()
	for i := range txs {
		t.removeTransaction(*txs[i].GetTxID())
	}
}

func (t *TxPool) addTransaction(tx *meta.Transaction) error {
	t.txPollMtx.Lock()
	defer t.txPollMtx.Unlock()
	newTx := *tx
	t.txPool = append(t.txPool, newTx)
	return nil
}

func (t *TxPool) getAllTransaction() []meta.Transaction {
	t.txPollMtx.RLock()
	defer t.txPollMtx.RUnlock()
	txs := make([]meta.Transaction, 0)
	for _, tx := range t.txPool {
		txs = append(txs, tx)
	}
	return txs
}

func (t *TxPool) removeTransaction(txID meta.TxID) error {
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
