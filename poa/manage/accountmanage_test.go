package manage

import (
	"testing"
	"time"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/poa/config"
	"github.com/linkchain/poa/meta"
	"github.com/linkchain/unittest"
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

func TestAccountManage_NewAccount(t *testing.T) {
	m := AccountManage{}
	a := m.NewAccount()
	t.Logf(t.Name(), a)
}

func TestAccountManage_AddAccount(t *testing.T) {
	m := AccountManage{}
	m.Init(nil)
	a := m.NewAccount()
	err := m.AddAccount(a)
	unittest.Equal(t, err, nil)
}

func TestAccountManage_GetAccount(t *testing.T) {
	m := AccountManage{}
	m.Init(nil)
	a := m.NewAccount()
	err := m.AddAccount(a)
	unittest.Equal(t, err, nil)

	na, err := m.GetAccount(a.GetAccountID())
	unittest.Equal(t, err, nil)
	t.Log(na)
}

func TestAccountManage_RemoveAccount(t *testing.T) {
	m := AccountManage{}
	m.Init(nil)
	a := m.NewAccount()
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
	fpriv, err := btcec.NewPrivateKey(btcec.S256())
	unittest.Equal(t, err, nil)
	tpriv, err := btcec.NewPrivateKey(btcec.S256())
	unittest.Equal(t, err, nil)
	fId := meta.NewAccountId(fpriv.PubKey())
	tId := meta.NewAccountId(tpriv.PubKey())

	fp := meta.NewTransactionPeer(*fId, nil)
	tp := meta.NewTransactionPeer(*tId, nil)
	amount := meta.NewAmout(10)
	tx := meta.NewTransaction(config.TransactionVersion, *fp, *tp, *amount, time.Now(), 1, nil, meta.FromSign{})
	sign, err := fpriv.Sign(tx.GetTxID().CloneBytes())
	unittest.Equal(t, err, nil)
	tx.SetSignature(sign.Serialize())
	t.Log("txid", tx.GetTxID())

	m := AccountManage{}
	m.Init(nil)
	fromA := meta.NewAccount(*fId, *meta.NewAmout(20), 0)
	m.AddAccount(fromA)

	_, err = m.GetAccountRelateTXs(tx, false)
	unittest.Equal(t, err, nil)
}

func TestAccountManage_GetAccountRelateTXs2(t *testing.T) {

}

func TestConvertAccount(t *testing.T) {

}
