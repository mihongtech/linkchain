package meta

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/protobuf"
)

type AccountID struct {
	ID []byte
}

func (id *AccountID) String() string {
	return hex.EncodeToString(id.ID)
}

//Serialize/Deserialize
func (a *AccountID) Serialize() serialize.SerializeStream {
	accountId := protobuf.AccountID{
		Id: proto.NewBuffer(a.ID).Bytes(),
	}
	return &accountId
}

func (a *AccountID) Deserialize(s serialize.SerializeStream) error {
	data := s.(*protobuf.AccountID)
	pk, err := btcec.ParsePubKey(data.Id, btcec.S256())
	if err != nil {
		log.Error("AccountID", "Deserialize failed", err)
		return err
	}
	a.ID = pk.SerializeCompressed()
	return nil
}

func NewAccountId(id *btcec.PublicKey) *AccountID {
	return &AccountID{ID: id.SerializeCompressed()}
}

func CreateAccountIdByPubKey(pubKey string) (*AccountID, error) {
	pkBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	pk, err := btcec.ParsePubKey(pkBytes, btcec.S256())
	if err != nil {
		return nil, err
	}
	return NewAccountId(pk), nil
}

func CreateAccountIdByPrivKey(privKey string) (*AccountID, error) {
	priv, err := hex.DecodeString(privKey)
	if err != nil {
		return nil, err
	}
	_, pk := btcec.PrivKeyFromBytes(btcec.S256(), priv)
	if err != nil {
		return nil, err
	}
	return NewAccountId(pk), nil
}

type Account struct {
	AccountID AccountID
	Value     Amount
	Nounce    uint32
}

func NewAccount(id AccountID, value Amount, nounce uint32) *Account {
	return &Account{AccountID: id, Value: value, Nounce: nounce}
}

func (a *Account) ChangeAmount(amount meta.IAmount) meta.IAmount {
	a.Value = *amount.(*Amount)
	return &a.Value
}

func (a *Account) GetAmount() meta.IAmount {
	return &(a.Value)
}

func (a *Account) GetAccountID() account.IAccountID {
	return &a.AccountID
}

func (a *Account) GetNounce() uint32 {
	return a.Nounce
}

func (a *Account) SetNounce(nounce uint32) error {
	if a.CheckNounce(nounce) {
		a.Nounce = nounce
		return nil
	}
	return errors.New("IAccount nounce is error")
}

func (a *Account) CheckNounce(nounce uint32) bool {
	return nounce-a.Nounce == 1
}

//Serialize/Deserialize
func (a *Account) Serialize() serialize.SerializeStream {
	return nil
}

func (a *Account) Deserialize(s serialize.SerializeStream) error {
	return nil
}

func (id *Account) String() string {
	data, err := json.Marshal(id)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
