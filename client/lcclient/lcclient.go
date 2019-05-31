package lcclient

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/mihongtech/linkchain/client/httpclient"
	"github.com/mihongtech/linkchain/common"
	"github.com/mihongtech/linkchain/common/hexutil"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/contract"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/protobuf"
	"github.com/mihongtech/linkchain/rpc/rpcjson"
	"github.com/mihongtech/linkchain/rpc/rpcobject"

	"github.com/golang/protobuf/proto"
)

var httpConfig = &httpclient.Config{
	RPCUser:     "lc",
	RPCPassword: "lc",
	RPCServer:   "localhost:8082",
}

// Client defines typed wrappers for the Ethereum RPC API.
type Client struct {
	c string
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c string) *Client {
	return &Client{c}
}

// Blockchain Access

// BlockByHash returns the given full block.
//
// Note that loading full blocks requires two requests. Use HeaderByHash
// if you don't need all transactions or uncle headers.
func (ec *Client) BlockByHash(ctx context.Context, hash math.Hash) (*meta.Block, error) {
	// return ec.getBlock(ctx, "eth_getBlockByHash", hash, true)
	return nil, nil
}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (ec *Client) BlockByNumber(ctx context.Context, number *big.Int) (*meta.Block, error) {
	// return ec.getBlock(ctx, "eth_getBlockByNumber", toBlockNumArg(number), true)
	return nil, nil
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ec *Client) CodeAt(ctx context.Context, account meta.AccountID, blockNumber *big.Int) ([]byte, error) {
	method := "getCode"
	//call
	if blockNumber == nil {
		blockNumber = big.NewInt(-1)
	}

	data, err := rpc(method, &rpcobject.GetCodeCmd{account.String(), blockNumber.Int64()})

	return data, err
}

// PendingCodeAt returns the contract code of the given account in the pending state.
func (ec *Client) PendingCodeAt(ctx context.Context, account meta.AccountID) ([]byte, error) {
	//	var result hexutil.Bytes
	//	err := ec.c.CallContext(ctx, &result, "eth_getCode", account, "pending")
	//	return result, err
	return nil, nil
}

func (ec *Client) TransactionReceipt(ctx context.Context, txHash math.Hash) (*core.Receipt, error) {
	method := "transactionReceipt"
	//call
	data, err := rpc(method, &rpcobject.GetTransactionReceiptCmd{txHash.String()})
	var receipt core.Receipt
	if err = json.Unmarshal(data, &receipt); err != nil {
		log.Error("Unmarshal json failed", "data", data)
		return nil, err
	}
	return &receipt, nil
}

// CallContract executes a message call transaction, which is directly executed in the VM
// of the node, but never mined into the chain.
//
// blockNumber selects the block height at which the call runs. It can be nil, in which
// case the code is taken from the latest known block. Note that state from very old
// blocks might not be available.
func (ec *Client) CallContract(ctx context.Context, msg contract.Message, blockNumber *big.Int) ([]byte, error) {
	method := "call"
	//call
	if blockNumber == nil {
		blockNumber = big.NewInt(-1)
	}
	data, err := rpc(method, &rpcobject.CallCmd{msg.From().String(), msg.To().String(), common.ToHex(msg.Data()), blockNumber.Int64()})
	if err != nil {
		return nil, err
	}
	var hex hexutil.Bytes
	json.Unmarshal(data, &hex)
	return hex, nil
}

// PendingCallContract executes a message call transaction using the EVM.
// The state seen by the contract call is the pending state.
func (ec *Client) PendingCallContract(ctx context.Context, msg contract.Message) ([]byte, error) {
	//	var hex hexutil.Bytes
	//	err := ec.c.CallContext(ctx, &hex, "eth_call", toCallArg(msg), "pending")
	//	if err != nil {
	//		return nil, err
	//	}
	//	return hex, nil
	return nil, nil
}

// SendTransaction injects a signed transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (ec *Client) SendTransaction(ctx context.Context, tx *meta.Transaction) (*meta.Transaction, error) {
	account := tx.GetFromCoins()[0].Id.String()

	amount := tx.GetToValue().GetBigInt().Int64()

	extraData := protobuf.TxData{}
	proto.Unmarshal(tx.Data, &extraData)
	data := contract.TxData{}
	data.Deserialize(&extraData)
	contractCode := common.Bytes2Hex(data.Payload)

	if tx.GetToCoins()[0].GetId().IsEmpty() {

		method := "publishContract"
		//call
		data, err := rpc(method, &rpcobject.PublishContractCmd{account, contractCode, amount, data.Price.Int64(), data.GasLimit})
		if err != nil {
			return nil, err
		}

		signedTx := meta.Transaction{}
		if err = json.Unmarshal(data, &signedTx); err != nil {
			log.Error("Unmarshal json failed", "data", data)
			return nil, err
		}

		return &signedTx, nil
	} else {
		method := "callContract"
		//call
		data, err := rpc(method, &rpcobject.CallContractCmd{account, tx.GetToCoins()[0].GetId().String(), contractCode, amount, data.Price.Int64(), data.GasLimit})
		if err != nil {
			return nil, err
		}
		signedTx := meta.Transaction{}
		if err = json.Unmarshal(data, &signedTx); err != nil {
			log.Error("Unmarshal json failed", "data", data)
			return nil, err
		}

		return &signedTx, nil
	}
	return nil, nil
}

//rpc call
func rpc(method string, cmd interface{}) ([]byte, error) {
	//param
	s, _ := rpcjson.MarshalCmd(1, method, cmd)
	//log.Info(method, "req", string(s))

	//response
	rawRet, err := httpclient.SendPostRequest(s, httpConfig)
	if err != nil {
		log.Error(method, "error", err)
		return nil, err
	}

	//log.Info(method, "rsp", string(rawRet))

	return rawRet, nil
}
