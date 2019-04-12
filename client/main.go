package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mihongtech/linkchain/client/cmd"
	"github.com/mihongtech/linkchain/common/util/log"
)

func main() {
	logLevel := flag.Int("loglevel", 3, "log level")

	//init log
	log.Root().SetHandler(
		log.LvlFilterHandler(log.Lvl(*logLevel),
			log.StreamHandler(os.Stdout, log.TerminalFormat(true))))

	log.Info("rpcserver client is running")

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
