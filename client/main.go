package main

import (
	"flag"
	"os"

	"github.com/linkchain/client/cmd"
	"github.com/linkchain/client/explorer"
	"github.com/linkchain/common/util/log"
)

func main() {
	logLevel := flag.Int("loglevel", 3, "log level")
	clientType := flag.String("clientType", "cmd", "client type: cmd、explorer")

	//init log
	log.Root().SetHandler(
		log.LvlFilterHandler(log.Lvl(*logLevel),
			log.StreamHandler(os.Stdout, log.TerminalFormat(true))))

	log.Info("rpcserver client is running, running at: " + *clientType)

	switch {
	case *clientType == "cmd":
		cmd.StartCmd()
	case *clientType == "explorer":
		explore.StartExplore()
	}
}
