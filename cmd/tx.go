package cmd

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/function/wallet"
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/node"
	"github.com/linkchain/poa/manage"
	"github.com/linkchain/poa/meta"
	"github.com/linkchain/protobuf"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(txCmd)
	txCmd.AddCommand(createTxCmd, signTxCmd, sendTxCmd, testTxCmd)
}

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "all tx related command",
}

var createTxCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new tx",
	Run: func(cmd *cobra.Command, args []string) {
		fromAccount := node.GetWallet().GetWAccount()
		if fromAccount == nil {
			println("cmd :can not find from")
			return
		}
		toAccount := node.GetConsensusService().GetAccountManager().NewAccount()
		amount := &meta.Amount{Value: 10}
		tx := manage.GetManager().TransactionManager.NewTransaction(fromAccount, toAccount, amount)
		buffer, err := proto.Marshal(tx.Serialize())
		if err != nil {
			log.Error("tx Serialize failed", "Marshaling error", err)
		}
		log.Info("createtx", "data", tx)
		log.Info("createtx", "hex", hex.EncodeToString(buffer))
	},
}

var signTxCmd = &cobra.Command{
	Use:   "sign",
	Short: "sign a new tx",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Error("signtx", "error", "please hex tx")
			return
		}
		buffer, err := hex.DecodeString(args[0])
		if err != nil {
			log.Error("signtx ", "error", "hex Decode failed")
			return
		}

		txData := protobuf.Transaction{}
		err = proto.Unmarshal(buffer, &txData)

		if err != nil {
			log.Error("signtx Deserialize failed", "Unmarshal error", err)
			return
		}
		log.Info("signtx", txData.String())

		var tx tx.ITx = &meta.Transaction{}
		tx.Deserialize(&txData)

		node.GetWallet().SignTransaction(tx)

		log.Info("signtx", "data", tx)
		signbuffer, err := proto.Marshal(tx.Serialize())
		log.Info("signtx", "hex", hex.EncodeToString(signbuffer))

		err = tx.Verify()

		if err != nil {
			log.Info("Verify tx", "successed", false)
		} else {
			log.Info("Verify tx", "successed", true)
		}
	},
}

var sendTxCmd = &cobra.Command{
	Use:   "send",
	Short: "send a new tx to network",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Error("sendtx", "error", "please hex tx")
			return
		}
		buffer, err := hex.DecodeString(args[0])
		if err != nil {
			log.Error("sendtx ", "error", "hex Decode failed")
			return
		}

		txData := protobuf.Transaction{}
		err = proto.Unmarshal(buffer, &txData)

		if err != nil {
			log.Error("sendtx Deserialize failed", "Unmarshal error", err)
			return
		}
		log.Info("sendtx", txData.String())

		var tx tx.ITx = &meta.Transaction{}
		tx.Deserialize(&txData)

		log.Info("sendtx", "data", tx)
		signbuffer, err := proto.Marshal(tx.Serialize())
		log.Info("sendtx", "hex", hex.EncodeToString(signbuffer))

		err = tx.Verify()

		if err != nil {
			log.Info("Verify tx", "successed", false)
		} else {
			log.Info("Verify tx", "successed", true)
		}

		manage.GetManager().TransactionManager.ProcessTx(tx)
	},
}

var testTxCmd = &cobra.Command{
	Use:   "test",
	Short: "send a new tx to network",
	Run: func(cmd *cobra.Command, args []string) {
		fromAccount := node.GetWallet().GetWAccount()
		if fromAccount == nil {
			println("cmd :can not find from")
			return
		}
		toAccount := node.GetConsensusService().GetAccountManager().NewAccount()
		amount := &meta.Amount{Value: 10}
		tx := manage.GetManager().TransactionManager.NewTransaction(fromAccount, toAccount, amount)
		tx.Deserialize(tx.Serialize())
		node.GetWallet().SignTransaction(tx)
		manage.GetManager().TransactionManager.ProcessTx(tx)
		account := wallet.NewWSAccount()
		account.GetAccountInfo()
	},
}
