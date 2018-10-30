package cmd

import (
	"github.com/linkchain/node"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(minerCmd)
	minerCmd.AddCommand(minerInfoCmd,
		startMineCmd,
		stopMineCmd)
}

var minerCmd = &cobra.Command{
	Use:   "miner",
	Short: "handle miner cmd",
}

var minerInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "get Miner info",
	Run: func(cmd *cobra.Command, args []string) {
		node.GetMiner().GetInfo()
	},
}

var startMineCmd = &cobra.Command{
	Use:   "start",
	Short: "get Miner info",
	Run: func(cmd *cobra.Command, args []string) {
		go node.GetMiner().StartMine()
	},
}

var stopMineCmd = &cobra.Command{
	Use:   "stop",
	Short: "get Miner info",
	Run: func(cmd *cobra.Command, args []string) {
		go node.GetMiner().StopMine()
	},
}
