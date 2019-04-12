package contract

import (
	"errors"
	"math/big"

	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/contract/vm"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/helper"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/normal"
)

func (p *Interpreter) ProcessTxState(tx *meta.Transaction, data interpreter.Params) (error, interpreter.Result) {
	var err error = nil
	ouput := Output{}
	ouput.TxFee = meta.NewAmount(0)

	if normal.IsNormal(tx.Type) {
		normal := normal.Interpreter{}
		err, newOut := normal.ProcessTxState(tx, &(data.(*Input).Input))
		if err != nil {
			return err, nil
		}
		ouput.TxFee = newOut.GetTxFee()
		ouput.Receipt = newOut.GetReceipt()
		return nil, &ouput
	} else if IsContract(tx.Type) {
		switch tx.Type {
		case ContractTx:
			err, txFee := processContractTx(tx, data)
			if err != nil {
				return err, nil
			}
			ouput.TxFee.Addition(*txFee)
			err, vmfee, receipt, resultTx, _ := processTxVm(tx, data)
			if err != nil {
				return err, nil
			}
			ouput.TxFee.Addition(*vmfee)
			ouput.Receipt = receipt
			ouput.ResultTx = resultTx
		}

		return err, &ouput
	} else {
		return interpreter.ErrKnownTxType, nil
	}

}

func processContractTx(tx *meta.Transaction, data interpreter.Params) (error, *meta.Amount) {
	fcValue := meta.NewAmount(0)
	err, fcValue := processTxFrom(tx, data)
	if err != nil {
		return err, fcValue
	}
	err = processContractTxTo(tx, data)
	tcValue := tx.GetToValue()
	return err, fcValue.Subtraction(*tcValue)
}

func processResultContractTx(tx *meta.Transaction, data interpreter.Params) (error, *meta.Amount) {
	fcValue := meta.NewAmount(0)
	err, fcValue := processTxFrom(tx, data)
	if err != nil {
		return err, fcValue
	}
	err = processResultContractTxTo(tx, data)
	tcValue := tx.GetToValue()
	return err, fcValue.Subtraction(*tcValue)
}

func processTxVm(tx *meta.Transaction, data interpreter.Params) (error, *meta.Amount, *core.Receipt, *meta.Transaction, []byte) {
	inputData := data.(*Input)

	fee := meta.NewAmount(0)
	msg := ConvertToMessage(tx)
	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, inputData.Header, inputData.Chain, &inputData.BlockSigner)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	adapter := NewStateAdapter(inputData.StateDB, *tx.GetTxID(), *inputData.Header.GetBlockID(), inputData.BlockSigner, int64(inputData.Header.Height))
	vmenv := vm.NewEVM(context, adapter, inputData.Config, inputData.VmCfg)
	// Apply the transaction to the current state (included in the env)
	ret, gas, failed, err := ApplyMessage(vmenv, msg, inputData.Gp)
	if err != nil {
		return err, fee, nil, nil, ret
	}
	*inputData.UsedGas += gas

	//Update system status.
	//1.Update the state with pending changes, the vm change push to stateDB.then have not update MPT tree.
	err = adapter.Commit()
	if err != nil {
		return err, fee, nil, nil, ret
	}

	//2.Get contract running result Tx for converting UTXO.it is prepared for update account UTXO change.
	contractResultTx, err := adapter.GetResultTransaction()
	if err != nil {
		return err, fee, nil, nil, ret
	}
	//3.Update account after vm run.
	err, contractFee := processResultContractTx(&contractResultTx, data)
	if err != nil {
		return err, fee, nil, nil, ret
	}

	//4.Update Account MPT tree.
	root := inputData.StateDB.IntermediateRoot()
	fee.Addition(*contractFee)

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := core.NewReceipt(root.CloneBytes(), failed, *inputData.UsedGas)
	receipt.TxHash = *tx.GetTxID()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To().IsEmpty() {
		receipt.ContractAddress = vm.CreateContractAccountID(msg.from, *tx.GetTxID())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = adapter.GetLogs(*tx.GetTxID())
	receipt.Bloom = core.CreateBloom(core.Receipts{receipt})
	log.Debug("contract", "addr", receipt.ContractAddress)
	log.Debug("contract resultTx", "txid", contractResultTx.GetTxID().String(), "tx", contractResultTx.String())
	log.Debug("check contract is failed", "failed", failed, "err", err)
	return err, fee, receipt, &contractResultTx, ret
}
func ConvertToMessage(tx *meta.Transaction) Message {
	// AsMessage returns the transaction as a core.Message.
	//
	// AsMessage requires a signer to derive the sender.
	//
	// XXX Rename message to something less arbitrary?
	txData := GetTxData(tx)
	msg := Message{
		from:     tx.From.Coins[0].Id,
		gasLimit: txData.GasLimit,
		gasPrice: new(big.Int).Set(txData.Price),
		to:       &tx.To.Coins[0].Id,
		amount:   tx.To.Coins[0].Value.GetBigInt(),
		data:     txData.Payload,
	}

	return msg
}

//Update contractTx fromAccount,unCoinBaseTx is not coinBase tx,fromAccount is tx from.
//Only update account which is related to tx from.
func processTxFrom(tx *meta.Transaction, data interpreter.Params) (error, *meta.Amount) {
	inputData := data.(*Input)
	fcValue := meta.NewAmount(0)
	for _, fc := range tx.From.Coins {
		fromObj := inputData.StateDB.GetObject(meta.GetAccountHash(fc.GetId()))
		if fromObj == nil {
			return errors.New("processTxFrom(contractTx)->can not find tx from"), fcValue
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
func processContractTxTo(tx *meta.Transaction, data interpreter.Params) error {
	inputData := data.(*Input)
	txId := tx.GetTxID()
	for index := range tx.To.Coins {
		if index == 0 {
			//the create or call contract. frist must be add to value to caller.
			//because transfer value to contract account will be excute in vm, then we must be add the same value to caller.
			fromObj := inputData.StateDB.GetObject(meta.GetAccountHash(tx.From.Coins[0].GetId()))
			nfTicket := meta.NewTicket(*txId, uint32(index))
			nfUTXO := meta.NewUTXO(nfTicket, inputData.Header.Height, inputData.Header.Height, *tx.To.Coins[index].GetValue())
			fromObj.GetAccount().UTXOs = append(fromObj.GetAccount().UTXOs, *nfUTXO)
			inputData.StateDB.SetObject(fromObj)
			continue
		}

		//the back change
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

//Update commonTx Account.
//CommonTx is coinBase and normal tx.
func processResultContractTxTo(tx *meta.Transaction, data interpreter.Params) error {
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
