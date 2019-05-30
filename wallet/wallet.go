package wallet

import (
	"encoding/hex"
	"errors"
	"path/filepath"

	"github.com/mihongtech/linkchain/accounts"
	"github.com/mihongtech/linkchain/accounts/keystore"
	"github.com/mihongtech/linkchain/app/context"
	"github.com/mihongtech/linkchain/common/btcec"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/util/event"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/helper"
	"github.com/mihongtech/linkchain/node"
)

type Wallet struct {
	keystore         *keystore.KeyStore
	password         string
	Name             string
	DataDir          string
	accounts         map[string]meta.Account
	nodeAPI          *node.PublicNodeAPI
	updateAccountSub *event.TypeMuxSubscription
}

func NewWallet() *Wallet {
	name := "wallet"
	password := "password"
	return &Wallet{accounts: make(map[string]meta.Account), Name: name, password: password}
}

func (w *Wallet) Setup(i interface{}) bool {
	globalConfig := i.(*context.Context).Config
	w.nodeAPI = i.(*context.Context).NodeAPI.(*node.PublicNodeAPI)
	w.DataDir = globalConfig.DataDir
	path := w.instanceDir(w.DataDir)
	w.keystore = keystore.NewKeyStore(path, keystore.StandardScryptN, keystore.StandardScryptP)
	w.nodeAPI = i.(*context.Context).NodeAPI.(*node.PublicNodeAPI)
	return true
}

func (w *Wallet) Start() bool {
	accountEvent := w.nodeAPI.GetAccountEvent()
	w.updateAccountSub = accountEvent.Subscribe(node.AccountEvent{})
	ksAccounts := w.keystore.Accounts()
	for i := range ksAccounts {
		account := helper.CreateTemplateAccount(ksAccounts[i].Address)
		w.accounts[account.Id.String()] = *account
	}
	w.reScanAllAccount()
	go w.updateWalletLoop()

	return true
}

func (w *Wallet) Stop() {
	log.Info("Stop wallet...")
	w.updateAccountSub.Unsubscribe()
}

func (w *Wallet) updateWalletLoop() {
	for obj := range w.updateAccountSub.Chan() {
		switch ev := obj.Data.(type) {
		case node.AccountEvent:
			if ev.IsUpdate {
				w.reScanAllAccount()
			}
		}
	}
}

func (w *Wallet) reScanAllAccount() {
	newWas := make([]meta.Account, 0)
	for key := range w.accounts {
		wa := w.accounts[key]
		newWa, err := w.queryAccount(wa.Id)
		if err != nil {
			continue
		}

		newWas = append(newWas, newWa)
	}
	for _, wa := range newWas {
		w.updateWalletAccount(wa)
	}
}

func (w *Wallet) updateWalletAccount(account meta.Account) error {
	a, ok := w.accounts[account.GetAccountID().String()]
	if !ok {
		return errors.New("GetAccountID can not find account")
	}

	a = account
	w.AddAccount(a)
	return nil
}

func (w *Wallet) NewAccount() (*meta.AccountID, error) {
	ksAccount, err := w.keystore.NewAccount(w.password)
	if err != nil {
		log.Error("wallet", "newAccount", err)
		return nil, err
	}
	account := helper.CreateTemplateAccount(ksAccount.Address)
	w.AddAccount(*account)
	return &ksAccount.Address, nil
}

func (w *Wallet) AddAccount(account meta.Account) {
	w.accounts[account.Id.String()] = account
}

func (w *Wallet) GetAllWAccount() []meta.Account {
	w.reScanAllAccount()
	var WAs []meta.Account
	for a := range w.accounts {
		WAs = append(WAs, w.accounts[a])
	}
	return WAs
}

func (w *Wallet) GetAccount(key string) (*meta.Account, error) {
	wa, ok := w.accounts[key]
	if ok {
		return &wa, nil
	} else {
		id, err := meta.HexToAccountID(key)
		if err != nil {
			return nil, err
		}
		newWa, err := w.queryAccount(id)
		return &newWa, err
	}
}

func (w *Wallet) queryAccount(id meta.AccountID) (meta.Account, error) {
	stateDB, err := w.nodeAPI.StateAt(w.nodeAPI.GetBestBlock().Header.Status)
	if err != nil {
		return meta.Account{}, err
	}

	stateObject := stateDB.GetObject(meta.GetAccountHash(id))
	if stateObject == nil {
		return meta.Account{}, errors.New("can not find IAccount")
	}
	return *stateObject.GetAccount(), nil
}

func (w *Wallet) SignTransaction(tx meta.Transaction) (*meta.Transaction, error) {
	for _, fc := range tx.GetFromCoins() {
		sign, err := w.SignMessage(fc.Id, tx.GetTxID().CloneBytes())
		if err != nil {
			return nil, err
		}
		tx.AddSignature(sign)
	}
	return &tx, nil
}

func (w *Wallet) SignMessage(accountId meta.AccountID, hash []byte) (math.ISignature, error) {
	_, ok := w.accounts[accountId.String()]
	if !ok {
		return nil, errors.New("SignMessage can not find account id")
	}

	ksAccount, err := w.keystore.Find(accounts.Account{Address: accountId})
	if err != nil {
		return nil, err
	}
	sign, err := w.keystore.SignHashWithPassphrase(ksAccount, w.password, hash)
	return meta.NewSignature(sign), nil
}

func (w *Wallet) importKey(privkeyStr string) (*meta.AccountID, error) {
	privkeyBuff, err := hex.DecodeString(privkeyStr)
	if err != nil {
		return nil, err
	}
	privkey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privkeyBuff)
	ksAccount, err := w.keystore.ImportECDSA(privkey, w.password)
	if err != nil {
		return nil, err
	}
	account := helper.CreateTemplateAccount(ksAccount.Address)
	w.AddAccount(*account)
	return &ksAccount.Address, err
}

func (w *Wallet) ImportAccount(privateKeyStr string) (*meta.AccountID, error) {
	a, err := w.importKey(privateKeyStr)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (w *Wallet) ExportAccount(id meta.AccountID) (string, error) {
	_, ok := w.accounts[id.String()]
	if !ok {
		return "", errors.New("export can not find account id")
	}
	ksAccount, err := w.keystore.Find(accounts.Account{Address: id})
	if err != nil {
		return "", err
	}

	return w.keystore.ExportECDSA(ksAccount, w.password)
}

func (w *Wallet) instanceDir(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Join(path, w.Name)
}
