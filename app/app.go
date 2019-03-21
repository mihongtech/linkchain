package app

import (
	"time"

	"github.com/linkchain/app/context"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/contract"
	"github.com/linkchain/interpreter"
	"github.com/linkchain/miner"
	"github.com/linkchain/node"
	"github.com/linkchain/normal"
	"github.com/linkchain/p2p"
	"github.com/linkchain/rpc/rpcserver"
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
	log.Info("App setup...")

	//prepare config
	appContext.Config = globalConfig

	//create interpreterAPI and Excutor by config choice different function
	appContext.InterpreterAPI = chooseInterpreterAPI(globalConfig.InterpreterAPI)

	//create service
	nodeSvc = node.NewNode()

	p2pSvc = p2p.NewP2P()
	minerSvc = miner.NewMiner()
	walletSvc = wallet.NewWallet()

	//node init
	if !nodeSvc.Setup(&appContext) {
		return false
	}

	//consensus api init
	appContext.NodeAPI = node.NewPublicNodeAPI(nodeSvc)

	//p2p init
	if !p2pSvc.Setup(&appContext) {
		return false
	}

	//p2p api init
	appContext.P2PAPI = p2pSvc

	//wallet init
	if !walletSvc.Setup(&appContext) {
		return false
	}
	//wallet api init
	appContext.WalletAPI = walletSvc

	//miner init
	if !minerSvc.Setup(&appContext) {
		return false
	}
	//miner api init
	appContext.MinerAPI = minerSvc

	return true
}

func Run() {
	//start all service
	nodeSvc.Start()
	p2pSvc.Start()
	walletSvc.Start()

	//start rpc
	startRPC()

	//here waiting for the interruption
	log.Info("App is running...")

	// listen the exit signal
	interrupt := interruptListener()
	<-interrupt
}

func Stop() {
	log.Info("Stopping app...")
	walletSvc.Stop()
	p2pSvc.Stop()
	nodeSvc.Stop()
	log.Info("App exit")
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

func startRPC() {
	//init rpc servce
	s, err := rpcserver.NewRPCServer(&rpcserver.Config{
		StartupTime: time.Now().Unix(),
	}, &appContext)
	if err != nil {
		return
	}

	s.Start()

	go func() {
		<-s.RequestedProcessShutdown()
		shutdownRequestChannel <- struct{}{}
	}()
}

func chooseInterpreterAPI(interpreter string) interpreter.Interpreter {
	log.Info("App", "interpreter", interpreter)
	switch interpreter {
	case "normal":
		return &normal.Interpreter{}
	case "contract":
		return &contract.Interpreter{}
	default:
		return &normal.Interpreter{}
	}
}
