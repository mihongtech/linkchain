package meta

import (
	"time"
	"errors"
	"encoding/json"

	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/poa/meta/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/btcec"
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

func (txpeer *POATransactionPeer) GetID() account.IAccountID {
	return &txpeer.AccountID
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

	txid math.Hash
}

func (tx *POATransaction) GetTxID() meta.DataID  {
	if tx.txid.IsEmpty() {
		tx.Deserialize(tx.Serialize())
	}
	return &tx.txid
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

func (tx *POATransaction) SetSignature(code []byte)  {
	tx.Signs = FromSign{Code:code}
}

func (tx *POATransaction) Verify()(error)  {
	signature, err := btcec.ParseSignature(tx.Signs.Code, btcec.S256())
	if err != nil {
		log.Error("POATransaction","VerifySign",err)
		return err
	}
	verified := signature.Verify(tx.txid.CloneBytes(), &tx.From.AccountID.ID)
	if verified {
		return nil
	}else {
		return errors.New("POATransaction VerifySign failed: Error Sign")
	}
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
		Sign:proto.NewBuffer(tx.Signs.Code).Bytes(),
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
	tx.Signs = FromSign{Code:data.Sign}

	t := protobuf.POATransaction{
		Version:data.Version,
		From:data.From,
		To:data.To,
		Time:data.Time,
		Amount:data.Amount,
		Extra:data.Extra,
	}
	tx.txid = math.MakeHash(&t)
}

func (tx *POATransaction) ToString()(string) {
	data, err := json.Marshal(tx);
	if  err != nil {
		return err.Error()
	}
	return string(data)
}
