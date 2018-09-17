package node

import (
	"github.com/linkchain/common"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/consensus"
	"github.com/linkchain/p2p"
)

var (
	//service collection
	svcList = []common.IService{
		&p2p.Service{},
		&consensus.Service{},
	}
)

func Init() {
	log.Info("Node init...")

	//init all service
	for _, v := range svcList {
		v.Init(nil)
	}
}

func Run() {
	log.Info("Node is running...")

	//start all service
	for _, v := range svcList {
		v.Start()
	}

	/*block :=svcList[1].(*consensus.Service).GetBlockManager().NewBlock()
	svcList[1].(*consensus.Service).GetBlockManager().ProcessBlock(block)*/
}

//get service
func GetConsensusService() *consensus.Service {
	return svcList[1].(*consensus.Service)
}

func GetP2pService() *p2p.Service {
	return svcList[0].(*p2p.Service)
}
