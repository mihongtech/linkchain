package app

import (
	"bitbucket.org/rollchain/miner"
	"time"

	"github.com/mihongtech/linkchain/app/context"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/contract"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/node"
	"github.com/mihongtech/linkchain/node/net/p2p"
	"github.com/mihongtech/linkchain/normal"
	"github.com/mihongtech/linkchain/rpc/rpcserver"
	"github.com/mihongtech/linkchain/wallet"
)

var (
	appContext context.Context
	nodeSvc    *node.Node
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

	minerSvc = miner.NewMiner()
	walletSvc = wallet.NewWallet()

	//node init
	if !nodeSvc.Setup(&appContext) {
		return false
	}

	//consensus api init
	appContext.NodeAPI = node.NewPublicNodeAPI(nodeSvc)

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
		Addr:        appContext.Config.RpcAddr,
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
