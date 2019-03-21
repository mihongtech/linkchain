package rpcserver

import (
	"errors"
	"fmt"
	"github.com/linkchain/contract"
	"math/big"
	"reflect"
	"time"

	"github.com/linkchain/common"
	"github.com/linkchain/common/hexutil"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/contract/vm"
	"github.com/linkchain/core"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/helper"
	"github.com/linkchain/node"
	"github.com/linkchain/rpc/rpcobject"

	"github.com/golang/protobuf/proto"
)

func publishContract(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.PublishContractCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	contractId, _ := meta.NewAccountIdFromStr("0000000000000000000000000000000000000000")
	bestHeight := GetNodeAPI(s).GetBestBlock().GetHeight()

	amount := meta.NewAmount(c.Amount)
	toCoin := helper.CreateToCoin(*contractId, amount)

	//make from
	from, err := GetWalletAPI(s).GetAccount(c.FromAccountId)
	if err != nil {
		return nil, err
	}
	fromCoin, fromAmount, err := from.MakeFromCoin(amount, bestHeight+1)
	if err != nil {
		return nil, err
	}

	transaction := helper.CreateTransaction(*fromCoin, *toCoin)
	backChange := helper.CreateToCoin(from.Id, fromAmount.Subtraction(*amount))
	if backChange.Value.GetInt64() > 0 {
		transaction.AddToCoin(*backChange)
	}

	transaction.Type = contract.ContractTx

	// generate tx data
	contractCode := common.Hex2Bytes(c.Contract)
	data := contract.TxData{Price: new(big.Int).SetInt64(c.GasPrice), GasLimit: c.GasLimit, Payload: contractCode}
	protobufMsg := data.Serialize()
	transaction.Data, err = proto.Marshal(protobufMsg)
	if err != nil {
		log.Error("Marshal tx contract paload error", "err", err)
		return nil, err
	}

	transaction, err = GetWalletAPI(s).SignTransaction(*transaction)
	if err != nil {
		return nil, err
	}

	if err = GetNodeAPI(s).ProcessTx(transaction); err == nil {
		GetNodeAPI(s).GetTxEvent().Send(node.TxEvent{transaction})
	}

	/*&rpcobject.PublishContractRSP{
		TxID:         transaction.GetTxID().String(),
		ContractAddr: vm.CreateContractAccountID(from.Id, *transaction.GetTxID()).String(),
		PlayLoad:     c.Contract,
		GasPrice:     int(data.Price.Int64()),
		GasLimit:     int(data.GasLimit),
	}*/
	return transaction, err
}

func getCode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.GetCodeCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	contractId, _ := meta.NewAccountIdFromStr(c.FromAccountId)
	// height := c.Height
	data, err := GetNodeAPI(s).GetCode(*contractId)
	if err != nil {
		fmt.Println("Get Code failed", err)
		return nil, err
	}

	return data, nil
}

func callContract(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.CallContractCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	contractId, _ := meta.NewAccountIdFromStr(c.Contract)
	bestHeight := GetNodeAPI(s).GetBestBlock().GetHeight()

	amount := meta.NewAmount(c.Amount)
	toCoin := helper.CreateToCoin(*contractId, amount)

	//make from
	from, err := GetWalletAPI(s).GetAccount(c.FromAccountId)
	if err != nil {
		return nil, err
	}
	fromCoin, fromAmount, err := from.MakeFromCoin(amount, bestHeight+1)
	if err != nil {
		return nil, err
	}

	transaction := helper.CreateTransaction(*fromCoin, *toCoin)
	backChange := helper.CreateToCoin(from.Id, fromAmount.Subtraction(*amount))
	if backChange.Value.GetInt64() > 0 {
		transaction.AddToCoin(*backChange)
	}

	transaction.Type = contract.ContractTx

	// generate tx data
	contractCode := common.Hex2Bytes(c.CallMethod)
	data := contract.TxData{Price: new(big.Int).SetInt64(1), GasLimit: 100000000, Payload: contractCode}
	protobufMsg := data.Serialize()
	transaction.Data, err = proto.Marshal(protobufMsg)
	if err != nil {
		log.Error("Marshal tx contract paload error", "err", err)
		return nil, err
	}

	transaction, err = GetWalletAPI(s).SignTransaction(*transaction)
	if err != nil {
		return nil, err
	}

	if err = GetNodeAPI(s).ProcessTx(transaction); err == nil {
		GetNodeAPI(s).GetTxEvent().Send(node.TxEvent{transaction})
	}

	return nil, err
}

