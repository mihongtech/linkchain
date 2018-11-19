package node

import (
	"errors"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
)

func (n *Node) initAccountManager() {
	n.stateDB.Setup(nil)
	n.stateDB.Start()
}

// add a account in accountmanage and if return nil , add success.
func (n *Node) addAccount(iAccount meta.Account) error {
	return n.stateDB.SetAccount(iAccount)
}

// get a account by accountid in accountmanage and return interface of account and error message,if error message is nil,search is success.
func (n *Node) getAccount(id meta.AccountID) (meta.Account, error) {
	a, ok := n.stateDB.GetAccount(id)
	if ok {
		return *a, nil
	}
	return *a, errors.New("can not find IAccount ")
}

//TODO need to test
//update accountmanage by tx when add block on chain and if error message is nil,update success.
func (n *Node) updateAccountsByBlock(block *meta.Block) error {
	err := n.stateDB.UpdateAccountsByBlock(block)
	//NewWalletEvent.Post(events.WAccountEvent{IsUpdate: true})
	return err
}

//TODO need to test
//restore accountmanage by tx when remove block on chain and if error message is nil,restore success.
func (n *Node) revertAccountsByBlock(block *meta.Block) error {
	err := n.stateDB.RollBack(block)
	//Notice wallet
	//NewWalletEvent.Post(events.WAccountEvent{IsUpdate: true})
	return err
}

func (n *Node) getAccountInfo() {
	log.Info("getAccountInfo")
	n.stateDB.GetAllAccount()
}
