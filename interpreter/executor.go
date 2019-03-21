package interpreter

import (
	"github.com/linkchain/common/math"
	"github.com/linkchain/core"
	"github.com/linkchain/core/meta"
)

type Executor interface {
	ExecuteResult(results []Result, txFee *meta.Amount, block *meta.Block) error                                                             //After executing block state,execute the result
	ChooseTransaction(txs []meta.Transaction, best *meta.Block, offChain OffChain, wallet Wallet, signer *meta.AccountID) []meta.Transaction //choose some of tx into block
}

type OffChain interface {
	core.Service
	UpdateMainChain(ev meta.ChainEvent)
	UpdateSideChain(ev meta.ChainSideEvent)
}

type Wallet interface {
	core.Service
	SignMessage(accountId meta.AccountID, hash []byte) (math.ISignature, error)
	SignTransaction(tx meta.Transaction) (*meta.Transaction, error)
	ImportAccount(privateKeyStr string) (*meta.AccountID, error)
	ExportAccount(id meta.AccountID) (string, error)
	GetAccount(key string) (*meta.Account, error)
	GetAllWAccount() []meta.Account
	AddAccount(account meta.Account)
	NewAccount() (*meta.AccountID, error)
}
