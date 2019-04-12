package rpcserver

import (
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/rpc/rpcobject"
)

func getBlockChainInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	info := GetNodeAPI(s).GetBlockChainInfo()
	return &rpcobject.ChainRSP{
		Chains: info.(*meta.ChainInfo),
	}, nil
}
