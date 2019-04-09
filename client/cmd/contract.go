package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/contract"
	"github.com/linkchain/contract/vm"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/rpc/rpcobject"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(contractCmd)
	contractCmd.AddCommand(
		publishContractCmd,
		callContractCmd,
		transactContractCmd,
		GetCallContractCmd)
}

var contractCmd = &cobra.Command{
	Use:   "contract",
	Short: "contract command",
	Long:  "This is all contract command for handling contract",
}

var publishContractCmd = &cobra.Command{
	Use:     "publish",
	Short:   "publish <from_address> <amount> <code>",
	Long:    "This is create contract command",
	Example: "contract publish 8dafd997b6e65e680768076d92821716fd7950ee 3 6060604052600a8060106000396000f360606040526008565b00",
	Run: func(cmd *cobra.Command, args []string) {
		example := []string{"example", "contract publish 8dafd997b6e65e680768076d92821716fd7950ee 6060604052600a8060106000396000f360606040526008565b00 3"}
		if len(args) != 3 {
			log.Error("publishContractCmd", "error", "please input address ,contract and value", example[0], example[1])
			return
		}

		account := args[0]
		code := args[2] //608060405260405160208061024e83398101806040528101908080519060200190929190505050806000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550506101d5806100796000396000f30060806040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806370a0823114610051578063a9059cbb146100a8575b600080fd5b34801561005d57600080fd5b50610092600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506100f5565b6040518082815260200191505060405180910390f35b3480156100b457600080fd5b506100f3600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061010d565b005b60006020528060005260406000206000915090505481565b806000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540392505081905550806000808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828254019250508190555050505600a165627a7a72305820c2ae0f9369f01615f3781c705c030f59698029756da6b54624c98b8c3f54381d00290000000000000000000000000000000000000000000000000000000000000064
		amount, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			log.Error("send", "error", "please input money:int")
			return
		}
		method := "publishContract"
		//call
		out, err := rpc(method, &rpcobject.PublishContractCmd{account, code, amount, 1, 100000000})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		tx := meta.Transaction{}
		if err := json.Unmarshal([]byte(out), &tx); err != nil {
			fmt.Println("json unmarshal tx err:%s", err.Error())
		}
		contractAddr := vm.CreateContractAccountID(tx.From.Coins[0].Id, *tx.GetTxID())
		contractData := contract.GetTxData(&tx)
		rsp := rpcobject.PublishContractRSP{
			TxID:         tx.GetTxID().String(),
			ContractAddr: contractAddr.String(),
			GasPrice:     contractData.Price.Int64(),
			GasLimit:     int64(contractData.GasLimit),
			PlayLoad:     code[:90] + "...",
		}

		jsonBuff, err := json.Marshal(rsp)
		if err != nil {
			fmt.Println("json marshal result err:%s", err.Error())
		}
		var jsonOut bytes.Buffer
		json.Indent(&jsonOut, jsonBuff, "", "  ")
		fmt.Println(jsonOut.String())
	},
}

var callContractCmd = &cobra.Command{
	Use:     "call",
	Short:   "call <from_address> <contract_address> <call_method>",
	Long:    "This is call contract command which is only run in local vm",
	Example: "contract call 8dafd997b6e65e680768076d92821716fd7950ee 91386e326c72b5d7f92431689d3ca921e13de072 70a082310000000000000000000000000a35c1bd74497c851265774e7e98027b46c27c41",
	Run: func(cmd *cobra.Command, args []string) {
		example := []string{"example", "contract call 8dafd997b6e65e680768076d92821716fd7950ee 91386e326c72b5d7f92431689d3ca921e13de072 70a082310000000000000000000000000a35c1bd74497c851265774e7e98027b46c27c41"}
		if len(args) != 3 {
			log.Error("callContractCmd", "error", "please input address and contract", example[0], example[1])
			return
		}

		account := args[0]
		contract := args[1]
		callMethod := args[2]
		method := "call"

		//call
		out, err := rpc(method, &rpcobject.CallCmd{account, contract, callMethod, 0})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

var transactContractCmd = &cobra.Command{
	Use:     "transact",
	Short:   "transact <from_address> <contract_address> <call_method> <amount>",
	Long:    "This is call contract command which is only run on-chain vm",
	Example: "contract transact 8dafd997b6e65e680768076d92821716fd7950ee 91386e326c72b5d7f92431689d3ca921e13de072 70a082310000000000000000000000000a35c1bd74497c851265774e7e98027b46c27c41 3",
	Run: func(cmd *cobra.Command, args []string) {
		example := []string{"example", "contract transact 8dafd997b6e65e680768076d92821716fd7950ee 98acd27a58c79eaab05ea4abd0daa8e63021df3bf2e65fcb38e2474fb706c3fe d27a58c79eaab05ea4abd0daa8e63021df3bf2e65fcb38e2474fb706c3fe"}
		if len(args) != 4 {
			log.Error("callContractCmd", "error", "please input address and contract", example[0], example[1])
			return
		}

		account := args[0]
		contract := args[1]
		callMethod := args[2]
		amount, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			log.Error("send", "error", "please input money:int")
			return
		}
		method := "callContract"

		//call
		out, err := rpc(method, &rpcobject.CallContractCmd{account, contract, callMethod, amount, 1, 100000000})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

var GetCallContractCmd = &cobra.Command{
	Use:     "get",
	Short:   "get <hash>",
	Long:    "This is get contract receipt command",
	Example: "contract get d27a58c79eaab05ea4abd0daa8e63021df3bf2e65fcb38e2474fb706c3fe",
	Run: func(cmd *cobra.Command, args []string) {
		example := []string{"example", "contract get d27a58c79eaab05ea4abd0daa8e63021df3bf2e65fcb38e2474fb706c3fe"}
		if len(args) != 1 {
			log.Error("callContractCmd", "error", "please input address and contract", example[0], example[1])
			return
		}

		txid := args[0]

		method := "transactionReceipt"
		//call
		out, err := rpc(method, &rpcobject.GetTransactionReceiptCmd{txid})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}
