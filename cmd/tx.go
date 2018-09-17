package cmd


import (
	"github.com/spf13/cobra"
	"github.com/linkchain/poa/poamanager"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/poa/meta"
	"github.com/linkchain/common/math"
	"github.com/golang/protobuf/proto"
	"encoding/hex"
)

func init() {
	RootCmd.AddCommand(txCmd)
	txCmd.AddCommand(createTxCmd, signTxCmd, sendTxCmd)
}

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "all tx related command",
}

var createTxCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new tx",
	Run: func(cmd *cobra.Command, args []string) {
		println("New tx generated")
		fromAddress := math.DoubleHashH([]byte("lf"))
		toAddress := math.DoubleHashH([]byte("lc"))

		formAccount := &meta.POAAccount{AccountID:meta.POAAccountID{ID:fromAddress}}
		toAccount := &meta.POAAccount{AccountID:meta.POAAccountID{ID:toAddress}}
		amount := &meta.POAAmount{Value:10}
		tx := poamanager.GetManager().TransactionManager.NewTransaction(formAccount,toAccount,amount)
		tx.Deserialize(tx.Serialize())
		log.Info("createtx","data",tx)
	},
}

var signTxCmd = &cobra.Command{
	Use:   "sign",
	Short: "sign a new tx",
	Run: func(cmd *cobra.Command, args []string) {
		println("Tx signed")
	},
}

var sendTxCmd = &cobra.Command{
	Use:   "send",
	Short: "send a new tx to network",
	Run: func(cmd *cobra.Command, args []string) {
		println("Tx send out")
		fromAddress := math.DoubleHashH([]byte("ls"))
		toAddress := math.DoubleHashH([]byte("lc"))
		formAccount := &meta.POAAccount{AccountID:meta.POAAccountID{ID:fromAddress}}
		toAccount := &meta.POAAccount{AccountID:meta.POAAccountID{ID:toAddress}}
		amount := &meta.POAAmount{Value:10}
		tx := poamanager.GetManager().TransactionManager.NewTransaction(formAccount,toAccount,amount)
		tx.Deserialize(tx.Serialize())
		buffer,err := proto.Marshal(tx.Serialize())
		if err != nil {
			log.Error("block 序列化不通过 marshaling error",err)
		}
		log.Info("send tx","txid",tx.GetTxID().GetString(),"data", hex.EncodeToString(buffer))
		poamanager.GetManager().TransactionManager.ProcessTx(tx)
	},
}
