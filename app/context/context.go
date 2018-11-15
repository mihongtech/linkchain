package context

import (
	"github.com/linkchain/core"
	"github.com/linkchain/config"
)

type Context struct{
	Node core.Service
	P2P  core.Service
	Config  *config.LinkChainConfig
}