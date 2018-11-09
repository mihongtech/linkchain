package util

import (
	"encoding/hex"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/poa/config"
	"github.com/linkchain/poa/meta"
)

func CreateAccountIdByPubKey(pubKey string) (*poameta.AccountID, error) {
	pkBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	pk, err := btcec.ParsePubKey(pkBytes, btcec.S256())
	if err != nil {
		return nil, err
	}
	return poameta.NewAccountId(pk), nil
}

func CreateAccountIdByPrivKey(privKey string) (*poameta.AccountID, error) {
	priv, err := hex.DecodeString(privKey)
	if err != nil {
		return nil, err
	}
	_, pk := btcec.PrivKeyFromBytes(btcec.S256(), priv)
	if err != nil {
		return nil, err
	}
	return poameta.NewAccountId(pk), nil
}

func CreateTempleteAccount(id poameta.AccountID) *poameta.Account {
	utxo := make([]poameta.UTXO, 0)
	a := poameta.NewAccount(id, config.NormalAccount, utxo, config.DafaultClearTime, poameta.AccountID{})
	return a
}

func CreateNormalAccount(key *btcec.PrivateKey) (*poameta.Account, error) {
	privStr := hex.EncodeToString(key.Serialize())
	id, err := CreateAccountIdByPrivKey(privStr)
	if err != nil {
		return nil, err
	}

	a := CreateTempleteAccount(*id)
	return a, nil
}
