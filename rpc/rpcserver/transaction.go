package rpcserver

import (
	"errors"
	"reflect"

	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/rpc/rpcobject"
)

func getTxByHash(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.GetTransactionByHashCmd)
	if !ok {
		log.Error("getTxByHash ", "Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	hash, err := math.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, err
	}

	transaction, _, _, _ := GetNodeAPI(s).GetTXByID(*hash)
	if transaction == nil {
		log.Error("getTxByHash ", "error", err)
		return nil, errors.New("getTxByHash failed")
	}

	return &rpcobject.TransactionWithIDRSP{transaction.GetTxID().GetString(), transaction}, nil
}
