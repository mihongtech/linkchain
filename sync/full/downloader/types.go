package downloader

import (
	"fmt"
	"github.com/linkchain/meta/block"
)

// peerDropFn is a callback type for dropping a peer detected as malicious.
type peerDropFn func(id string)

// dataPack is a data message returned by a peer for some query.
type dataPack interface {
	PeerId() string
	Items() int
	Stats() string
}

// headerPack is a batch of block headers returned by a peer.
type blockPack struct {
	peerId string
	blocks []block.IBlock
}

func (p *blockPack) PeerId() string { return p.peerId }
func (p *blockPack) Items() int     { return len(p.blocks) }
func (p *blockPack) Stats() string  { return fmt.Sprintf("%d", len(p.blocks)) }
