package manage

import (
	"errors"
	"sync"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/tx"
	poameta "github.com/linkchain/poa/meta"
)

type AccountManage struct {
	accountMtx sync.RWMutex
	accountMap map[string]poameta.Account
}

func (m *AccountManage) readAccount(key string) (poameta.Account, bool) {
	m.accountMtx.RLock()
	defer m.accountMtx.RUnlock()
	value, ok := m.accountMap[key]
	return value, ok
}

func (m *AccountManage) writeAccount(key string, value poameta.Account) {
	m.accountMtx.Lock()
	defer m.accountMtx.Unlock()
	m.accountMap[key] = value
}

func (m *AccountManage) removeAccount(key string) {
	m.accountMtx.Lock()
	defer m.accountMtx.Unlock()
	delete(m.accountMap, key)
}

/** interface: common.IService **/
func (m *AccountManage) Init(i interface{}) bool {
	log.Info("AccountManage init...")
	m.accountMap = make(map[string]poameta.Account)
	return true
}

func (m *AccountManage) Start() bool {
	log.Info("AccountManage start...")
	return true
}

func (m *AccountManage) Stop() {
	log.Info("AccountManage stop...")
}

func (m *AccountManage) NewAccount() account.IAccount {
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Info("AccountManage", "NewAccount - generate private key failed", err)
	}
	id := *poameta.NewAccountId(priv.PubKey())
	amount := *poameta.NewAmout(0)
	a := poameta.NewAccount(id, amount, 0)
	return a
}

func (m *AccountManage) AddAccount(iAccount account.IAccount) error {
	a := *iAccount.(*poameta.Account)
	m.writeAccount(iAccount.GetAccountID().GetString(), a)
	return nil
}

func (m *AccountManage) GetAccount(id account.IAccountID) (account.IAccount, error) {
	a, ok := m.readAccount(id.GetString())
	if ok {
		return &a, nil
	}
	return nil, errors.New("Can not find Account ")
}

func (m *AccountManage) RemoveAccount(id account.IAccountID) error {
	m.removeAccount(id.GetString())
	return nil
}

func (m *AccountManage) UpdateAccountByTX(tx tx.ITx, isMine bool) error {
	fromAccountId := tx.GetFrom().GetID()
	toAccountId := tx.GetTo().GetID()
	amount := tx.GetAmount()
	if !isMine {
		fromAccount, err := m.GetAccount(fromAccountId)
		if err != nil {
			return err
		}

		err = m.checkFromAccount(fromAccountId, amount, tx.GetNounce(), true)
		if err != nil {
			return err
		}

		//update from an to Account Status
		fromAccount.ChangeAmount(fromAccount.GetAmount().Subtraction(amount))
		fromAccount.SetNounce(tx.GetNounce())
		m.AddAccount(fromAccount)
	}

	toAccount, err := m.GetAccount(toAccountId)
	if err != nil {
		toAccount = poameta.NewAccount(*toAccountId.(*poameta.AccountID), *amount.(*poameta.Amount), 0)
	} else {
		toAccount.ChangeAmount(toAccount.GetAmount().Addition(amount))
	}

	m.AddAccount(toAccount)
	return nil
}

func (m *AccountManage) CheckTxFromAccount(tx tx.ITx) error {
	fromAccountId := tx.GetFrom().GetID()
	amount := tx.GetAmount()

	err := m.checkFromAccount(fromAccountId, amount, tx.GetNounce(), true)
	if err != nil {
		return err
	}
	return nil
}

func (m *AccountManage) CheckTxFromNounce(tx tx.ITx) error {
	fromAccountId := tx.GetFrom().GetID()
	amount := tx.GetAmount()

	err := m.checkFromAccount(fromAccountId, amount, tx.GetNounce(), false)
	if err != nil {
		return err
	}

	return nil
}

func (m *AccountManage) checkFromAccount(fromId account.IAccountID, amount meta.IAmount, txNounce uint32, isStrict bool) error {
	fromAccount, err := m.GetAccount(fromId)
	if err != nil {
		log.Error("AccountManage", "Check from account", "can not find the account of the tx's")
		log.Error("AccountManage", "tx from", fromId.GetString())
		return err
	}

	if fromAccount.GetAmount().IsLessThan(amount) {
		return errors.New("Check from account the from of tx doesn't have enough money to pay")
	}

	if isStrict {
		if !fromAccount.CheckNounce(txNounce) {

			return errors.New("Check from account the from of tx doesn't have corrent nounce")
		}
	} else {
		if fromAccount.GetNounce() >= txNounce {
			log.Error("AccountManage", "Check from account", "the from of tx should be more than fromAccount nounce")
			return errors.New("Check from account the from of tx should be more than fromAccount nounce")
		}
	}

	return nil
}

func (m *AccountManage) GetAllAccounts() {
	m.accountMtx.RLock()
	defer m.accountMtx.RUnlock()

	for _, accountId := range m.accountMap {
		log.Info("AccountManage", accountId.AccountID.GetString(), accountId.Value.GetString(), "nounce", accountId.Nounce)
	}
}
