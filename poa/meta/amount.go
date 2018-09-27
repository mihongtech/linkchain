package meta

import (
	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/protobuf"
	"strconv"
)

type Amount struct {
	Value int32
}

func NewAmout(value int32) *Amount {
	return &Amount{Value: value}
}

func (a *Amount) GetInt() int {
	return int(a.Value)
}

func (a *Amount) GetFloat() float32 {
	return float32(a.Value)
}

func (a *Amount) GetString() string {
	return strconv.Itoa(a.GetInt())
}

func (a *Amount) IsLessThan(otherAmount meta.IAmount) bool {
	return a.GetFloat() < otherAmount.GetFloat()
}

func (a *Amount) Subtraction(otherAmount meta.IAmount) meta.IAmount {
	a.Value = int32(a.GetInt() - otherAmount.GetInt())
	return a
}

func (a *Amount) Addition(otherAmount meta.IAmount) meta.IAmount {
	a.Value = int32(a.GetInt() + otherAmount.GetInt())
	return a
}

func (a *Amount) Reverse() meta.IAmount {
	a.Value = int32(0 - a.GetInt())
	return a
}

//Serialize/Deserialize
func (a *Amount) Serialize() serialize.SerializeStream {
	amount := protobuf.Amount{
		Value: proto.Int32(a.Value),
	}
	return &amount
}

func (a *Amount) Deserialize(s serialize.SerializeStream) {
	data := s.(*protobuf.Amount)
	a.Value = *data.Value
}
