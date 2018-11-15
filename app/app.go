package app

import (
	"github.com/linkchain/common"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/p2p"
	"github.com/linkchain/node"
)

var (
	//service collection
	svcList = []common.Service{
		&p2p.Service{},
		&node.Node{},
	}
)

func Init(globalConfig *config.LinkChainConfig) bool {
	log.Info("Node init...")

	//p2p init
	if !svcList[0].Init(globalConfig) {
		return false
	}

	//node init
	if !svcList[1].Init(globalConfig) {
		return false
	}

	return true
}

func Run() {
	log.Info("Node is running...")

	//start all service
	for _, v := range svcList {
		if !v.Start() {
			return
		}
	}
}

func Stop() {
	// TODO implement me
}

func GetP2pService() *p2p.Service {
	return svcList[0].(*p2p.Service)
}
