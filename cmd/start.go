package cmd

import (
	"github.com/linkchain/node"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start linkchain node",
	Run: func(cmd *cobra.Command, args []string) {
		if !node.Init(nil) {
			return
		}
		node.Run()
	},
}
