package app

import (
	"github.com/linkchain/app/context"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/node"
	"github.com/linkchain/p2p"
)

var appContext context.Context

func Setup(globalConfig *config.LinkChainConfig) bool {
	log.Info("Node init...")

	appContext.Node = &node.Node{}
	appContext.P2P = &p2p.Service{}
	appContext.Config = globalConfig

	//node init
	if !appContext.Node.Setup(&appContext) {
		return false
	}

	//p2p init
	if !appContext.P2P.Setup(&appContext) {
		return false
	}

	return true
}

func Run() {
	log.Info("Node is running...")

	//start all service
	appContext.Node.Start()
	appContext.P2P.Start()
}

func Stop() {
	// TODO implement me
}

func GetP2pService() *p2p.Service {
	return appContext.P2P.(*p2p.Service)
}