func doCall(s *Server, args *rpcobject.CallCmd, vmCfg vm.Config, timeout time.Duration) ([]byte, bool, error) {
	defer func(start time.Time) { log.Debug("Executing EVM call finished", "runtime", time.Since(start)) }(time.Now())

	block := GetNodeAPI(s).GetBestBlock()
	header := block.Header

	state, err := GetNodeAPI(s).StateAt(header.Status)
	if err != nil {
		return nil, false, err
	}
	contractId, _ := meta.NewAccountIdFromStr(args.Contract)

	from, _ := meta.NewAccountIdFromStr(args.FromAccountId)

	gas := uint64(1000000000)
	gasPrice := big.NewInt(10)

	var addr meta.AccountID
	if from.IsEmpty() {
		if wallets := GetWalletAPI(s).GetAllWAccount(); len(wallets) > 0 {
			addr = *(wallets[0].GetAccountID())
		}
	} else {
		addr = *from
	}

	msg := contract.NewMessage(addr, contractId, big.NewInt(0), gas, gasPrice, common.FromHex(args.Data), false)

	statedb := contract.NewStateAdapter(state, math.Hash{}, *header.GetBlockID(), meta.AccountID{}, int64(header.Height)+1)

	// Get a new instance of the EVM.
	evm, vmError, err := GetEVM(msg, statedb, &header, vmCfg, GetNodeAPI(s))
	if err != nil {
		return nil, false, err
	}

	// Setup the gas pool (also for unmetered requests)
	// and apply the message.
	res, _, failed, err := contract.ApplyMessage(evm, msg, new(core.GasPool).AddGas(math.MaxUint64))
	if err := vmError(); err != nil {
		return nil, false, err
	}
	return res, failed, err
}

func GetEVM(msg contract.Message, state *contract.StateAdapter, header *meta.BlockHeader, vmCfg vm.Config, node *node.PublicNodeAPI) (*vm.EVM, func() error, error) {
	from := msg.From()

	state.AddBalance(from, big.NewInt(math.MaxInt64/2-1))
	//	if err != nil {
	//		log.Error("initial balance faled", "err", err)
	//	}
	vmError := func() error { return nil }

	context := contract.NewEVMContext(msg, header, node, nil)
	return vm.NewEVM(context, state, node.GetChainConfig(), vmCfg), vmError, nil
}

// Call executes the given transaction on the state for the given block number.
// It doesn't make and changes in the state/blockchain and is useful to execute and retrieve values.
func call(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.CallCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	// contractId, _ := meta.NewAccountIdFromStr(c.FromAccountId)
	// height := c.Height

	result, _, err := doCall(s, c, vm.Config{}, 5*time.Second)
	if err != nil {

		log.Error("result is", "result", hexutil.Encode(result), "err", err)
		return nil, err
	}
	log.Info("call result is", "result", hexutil.Encode(result))
	return hexutil.Bytes(result), err
}

func GetTransactionReceipt(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.GetTransactionReceiptCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}
	hash, err := math.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, err
	}

	tx, blockHash, _, _ := GetNodeAPI(s).GetTXByID(*hash)
	if tx == nil {
		return nil, errors.New("The contract tx have not been mined in block")
	}

	if tx == nil {
		return nil, nil
	}
	receipts := GetNodeAPI(s).GetReceiptsByHash(blockHash)

	for i, data := range receipts {
		if data.TxHash.IsEqual(tx.GetTxID()) {
			return receipts[i], nil
		}
	}

	//	from := tx.GetFromCoins()[0].GetId()
	//
	//	fields := core.Receipt{
	//		"blockHash":        blockHash,
	//		"blockNumber":      hexutil.Uint64(blockNumber),
	//		"transactionHash":  hash,
	//		"transactionIndex": hexutil.Uint64(index),
	//		"from":             from,
	//		"to":               tx.GetToCoins()[0].Id,
	//		"contractAddress":  nil,
	//		"logs":             receipt.Logs,
	//		"logsBloom":        receipt.Bloom,
	//	}
	//
	//	// Assign receipt status or post state.
	//	if len(receipt.PostState) > 0 {
	//		fields["root"] = hexutil.Bytes(receipt.PostState)
	//	} else {
	//		fields["status"] = hexutil.Uint(receipt.Status)
	//	}
	//	if receipt.Logs == nil {
	//		fields["logs"] = [][]*meta.Log{}
	//	}
	//	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	//	if receipt.ContractAddress != (meta.AccountID{}) {
	//		fields["contractAddress"] = receipt.ContractAddress
	//	}
	return nil, errors.New("get receipt failed")
}
