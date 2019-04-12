package peer

import (
	_ "fmt"

	"github.com/mihongtech/linkchain/p2p/discover"
	"github.com/mihongtech/linkchain/p2p/message"
)

// Protocol represents a P2P subprotocol implementation.
type Protocol struct {
	// Name should contain the official protocol name,
	// often a three-letter word.
	Name string

	// Version should contain the version number of the protocol.
	Version uint64

	// Length should contain the number of message codes used
	// by the protocol.
	Length uint64

	// Run is called in a new groutine when the protocol has been
	// negotiated with a peer. It should read and write messages from
	// rw. The Payload for each message must be fully consumed.
	//
	// The peer connection is closed when Start returns. It should return
	// any protocol-level error (such as an I/O error) that is
	// encountered.
	Run func(peer *Peer, rw message.MsgReadWriter) error

	// NodeInfo is an optional helper method to retrieve protocol specific metadata
	// about the host discover.
	NodeInfo func() interface{}

	// PeerInfo is an optional helper method to retrieve protocol specific metadata
	// about a certain peer in the network. If an info retrieval function is set,
	// but returns nil, it is assumed that the protocol handshake is still running.
	PeerInfo func(id discover.NodeID) interface{}
}

func (p Protocol) Cap() message.Cap {
	return message.Cap{p.Name, p.Version}
}
