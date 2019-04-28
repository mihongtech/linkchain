package context

import (
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/interpreter"
)

type Context struct {
	NodeAPI        core.Service
	P2PAPI         core.Service
	MinerAPI       core.Service
	TxpoolAPI      core.Service
	WalletAPI      interpreter.Wallet
	InterpreterAPI interpreter.Interpreter
	Config         *config.LinkChainConfig
}
