package node

import (
	"github.com/linkchain/common"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/consensus"
	"github.com/linkchain/function/wallet"
	"github.com/linkchain/p2p"
	"github.com/linkchain/sync"
)

var (
	//service collection
	svcList = []common.IService{
		&consensus.Service{},
		&wallet.Wallet{},
		&p2p.Service{},
		&sync.Service{},
	}
)

func Init() {
	log.Info("Node init...")

	//init all service
	for index, v := range svcList {
		if index == 3 {
			v.Init(svcList[0].(*consensus.Service))
			continue
		}
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

func GetWallet() *wallet.Wallet {
	return svcList[2].(*wallet.Wallet)
}
