package meta

import (
	"encoding/json"
	"errors"
	"github.com/mihongtech/linkchain/common"
	"math/big"
)

// A Amount represents a signed multi-precision integer.
type Amount struct {
	bigint *big.Int
}

func (bi *Amount) GetBigInt() *big.Int {
	return bi.bigint
}

// NewBigInt allocates and returns a new Amount set to x.
func NewAmount(x int64) *Amount {
	return &Amount{big.NewInt(x)}
}

// GetBytes returns the absolute value of x as a big-endian byte slice.
func (bi *Amount) GetBytes() []byte {
	if bi.GetBigInt() == nil {
		return big.NewInt(0).Bytes()
	}

	return bi.bigint.Bytes()
}

// String returns the value of x as a formatted decimal string.
func (bi *Amount) String() string {
	return bi.bigint.String()
}

// GetInt64 returns the int64 representation of x. If x cannot be represented in
// an int64, the result is undefined.
func (bi *Amount) GetInt64() int64 {
	return bi.bigint.Int64()
}

// SetBytes interprets buf as the bytes of a big-endian unsigned integer and sets
// the big int to that value.
func (bi *Amount) SetBytes(buf []byte) {
	bi.bigint.SetBytes(common.CopyBytes(buf))
}

// SetInt64 sets the big int to x.
func (bi *Amount) SetInt64(x int64) {
	if bi.bigint == nil {
		bi.bigint = big.NewInt(0)
	}
	bi.bigint.SetInt64(x)
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
//
func (bi *Amount) Sign() int {
	return bi.bigint.Sign()
}

// SetString sets the big int to x.
//
// The string prefix determines the actual conversion base. A prefix of "0x" or
// "0X" selects base 16; the "0" prefix selects base 8, and a "0b" or "0B" prefix
// selects base 2. Otherwise the selected base is 10.
func (bi *Amount) SetString(x string, base int) {
	bi.bigint.SetString(x, base)
}

// Amounts represents a slice of big ints.
type Amounts struct{ bigints []*big.Int }

// Size returns the number of big ints in the slice.
func (bi *Amounts) Size() int {
	return len(bi.bigints)
}

// Get returns the bigint at the given index from the slice.
func (bi *Amounts) Get(index int) (bigint *Amount, _ error) {
	if index < 0 || index >= len(bi.bigints) {
		return nil, errors.New("index out of bounds")
	}
	return &Amount{bi.bigints[index]}, nil
}

// Set sets the big int at the given index in the slice.
func (bi *Amounts) Set(index int, bigint *Amount) error {
	if index < 0 || index >= len(bi.bigints) {
		return errors.New("index out of bounds")
	}
	bi.bigints[index] = bigint.bigint
	return nil
}

// GetString returns the value of x as a formatted string in some number base.
func (bi *Amount) GetString(base int) string {
	return bi.bigint.Text(base)
}

func (a *Amount) IsLessThan(otherAmount Amount) bool {
	return a.GetInt64() < otherAmount.GetInt64()
}

func (a *Amount) Subtraction(otherAmount Amount) *Amount {
	a.SetInt64(a.GetInt64() - otherAmount.GetInt64())
	return a
}

func (a *Amount) Addition(otherAmount Amount) *Amount {
	a.SetInt64(a.GetInt64() + otherAmount.GetInt64())
	return a
}

func (a *Amount) Reverse() *Amount {
	a.SetInt64(0 - a.GetInt64())
	return a
}

//Json Hash convert to Hex
func (a Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.GetInt64())
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	value := int64(0)
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	a.SetInt64(value)
	return nil
}
