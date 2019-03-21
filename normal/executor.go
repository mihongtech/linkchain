package normal

import (
	"errors"

	"github.com/linkchain/config"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/interpreter"
)

func (e *Interpreter) ExecuteResult(results []interpreter.Result, txFee *meta.Amount, block *meta.Block) error {
	//push txfee into coinbase
	if block.TXs[0].Type != config.CoinBaseTx {
		return errors.New("the frist tx of block must be coinbase")
	}

	block.TXs[0].To.Coins[0].Value.Addition(*txFee)
	block.TXs[0].RebuildTxID()
	//The Interpreter result is useless
	return nil
}

func (e *Interpreter) ChooseTransaction(txs []meta.Transaction, best *meta.Block, offChain interpreter.OffChain, wallet interpreter.Wallet, signer *meta.AccountID) []meta.Transaction {
	//normal not remove transaction
	return txs
}
