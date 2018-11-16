package meta

import "github.com/linkchain/common/math"

type Amount = math.BigInt

func NewAmount(x int64) *Amount {
	return math.NewBigInt(x)
}