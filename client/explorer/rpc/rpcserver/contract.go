package rpcserver

import (
	"fmt"
	"reflect"

	"github.com/linkchain/client/explorer/rpc/rpcobject"
)

func publishContract(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.PublishContractCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	account := c.FromAccountId
	contract := c.Contract
	amount := c.Amount

	method := "publishContract"

	//call
	out, err := rpc(method, &rpcobject.PublishContractCmd{account, contract, amount, 1, 100000000})
	return out, err
}

func call(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.CallContractCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	account := c.FromAccountId
	contract := c.Contract
	callMethod := c.CallMethod

	method := "call"

	//call
	out, err := rpc(method, &rpcobject.CallCmd{account, contract, callMethod, 0})
	return out, err
}

func callContract(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.CallContractCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	account := c.FromAccountId
	contract := c.Contract
	callMethod := c.CallMethod
	amount := c.Amount

	method := "callContract"

	//call
	out, err := rpc(method, &rpcobject.CallContractCmd{account, contract, callMethod, amount, 1, 100000000})
	return out, err
}

func getTransactionReceipt(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.GetTransactionReceiptCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	hash := c.Hash

	method := "transactionReceipt"

	//call
	out, err := rpc(method, &rpcobject.GetTransactionReceiptCmd{hash})
	return out, err
}
