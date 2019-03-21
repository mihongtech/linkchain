package node

import (
	"testing"
)

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
	m.addAccount(fromA)

	_, err = m.GetAccountRelateTXs(tx, false)
	unittest.Equal(t, err, nil)*/
}

func TestAccountManage_GetAccountRelateTXs2(t *testing.T) {

}

func TestConvertAccount(t *testing.T) {

}
