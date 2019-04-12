package rpcserver

import (
	"fmt"
	"reflect"

	"github.com/mihongtech/linkchain/client/explorer/rpc/rpcobject"
)

func newAcount(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	method := "newAcount"

	//call
	out, err := rpc(method, nil)
	return out, err
}

func getAccountInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.SingleCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	accountID := c.Key

	method := "getAccountInfo"

	//call
	out, err := rpc(method, &rpcobject.SingleCmd{accountID})
	return out, err
}

func getWalletInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	method := "getWalletInfo"

	//call
	out, err := rpc(method, nil)
	return out, err
}

func importAccount(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.ImportAccountCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	signer := c.Signer

	method := "importAccount"

	//call
	out, err := rpc(method, rpcobject.ImportAccountCmd{signer})
	return out, err
}

func exportAccount(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.ExportAccountCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	accountId := c.AccountId

	method := "exportAccount"

	//call
	out, err := rpc(method, rpcobject.ExportAccountCmd{accountId})
	return out, err
}

func sendMoneyTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.SendToTxCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	fromAccountId := c.FromAccountId
	toAccountId := c.ToAccountId
	amount := c.Amount

	method := "sendMoneyTransaction"

	//call
	out, err := rpc(method, rpcobject.SendToTxCmd{fromAccountId, toAccountId, amount})
	return out, err
}
