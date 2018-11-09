package meta

import "github.com/linkchain/common/math"

/**/

//TODO need give up
/*type DataID interface {
	GetString() string
	IsEqual(id DataID) bool
	IsEmpty() bool
	CloneBytes() []byte
	SetBytes(newHash []byte) error

	serialize.ISerialize
}*/

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

func MakeTreeID(b []byte) (*TreeID, error) {
	return math.NewHash(b)
}

type IAccountID interface {
	IsEqual(other IAccountID) bool
	String() string
}
