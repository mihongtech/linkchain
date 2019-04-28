package txpool

import (
	"errors"

	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/core/meta"
)

func (tp *TxPool) ProcessTx(tx *meta.Transaction) error {
	return tp.processTx(tx)
}

func (tp *TxPool) GetAllTransaction() []meta.Transaction {
	return tp.getAllTransaction()
}

func (tp *TxPool) RemoveTransaction(txID meta.TxID) error {
	return tp.removeTransaction(txID)
}

func (tp *TxPool) checkTx(tx *meta.Transaction) error {
	err := tp.validatorAPI.CheckTx(tx)
	if err != nil {
		return errors.New("CheckTx" + "\ttx:" + tx.GetTxID().String() + "\nerror:" + err.Error())
	}
	return err
}

func (tp *TxPool) processTx(tx *meta.Transaction) error {
	log.Info("ProcessTx ...")
	//1.checkTx
	if err := tp.checkTx(tx); err != nil {
		return err
	}
	//2.push Tx into storage
	err := tp.addTransaction(tx)
	if err != nil {
		return err
	}
	log.Info("Add Tranasaction Pool  ...", "txid", tx.GetTxID(), "tx", tx)
	return nil
}
