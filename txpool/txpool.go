package txpool

import (
	"sync"

	"github.com/mihongtech/linkchain/app/context"
	"github.com/mihongtech/linkchain/common/util/event"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/node"
)

type TxPool struct {
	txPool       []meta.Transaction
	nodeAPI      *node.PublicNodeAPI
	validatorAPI interpreter.Validator

	txPollMtx sync.RWMutex

	insertBlockSub *event.TypeMuxSubscription
}

func NewTxPool() *TxPool {
	return &TxPool{
		txPool: make([]meta.Transaction, 0),
	}
}

func (tp *TxPool) Setup(i interface{}) bool {
	tp.nodeAPI = i.(*context.Context).NodeAPI.(*node.PublicNodeAPI)
	tp.validatorAPI = i.(*context.Context).InterpreterAPI.(interpreter.Validator)
	return true
}

func (tp *TxPool) Start() bool {
	txPoolEvent := tp.nodeAPI.GetTxPoolEvent()
	tp.insertBlockSub = txPoolEvent.Subscribe(node.InsertBlockEvent{})
	go tp.updateTxLoop()
	return true
}

func (tp *TxPool) Stop() {

}

func (tp *TxPool) updateTxLoop() {
	for {
		select {
		case ev := <-tp.insertBlockSub.Chan():
			switch ev := ev.Data.(type) {
			case node.InsertBlockEvent:
				tp.updateTransaction(ev.Block)
			}
		}
	}
}
func (tp *TxPool) updateTransaction(block *meta.Block) {
	txs := block.GetTxs()
	for i := range txs {
		tp.removeTransaction(*txs[i].GetTxID())
	}
}

func (tp *TxPool) addTransaction(tx *meta.Transaction) error {
	tp.txPollMtx.Lock()
	defer tp.txPollMtx.Unlock()
	newTx := *tx
	tp.txPool = append(tp.txPool, newTx)
	return nil
}

func (tp *TxPool) getAllTransaction() []meta.Transaction {
	tp.txPollMtx.RLock()
	defer tp.txPollMtx.RUnlock()
	txs := make([]meta.Transaction, 0)
	for _, tx := range tp.txPool {
		txs = append(txs, tx)
	}
	return txs
}

func (tp *TxPool) removeTransaction(txID meta.TxID) error {
	tp.txPollMtx.Lock()
	defer tp.txPollMtx.Unlock()

	txLen := len(tp.txPool)
	for i := 0; i < txLen; i++ {
		txHash := tp.txPool[i].GetTxID()
		if txHash.IsEqual(&txID) {
			tp.txPool = append(tp.txPool[:i], tp.txPool[i+1:]...)
			i--
			txLen--
		}
	}

	return nil
}
