package poamanager

import (
	"github.com/linkchain/meta/account"
	"github.com/linkchain/common/util/log"
	poameta "github.com/linkchain/poa/meta"
	"crypto/sha256"
	"github.com/linkchain/common/math"
	"time"
	"errors"
	"github.com/linkchain/meta/tx"
)

type POAAccountManager struct {
	accountMap map[poameta.POAAccountID]poameta.POAAccount
}

/** interface: common.IService **/
func (m *POAAccountManager) Init(i interface{}) bool{
	log.Info("POABlockManager init...");
	m.accountMap = make(map[poameta.POAAccountID]poameta.POAAccount)

	return true
}

func (m *POAAccountManager) Start() bool{
	log.Info("POABlockManager start...");
	return true
}

func (m *POAAccountManager) Stop(){
	log.Info("POABlockManager stop...");
}


func (m *POAAccountManager) NewAccount() account.IAccount  {
	t := time.Now()
	accountID := math.Hash(sha256.Sum256([]byte(t.String())))
	account := poameta.POAAccount{AccountID:poameta.POAAccountID{ID:accountID},Value:poameta.POAAmount{Value:int32(t.Day())}}
	return &account
}


func (m *POAAccountManager) AddAccount(iAccount account.IAccount) error  {
	accountId := *iAccount.GetAccountID().(*poameta.POAAccountID)
	account := *iAccount.(*poameta.POAAccount)
	m.accountMap[accountId] = account
	return nil
}

func (m *POAAccountManager) GetAccount(id account.IAccountID) (account.IAccount,error) {
	account,ok := m.accountMap[*id.(*poameta.POAAccountID)]
	if ok {
		return &account,nil
	}
	return nil,errors.New("Can not find Account ")
}

func (m *POAAccountManager) RemoveAccount(id account.IAccountID) error  {
	delete(m.accountMap,*id.(*poameta.POAAccountID))
	return nil
}

func (m *POAAccountManager) UpdateAccountByTX(tx tx.ITx) error {
	fromAccountId := tx.GetFrom().(*poameta.POATransactionPeer).AccountID
	toAccountId := tx.GetTo().(*poameta.POATransactionPeer).AccountID

	fromAccount,error1 := m.GetAccount(&fromAccountId)


	if error1 != nil {
		log.Error("POAAccountManager","update account status","can not find the account of the tx's")
		return error1
	}

	amount := tx.GetAmount()

	if fromAccount.GetAmount().IsLessThan(amount) {
		log.Error("POAAccountManager","update account status","the from of tx doesn't have enough money to pay")
		return errors.New("update account status the from of tx doesn't have enough money to pay")
	}

	fromAmount := poameta.POAAmount{Value:0}
	fromAmount.Subtraction(amount)
	fromAccount.ChangeAmount(&fromAmount)
	m.UpdateAccount(fromAccount)

	toAccount := &poameta.POAAccount{AccountID:toAccountId,Value:*amount.(*poameta.POAAmount)}
	m.UpdateAccount(toAccount)
	return nil
}

func (m *POAAccountManager) UpdateAccount(iAccount account.IAccount) error {
	account := *iAccount.(*poameta.POAAccount)

	newAccount,error := m.GetAccount(iAccount.GetAccountID())
	if error == nil {
		newAccount.GetAmount().Addition(account.GetAmount())
		m.AddAccount(newAccount)
	} else {
		m.AddAccount(iAccount)
	}

	return nil
}

func (m *POAAccountManager) CheckTxFromAccount(tx tx.ITx) error {
	fromAccountId := tx.GetFrom().(*poameta.POATransactionPeer).AccountID
	fromAccount,error1 := m.GetAccount(&fromAccountId)
	amount := tx.GetAmount()

	if error1 != nil {
		log.Error("POAAccountManager","update account status","can not find the account of the tx's")
		return error1
	}

	if fromAccount.GetAmount().IsLessThan(amount) {
		log.Error("POAAccountManager","update account status","the from of tx doesn't have enough money to pay")
		return errors.New("update account status the from of tx doesn't have enough money to pay")
	}
	return nil
}

func (m *POAAccountManager) GetAllAccounts()  {
	for _,accountId := range m.accountMap{
		log.Info("POAAccountManager",accountId.AccountID.GetString(),accountId.Value.GetString())
	}
}


