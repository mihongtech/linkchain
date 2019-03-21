package cmd

import (
	"fmt"
	"strconv"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/rpc/rpcobject"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(walletCmd)
	walletCmd.AddCommand(getWalletInfoCmd,
		getAccountCmd,
		newAddressCmd,
		sendMoneyCmd,
		importCmd,
		exportCmd)
}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "handle wallet cmd",
}

// get wallet information
var getWalletInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "get wallet info",
	Run: func(cmd *cobra.Command, args []string) {
		method := "getWalletInfo"

		//call
		out, err := rpc(method, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

// get account information
var getAccountCmd = &cobra.Command{
	Use:     "account <accountID>",
	Short:   "get account info",
	Example: "wallet getaccount <accountID>, wallet getAccount 55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Error("getAccount", "error", "please input <accountId>")
			return
		}

		accountID := args[0]
		method := "getAccountInfo"

		//call
		out, err := rpc(method, &rpcobject.SingleCmd{accountID})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

// create new account
var newAddressCmd = &cobra.Command{
	Use:   "newaccount",
	Short: "generate new wallet account address",
	Run: func(cmd *cobra.Command, args []string) {
		method := "newAcount"

		//call
		out, err := rpc(method, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

//normal transaction
var sendMoneyCmd = &cobra.Command{
	Use:     "send <from account> <target account> <money>",
	Short:   "send money to other account",
	Example: "wallet send 02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50 55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b 10",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 3 {
			log.Error("send", "error", "incorrect parameter number")
			return
		}
		fromAccountID := args[0]
		toAccountID := args[1]
		amount, err := strconv.Atoi(args[2])
		if err != nil {
			log.Error("send", "error", "please input money:int")
			return
		}

		method := "sendMoneyTransaction"

		//call
		out, err := rpc(method, &rpcobject.SendToTxCmd{fromAccountID, toAccountID, amount})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

var importCmd = &cobra.Command{
	Use:     "import",
	Short:   "import account",
	Example: "wallet import 55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b",
	Run: func(cmd *cobra.Command, args []string) {
		method := "importAccount"
		if len(args) != 1 {
			log.Error("wallet", "input error", "please input <privkey hex str>")
			return
		}
		//call
		out, err := rpc(method, rpcobject.ImportAccountCmd{Signer: args[0]})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "export account",
	Example: "wallet export 025aa040dddd8f873ac5d02dfd249adc4d2c9d6def472a4405252fa6f6650ee1f0",
	Run: func(cmd *cobra.Command, args []string) {
		method := "exportAccount"
		if len(args) != 1 {
			log.Error("export", "error", "please input <accountId hex str>", "example", "wallet export 025aa040dddd8f873ac5d02dfd249adc4d2c9d6def472a4405252fa6f6650ee1f0")
			return
		}
		//call
		out, err := rpc(method, rpcobject.ExportAccountCmd{AccountId: args[0]})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}
