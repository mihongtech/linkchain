package main

import (
	"flag"
	"os"

	"github.com/mihongtech/linkchain/client/cmd"
	"github.com/mihongtech/linkchain/client/explorer"
	"github.com/mihongtech/linkchain/common/util/log"
)

func main() {
	logLevel := flag.Int("loglevel", 3, "log level")
	isExplorer := flag.Bool("explorer", false, "is explorer mode")
	flag.Parse()

	//init log
	log.Root().SetHandler(
		log.LvlFilterHandler(log.Lvl(*logLevel),
			log.StreamHandler(os.Stdout, log.TerminalFormat(true))))

	log.Info("rpcserver client is running")

	if *isExplorer {
		explorer.StartExplore()
	} else {
		cmd.StartCmd()
	}
}
