package manage

import (
	"errors"
	"sync"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/events"
	"github.com/linkchain/meta/tx"
	poameta "github.com/linkchain/poa/meta"
)

type AccountManage struct {
	accountMtx     sync.RWMutex
	accountMap     map[string]poameta.Account
	NewWalletEvent *event.TypeMux
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
	m.writeAccount(iAccount.GetAccountID().String(), a)
	return nil
}

func (m *AccountManage) GetAccount(id account.IAccountID) (account.IAccount, error) {
	a, ok := m.readAccount(id.String())
	if ok {
		return &a, nil
	}
	return nil, errors.New("Can not find IAccount ")
}

func (m *AccountManage) RemoveAccount(id account.IAccountID) error {
	m.removeAccount(id.String())
	return nil
}

func (m *AccountManage) GetAccountRelateTXs(tx tx.ITx, isMine bool) ([]account.IAccount, error) {
	fId := tx.GetFrom().GetID()
	tId := tx.GetTo().GetID()
	a := make([]account.IAccount, 0)
	if !isMine {
		fA, err := m.GetAccount(fId)
		if err != nil {
			return nil, err
		}
		a = append(a, fA)
	}

	tA, err := m.GetAccount(tId)
	if err == nil {
		a = append(a, tA)
	}
	return a, nil
}

func (m *AccountManage) ConvertAccount(tx tx.ITx, isMine bool) (account.IAccount, account.IAccount) {
	fId := tx.GetFrom().GetID()
	tId := tx.GetTo().GetID()
	var fA, tA account.IAccount
	toAmount := *tx.GetAmount().(*poameta.Amount)
	if !isMine {
		fromAmount := poameta.NewAmout(toAmount.Value)
		fromAmount.Reverse()
		fA = poameta.NewAccount(*fId.(*poameta.AccountID), *fromAmount, tx.GetNounce())
	}
	tA = poameta.NewAccount(*tId.(*poameta.AccountID), toAmount, 0)

	return fA, tA
}

//TODO need to test
func (m *AccountManage) UpdateAccountsByTxs(txs []tx.ITx, mineIndex int) error {
	cache := make(map[string]account.IAccount)
	for index, t := range txs {
		txA, err := m.GetAccountRelateTXs(t, index == mineIndex)
		if err != nil {
			return err
		}
		for _, a := range txA {
			key := a.GetAccountID().String()
			cache[key] = a
		}
	}

	//when cache is empty,only update mineTx
	//if not txs only contain mineTx,then update
	if len(cache) == 0 {
		if len(txs) != 1 || mineIndex != 0 {
			return errors.New("When cache is empty,only update mineTx")
		}
	}
	//Check Tx IAccount
	for index, t := range txs {
		fA, tA := m.ConvertAccount(t, index == mineIndex)

		if index != mineIndex {
			fKey := fA.GetAccountID().String()
			cachefA, _ := cache[fKey]
			err := m.checkAccount(cachefA, t.GetAmount(), t.GetNounce(), true)
			if err != nil {
				return err
			}
			cachefA.SetNounce(fA.GetNounce())
			cachefA.ChangeAmount(cachefA.GetAmount().Addition(fA.GetAmount()))
			cache[fKey] = cachefA
		}
		tKey := tA.GetAccountID().String()
		cachetA, ok := cache[tKey]
		if ok {
			cachetA.ChangeAmount(cachetA.GetAmount().Addition(tA.GetAmount()))
		} else {
			cachetA = tA
		}
		cache[tKey] = cachetA
	}

	//Update Accounts
	for _, a := range cache {
		m.AddAccount(a)
	}

	//Notice wallet
	m.NewWalletEvent.Post(events.WAccountEvent{IsUpdate: true})
	return nil
}

//TODO need to test
func (m *AccountManage) RevertAccountsByTxs(txs []tx.ITx, mineIndex int) error {
	cache := make(map[string]account.IAccount)
	for index, t := range txs {
		txA, _ := m.GetAccountRelateTXs(t, index == mineIndex)
		for _, a := range txA {
			key := a.GetAccountID().String()
			cache[key] = a
		}
	}

	//Check Tx IAccount
	for index, t := range txs {
		t.GetAmount().Reverse()
		t.SetNounce(t.GetNounce() - 1)
		fA, tA := m.ConvertAccount(t, index == mineIndex)

		if index != mineIndex {
			fKey := fA.GetAccountID().String()
			cachefA, _ := cache[fKey]
			cachefA.SetNounce(fA.GetNounce())
			cachefA.ChangeAmount(cachefA.GetAmount().Addition(fA.GetAmount()))
			cache[fKey] = cachefA
		}
		tKey := tA.GetAccountID().String()
		cachetA, _ := cache[tKey]
		cachetA.ChangeAmount(cachetA.GetAmount().Addition(tA.GetAmount()))
		cache[tKey] = cachetA
	}

	//Update Accounts
	for _, a := range cache {
		m.AddAccount(a)
	}
	return nil
}

func (m *AccountManage) CheckTxAccount(tx tx.ITx) error {
	fromAccountId := tx.GetFrom().GetID()
	amount := tx.GetAmount()
	fromAccount, err := m.GetAccount(fromAccountId)
	if err != nil {
		return err
	}
	return m.checkAccount(fromAccount, amount, tx.GetNounce(), false)
}

func (m *AccountManage) checkAccount(fromAccount account.IAccount, amount meta.IAmount, txNounce uint32, isStrict bool) error {
	if fromAccount.GetAmount().IsLessThan(amount) {
		return errors.New("checkAccount() the from of tx doesn't have enough money to pay")
	}

	if isStrict {
		if !fromAccount.CheckNounce(txNounce) {
			return errors.New("checkAccount() the from of tx doesn't have corrent nounce")
		}
	} else {
		if fromAccount.GetNounce() >= txNounce {
			log.Error("checkAccount", "from", fromAccount.GetNounce(), "tx", txNounce)
			return errors.New("checkAccount() the from of tx should be more than fromAccount nounce")
		}
	}
	return nil
}

func (m *AccountManage) GetAllAccounts() {
	m.accountMtx.RLock()
	defer m.accountMtx.RUnlock()

	for _, accountId := range m.accountMap {
		log.Info("AccountManage", "account", accountId.GetAccountID().String(), "amount", accountId.GetAmount().GetInt(), "nounce", accountId.GetNounce())
	}
}
