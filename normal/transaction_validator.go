package normal

import (
	"errors"
	"github.com/mihongtech/linkchain/interpreter"

	"github.com/mihongtech/linkchain/common"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core/meta"
)

func (n *Interpreter) CheckTx(tx *meta.Transaction) error {
	if IsNormal(tx.Type) {
		var err error = nil
		if err = commonCheck(tx); err != nil {
			return err
		}

		switch tx.Type {
		case config.CoinBaseTx:
			err = checkCoinBaseTx(tx)
		case config.NormalTx:
			err = checkNormalTx(tx)
		}
		return err
	} else {
		return interpreter.ErrKnownTxType
	}
}

func commonCheck(tx *meta.Transaction) error {
	//check toCount/fromCount/signCount must be >0
	if toCount := len(tx.To.Coins); toCount <= 0 {
		return errors.New("the tx from/to/sign count must be more than 0")
	}

	return nil
}

func CheckToZero(tx *meta.Transaction) error {
	//check output value must be > 0
	for _, tc := range tx.To.Coins {
		if !tc.CheckValue() {
			return errors.New("the tx toValue must be more than 0")
		}
	}
	return nil
}

func checkCoinBaseTx(tx *meta.Transaction) error {
	toCount := len(tx.To.Coins)
	if toCount != 1 {
		return errors.New("the coin base tx must be only one to")
	}

	fromCount := len(tx.From.Coins)
	signCount := len(tx.Sign)
	if fromCount > 0 || signCount > 0 {
		return errors.New("the coin base tx from/sign count must be 0")
	}

	if err := CheckToZero(tx); err != nil {
		return err
	}
	return nil
}

func checkNormalTx(tx *meta.Transaction) error {
	if err := CheckToZero(tx); err != nil {
		return err
	}
	if err := CheckFromCount(tx); err != nil {
		return err
	}

	if err := checkUnCoinBaseTx(tx); err != nil {
		return err
	}
	return tx.Verify()
}

func checkUnCoinBaseTx(tx *meta.Transaction) error {
	fromCount := len(tx.From.Coins)

	//all of from id in tx must be different
	for i := 0; i < fromCount; i++ {
		for j := i + 1; j < fromCount; j++ {
			tempAccountId := tx.From.Coins[j].Id
			if tx.From.Coins[i].Id.IsEqual(tempAccountId) {
				return errors.New("all of from ticket in tx with must have different from account id")
			}
		}
	}
	//all of from ticket in tx must be different
	tickets := make([]meta.Ticket, 0)
	for _, fc := range tx.From.Coins {
		tickets = append(tickets, fc.Ticket...)
	}

	count := len(tickets)
	for i := 0; i < count; i++ {
		for j := i + 1; j < count; j++ {
			tempTxid := tickets[j].Txid
			if tickets[i].Txid.IsEqual(&tempTxid) && tickets[i].Index == tickets[j].Index {
				return errors.New("all of from ticket in tx with must have different ticket")
			}
		}
	}
	return nil
}

func CheckFromCount(tx *meta.Transaction) error {
	fromCount := len(tx.From.Coins)
	signCount := len(tx.Sign)
	if fromCount <= 0 || signCount <= 0 {
		return errors.New("the tx from/sign count must be more than 0")
	}
	return nil
}

func (n *Interpreter) VerifyTx(tx *meta.Transaction, data interpreter.Params) error {
	if IsNormal(tx.Type) {
		var err error = nil
		switch tx.Type {
		case config.CoinBaseTx:
			err = verifyCoinBaseTx(tx, data)
		case config.NormalTx:
			err = verifyNormalTx(tx, data)
		}
		return err
	} else {
		return interpreter.ErrKnownTxType
	}
}

func verifyNormalTx(tx *meta.Transaction, data interpreter.Params) error {
	inputData := data.(*Input)

	if err := VerifyUnCoinBaseTx(tx, data); err != nil {
		return err
	}

	for _, fc := range tx.From.Coins {
		fromObj := inputData.StateDB.GetObject(meta.GetAccountHash(fc.GetId()))
		if fromObj == nil {
			return errors.New("verifyNormalTx()->can not find tx from")
		}

		if fromObj.GetAccount().AccountType != config.NormalAccount {
			return errors.New("the from of normal tx must be normal account")
		}

		if !fromObj.GetAccount().IsFromEffect(&fc, inputData.Header.Height) {
			return errors.New("verifyDelayTx()->the from ticket had not reach to effect height")
		}
	}
	return nil
}

func verifyCoinBaseTx(tx *meta.Transaction, data interpreter.Params) error {
	inputData := data.(*Input)
	if height, err := common.BytesToUInt32(tx.Data); height != inputData.Header.Height || err != nil {
		return errors.New("the coin base tx height must be equal to block height")
	}

	for _, tc := range tx.To.Coins {
		toObj := inputData.StateDB.GetObject(meta.GetAccountHash(tc.GetId()))
		if toObj != nil {
			if toObj.GetAccount().AccountType != config.NormalAccount {
				return errors.New("the to of normal tx must be normal account")
			}
		}
	}
	return nil
}

func VerifyUnCoinBaseTx(tx *meta.Transaction, data interpreter.Params) error {
	inputData := data.(*Input)
	fcValue := meta.NewAmount(0)
	tcValue := tx.GetToValue()
	//Interpreter verify
	for _, fc := range tx.From.Coins {
		fromObj := inputData.StateDB.GetObject(meta.GetAccountHash(fc.GetId()))
		if fromObj == nil {
			return errors.New("verifyUnCoinBaseTx()->can not find tx from")
		}

		if ok := fromObj.GetAccount().CheckFromCoin(&fc); !ok {
			return errors.New("cache account can not contain fromCoin")
		}
		value, err := fromObj.GetAccount().GetFromCoinValue(&fc)
		if err != nil {
			return err
		}
		fcValue.Addition(*value)
	}

	if fcValue.IsLessThan(*tcValue) {
		return errors.New("the tx from value < to value")
	}
	return nil
}
