package rpcobject

import (
	"github.com/mihongtech/linkchain/core/meta"
)

//block
type BlockRSP struct {
	Height uint32            `json:"height"`
	Hash   string            `json:"hash"`
	Header *meta.BlockHeader `json:"header"`
	TXIDs  []string          `json:"txids"`
	Hex    string            `json:"hex"`
}

//chain
type ChainRSP struct {
	Chains *meta.ChainInfo `json:"chains"`
}

//wallet
type WalletAccountRSP struct {
	ID     string `json:"id"`
	Type   uint32 `json:"type"`
	Amount int64  `json:"amount"`
}

type WalletInfoRSP struct {
	Accounts []*WalletAccountRSP `json:"accounts"`
}

//account
type TxRSP struct {
	TxID          string `json:"txid"`
	Index         uint32 `json:"index"`
	Value         int64  `json:"value"`
	LocatedHeight uint32 `json:"locatedHeight"`
	EffectHeight  uint32 `json:"effectHeight"`
}

type AccountRSP struct {
	ID          string         `json:"id"`
	Type        uint32         `json:"type"`
	Amount      int64          `json:"amount"`
	SecurityID  string         `json:"securityID"`
	ClearTime   int64          `json:"clearTime"`
	ClearDetail meta.ClearTime `json:"clearDetail"`
	UTXO        []*TxRSP       `json:"utxo"`
	StorageRoot string         `json:"storageRoot"`
	CodeHash    string         `json:"codeHash"`
	Code        string         `json:"code"`
}

//sendmoney
type TransactionWithIDRSP struct {
	ID string            `json:"id"`
	Tx *meta.Transaction `json:"tx"`
}

type PublishContractRSP struct {
	TxID         string `json:"txid"`
	ContractAddr string `json:"contractAddr"`
	PlayLoad     string `json:"playLoad"`
	GasPrice     int64  `json:"gasPrice"`
	GasLimit     int64  `json:"gasLimit"`
}

type CallContractRSP struct {
	TxID         string `json:"txid"`
	ContractAddr string `json:"contractAddr"`
	GasPrice     int    `json:"gasPrice"`
	GasLimit     int    `json:"gasLimit"`
}
