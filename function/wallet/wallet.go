package wallet

import (
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta/account"
	"errors"
	poameta "github.com/linkchain/poa/meta"
	"encoding/hex"
	"github.com/linkchain/meta/tx"
)

var minePriv,_ = hex.DecodeString("fe38240982f313ae5afb3e904fb8215fb11af1200592b" +
	"fca26c96c4738e4bf8f")

type WAccount struct {
	privKey btcec.PrivateKey
	amount int
}

func NewWSAccount() WAccount {
	priv,err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Info("POAAccountManager","NewAccount - generate private key failed",err)
	}
	return WAccount{privKey:*priv,amount:0}
}

func CreateWAccountFromBytes(pb []byte) WAccount {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), pb)
	return WAccount{privKey:*priv,amount:0}
}

func (wa *WAccount) ConvertAccount() account.IAccount  {
	accountID := *poameta.NewAccountId(wa.privKey.PubKey().SerializeCompressed()).(*poameta.POAAccountID)
	a := poameta.POAAccount{AccountID:accountID,Value:poameta.POAAmount{Value:int32(wa.amount)}}
	return &a
}

func (wa *WAccount) Sign(messageHash []byte) []byte {
	signature, err := wa.privKey.Sign(messageHash)
	if err != nil {
		log.Error("WAccount","Sign",err)
		return nil
	}
	return signature.Serialize()
}

type Wallet struct {
	accounts map[string]WAccount
}

func (w *Wallet) Init(i interface{}) bool{
	log.Info("Wallet init...");
	w.accounts = make(map[string]WAccount)

	return true
}

func (w *Wallet) Start() bool{
	log.Info("Wallet start...");
	gensisWA := CreateWAccountFromBytes(minePriv)
	gensisKey := hex.EncodeToString(gensisWA.privKey.PubKey().SerializeCompressed())
	w.accounts[gensisKey] = gensisWA
	return true
}

func (w *Wallet) Stop(){
	log.Info("Wallet stop...");
}

func (w *Wallet) UpdateWalletAccount(account account.IAccount) (error) {
	a,ok := w.accounts[account.GetAccountID().GetString()]
	if !ok {
		return errors.New("ConvertAccount can not find account")
	}
	a.amount = account.GetAmount().GetInt()
	return nil
}

func (w *Wallet) SignTransaction(tx tx.ITx) (tx.ITx,error) {
	a,ok := w.accounts[tx.GetFrom().GetID().GetString()]
	if !ok {
		return nil,errors.New("SignTransaction can not find tx from account")
	}
	sign := a.Sign(tx.GetTxID().CloneBytes())
	if sign == nil {
		return nil,errors.New("SignTransaction failed")
	}
	tx.SetSignature(sign)
	return tx,nil
}

func (w *Wallet) AddWAccount(wa WAccount)  {
	key := hex.EncodeToString(wa.privKey.PubKey().SerializeCompressed())
	w.accounts[key] = wa
}

func (w *Wallet) GetWAccount() account.IAccount  {
	var randWA WAccount
	if len(w.accounts) > 0{
		for a := range w.accounts {
			randWA = w.accounts[a]
			break
		}
		return randWA.ConvertAccount()
	}
	return nil
}