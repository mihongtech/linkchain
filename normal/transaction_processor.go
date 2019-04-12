package normal

import (
	"errors"
	"github.com/mihongtech/linkchain/interpreter"

	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/helper"
)

func (n *Interpreter) ProcessTxState(tx *meta.Transaction, data interpreter.Params) (error, interpreter.Result) {

	switch tx.Type {
	case config.CoinBaseTx:
		err, fee := n.processCoinBaseTxState(tx, data)
		output := &Output{}
		output.TxFee = fee
		return err, output
	case config.NormalTx:
		err, fee := n.processNormalTxState(tx, data)
		output := &Output{}
		output.TxFee = fee
		return err, output
	}
	return nil, nil
}

//Process CoinBaseTx account,only update to account.
func (n *Interpreter) processCoinBaseTxState(tx *meta.Transaction, data interpreter.Params) (error, *meta.Amount) {
	err := processTxTo(tx, data)
	return err, meta.NewAmount(0)
}

func (n *Interpreter) processNormalTxState(tx *meta.Transaction, data interpreter.Params) (error, *meta.Amount) {
	fcValue := meta.NewAmount(0)
	err, fcValue := processTxFrom(tx, data)
	if err != nil {
		return err, fcValue
	}
	err = processTxTo(tx, data)
	tcValue := tx.GetToValue()
	return err, fcValue.Subtraction(*tcValue)
}

//Update unCoinBaseTx fromAccount,unCoinBaseTx is not coinBase tx,fromAccount is tx from.
//Only update account which is related to tx from.
func processTxFrom(tx *meta.Transaction, data interpreter.Params) (error, *meta.Amount) {
	inputData := data.(*Input)
	fcValue := meta.NewAmount(0)
	for _, fc := range tx.From.Coins {
		fromObj := inputData.StateDB.GetObject(meta.GetAccountHash(fc.GetId()))
		if fromObj == nil {
			return errors.New("verifyUnCoinBaseTx()->can not find tx from"), fcValue
		}

		value, err := fromObj.GetAccount().GetFromCoinValue(&fc)
		if err != nil {
			return err, fcValue
		}
		fcValue.Addition(*value)

		if err := fromObj.GetAccount().RemoveUTXOByFromCoin(&fc); err != nil {
			return err, fcValue
		}
		inputData.StateDB.SetObject(fromObj)
	}
	return nil, fcValue
}

//Update commonTx Account.
//CommonTx is coinBase and normal tx.
func processTxTo(tx *meta.Transaction, data interpreter.Params) error {
	inputData := data.(*Input)
	txId := tx.GetTxID()
	for index := range tx.To.Coins {
		toObj := inputData.StateDB.GetObject(meta.GetAccountHash(tx.To.Coins[index].GetId()))

		if toObj == nil {
			a := *helper.CreateTemplateAccount(tx.To.Coins[index].GetId())
			toObj = inputData.StateDB.NewObject(meta.GetAccountHash(a.Id), a)
		}

		nTicket := meta.NewTicket(*txId, uint32(index))
		nUTXO := meta.NewUTXO(nTicket, inputData.Header.Height, inputData.Header.Height, *tx.To.Coins[index].GetValue())
		toObj.GetAccount().UTXOs = append(toObj.GetAccount().UTXOs, *nUTXO)
		inputData.StateDB.SetObject(toObj)
	}
	return nil
}
