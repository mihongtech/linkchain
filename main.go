package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mihongtech/linkchain/app"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/config"
)

func main() {
	var (
		logLevel    = flag.Int("loglevel", 3, "log level")
		listenPort  = flag.Int("port", 40000, "linkchain listen port")
		rpcIp       = flag.String("rpcip", "127.0.0.1", "linkchain rpc ip")
		rpcPort     = flag.Int("rpcport", 8082, "linkchain rpc port")
		dataDir     = flag.String("datadir", config.DefaultDataDir(), "linkchain data dir")
		console     = flag.Bool("console", false, "log out put console(default=false).")
		nodiscovery = flag.Bool("nodiscovery", false, "default = false means use discovery protocol")
		genesispath = flag.String("genesis", "genesis.json", "linkchain genesis config file path")
		bootnodes   = flag.String("bootnodes", "", "Comma separated enode URLs for P2P discovery bootstrap")
		interpreter = flag.String("interpreter", "contract", "choose interprete api")
	)
	flag.Parse()

	if err := initLog(logLevel, *console, *dataDir); err != nil {
		log.Error("initLog failed, exit", "err", err)
		return
	}

	// init config
	globalConfig := &config.LinkChainConfig{}
	globalConfig.ListenAddress = fmt.Sprintf(":%d", *listenPort)
	globalConfig.DataDir = *dataDir
	globalConfig.GenesisPath = *genesispath
	globalConfig.NoDiscovery = *nodiscovery
	globalConfig.BootstrapNodes = *bootnodes
	globalConfig.InterpreterAPI = *interpreter
	globalConfig.RpcAddr = *rpcIp + ":" + strconv.Itoa(*rpcPort)
	// start node
	if !app.Setup(globalConfig) {
		log.Error("app setup failed, exit")
		return
	}

	app.Run()
	defer app.Stop()
}

func initLog(logLevel *int, console bool, dataDir string) error {
	//init log
	ostream := log.StreamHandler(os.Stdout, log.TerminalFormat(true))
	glogger := log.NewGlogHandler(ostream)

	if !console {
		rfh, err := log.RotatingFileHandler(
			filepath.Join(dataDir, "log"),
			512*1024*1024,
			log.TerminalFormat(true),
		)

		if err != nil {
			fmt.Println("Init loghandler failed, err is ", err)
			return err
		}
		glogger.SetHandler(rfh)
		log.PrintOrigins(true)
	}

	glogger.Verbosity(log.Lvl(*logLevel))
	log.Root().SetHandler(glogger)
	return nil
}
