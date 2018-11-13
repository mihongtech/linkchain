package manage

import (
	"errors"
	"sync"

	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/consensus/state"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/block"
	"github.com/linkchain/meta/events"
	poastate "github.com/linkchain/poa/state"
)

// AccountManage is a important manager which is handling  all th account status of the whole system.
type AccountManage struct {
	manageMtx      sync.Mutex
	stateDB        state.StateDBer
	NewWalletEvent *event.TypeMux
}

/** interface: common.IService **/
func (m *AccountManage) Init(i interface{}) bool {
	log.Info("AccountManage init...")
	m.stateDB = &poastate.StateDB{}
	m.stateDB.Init(nil)
	return true
}

func (m *AccountManage) Start() bool {
	log.Info("AccountManage start...")
	m.stateDB.Start()
	return true
}

func (m *AccountManage) Stop() {
	log.Info("AccountManage stop...")
	m.stateDB.Stop()
}

// add a account in accountmanage and if return nil , add success.
func (m *AccountManage) AddAccount(iAccount account.IAccount) error {
	return m.stateDB.SetAccount(iAccount)
}

// get a account by accountid in accountmanage and return interface of account and error message,if error message is nil,search is success.
func (m *AccountManage) GetAccount(id meta.IAccountID) (account.IAccount, error) {
	a, ok := m.stateDB.GetAccount(id)
	if ok {
		return a, nil
	}
	return nil, errors.New("Can not find IAccount ")
}

// get a account by accountid in accountmanage and if return nil , add success.
func (m *AccountManage) RemoveAccount(id meta.IAccountID) error {
	//TODO
	return nil
}

//TODO need to test
//update accountmanage by tx when add block on chain and if error message is nil,update success.
func (m *AccountManage) UpdateAccountsByBlock(block block.IBlock) error {
	err := m.stateDB.UpdateAccountsByBlock(block)
	m.NewWalletEvent.Post(events.WAccountEvent{IsUpdate: true})
	return err
}

//TODO need to test
//restore accountmanage by tx when remove block on chain and if error message is nil,restore success.
func (m *AccountManage) RevertAccountsByBlock(block block.IBlock) error {
	err := m.stateDB.RollBack(block)
	//Notice wallet
	m.NewWalletEvent.Post(events.WAccountEvent{IsUpdate: true})
	return err
}

//TODO need to remove
//get all account in account.
func (m *AccountManage) GetAllAccounts() {
	m.stateDB.GetAllAccount()
}
