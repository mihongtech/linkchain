package rpcserver

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/math"
	"github.com/linkchain/core/meta"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/rpc/rpcobject"
)

func getBlockByHeight(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.GetBlockByHeightCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	height := c.Height

	if uint32(height) > GetNodeAPI(s).GetBestBlock().GetHeight() || height < 0 {
		log.Error("getblockbyheight ", "error", "height is out of range", "best", GetNodeAPI(s).GetBestBlock().GetHeight())
		return nil, nil
	}

	// get block
	block, err := GetNodeAPI(s).GetBlockByHeight(uint32(height))
	if err != nil {
		log.Error("getblockbyheight ", "error", err)
		return nil, err
	}

	b := getBlockObject(block)
	return b, nil
}

func getBlockByHash(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.GetBlockByHashCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	hash, err := math.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, err
	}

	block, err := GetNodeAPI(s).GetBlockByID(*hash)
	if err != nil {
		log.Error("getblockbyhash ", "error", err)
		return nil, err
	}

	b := getBlockObject(block)
	return b, nil
}

func getBlockObject(block *meta.Block) *rpcobject.BlockRSP {

	buffer, _ := proto.Marshal(block.Serialize())
	txids := make([]string, 0)
	for i := range block.TXs {
		txids = append(txids, block.TXs[i].GetTxID().String())
	}

	return &rpcobject.BlockRSP{
		block.GetHeight(),
		block.GetBlockID().String(),
		&block.Header,
		txids,
		hex.EncodeToString(buffer),
	}
}
