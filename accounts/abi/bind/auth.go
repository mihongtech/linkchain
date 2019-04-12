package bind

import (
	"errors"
	"io"
	"io/ioutil"

	"github.com/mihongtech/linkchain/accounts/keystore"
	"github.com/mihongtech/linkchain/common/btcec"
	"github.com/mihongtech/linkchain/core/meta"
)

// NewTransactor is a utility method to easily create a transaction signer from
// an encrypted json key stream and the associated passphrase.
func NewTransactor(keyin io.Reader, passphrase string) (*TransactOpts, error) {
	json, err := ioutil.ReadAll(keyin)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(json, passphrase)
	if err != nil {
		return nil, err
	}
	return NewKeyedTransactor(key.PrivateKey), nil
}

// NewKeyedTransactor is a utility method to easily create a transaction signer
// from a single private key.
func NewKeyedTransactor(key *btcec.PrivateKey) *TransactOpts {
	keyAddr := meta.NewAccountId((*btcec.PublicKey)(&key.PublicKey))
	return &TransactOpts{
		From: *keyAddr,
		Signer: func(address meta.AccountID, tx *meta.Transaction) (*meta.Transaction, error) {
			if !keyAddr.IsEqual(address) {
				return nil, errors.New("not authorized to sign this account")
			}
			hash := tx.GetTxID().CloneBytes()

			sign, err := btcec.SignCompact(btcec.S256(), key, hash, true)
			if err != nil {
				return nil, err
			}
			signature := meta.NewSignature(sign)
			tx.AddSignature(signature)

			return tx, nil
		},
	}
}
