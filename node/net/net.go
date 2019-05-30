package net

import (
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/node/net/p2p/discover"
	"github.com/mihongtech/linkchain/node/net/p2p/peer"
)

type Net interface {
	core.Service
	Self() *discover.Node
	AddPeer(node *discover.Node)
	Peers() []*peer.Peer
	RemovePeer(node *discover.Node)
}
