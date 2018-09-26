package cmd

import (
	"github.com/linkchain/poa/manage"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(chainCmd)
	chainCmd.AddCommand(chainInfoCmd)
}

var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "handle chain",
}

var chainInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "getBlockChainInfo",
	Run: func(cmd *cobra.Command, args []string) {
		manage.GetManager().ChainManager.GetBlockChainInfo()
	},
}
