package meta

import (
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"encoding/json"
	"github.com/linkchain/poa/meta/protobuf"
)


type POAAccountID struct {
	ID math.Hash
}

func (id *POAAccountID) GetString() string  {
	return id.ID.GetString()
}

//Serialize/Deserialize
func (a *POAAccountID)Serialize()(serialize.SerializeStream){
	id := a.ID.Serialize().(*protobuf.Hash)
	accountId := protobuf.POAAccountID{
		Id:id,
	}
	return &accountId
}

func (a *POAAccountID)Deserialize(s serialize.SerializeStream){
	data := s.(*protobuf.POAAccountID)
	a.ID.Deserialize(data.Id)
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


