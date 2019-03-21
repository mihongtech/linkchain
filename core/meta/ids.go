package meta

import (
	"github.com/linkchain/common/math"
)

type TxID = math.Hash

func MakeTxID(b []byte) *TxID {
	hash := math.DoubleHashH(b)
	return &hash
}

type BlockID = math.Hash

func MakeBlockId(b []byte) *BlockID {
	hash := math.DoubleHashH(b)
	return &hash
}

type TreeID = math.Hash
