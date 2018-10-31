package meta

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/tx"
)

type OldTransactionPeer struct {
	AccountID AccountID
	Extra     []byte
}

func NewOldTransactionPeer(id AccountID, extra []byte) *OldTransactionPeer {
	return &OldTransactionPeer{AccountID: id, Extra: extra}
}

//Serialize/Deserialize
func (txpeer *OldTransactionPeer) Serialize() serialize.SerializeStream {
	/*accountID := txpeer.AccountID.Serialize().(*protobuf.AccountID)
	peer := protobuf.OldTransactionPeer{
		AccountID: accountID,
		Extra:     proto.NewBuffer(txpeer.Extra).Bytes(),
	}
	return &peer*/
	return nil
}

func (txpeer *OldTransactionPeer) Deserialize(s serialize.SerializeStream) error {
	/*data := *s.(*protobuf.OldTransactionPeer)
	err := txpeer.AccountID.Deserialize(data.AccountID)
	if err != nil {
		return err
	}
	txpeer.Extra = data.Extra*/
	return nil
}

func (txpeer *OldTransactionPeer) GetID() account.IAccountID {
	return &txpeer.AccountID
}

type OldFromSign struct {
	Code []byte
}

type OldTransaction struct {
	// Version of the OldTransaction.  This is not the same as the Blocks version.
	Version uint32

	From OldTransactionPeer

	To OldTransactionPeer

	Amount Amount

	Time time.Time
	// Data used to extenion the block.

	Nounce uint32

	Data []byte

	Signs OldFromSign

	txid math.Hash
}

func NewOldTransaction(version uint32, from OldTransactionPeer, to OldTransactionPeer, amount Amount, time time.Time, nounce uint32, extra []byte, signs OldFromSign) *OldTransaction {
	return &OldTransaction{
		Version: version,
		From:    from,
		To:      to,
		Amount:  amount,
		Time:    time,
		Nounce:  nounce,
		Data:    extra,
		Signs:   signs,
	}
}

func (tx *OldTransaction) GetTxID() meta.DataID {
	if tx.txid.IsEmpty() {
		err := tx.Deserialize(tx.Serialize())
		if err != nil {
			log.Error("OldTransaction", "GetTxID() error", err)
			return nil
		}
	}
	return &tx.txid
}

func (tx *OldTransaction) SetFrom(from tx.ITxPeer) {
	txin := *from.(*OldTransactionPeer)
	tx.From = txin
}

func (tx *OldTransaction) SetTo(to tx.ITxPeer) {
	txout := *to.(*OldTransactionPeer)
	tx.To = txout
}

func (tx *OldTransaction) ChangeFromTo() tx.ITx {
	temp := tx.From
	tx.From = tx.To
	tx.To = temp
	tx.Nounce -= 1
	return tx
}

func (tx *OldTransaction) SetAmount(iAmount meta.IAmount) {
	amount := *iAmount.(*Amount)
	tx.Amount = amount
}

func (tx *OldTransaction) SetNounce(nounce uint32) {
	tx.Nounce = nounce
}

func (tx *OldTransaction) GetFrom() tx.ITxPeer {
	return &tx.From
}

func (tx *OldTransaction) GetTo() tx.ITxPeer {
	return &tx.To
}

func (tx *OldTransaction) GetAmount() meta.IAmount {
	return &tx.Amount
}

func (tx *OldTransaction) GetNounce() uint32 {
	return tx.Nounce
}

func (tx *OldTransaction) Sign() (math.ISignature, error) {
	//TODO sign need to finish
	return nil, nil
}

func (tx *OldTransaction) GetSignature() math.ISignature {
	return nil
}

func (tx *OldTransaction) SetSignature(code []byte) {
	tx.Signs = OldFromSign{Code: code}
}

func (tx *OldTransaction) Verify() error {
	signature, err := btcec.ParseSignature(tx.Signs.Code, btcec.S256())
	if err != nil {
		log.Error("OldTransaction", "VerifySign", err)
		return err
	}

	pk, err := btcec.ParsePubKey(tx.From.AccountID.ID, btcec.S256())
	if err != nil {
		return errors.New("OldTransaction VerifySign ParsePubKey is error")
	}

	verified := signature.Verify(tx.GetTxID().CloneBytes(), pk)
	if verified {
		return nil
	} else {
		return errors.New("OldTransaction VerifySign failed: Error Sign")
	}
}

//Serialize/Deserialize
func (tx *OldTransaction) Serialize() serialize.SerializeStream {
	/*from := tx.From.Serialize().(*protobuf.OldTransactionPeer)
	to := tx.To.Serialize().(*protobuf.OldTransactionPeer)
	amount := tx.Amount.Serialize().(*protobuf.Amount)

	t := protobuf.OldTransaction{
		Version: proto.Uint32(tx.Version),
		From:    from,
		To:      to,
		Time:    proto.Int64(tx.Time.Unix()),
		Amount:  amount,
		Nounce:  proto.Uint32(tx.Nounce),
		Extra:   proto.NewBuffer(tx.Data).Bytes(),
		Sign:    proto.NewBuffer(tx.Signs.Code).Bytes(),
	}
	return &t*/
	return nil
}

func (tx *OldTransaction) Deserialize(s serialize.SerializeStream) error {
	/*data := *s.(*protobuf.OldTransaction)
	tx.Version = *data.Version
	err := tx.From.Deserialize(data.From)
	if err != nil {
		return err
	}
	err = tx.To.Deserialize(data.To)
	if err != nil {
		return err
	}
	tx.Time = time.Unix(*data.Time, 0)
	tx.Nounce = *data.Nounce
	err = tx.Amount.Deserialize(data.Amount)
	if err != nil {
		return err
	}
	tx.Data = data.Extra
	tx.Signs = OldFromSign{Code: data.Sign}

	t := protobuf.OldTransaction{
		Version: data.Version,
		From:    data.From,
		To:      data.To,
		Time:    data.Time,
		Nounce:  data.Nounce,
		Amount:  data.Amount,
		Extra:   data.Extra,
	}
	tx.txid = math.MakeHash(&t)*/
	return nil
}

func (tx *OldTransaction) String() string {
	data, err := json.Marshal(tx)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
