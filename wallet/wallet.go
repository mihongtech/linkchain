package wallet

import (
	"encoding/hex"
	"errors"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/helper"
	"github.com/linkchain/node"
)

var minePriv, _ = hex.DecodeString("7a9c6f2b865c98c9fe174869de5818f4c62bc845441c08269487cdba6688f6b1")

type WAccount struct {
	privKey btcec.PrivateKey
	account meta.Account
}

func NewWSAccount() WAccount {
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Info("Wallet", "NewAccount - generate private key failed", err)
	}

	a, err := helper.CreateNormalAccount(priv)
	if err != nil {
		log.Info("Wallet", "NewAccount - failed", err)
	}
	return WAccount{privKey: *priv, account: *a}
}

func CreateWAccountFromBytes(privb []byte) WAccount {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), privb)
	a, err := helper.CreateNormalAccount(priv)
	if err != nil {
		log.Info("Wallet", "NewAccount - failed", err)
	}
	return WAccount{*priv, *a}
}

func (wa *WAccount) UpdateWAccount(account meta.Account) error {
	if !wa.account.GetAccountID().IsEqual(*account.GetAccountID()) {
		return errors.New("IAccount is error")
	}
	wa.account = account
	return nil
}

func (wa *WAccount) GetAccountID() meta.AccountID {
	id := meta.NewAccountId(wa.privKey.PubKey())
	return *id
}

func (wa *WAccount) getAccount() *meta.Account {
	return &wa.account
}

func (wa *WAccount) MakeFromCoin(value *meta.Amount) (*meta.FromCoin, *meta.Amount, error) {
	if wa.GetAmount() < value.GetInt64() {
		return nil, nil, errors.New("WAccount MakeFromCoin() amount is too large")
	}
	fc := helper.CreateFromCoin(wa.GetAccountID())
	fromAmount := meta.NewAmount(0)
	for _, v := range wa.account.UTXOs {
		fromAmount.Addition(v.Value)
		t := meta.NewTicket(v.Txid, v.Index)
		fc.AddTicket(t)
	}

	return fc, fromAmount, nil
}

func (wa *WAccount) GetAccountInfo() {
	log.Info("Wallet Info", "account", wa.account.GetAccountID().String(), "amount", wa.GetAmount(), "accounts", wa.account)
	for _, c := range wa.account.UTXOs {
		log.Info("Wallet Info", "Tickets", c.String())
	}
}

func (wa *WAccount) GetAccountPubkey() string {
	return hex.EncodeToString(wa.privKey.PubKey().SerializeCompressed())
}

func (wa *WAccount) GetAccountPrivkey() string {
	return hex.EncodeToString(wa.privKey.Serialize())
}

func (wa *WAccount) GetAmount() int64 {
	return wa.account.GetAmount().GetInt64()
}

func (wa *WAccount) Sign(messageHash []byte) math.ISignature {
	signature, err := wa.privKey.Sign(messageHash)
	if err != nil {
		log.Error("WAccount", "Sign", err)
		return nil
	}
	return meta.NewSignatrue(signature.Serialize())
}

type Wallet struct {
	accounts         map[string]WAccount
	nodeAPI          *node.PublicNodeAPI
	updateAccountSub *event.TypeMuxSubscription
}

func (w *Wallet) Setup(i interface{}) bool {
	log.Info("Wallet init...")
	w.accounts = make(map[string]WAccount)
	//w.nodeAPI = app.GetNodeAPI()
	return true
}

func (w *Wallet) Start() bool {
	log.Info("Wallet start...")
	gensisWA := CreateWAccountFromBytes(minePriv)
	gensisKey := hex.EncodeToString(gensisWA.privKey.PubKey().SerializeCompressed())

	w.accounts[gensisKey] = gensisWA
	w.updateAccountSub = w.nodeAPI.GetAccountEvent().Subscribe(node.AccountEvent{})
	w.ReScanAllAccount()
	go w.updateWalletLoop()

	return true
}

func (w *Wallet) Stop() {
	log.Info("Wallet stop...")
	w.updateAccountSub.Unsubscribe()
}

func (w *Wallet) updateWalletLoop() {
	for obj := range w.updateAccountSub.Chan() {
		switch ev := obj.Data.(type) {
		case node.AccountEvent:
			if ev.IsUpdate {
				w.ReScanAllAccount()
			}
		}
	}
}

func (w *Wallet) ReScanAllAccount() {
	newWas := make([]meta.Account, 0)
	for key, _ := range w.accounts {
		wa := w.accounts[key]
		newWa, err := w.nodeAPI.GetAccount(wa.GetAccountID())
		if err != nil {
			continue
		}

		newWas = append(newWas, newWa)
	}
	for _, wa := range newWas {
		w.UpdateWalletAccount(wa)
	}
}

func (w *Wallet) UpdateWalletAccount(account meta.Account) error {
	a, ok := w.accounts[account.GetAccountID().String()]
	if !ok {
		return errors.New("GetAccountID can not find account")
	}
	err := a.UpdateWAccount(account)
	if err != nil {
		log.Error("UpdateWalletAccount", "error", err)
		return err
	}
	w.AddWAccount(a)
	return nil
}

func (w *Wallet) AddWAccount(wa WAccount) {
	key := hex.EncodeToString(wa.privKey.PubKey().SerializeCompressed())
	w.accounts[key] = wa
}

func (w *Wallet) ChooseWAccount(amount *meta.Amount) (WAccount, error) {
	if len(w.accounts) > 0 {
		for key := range w.accounts {
			wa := w.accounts[key]
			if wa.GetAmount() >= amount.GetInt64() {
				return w.accounts[key], nil
			}
		}
	}
	return NewWSAccount(), errors.New("Wallet can not find legal account")
}

func (w *Wallet) GetAllWAccount() []WAccount {
	var WAs []WAccount
	for a := range w.accounts {
		WAs = append(WAs, w.accounts[a])
	}
	return WAs
}

func (w *Wallet) GetWAccount(key string) (WAccount, error) {
	wa, ok := w.accounts[key]
	if ok {
		return wa, nil
	}
	return WAccount{}, errors.New("can not find waccount")
}

func (w *Wallet) SignTransaction(tx meta.Transaction) (*meta.Transaction, error) {
	for _, fc := range tx.GetFromCoins() {
		sign, err := w.signByFromCoin(fc, tx.GetTxID())
		if err != nil {
			return nil, err
		}
		tx.AddSignature(sign)
	}
	return &tx, nil
}

func (w *Wallet) signByFromCoin(fromCoin meta.FromCoin, data *meta.TxID) (math.ISignature, error) {
	a, ok := w.accounts[fromCoin.GetId().String()]
	if !ok {
		return nil, errors.New("signByFromCoin can not find tx from account")
	}
	sign := a.Sign(data.CloneBytes())
	if sign == nil {
		return nil, errors.New("signByFromCoin failed")
	}

	return sign, nil
}
