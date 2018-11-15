package cmd

import (
	"encoding/hex"
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/node"
	"github.com/linkchain/core/meta"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(accountCmd)
	accountCmd.AddCommand(getAccountByPubCmd, allCmd)
}

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "watch account manage status",
}

var getAccountByPubCmd = &cobra.Command{
	Use:   "pubkey",
	Short: "get account by pubkey",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Error("send", "error", "please input account", "example", "tx send 55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b 10")
			return
		}
		buffer, err := hex.DecodeString(args[0])
		if err != nil {
			log.Error("send ", "error", "hex Decode failed")
			return
		}
		pb, err := btcec.ParsePubKey(buffer, btcec.S256())
		if err != nil {
			log.Error("send ", "error", "account is error", "season", err)
			return
		}

		id := meta.NewAccountId(pb)

		a, err := node.GetAccount(*id)
		if err != nil {
			log.Error("send ", "error", "get account is error", "season", err)
			return
		}
		log.Info("send", "account", a.GetAccountID(), "amount", a.GetAmount().GetInt64(), "nounce", a.GetAmount().GetInt64())
	},
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "get all account",
	Run: func(cmd *cobra.Command, args []string) {
		//TODO need to give up ,because get all accountmessage is too waste
		//node.GetAllAccounts()
	},
}
