package contract

import (
	"github.com/linkchain/config"
	"github.com/linkchain/contract/vm"
	"github.com/linkchain/core"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/normal"
)

type Input struct {
	normal.Input
	Chain   ChainContext
	Config  *config.ChainConfig
	VmCfg   vm.Config
	UsedGas *uint64
	Gp      *core.GasPool
}

type Interpreter struct {
	normal.Interpreter
}

type Output struct {
	normal.Output
	ResultTx *meta.Transaction
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
