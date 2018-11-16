package app

import (
	"github.com/linkchain/app/context"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/node"
	"github.com/linkchain/p2p"
)

var (
	appContext context.Context
	nodeSvc    *node.Node
	p2pSvc     *p2p.Service
)

func Setup(globalConfig *config.LinkChainConfig) bool {
	log.Info("Node init...")

	//prepare config
	appContext.Config = globalConfig

	//create service
	nodeSvc = node.NewNode()
	p2pSvc = p2p.NewP2P()

	//node init
	if !nodeSvc.Setup(&appContext) {
		return false
	}

	//node api init
	appContext.NodeAPI = node.NewPublicNodeAPI(nodeSvc)

	//p2p init
	if !p2pSvc.Setup(&appContext) {
		return false
	}

	//p2p api init
	appContext.P2PAPI = p2pSvc

	return true
}

func Run() {
	log.Info("Node is running...")

	//start all service
	nodeSvc.Start()
	p2pSvc.Start()
}

func Stop() {
	// TODO implement me
}

func GetAppContext() *context.Context {
	return &appContext
}

func GetNodeAPI() *node.PublicNodeAPI {
	return appContext.NodeAPI.(*node.PublicNodeAPI)
}

func GetP2PAPI() *p2p.Service {
	return p2pSvc
}
