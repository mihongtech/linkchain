package wallet

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/consensus/manager"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/tx"
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

func CreateWAccountFromBytes(pb []byte) WAccount {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), pb)
	return WAccount{privKey: *priv, amount: 0}
}

func (wa *WAccount) UpdateWAccount(iAccount account.IAccount) error {
	if strings.Compare(iAccount.GetAccountID().GetString(), string(wa.privKey.PubKey().SerializeCompressed())) != 0 {
		return errors.New("Account is error")
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
	log.Info("GetAccountInfo", "privkey", hex.EncodeToString(wa.privKey.Serialize()))
	log.Info("GetAccountInfo", "pubkey", hex.EncodeToString(wa.privKey.PubKey().SerializeCompressed()))
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
	accounts map[string]WAccount
	am       manager.AccountManager
}

func (w *Wallet) Init(i interface{}) bool {
	log.Info("Wallet init...")
	w.accounts = make(map[string]WAccount)
	w.am = i.(manager.AccountManager)
	return true
}

func (w *Wallet) Start() bool {
	log.Info("Wallet start...")
	gensisWA := CreateWAccountFromBytes(minePriv)
	gensisKey := hex.EncodeToString(gensisWA.privKey.PubKey().SerializeCompressed())
	w.accounts[gensisKey] = gensisWA
	ga, err := w.am.GetAccount(gensisWA.ConvertAccount().GetAccountID())
	if err == nil {
		gensisWA.UpdateWAccount(ga)
	}
	return true
}

func (w *Wallet) Stop() {
	log.Info("Wallet stop...")
}

func (w *Wallet) UpdateWalletAccount(account account.IAccount) error {
	a, ok := w.accounts[account.GetAccountID().GetString()]
	if !ok {
		return errors.New("ConvertAccount can not find account")
	}
	a.amount = account.GetAmount().GetInt()
	return nil
}

func (w *Wallet) AddWAccount(wa WAccount) {
	key := hex.EncodeToString(wa.privKey.PubKey().SerializeCompressed())
	w.accounts[key] = wa
}

func (w *Wallet) GetWAccount() account.IAccount {
	var randWA WAccount
	if len(w.accounts) > 0 {
		for a := range w.accounts {
			randWA = w.accounts[a]
			break
		}
		return randWA.ConvertAccount()
	}
	return nil
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
