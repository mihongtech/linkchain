package block

import (
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/tx"
)

type NewMinedBlockEvent struct{ Block IBlock }

type IBlock interface {
	//block content
	SetTx(...tx.ITx) error

	SetSign(signature math.ISignature)

	GetTxs() []tx.ITx

	GetHeight() uint32

	GetBlockID() *meta.BlockID

	GetPrevBlockID() *meta.BlockID

	GetMerkleRoot() *meta.TreeID

	CalculateTxTreeRoot() meta.TreeID

	//verifiy
	Verify(string) error

	IsGensis() bool

	GetBlockInfo() string

	//serialize
	serialize.ISerialize
}
