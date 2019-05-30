package pool

import "github.com/mihongtech/linkchain/core/meta"

type TxPool interface {
	CheckTx(tx *meta.Transaction) error
	GetAllTransaction() []meta.Transaction
	AddTransaction(tx *meta.Transaction) error
	RemoveTransaction(txID meta.TxID) error
	ProcessTx(tx *meta.Transaction) error
}
