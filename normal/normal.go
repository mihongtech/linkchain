package normal

import (
	"github.com/mihongtech/linkchain/common/lcdb"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/storage/state"
)

type Interpreter struct {
}

func (i *Interpreter) CreateOffChain(db lcdb.Database) interpreter.OffChain {
	return &OffChainState{}
}

type Input struct {
	Header      *meta.BlockHeader
	StateDB     *state.StateDB
	ChainReader meta.ChainReader
	BlockSigner meta.AccountID
}

func (i *Input) GetBlockSigner() meta.AccountID {
	return i.BlockSigner
}

func (i *Input) GetBlockHeader() *meta.BlockHeader {
	return i.Header
}

func (i *Input) GetStateDB() *state.StateDB {
	return i.StateDB
}

func (i *Input) GetChainReader() meta.ChainReader {
	return i.ChainReader
}

type Output struct {
	TxFee   *meta.Amount
	Receipt *core.Receipt
}

func (o *Output) GetTxFee() *meta.Amount {
	return o.TxFee
}

func (o *Output) GetReceipt() *core.Receipt {
	return o.Receipt
}

func (o *Output) WriteResult() error {
	return nil
}

func IsNormal(txType uint32) bool {
	return txType == config.CoinBaseTx || txType == config.NormalTx
}

func GetReceiptsByResult(results []interpreter.Result) []*core.Receipt {
	receipts := make([]*core.Receipt, 0)
	for i := range results {
		if results[i].GetReceipt() != nil {
			receipts = append(receipts, results[i].GetReceipt())
		}
	}
	return receipts
}
