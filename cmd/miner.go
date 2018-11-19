package cmd

import (
	"github.com/linkchain/app"
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
		go app.GetMinerAPI().GetInfo()
	},
}

var startMineCmd = &cobra.Command{
	Use:   "start",
	Short: "get Miner info",
	Run: func(cmd *cobra.Command, args []string) {
		go app.GetMinerAPI().StartMine()
	},
}

var stopMineCmd = &cobra.Command{
	Use:   "stop",
	Short: "get Miner info",
	Run: func(cmd *cobra.Command, args []string) {
		go app.GetMinerAPI().StopMine()
	},
}
