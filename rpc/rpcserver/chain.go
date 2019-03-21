package rpcserver

import (
	"github.com/linkchain/core/meta"
	"github.com/linkchain/rpc/rpcobject"
)

func getBlockChainInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	info := GetNodeAPI(s).GetBlockChainInfo()
	return &rpcobject.ChainRSP{
		Chains: info.(*meta.ChainInfo),
	}, nil
}
