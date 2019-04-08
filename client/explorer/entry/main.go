package main

import (
	"flag"
	"os"

	"github.com/linkchain/client/explorer"
	"github.com/linkchain/common/util/log"
)

func main() {
	logLevel := flag.Int("loglevel", 3, "log level")

	//init log
	log.Root().SetHandler(
		log.LvlFilterHandler(log.Lvl(*logLevel),
			log.StreamHandler(os.Stdout, log.TerminalFormat(true))))

	log.Info("rpcserver explorer is running")

	// start console cmd
	explorer.StartExplore()
}
