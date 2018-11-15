package cmd

import (
	"encoding/hex"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(txCmd)
	txCmd.AddCommand(createTxCmd, signTxCmd, sendTxCmd, decodeTxCmd)
}

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "all tx related command",
}

var createTxCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new tx",
	Run: func(cmd *cobra.Command, args []string) {
		/*amount := amount2.NewAmount(10)
		from, err := node.GetWallet().ChooseWAccount(*amount)
		if err != nil {
			println("cmd :can not find from")
			return
		}
		fromAccount := from.GetAccountID()
		toAccount := node.GetConsensusService().GetAccountManager().NewAccount()

		tx := manage.GetManager().TransactionManager.CreateTransaction(fromAccount, toAccount, amount)
		buffer, err := proto.Marshal(tx.Serialize())
		if err != nil {
			log.Error("tx Serialize failed", "Marshaling error", err)
		}
		log.Info("createtx", "data", tx)
		log.Info("createtx", "hex", hex.EncodeToString(buffer))*/
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

		var tx *meta.Transaction = &meta.Transaction{}
		err = tx.Deserialize(&txData)
		if err != nil {
			log.Error("signtx Deserialize failed", "Deserialize error", err)
			return
		}
		//app.GetWallet().SignTransaction(tx)

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

		var tx *meta.Transaction = &meta.Transaction{}
		err = tx.Deserialize(&txData)
		if err != nil {
			log.Error("sendtx Deserialize failed", "Deserialize error", err)
			return
		}

		log.Info("sendtx", "data", tx)
		signbuffer, err := proto.Marshal(tx.Serialize())
		log.Info("sendtx", "hex", hex.EncodeToString(signbuffer))

		err = tx.Verify()

		if err != nil {
			log.Info("Verify tx", "successed", false)
		} else {
			log.Info("Verify tx", "successed", true)
		}

		//node.ProcessTx(tx)
		//node.NewTxEvent.Send(meta_tx.TxEvent{tx})

	},
}

var decodeTxCmd = &cobra.Command{
	Use:   "decode",
	Short: "decode tx",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Error("decode", "error", "please hex tx")
			return
		}
		buffer, err := hex.DecodeString(args[0])
		if err != nil {
			log.Error("decode ", "error", "hex Decode failed")
			return
		}

		txData := protobuf.Transaction{}
		err = proto.Unmarshal(buffer, &txData)

		if err != nil {
			log.Error("decode Deserialize failed", "Unmarshal error", err)
			return
		}
		log.Info("decode", txData.String())

		var tx *meta.Transaction = &meta.Transaction{}
		err = tx.Deserialize(&txData)
		if err != nil {
			log.Error("decode Deserialize failed", "Deserialize error", err)
			return
		}

		log.Info("decode", "data", tx)
	},
}
