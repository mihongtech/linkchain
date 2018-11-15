package node

import (
	"errors"
	"github.com/linkchain/storage"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/config"
	"encoding/hex"
)


var (
	stateDB  = &storage.StateDB{}
)

func Init() {

}

func initAccountManager(){
	stateDB.Init(nil)
	stateDB.Start()
}

// add a account in accountmanage and if return nil , add success.
func addAccount(iAccount meta.Account) error {
	return stateDB.SetAccount(iAccount)
}

// get a account by accountid in accountmanage and return interface of account and error message,if error message is nil,search is success.
func GetAccount(id meta.AccountID) (meta.Account, error) {
	a, ok := stateDB.GetAccount(id)
	if ok {
		return *a, nil
	}
	return *a, errors.New("can not find IAccount ")
}

//TODO need to test
//update accountmanage by tx when add block on chain and if error message is nil,update success.
func updateAccountsByBlock(block *meta.Block) error {
	err := stateDB.UpdateAccountsByBlock(block)
	//NewWalletEvent.Post(events.WAccountEvent{IsUpdate: true})
	return err
}

//TODO need to test
//restore accountmanage by tx when remove block on chain and if error message is nil,restore success.
func revertAccountsByBlock(block *meta.Block) error {
	err := stateDB.RollBack(block)
	//Notice wallet
	//NewWalletEvent.Post(events.WAccountEvent{IsUpdate: true})
	return err
}

func CreateAccountIdByPubKey(pubKey string) (*meta.AccountID, error) {
	pkBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	pk, err := btcec.ParsePubKey(pkBytes, btcec.S256())
	if err != nil {
		return nil, err
	}
	return meta.NewAccountId(pk), nil
}

func CreateAccountIdByPrivKey(privKey string) (*meta.AccountID, error) {
	priv, err := hex.DecodeString(privKey)
	if err != nil {
		return nil, err
	}
	_, pk := btcec.PrivKeyFromBytes(btcec.S256(), priv)
	if err != nil {
		return nil, err
	}
	return meta.NewAccountId(pk), nil
}

func CreateTempleteAccount(id meta.AccountID) *meta.Account {
	utxo := make([]meta.UTXO, 0)
	a := meta.NewAccount(id, config.NormalAccount, utxo, config.DafaultClearTime, meta.AccountID{})
	return a
}

func CreateNormalAccount(key *btcec.PrivateKey) (*meta.Account, error) {
	privStr := hex.EncodeToString(key.Serialize())
	id, err := CreateAccountIdByPrivKey(privStr)
	if err != nil {
		return nil, err
	}

	a := CreateTempleteAccount(*id)
	return a, nil
}