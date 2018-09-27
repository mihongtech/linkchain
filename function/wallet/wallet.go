package wallet

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/events"
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/poa/manage"
	poameta "github.com/linkchain/poa/meta"
)

var minePriv, _ = hex.DecodeString("55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b")

type WAccount struct {
	privKey btcec.PrivateKey
	amount  int
	nounce  uint32
}

func NewWSAccount() WAccount {
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Info("POAAccountManager", "NewAccount - generate private key failed", err)
	}
	return WAccount{privKey: *priv, amount: 0}
}

func CreateWAccountFromBytes(privb []byte, amount int) WAccount {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), privb)
	return WAccount{privKey: *priv, amount: amount}
}

func (wa *WAccount) UpdateWAccount(iAccount account.IAccount) error {
	if strings.Compare(iAccount.GetAccountID().GetString(), hex.EncodeToString(wa.privKey.PubKey().SerializeCompressed())) != 0 {
		return errors.New("Account is error")
	}
	if wa.amount != iAccount.GetAmount().GetInt() || wa.nounce != iAccount.GetNounce() {
		log.Info("updateWallet", "account", iAccount.GetAccountID().GetString(), "amount", iAccount.GetAmount().GetInt(), "nounce", iAccount.GetNounce())
	}
	wa.amount = iAccount.GetAmount().GetInt()
	wa.nounce = iAccount.GetNounce()
	return nil
}

func (wa *WAccount) ConvertAccount() account.IAccount {
	id := *poameta.NewAccountId(wa.privKey.PubKey())
	amount := *poameta.NewAmout(int32(wa.amount))
	return poameta.NewAccount(id, amount, wa.nounce)
}

func (wa *WAccount) GetAccountInfo() {
	log.Info("Wallet Info", "account", hex.EncodeToString(wa.privKey.PubKey().SerializeCompressed()), "amount", wa.GetAmount(), "nounce", wa.GetNounce())
}

func (wa *WAccount) GetAccountPubkey() string {
	return hex.EncodeToString(wa.privKey.PubKey().SerializeCompressed())
}

func (wa *WAccount) GetAccountPrivkey() string {
	return hex.EncodeToString(wa.privKey.Serialize())
}

func (wa *WAccount) GetAmount() int {
	return wa.amount
}

func (wa *WAccount) GetNounce() uint32 {
	return wa.nounce
}

func (wa *WAccount) Sign(messageHash []byte) []byte {
	signature, err := wa.privKey.Sign(messageHash)
	if err != nil {
		log.Error("WAccount", "Sign", err)
		return nil
	}
	return signature.Serialize()
}

type Wallet struct {
	accounts         map[string]WAccount
	am               manage.AccountManage
	updateAccountSub *event.TypeMuxSubscription
}

func (w *Wallet) Init(i interface{}) bool {
	log.Info("Wallet init...")
	w.accounts = make(map[string]WAccount)
	w.am = *i.(*manage.AccountManage)
	return true
}

func (w *Wallet) Start() bool {
	log.Info("Wallet start...")
	gensisWA := CreateWAccountFromBytes(minePriv, 50)
	gensisKey := hex.EncodeToString(gensisWA.privKey.PubKey().SerializeCompressed())
	w.accounts[gensisKey] = gensisWA
	w.updateAccountSub = w.am.NewWalletEvent.Subscribe(events.WAccountEvent{})
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
		case events.WAccountEvent:
			log.Info("POST EVENT")
			if ev.IsUpdate {
				newWas := make([]account.IAccount, 0)
				for key := range w.accounts {
					wa := w.accounts[key]
					newWa, err := w.am.GetAccount(wa.ConvertAccount().GetAccountID())
					if err != nil {
						continue
					}
					newWas = append(newWas, newWa)
				}
				for _, wa := range newWas {
					w.UpdateWalletAccount(wa)
				}
			}
		}
	}
}

func (w *Wallet) UpdateWalletAccount(account account.IAccount) error {
	a, ok := w.accounts[account.GetAccountID().GetString()]
	if !ok {
		return errors.New("ConvertAccount can not find account")
	}
	err := a.UpdateWAccount(account)
	if err != nil {
		return err
	}
	w.AddWAccount(a)
	return nil
}

func (w *Wallet) AddWAccount(wa WAccount) {
	key := hex.EncodeToString(wa.privKey.PubKey().SerializeCompressed())
	w.accounts[key] = wa
}

func (w *Wallet) ChooseWAccount(amount meta.IAmount) (WAccount, error) {
	if len(w.accounts) > 0 {
		for key := range w.accounts {
			wa := w.accounts[key]
			if wa.GetAmount() >= amount.GetInt() {
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
func (w *Wallet) SignTransaction(tx tx.ITx) (tx.ITx, error) {
	a, ok := w.accounts[tx.GetFrom().GetID().GetString()]
	if !ok {
		return nil, errors.New("SignTransaction can not find tx from account")
	}
	sign := a.Sign(tx.GetTxID().CloneBytes())
	if sign == nil {
		return nil, errors.New("SignTransaction failed")
	}
	tx.SetSignature(sign)
	return tx, nil
}
