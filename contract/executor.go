package contract

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/normal"
)

func (e *Interpreter) ExecuteResult(results []interpreter.Result, txFee *meta.Amount, block *meta.Block) error {

	//push txfee into coinbase
	if block.TXs[0].Type != config.CoinBaseTx {
		return errors.New("the frist tx of block must be coinbase")
	}

	block.TXs[0].To.Coins[0].Value.Addition(*txFee)
	block.TXs[0].RebuildTxID()

	//push block header data
	useGas := uint64(0)
	for i := range results {
		if results[i].(*Output).ResultTx != nil {
			useGas += results[i].GetReceipt().GasUsed
		}
	}
	blockHeaderData := NewBlockHeaderData(normal.GetReceiptsByResult(results), useGas)
	headerData, err := proto.Marshal(blockHeaderData.Serialize())
	if err != nil {
		return err
	}

	//push contract result tx
	for i := range results {
		result := results[i].(*Output)
		if result.ResultTx != nil {
			block.SetTx(*result.ResultTx)
		}
	}
	block.Header.Data = headerData

	return err
}
