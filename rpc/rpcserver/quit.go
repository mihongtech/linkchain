package rpcserver

func shutdown(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.Stop(), nil
}
