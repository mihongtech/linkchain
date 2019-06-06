package meta

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"github.com/mihongtech/linkchain/common/btcec"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/serialize"
	"github.com/mihongtech/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
)

const AddressLength = 20

type Address [AddressLength]byte

func CreateAddress(b []byte) Address {
	hash := math.HashB(b)
	return BytesToAddress(hash[12:])
}

func BytesToAddress(b []byte) Address {
	var a Address
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)

	return a
}

func HexToAddress(str string) (Address, error) {
	if len(str) > 1 {
		if str[0:2] == "0x" || str[0:2] == "0X" {
			str = str[2:]
		}
	}
	if len(str)%2 == 1 {
		str = "0" + str
	}

	data, err := hex.DecodeString(str)
	if err != nil {
		return Address{}, err
	}
	return BytesToAddress(data), nil
}

func (a Address) String() string {
	return hex.EncodeToString(a[:])
}

func (a Address) IsEqual(other Address) bool {
	return strings.Compare(a.String(), other.String()) == 0
}

func (a *Address) IsEmpty() bool {
	isEmpty := true
	l := len(a)
	for i := 0; i < l; i++ {
		if a[i] != 0 {
			isEmpty = false
			break
		}
	}
	return isEmpty
}

//Serialize to proto
func (a *Address) Serialize() serialize.SerializeStream {
	address := protobuf.Address{
		Data: proto.NewBuffer(a[:]).Bytes(),
	}
	return &address
}

//Deserialize from proto
func (a *Address) Deserialize(s serialize.SerializeStream) error {
	protoAddress := s.(*protobuf.Address)

	return a.SetBytes(protoAddress.Data)
}

func (a *Address) SetBytes(b []byte) error {
	if len(b) > AddressLength {
		return errors.New("byte's len more than max account length")
	}
	copy(a[:], b)
	return nil
}

func (a Address) CloneBytes() []byte {
	return a[:]
}

// Big converts an address to a big integer.
func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

// BigToAddress returns Address with byte values of b.
// If b is larger than len(h), b will be cropped from the left.
func BigToAddress(b *big.Int) Address { return BytesToAddress(b.Bytes()) }

//Json Hash convert to Hex
func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Address) UnmarshalJSON(data []byte) error {
	var str = ""
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	account, err := HexToAddress(str)
	if err != nil {
		return err
	}
	return a.SetBytes(account.CloneBytes())
}

func NewAddress(pubkey *btcec.PublicKey) *Address {
	// TODO: maybe use bitcion account generate function
	id := BytesToAddress(math.HashB(pubkey.SerializeCompressed())[12:])
	return &id
}

func NewAddressFromStr(str string) (*Address, error) {
	buff, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	id := BytesToAddress(buff)
	return &id, nil
}
