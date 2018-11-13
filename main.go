package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/linkchain/cmd"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/node"
)

func main() {
	var (
		logLevel   = flag.Int("loglevel", 3, "log level")
		listenPort = flag.Int("listenport", 40000, "linkchain listen port")
		dataDir    = flag.String("datadir", config.DefaultDataDir(), "linkchain data dir")
	)
	flag.Parse()

	//init log
	log.Root().SetHandler(
		log.LvlFilterHandler(log.Lvl(*logLevel),
			log.StreamHandler(os.Stdout, log.TerminalFormat(true))))

	// init config
	globalConfig := &config.LinkChainConfig{}
	globalConfig.ListenAddress = fmt.Sprintf(":%d", *listenPort)
	globalConfig.DataDir = *dataDir

	// start node
	if !node.Init(globalConfig) {
		return
	}
	node.Run()
	defer node.Stop()

	// start console cmd
	startCmd()
}

func startCmd() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">")
		// Scans a line from Stdin(Console)
		scanner.Scan()
		// Holds the string that scanned
		text := scanner.Text()
		if len(text) != 0 {
			words := strings.Fields(text)

			cmd.RootCmd.SetArgs(words)
			cmd.RootCmd.Execute()
		}

	}
}
