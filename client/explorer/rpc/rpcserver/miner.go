package rpcserver

func getMineInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	method := "getMineInfo"

	//call
	out, err := rpc(method, nil)
	return out, err
}

func startMine(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	method := "startMine"

	//call
	out, err := rpc(method, nil)
	return out, err
}

func stopMine(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	method := "stopMine"

	//call
	out, err := rpc(method, nil)
	return out, err
}

func mine(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	method := "mine"

	//call
	out, err := rpc(method, nil)
	return out, err
}
