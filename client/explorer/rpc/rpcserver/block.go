package rpcserver

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/linkchain/client/explorer/rpc/rpcobject"
)

func getBestBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	method := "getBestBlock"

	//call
	out, err := rpc(method, nil)
	block := &rpcobject.BlockRSP{}
	json.Unmarshal(out, block)
	return block, err
}

func getBlockByHeight(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.GetBlockByHeightCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	height := c.Height

	method := "getBlockByHeight"

	//call
	out, err := rpc(method, &rpcobject.GetBlockByHeightCmd{height})
	block := &rpcobject.BlockRSP{}
	json.Unmarshal(out, block)
	return block, err
}

func getBlockByHash(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.GetBlockByHashCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	hash := c.Hash

	method := "getBlockByHash"

	//call
	out, err := rpc(method, &rpcobject.GetBlockByHashCmd{hash})
	block := &rpcobject.BlockRSP{}
	json.Unmarshal(out, block)
	return block, err
}
