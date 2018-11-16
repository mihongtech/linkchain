package app

import (
	"github.com/linkchain/app/context"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/miner"
	"github.com/linkchain/node"
	"github.com/linkchain/p2p"
	"github.com/linkchain/wallet"
)

var (
	appContext context.Context
	nodeSvc    *node.Node
	p2pSvc     *p2p.Service
	minerSvc   *miner.Miner
	walletSvc  *wallet.Wallet
)

func Setup(globalConfig *config.LinkChainConfig) bool {
	log.Info("Node init...")

	//prepare config
	appContext.Config = globalConfig

	//create service
	nodeSvc = node.NewNode()
	p2pSvc = p2p.NewP2P()
	minerSvc = miner.NewMiner()
	walletSvc = wallet.NewWallet()

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

	//miner init
	if !minerSvc.Setup(&appContext) {
		return false
	}
	//miner api init
	appContext.MinerAPI = minerSvc

	//wallet init
	if !walletSvc.Setup(&appContext) {
		return false
	}
	//wallet api init
	appContext.WalletAPI = walletSvc
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

func GetMinerAPI() *miner.Miner {
	return minerSvc
}

func GetWalletAPI() *wallet.Wallet {
	return walletSvc
}
