package node

import (
	"errors"

	"github.com/mihongtech/linkchain/core/meta"
)

func (n *Node) initAccountManager() {
	// TODO: add implement code
}

// get a account by accountid in accountmanage and return interface of account and error message,if error message is nil,search is success.
func (n *Node) getAccount(id meta.AccountID) (meta.Account, error) {
	stateDB, err := n.blockchain.State()
	if err != nil {
		return meta.Account{}, err
	}

	stateObject := stateDB.GetObject(meta.GetAccountHash(id))
	if stateObject == nil {
		return meta.Account{}, errors.New("can not find IAccount")
	}
	return *stateObject.GetAccount(), nil
}
