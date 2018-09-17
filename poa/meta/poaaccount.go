package meta

import (
	"encoding/json"

	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/poa/meta/protobuf"
	"github.com/linkchain/common/btcec"
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/util/log"
)


type POAAccountID struct {
	ID btcec.PublicKey
}

func (id *POAAccountID) GetString() string  {
	return hex.EncodeToString(id.ID.SerializeUncompressed())
}

//Serialize/Deserialize
func (a *POAAccountID)Serialize()(serialize.SerializeStream){
	accountId := protobuf.POAAccountID{
		Id:proto.NewBuffer(a.ID.SerializeUncompressed()).Bytes(),
	}
	return &accountId
}

func (a *POAAccountID)Deserialize(s serialize.SerializeStream){
	data := s.(*protobuf.POAAccountID)
	pk,err := btcec.ParsePubKey(data.Id, btcec.S256())
	if err != nil {
		log.Error("POAAccountID","Deserialize failed",err)
		return
	}
	a.ID = *pk
}

func NewAccountId(pkBytes []byte) account.IAccountID {
	pk,err := btcec.ParsePubKey(pkBytes, btcec.S256())
	if err != nil {
		log.Error("POAAccountID","Deserialize failed",err)
		return nil
	}
	return &POAAccountID{ID:*pk}
}

type POAAccount struct {
	AccountID POAAccountID
	Value POAAmount
}

func (a *POAAccount) ChangeAmount(amount meta.IAmount) meta.IAmount{
	a.Value = *amount.(*POAAmount)
	return &a.Value
}

func (a *POAAccount) GetAmount() meta.IAmount{
	return &(a.Value)
}

func (a *POAAccount) GetAccountID() account.IAccountID{
	return &a.AccountID
}


//Serialize/Deserialize
func (a *POAAccount)Serialize()(serialize.SerializeStream){
	return nil
}

func (a *POAAccount)Deserialize(s serialize.SerializeStream){
}

func (a *POAAccount) ToString()(string) {
	data, err := json.Marshal(a);
	if  err != nil {
		return err.Error()
	}
	return string(data)
}


