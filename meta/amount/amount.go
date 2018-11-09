package amount

import "github.com/linkchain/common/math"

/*type IAmount interface {
	GetInt() int
	GetFloat() float32
	GetString() string

	IsLessThan(otherAmount IAmount) bool

	Subtraction(otherAmount IAmount) IAmount
	Addition(otherAmount IAmount) IAmount
	Reverse() IAmount
}*/

type Amount = math.BigInt

func NewAmount(x int64) *Amount {
	return math.NewBigInt(x)
}
