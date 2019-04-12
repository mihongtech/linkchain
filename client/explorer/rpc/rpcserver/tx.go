package rpcserver

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mihongtech/linkchain/client/explorer/rpc/rpcobject"
)

func getTxByHash(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.GetTransactionByHashCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	hash := c.Hash

	method := "getTxByHash"

	//call
	out, err := rpc(method, &rpcobject.GetTransactionByHashCmd{hash})
	transaction := &rpcobject.TransactionWithIDRSP{}
	json.Unmarshal(out, transaction)
	return transaction, err
}
