package node

import (
	"github.com/linkchain/common"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/consensus"
	"github.com/linkchain/function/miner"
	"github.com/linkchain/function/wallet"
	"github.com/linkchain/p2p"
	"github.com/linkchain/storage"
)

var (
	//service collection
	svcList = []common.IService{
		&storage.Storage{},
		&consensus.Service{},
		&wallet.Wallet{},
		&p2p.Service{},
		&miner.Miner{},
	}
)

func Init(globalConfig *config.LinkChainConfig) bool {
	log.Info("Node init...")

	//storage init
	if !svcList[0].Init(globalConfig) {
		return false
	} else {
		globalConfig.StorageService = svcList[0]
	}

	//consensus init
	if !svcList[1].Init(globalConfig) {
		return false
	}

	globalConfig.ConsensusService = svcList[1]

	//wallet init
	if !svcList[2].Init(globalConfig) {
		return false
	}

	//p2p init
	if !svcList[3].Init(globalConfig) {
		return false
	}

	//storage init
	if !svcList[4].Init(nil) {
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

	/*block :=svcList[1].(*consensus.Service).GetBlockManager().CreateBlock()
	svcList[1].(*consensus.Service).GetBlockManager().ProcessBlock(block)*/
}

func Stop() {
	// TODO implement me
}

//get service

func GetStorage() *storage.Storage {
	return svcList[0].(*storage.Storage)
}

func GetConsensusService() *consensus.Service {
	return svcList[1].(*consensus.Service)
}

func GetWallet() *wallet.Wallet {
	return svcList[2].(*wallet.Wallet)
}

func GetP2pService() *p2p.Service {
	return svcList[3].(*p2p.Service)
}

func GetMiner() *miner.Miner {
	return svcList[4].(*miner.Miner)
}
