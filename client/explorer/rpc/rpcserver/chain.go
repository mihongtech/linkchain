package rpcserver

func getBlockChainInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	method := "getBlockChainInfo"

	//call
	out, err := rpc(method, nil)
	return out, err
}
