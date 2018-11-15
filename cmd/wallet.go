package cmd

import (
	_ "encoding/hex"

	_ "github.com/linkchain/common/btcec"
	_ "github.com/linkchain/common/util/log"
	_ "github.com/linkchain/wallet"
	_ "github.com/linkchain/core/meta"
	_ "github.com/linkchain/app"
	_ "github.com/linkchain/node"
	"github.com/spf13/cobra"
	_ "strconv"
)

func init() {
	RootCmd.AddCommand(walletCmd)
	walletCmd.AddCommand(getWalletInfoCmd,
		getNewAccountCmd,
		sendMoneyCmd,
		importKeyCmd,
		exportCmd)
}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "handle wallet cmd",
}

var getWalletInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "get wallet info",
	Run: func(cmd *cobra.Command, args []string) {
		//was := app.GetWallet().GetAllWAccount()
		//for _, wa := range was {
		//	wa.GetAccountInfo()
		//}
	},
}

var getNewAccountCmd = &cobra.Command{
	Use:   "new",
	Short: "generate new wallet account",
	Run: func(cmd *cobra.Command, args []string) {
		//wa := wallet.NewWSAccount()
		//app.GetWallet().AddWAccount(wa)
		//log.Info("Wallet Info", "new wallet account", wa.GetAccountPubkey())
	},
}

var sendMoneyCmd = &cobra.Command{
	Use:   "send",
	Short: "send money to other account",
	Run: func(cmd *cobra.Command, args []string) {
		//if len(args) != 2 {
		//	log.Error("send", "error", "please input account", "example", "tx send 55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b 10")
		//	return
		//}
		//buffer, err := hex.DecodeString(args[0])
		//if err != nil {
		//	log.Error("send ", "error", "hex Decode failed")
		//	return
		//}
		//pb, err := btcec.ParsePubKey(buffer, btcec.S256())
		//if err != nil {
		//	log.Error("send ", "error", "account is error", "season", err)
		//	return
		//}
		//a, err := strconv.Atoi(args[1])
		//if err != nil {
		//	log.Error("send", "error", "please input money:int", "example", "tx send 55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b 10")
		//	return
		//}
		//amount := amount.NewAmount(int64(a))
		//toID, err := util.CreateAccountIdByPubKey(hex.EncodeToString(pb.SerializeCompressed()))
		//toCoin := util.CreateToCoin(toID, amount)
		//
		//from, err := app.GetWallet().ChooseWAccount(amount)
		//if err != nil {
		//	log.Error("send ", "error", "input is more than account's amount", "season", err)
		//	return
		//}
		//fromCoin, fromAmount, err := from.MakeFromCoin(amount)
		//if err != nil {
		//	log.Error("send ", "error", "input is more than account's amount", "season", err)
		//	return
		//}
		//toFromCoin := util.CreateToCoin(from.GetAccountID(), fromAmount.Subtraction(*amount))
		//
		//transaction := util.CreateTransaction(fromCoin, toCoin)
		//transaction.AddToCoin(toFromCoin)
		//
		//transaction, err = app.GetWallet().SignTransaction(transaction)
		//if err != nil {
		//	log.Error("send ", "error", "sign tx is failed", "season", err)
		//	return
		//}
		//log.Info("send", "txid", transaction.GetTxID().GetString())
		//node.GetManager().TransactionManager.ProcessTx(transaction)
		//node.GetManager().NewTxEvent.Send(tx.TxEvent{transaction})
	},
}

var importKeyCmd = &cobra.Command{
	Use:   "import",
	Short: "import privkey",
	Run: func(cmd *cobra.Command, args []string) {
		//if len(args) != 1 {
		//	log.Error("importprivkey", "error", "please input privkey", "example", "wallet import 6647e717248720f1b50f3f1f765b731783205f2de2fedc9e447438966af7df82")
		//	return
		//}
		//buffer, err := hex.DecodeString(args[0])
		//if err != nil {
		//	log.Error("importprivkey ", "error", "hex Decode failed")
		//	return
		//}
		//wa := wallet.CreateWAccountFromBytes(buffer, amount.NewAmount(0))
		//app.GetWallet().AddWAccount(wa)
		//log.Info("Wallet Info", "new wallet account", wa.GetAccountPubkey())
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export privkey",
	Run: func(cmd *cobra.Command, args []string) {
		//if len(args) != 1 {
		//	log.Error("export", "error", "please input pubkey", "example", "wallet export 025aa040dddd8f873ac5d02dfd249adc4d2c9d6def472a4405252fa6f6650ee1f0")
		//	return
		//}
		//
		//wa, err := app.GetWallet().GetWAccount(args[0])
		//if err != nil {
		//	log.Error("export", "error", err, "example", "wallet export 025aa040dddd8f873ac5d02dfd249adc4d2c9d6def472a4405252fa6f6650ee1f0")
		//	return
		//}
		//log.Info("Wallet Info", "new wallet account", wa.GetAccountPrivkey())
	},
}
