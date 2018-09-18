package meta

import (
	"strconv"
	"github.com/linkchain/meta"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/poa/meta/protobuf"
	"github.com/golang/protobuf/proto"
)

type POAAmount struct {
	Value int32
}

func NewPOAAmout(value int32) POAAmount {
	return POAAmount{Value:value}
}

func (a *POAAmount) GetInt() int  {
	return int(a.Value)
}

func (a *POAAmount) GetFloat() float32  {
	return float32(a.Value)
}

func (a *POAAmount) GetString() string  {
	return strconv.Itoa(a.GetInt())
}

func (a *POAAmount)IsLessThan(otherAmount meta.IAmount) bool  {
	return a.GetFloat() < otherAmount.GetFloat()
}

func (a *POAAmount)Subtraction(otherAmount meta.IAmount)  {
	a.Value = int32(a.GetInt() - otherAmount.GetInt())
}

func (a *POAAmount)Addition(otherAmount meta.IAmount)  {
	a.Value = int32(a.GetInt() + otherAmount.GetInt())
}

//Serialize/Deserialize
func (a *POAAmount) Serialize()(serialize.SerializeStream){
	amount := protobuf.POAAmount{
		Value:proto.Int32(a.Value),
	}
	return &amount
}

func (a *POAAmount) Deserialize(s serialize.SerializeStream){
	data := s.(*protobuf.POAAmount)
	a.Value = *data.Value
}


