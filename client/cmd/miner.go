package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(minerCmd)
	minerCmd.AddCommand(minerInfoCmd,
		startMineCmd,
		stopMineCmd,
		mineCmd)
}

var minerCmd = &cobra.Command{
	Use:   "miner",
	Short: "handle miner cmd",
}

var minerInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "get Miner info",
	Run: func(cmd *cobra.Command, args []string) {

		method := "getMineInfo"

		//call
		out, err := rpc(method, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

var startMineCmd = &cobra.Command{
	Use:   "start",
	Short: "start mine loop",
	Run: func(cmd *cobra.Command, args []string) {
		method := "startMine"

		//call
		out, err := rpc(method, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

var stopMineCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop mine loop",
	Run: func(cmd *cobra.Command, args []string) {
		method := "stopMine"

		//call
		out, err := rpc(method, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

var mineCmd = &cobra.Command{
	Use:   "mine",
	Short: "mine a single block",
	Run: func(cmd *cobra.Command, args []string) {
		method := "mine"

		//call
		out, err := rpc(method, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}
