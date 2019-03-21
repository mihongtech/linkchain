package context

import (
	"github.com/linkchain/config"
	"github.com/linkchain/core"
	"github.com/linkchain/interpreter"
)

type Context struct {
	NodeAPI        core.Service
	P2PAPI         core.Service
	MinerAPI       core.Service
	WalletAPI      interpreter.Wallet
	InterpreterAPI interpreter.Interpreter
	Config         *config.LinkChainConfig
}
