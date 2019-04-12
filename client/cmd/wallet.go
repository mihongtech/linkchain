package cmd

import (
	"fmt"
	"strconv"

	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/rpc/rpcobject"

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
	Short: "wallet command",
	Long:  "This is all wallet command for handling wallet",
}

// get wallet information
var getWalletInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "wallet info",
	Long:    "This is get wallet info command",
	Example: "wallet info",
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
	Use:     "account",
	Short:   "wallet account <address>",
	Long:    "This is get account info command",
	Example: "wallet account 55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b",
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
	Use:     "newaccount",
	Short:   "wallet newaccount",
	Long:    "This is generate new account command",
	Example: "wallet newaccount",
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
	Use:     "send ",
	Short:   "send <from_address> <target_address> <amount>",
	Long:    "This is send money to account command(normal tx)",
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
	Short:   "import <privkey>",
	Long:    "This is import privkey into wallet command",
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
	Short:   "export <address>",
	Long:    "This is export privkey from wallet command",
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
