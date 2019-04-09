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
	Short: "miner command",
	Long:  "This is all miner command for handling miner",
}

var minerInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "miner info",
	Long:    "This is get miner info command",
	Example: "miner info",
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
	Use:     "start",
	Short:   "miner start",
	Long:    "This is start auto-miner command",
	Example: "miner start",
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
	Use:     "stop",
	Short:   "miner stop",
	Long:    "This is stop auto-miner command",
	Example: "miner stop",
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
	Use:     "single",
	Short:   "miner single",
	Long:    "This is mine single block command",
	Example: "miner single",
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
