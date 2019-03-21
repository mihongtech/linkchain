package rpcserver

func getMineInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return GetMinerAPI(s).GetInfo(), nil
}

func startMine(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	go GetMinerAPI(s).StartMine()
	return nil, nil
}

func stopMine(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	go GetMinerAPI(s).StopMine()
	return nil, nil
}

func mine(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	block, err := GetMinerAPI(s).MineBlock()
	if err != nil {
		return nil, err
	}

	b := getBlockObject(block)
	return b, nil
}
