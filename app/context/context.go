package context

import (
	"github.com/linkchain/config"
	"github.com/linkchain/core"
)

type Context struct {
	NodeAPI  core.Service
	P2PAPI   core.Service
	MinerAPI core.Service
	Config   *config.LinkChainConfig
}
