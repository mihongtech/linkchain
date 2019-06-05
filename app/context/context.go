package context

import (
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/node"
)

type Context struct {
	NodeAPI        *node.CoreAPI
	WalletAPI      interpreter.Wallet
	InterpreterAPI interpreter.Interpreter
	Config         *config.LinkChainConfig
}
