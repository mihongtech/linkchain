package rpcserver

import (
	"fmt"
	"reflect"

	"github.com/linkchain/client/explorer/rpc/rpcobject"
)

func addPeer(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.PeerCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	peer := c.Peer

	method := "addPeer"

	//call
	out, err := rpc(method, &rpcobject.PeerCmd{peer})
	return out, err
}

func removePeer(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.PeerCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	peer := c.Peer

	method := "removePeer"

	//call
	out, err := rpc(method, &rpcobject.PeerCmd{peer})
	return out, err
}

func listPeer(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	method := "listPeer"

	//call
	out, err := rpc(method, nil)
	return out, err
}

func selfPeer(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	method := "selfPeer"

	//call
	out, err := rpc(method, nil)
	return out, err
}
