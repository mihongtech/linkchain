package meta

import "github.com/linkchain/common/serialize"

type IAmount interface {
	GetInt() int
	GetFloat() float32
	GetString() string

	IsLessThan(otherAmount IAmount) bool

	Subtraction(otherAmount IAmount) IAmount
	Addition(otherAmount IAmount) IAmount
}

type DataID interface {
	GetString() string
	IsEqual(id DataID) bool
	IsEmpty() bool
	CloneBytes() []byte
	SetBytes(newHash []byte) error

	serialize.ISerialize
}
