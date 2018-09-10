package meta

import (
	"strconv"
	"github.com/linkchain/meta"
)

type POAAmount struct {
	Value int
}

func (a *POAAmount) GetInt() int  {
	return a.Value
}

func (a *POAAmount) GetFloat() float32  {
	return float32(a.Value)
}

func (a *POAAmount) GetString() string  {
	return strconv.Itoa(a.Value)
}

func (a *POAAmount)IsLessThan(otherAmount meta.IAmount) bool  {
	return a.GetFloat() < otherAmount.GetFloat()
}

func (a *POAAmount)Subtraction(otherAmount meta.IAmount)  {
	a.Value = a.GetInt() - otherAmount.GetInt()
}

func (a *POAAmount)Addition(otherAmount meta.IAmount)  {
	a.Value = a.GetInt() + otherAmount.GetInt()
}


