package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(quitCmd)
}

var quitCmd = &cobra.Command{
	Use:   "quit",
	Short: "quit linkchain client",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(0)
	},
}
