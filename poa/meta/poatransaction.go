package meta

import (
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/serialize"
	"crypto/sha256"
	"encoding/json"
	"time"
	"github.com/linkchain/poa/meta/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/util/log"
)

type POATransactionPeer struct {
	AccountID POAAccountID
	Extra []byte
}

func GetPOATransactionPeer(iaccount account.IAccount, extra []byte) POATransactionPeer {
	id := *iaccount.GetAccountID().(*POAAccountID)
	return POATransactionPeer{AccountID:id,Extra:extra}
}

//Serialize/Deserialize
func (txpeer *POATransactionPeer) Serialize()(serialize.SerializeStream){
	accountID := txpeer.AccountID.Serialize().(*protobuf.POAAccountID)
	peer := protobuf.POATransactionPeer{
		AccountID:accountID,
		Extra:proto.NewBuffer(txpeer.Extra).Bytes(),
	}
	return &peer
}

func (txpeer *POATransactionPeer) Deserialize(s serialize.SerializeStream){
	data := *s.(*protobuf.POATransactionPeer)
	txpeer.AccountID.Deserialize(data.AccountID)
	txpeer.Extra = data.Extra
}


type FromSign struct {
	Code []byte
}

type POATransaction struct {
	// Version of the Transaction.  This is not the same as the Blocks version.
	Version uint32

	From POATransactionPeer

	To POATransactionPeer

	Amount POAAmount

	Time time.Time
	// Extra used to extenion the block.
	Extra []byte

	Signs FromSign
}

func (tx *POATransaction) GetTxID() tx.ITxID  {
	newTx := tx.Serialize().(*protobuf.POATransaction)
	buffer,err := proto.Marshal(newTx)
	if err != nil {
		log.Error("header marshaling error: ", err)
	}
	first := sha256.Sum256(buffer)
	return math.Hash(sha256.Sum256(first[:]))
}

func (tx *POATransaction) SetFrom(from tx.ITxPeer)  {
	txin := *from.(*POATransactionPeer)
	tx.From = txin
}

func (tx *POATransaction) SetTo(to tx.ITxPeer)  {
	txout := *to.(*POATransactionPeer)
	tx.To = txout
}

func (tx *POATransaction) ChangeFromTo() tx.ITx  {
	temp := tx.From
	tx.From = tx.To
	tx.To = temp
	return tx
}

func (tx *POATransaction) SetAmount(iAmount meta.IAmount)  {
	amount := *iAmount.(*POAAmount)
	tx.Amount = amount
}

func (tx *POATransaction) GetFrom() tx.ITxPeer  {
	return &tx.From
}

func (tx *POATransaction) GetTo() tx.ITxPeer  {
	return &tx.To
}

func (tx *POATransaction) GetAmount() meta.IAmount  {
	return &tx.Amount
}

func (tx *POATransaction) Sign()(math.ISignature, error)  {
	//TODO sign need to finish
	return nil,nil
}

func (tx *POATransaction) GetSignature()(math.ISignature)  {
	return nil
}

func (tx *POATransaction) Verify()(error)  {
	return nil
}

//Serialize/Deserialize
func (tx *POATransaction) Serialize()(serialize.SerializeStream){
	from := tx.From.Serialize().(*protobuf.POATransactionPeer)
	to := tx.To.Serialize().(*protobuf.POATransactionPeer)
	amount := tx.Amount.Serialize().(*protobuf.POAAmount)

	t := protobuf.POATransaction{
		Version:proto.Uint32(tx.Version),
		From:from,
		To:to,
		Time:proto.Int64(tx.Time.Unix()),
		Amount:amount,
		Extra:proto.NewBuffer(tx.Extra).Bytes(),
	}
	return &t
}

func (tx *POATransaction) Deserialize(s serialize.SerializeStream){
	data := *s.(*protobuf.POATransaction)
	tx.Version = *data.Version
	tx.From.Deserialize(data.From)
	tx.To.Deserialize(data.To)
	tx.Time = time.Unix(*data.Time,0)
	tx.Amount.Deserialize(data.Amount)
	tx.Extra = data.Extra
}

func (tx *POATransaction) ToString()(string) {
	data, err := json.Marshal(tx);
	if  err != nil {
		return err.Error()
	}
	return string(data)
}


