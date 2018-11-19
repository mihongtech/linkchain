package node

import (
	"testing"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/unittest"
	"github.com/linkchain/util"
)

func TestAccountManage_Init(t *testing.T) {
	m := AccountManage{}
	unittest.Assert(t, m.Init(nil), "TestAccountManage_Init")
}

func TestAccountManage_Start(t *testing.T) {
	m := AccountManage{}
	unittest.Assert(t, m.Start(), "TestAccountManage_Start")
}

func TestAccountManage_Stop(t *testing.T) {
	m := AccountManage{}
	m.Stop()
}

func TestAccountManage_AddAccount(t *testing.T) {
	m := AccountManage{}
	m.Init(nil)
	priv, _ := btcec.NewPrivateKey(btcec.S256())
	a, _ := util.CreateNormalAccount(priv)
	err := m.AddAccount(a)
	unittest.Equal(t, err, nil)
}

func TestAccountManage_GetAccount(t *testing.T) {
	m := AccountManage{}
	m.Init(nil)
	priv, _ := btcec.NewPrivateKey(btcec.S256())
	a, _ := util.CreateNormalAccount(priv)
	err := m.AddAccount(a)
	unittest.Equal(t, err, nil)

	na, err := m.GetAccount(a.GetAccountID())
	unittest.Equal(t, err, nil)
	t.Log(na)
}

func TestAccountManage_RemoveAccount(t *testing.T) {
	m := AccountManage{}
	m.Init(nil)
	priv, _ := btcec.NewPrivateKey(btcec.S256())
	a, _ := util.CreateNormalAccount(priv)
	err := m.AddAccount(a)
	unittest.Equal(t, err, nil)

	err = m.RemoveAccount(a.GetAccountID())
	unittest.Equal(t, err, nil)
	_, err = m.GetAccount(a.GetAccountID())
	unittest.NotEqual(t, err, nil)
}

func TestAccountManage_UpdateAccountsByTxs(t *testing.T) {

}

func TestAccountManage_RevertAccountsByTxs(t *testing.T) {

}

func TestAccountManage_CheckTxAccount(t *testing.T) {

}

func TestAccountManage_GetAccountRelateTXs(t *testing.T) {
	/*fpriv, err := btcec.NewPrivateKey(btcec.S256())
	unittest.Equal(t, err, nil)
	tpriv, err := btcec.NewPrivateKey(btcec.S256())
	unittest.Equal(t, err, nil)
	fId := poameta.NewAccountId(fpriv.PubKey())
	tId := poameta.NewAccountId(tpriv.PubKey())

	fp := poameta.NewTransactionPeer(*fId, nil)
	tp := poameta.NewTransactionPeer(*tId, nil)
	amount := poameta.NewAmout(10)
	tx := poameta.NewTransaction(config.TransactionVersion, *fp, *tp, *amount, time.Now(), 1, nil, poameta.FromSign{})
	sign, err := fpriv.Sign(tx.GetTxID().CloneBytes())
	unittest.Equal(t, err, nil)
	tx.SetSignature(sign.Serialize())
	t.Log("txid", tx.GetTxID())

	m := AccountManage{}
	m.Init(nil)
	fromA := poameta.NewAccount(*fId, *poameta.NewAmout(20), 0)
	m.AddAccount(fromA)

	_, err = m.GetAccountRelateTXs(tx, false)
	unittest.Equal(t, err, nil)*/
}

func TestAccountManage_GetAccountRelateTXs2(t *testing.T) {

}

func TestConvertAccount(t *testing.T) {

}
