package cmd

import (
	//"github.com/linkchain/app"

	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(chainCmd)
	chainCmd.AddCommand(chainInfoCmd)
}

var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "chain command",
	Long:  "This is all chain command for handling chain",
}

var chainInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "info",
	Long:    "This is get blockChain info command",
	Example: "chain info",
	Run: func(cmd *cobra.Command, args []string) {

		method := "getBlockChainInfo"

		//call
		out, err := rpc(method, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}
