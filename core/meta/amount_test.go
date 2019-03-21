package meta

import (
	"github.com/linkchain/unittest"
	"testing"
)

func TestNewAmount(t *testing.T) {
	amount := NewAmount(0)
	unittest.Equal(t, amount.GetInt64(), int64(0))

	amount1 := NewAmount(10)
	unittest.Equal(t, amount1.GetInt64(), int64(10))

	amount2 := NewAmount(-10)
	unittest.Equal(t, amount2.GetInt64(), int64(-10))
}
