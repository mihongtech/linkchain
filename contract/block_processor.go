package contract

import (
	"errors"
	"github.com/linkchain/interpreter"

	"github.com/linkchain/common/math"
	"github.com/linkchain/config"
	"github.com/linkchain/contract/vm"
	"github.com/linkchain/core"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/normal"
	"github.com/linkchain/storage/state"

	"github.com/golang/protobuf/proto"
)

//Process Block State tire.
//Return error and block running receipts(result).
func (p *Interpreter) ProcessBlockState(block *meta.Block, stateDb *state.StateDB, chain core.Chain, validate interpreter.Validator) (error, []interpreter.Result) {
	//update mine account status
	actualReward, fee, results, root, err := p.processBlockState(block, stateDb, chain, validate)
	if err != nil {
		return err, nil
	}

	outputs := make([]Output, 0)
	for i := range results {
		outputs = append(outputs, *results[i].(*Output))
	}

	receipts := make([]*core.Receipt, 0)
	useGas := uint64(0)
	for i := range outputs {
		if outputs[i].ResultTx != nil {
			receipts = append(receipts, outputs[i].Receipt)
			useGas += outputs[i].Receipt.GasUsed
		}
	}
	blockHeaderData := NewBlockHeaderData(receipts, useGas)
	headerData, err := proto.Marshal(blockHeaderData.Serialize())
	if err != nil {
		return err, nil
	}
	//check status with header status root.
	if err := validate.VerifyBlockState(block, *root, actualReward, fee, headerData); err != nil {
		return err, nil
	}
	return nil, results
}

func (p *Interpreter) ExecuteBlockState(block *meta.Block, stateDb *state.StateDB, chain core.Chain, validate interpreter.Validator) (error, []interpreter.Result, math.Hash, *meta.Amount) {
	//update mine account status
	_, fee, results, root, err := p.processBlockState(block, stateDb, chain, validate)
	if err != nil {
		return err, nil, math.Hash{}, nil
	}

	return nil, results, *root, fee
}

//Process Block State tire.
//Return error and block running receipts(result).
func (p *Interpreter) processBlockState(block *meta.Block, stateDb *state.StateDB, chain core.Chain, validat interpreter.Validator) (*meta.Amount, *meta.Amount, []interpreter.Result, *math.Hash, error) {
	txs := block.GetTxs()

	coinBase := meta.NewAmount(0)
	txFee := meta.NewAmount(0)
	headerData := GetHeaderData(&block.Header)
	gp := new(core.GasPool).AddGas(headerData.GasLimit)
	inputData := Input{normal.Input{&block.Header, stateDb, chain, block.TXs[0].To.Coins[0].Id},
		chain,
		chain.Config(),
		vm.Config{},
		new(uint64),
		gp,
	}
	outputDatas := make([]interpreter.Result, 0)
	for index := range txs {
		if err := validat.VerifyTx(&txs[index], &inputData); err != nil {
			return nil, nil, nil, nil, errors.New(err.Error() + ",txid=" + txs[index].GetTxID().String())
		}
		err, outputData := p.ProcessTxState(&txs[index], &inputData)
		if err != nil {
			return nil, nil, nil, nil, errors.New(err.Error() + ",txid=" + txs[index].GetTxID().String())
		}
		outputDatas = append(outputDatas, outputData)
		if txs[index].GetType() != config.CoinBaseTx {
			txFee.Addition(*outputData.GetTxFee())
		} else {
			coinBase.Addition(*txs[index].GetToValue())
		}
	}

	root := stateDb.IntermediateRoot()
	return coinBase, txFee, outputDatas, &root, nil
}
