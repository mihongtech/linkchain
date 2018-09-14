package block

import (
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
)

type IBlock interface {
	//block content
	SetTx([]tx.ITx)(error)

	GetTxs() []tx.ITx

	GetHeight() uint32

	GetBlockID() meta.DataID

	GetPrevBlockID() meta.DataID
	//verifiy
	Verify()(error)

	IsGensis() bool

	//serialize
	serialize.ISerialize
}