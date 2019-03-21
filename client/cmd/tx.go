package cmd

import (
	"fmt"
	"github.com/linkchain/common/util/log"

	"github.com/linkchain/rpc/rpcobject"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(txCmd)
	txCmd.AddCommand(getTxCmd)
}

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "all tx related command",
}

var getTxCmd = &cobra.Command{
	Use:   "get",
	Short: "get tx",
	Run: func(cmd *cobra.Command, args []string) {
		example := []string{"example", "tx hash 98acd27a58c79eaab05ea4abd0daa8e63021df3bf2e65fcb38e2474fb706c3fe"}
		if len(args) != 1 {
			log.Error("gettxbyhash", "error", "please input hash", example[0], example[1])
			return
		}

		hash := args[0]
		method := "getTxByHash"

		//call
		out, err := rpc(method, &rpcobject.GetTransactionByHashCmd{hash})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}
