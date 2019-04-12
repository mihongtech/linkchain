package contract

import (
	"errors"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/normal"
)

func IsContract(txType uint32) bool {
	return txType == ContractTx || txType == ContractResultTx
}

//check tx which related to contract
func (v *Interpreter) CheckTx(tx *meta.Transaction) error {
	if normal.IsNormal(tx.Type) {
		normal := normal.Interpreter{}
		return normal.CheckTx(tx)
	} else if IsContract(tx.Type) {
		var err error = nil
		switch tx.Type {
		case ContractTx:
			err = checkContractTx(tx)
		}
		return err
	} else {
		return interpreter.ErrKnownTxType
	}
}

func checkContractTx(tx *meta.Transaction) error {
	if len(tx.From.Coins) != 1 {
		return errors.New("the contract tx from must be only one")
	}

	if len(tx.To.Coins) > 2 {
		return errors.New("the contract tx to must be less than two")
	}

	//If have backchange then check backchange id.
	if len(tx.To.Coins) == 2 {
		if !tx.From.Coins[0].Id.IsEqual(tx.To.Coins[1].Id) {
			return errors.New("the contract tx backchange addr must be same as from")
		}
	}

	//check tx data
	if len(tx.Data) == 0 {
		return errors.New("the len of contract tx data must be more than zero")
	} else {
		if GetTxData(tx) == nil {
			return errors.New("the data of contract tx is error")
		}
	}
	return nil
}

//verify tx which related to contract
func (v *Interpreter) VerifyTx(tx *meta.Transaction, data interpreter.Params) error {
	if normal.IsNormal(tx.Type) {
		normal := normal.Interpreter{}
		return normal.VerifyTx(tx, &(data.(*Input).Input))
	} else if IsContract(tx.Type) {
		if err := normal.VerifyUnCoinBaseTx(tx, &(data.(*Input).Input)); err != nil && ContractTx == tx.Type {
			return err
		}
		var err error = nil
		switch tx.Type {
		case ContractTx:
			err = verifyContractTx(tx, data)
		}

		return err
	} else {
		return interpreter.ErrKnownTxType
	}

}

func verifyContractTx(tx *meta.Transaction, data interpreter.Params) error {
	inputData := data.(*Input)
	//all of from in CreateOrReChargeTx must be normal account
	for _, fc := range tx.From.Coins {
		fromObj := inputData.StateDB.GetObject(meta.GetAccountHash(fc.GetId()))
		if fromObj == nil {
			return errors.New("verifyContractTx()->can not find tx from")
		}

		if fromObj.GetAccount().AccountType != config.NormalAccount {
			return errors.New("the from of contract tx must be normal account")
		}

		if !fromObj.GetAccount().IsFromEffect(&fc, inputData.Header.Height) {
			return errors.New("verifyCreateTx()->the from ticket had not reach to effect height")
		}
	}

	if !tx.To.Coins[0].Id.IsEmpty() {
		obj := inputData.StateDB.GetObject(meta.GetAccountHash(tx.To.Coins[0].Id))
		if obj.GetAccount().AccountType != ContractAccount {
			return errors.New("the frist to of contract tx must be contract account")
		}
	}
	return nil
}
