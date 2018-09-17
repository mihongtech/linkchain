package poamanager

import (
	"time"
	"errors"

	"github.com/linkchain/meta/account"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/common/math"
	"github.com/linkchain/meta/tx"

	poameta "github.com/linkchain/poa/meta"
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
	accountID := math.DoubleHashH([]byte(t.String()))
	a := poameta.POAAccount{AccountID:poameta.POAAccountID{ID:accountID},Value:poameta.POAAmount{Value:int32(t.Day())}}
	return &a
}


func (m *POAAccountManager) AddAccount(iAccount account.IAccount) error  {
	aId := *iAccount.GetAccountID().(*poameta.POAAccountID)
	a := *iAccount.(*poameta.POAAccount)
	m.accountMap[aId] = a
	return nil
}

func (m *POAAccountManager) GetAccount(id account.IAccountID) (account.IAccount,error) {
	a,ok := m.accountMap[*id.(*poameta.POAAccountID)]
	if ok {
		return &a,nil
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

	fromAccount,err := m.GetAccount(&fromAccountId)


	if err != nil {
		log.Error("POAAccountManager","update account status","can not find the account of the tx's")
		return err
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
	a := *iAccount.(*poameta.POAAccount)

	newAccount,err := m.GetAccount(iAccount.GetAccountID())
	if err == nil {
		newAccount.GetAmount().Addition(a.GetAmount())
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
		log.Error("POAAccountManager","update account status","can not find the account of the tx's")
		return err
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


