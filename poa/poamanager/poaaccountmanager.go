package poamanager

import (
	"errors"

	"github.com/linkchain/meta/account"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta/tx"

	poameta "github.com/linkchain/poa/meta"
	"github.com/linkchain/common/btcec"
)

type POAAccountManager struct {
	accountMap map[string]poameta.POAAccount
}

/** interface: common.IService **/
func (m *POAAccountManager) Init(i interface{}) bool{
	log.Info("POAAccountManager init...");
	m.accountMap = make(map[string]poameta.POAAccount)

	return true
}

func (m *POAAccountManager) Start() bool{
	log.Info("POAAccountManager start...");
	return true
}

func (m *POAAccountManager) Stop(){
	log.Info("POAAccountManager stop...");
}


func (m *POAAccountManager) NewAccount() account.IAccount  {
	priv,err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Info("POAAccountManager","NewAccount - generate private key failed",err)
	}
	accountID := *poameta.NewAccountId(priv.PubKey().SerializeCompressed()).(*poameta.POAAccountID)
	a := poameta.POAAccount{AccountID:accountID,Value:poameta.POAAmount{Value:int32(0)}}
	return &a
}


func (m *POAAccountManager) AddAccount(iAccount account.IAccount) error  {
	a := *iAccount.(*poameta.POAAccount)
	m.accountMap[iAccount.GetAccountID().GetString()] = a
	return nil
}

func (m *POAAccountManager) GetAccount(id account.IAccountID) (account.IAccount,error) {
	a,ok := m.accountMap[id.GetString()]
	if ok {
		return &a,nil
	}
	return nil,errors.New("Can not find Account ")
}

func (m *POAAccountManager) RemoveAccount(id account.IAccountID) error  {
	delete(m.accountMap,id.GetString())
	return nil
}

func (m *POAAccountManager) UpdateAccountByTX(tx tx.ITx) error {
	fromAccountId := tx.GetFrom().GetID()
	toAccountId := tx.GetTo().GetID()

	fromAccount,err := m.GetAccount(fromAccountId)
	if err != nil {
		log.Error("POAAccountManager","UpdateAccountByTX","can not find the account of the tx's")
		return err
	}

	amount := tx.GetAmount()

	if fromAccount.GetAmount().IsLessThan(amount) {
		log.Error("POAAccountManager","UpdateAccountByTX","the from of tx doesn't have enough money to pay")
		return errors.New("UpdateAccountByTX the from of tx doesn't have enough money to pay")
	}

	log.Info("POAAccountManager","fromAccount nounce",fromAccount.GetNounce(),"tx nounce",tx.GetNounce())
	if !fromAccount.CheckNounce(tx.GetNounce()) {
		log.Error("POAAccountManager","CheckTxFromAccount","the from of tx doesn't have corrent nounce")
		return errors.New("CheckTxFromAccount the from of tx doesn't have corrent nounce")
	}

	fromAmount := poameta.NewPOAAmout(0)
	fromAmount.Subtraction(amount)
	fromAccount.ChangeAmount(&fromAmount)
	fromAccount.SetNounce(tx.GetNounce())
	m.UpdateAccount(fromAccount)

	toNounce := uint32(0)
	toAccount, err := m.GetAccount(toAccountId)
	if err == nil {
		toNounce = toAccount.GetNounce()
	}
	a := poameta.NewPOAAccount(toAccountId,amount,toNounce)
	m.UpdateAccount(&a)

	return nil
}

func (m *POAAccountManager) UpdateAccount(iAccount account.IAccount) error {
	newAccount,err := m.GetAccount(iAccount.GetAccountID())
	if err == nil {
		newAccount.GetAmount().Addition(iAccount.GetAmount())
		m.AddAccount(newAccount)
	} else {
		m.AddAccount(iAccount)
	}

	return nil
}

func (m *POAAccountManager) CheckTxFromAccount(tx tx.ITx) error {
	fromAccountId := tx.GetFrom().(*poameta.POATransactionPeer).AccountID
	fromAccount,err := m.GetAccount(&fromAccountId)
	amount := tx.GetAmount()

	if err != nil {
		log.Error("POAAccountManager","CheckTxFromAccount","can not find the account of the tx's")
		log.Error("POAAccountManager","tx from",fromAccountId.GetString())
		return err
	}

	if fromAccount.GetAmount().IsLessThan(amount) {
		log.Error("POAAccountManager","CheckTxFromAccount","the from of tx doesn't have enough money to pay")
		return errors.New("CheckTxFromAccount the from of tx doesn't have enough money to pay")
	}

	if !fromAccount.CheckNounce(tx.GetNounce()) {
		log.Error("POAAccountManager","CheckTxFromAccount","the from of tx doesn't have corrent nounce")
		return errors.New("CheckTxFromAccount the from of tx doesn't have corrent nounce")
	}

	//checkSign
	err = tx.Verify()
	if err != nil {
		return err
	}

	return nil
}

func (m *POAAccountManager) GetAllAccounts()  {
	for _,accountId := range m.accountMap{
		log.Info("POAAccountManager",accountId.AccountID.GetString(),accountId.Value.GetString())
	}
}


