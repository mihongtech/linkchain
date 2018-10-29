package meta

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/protobuf"
)

type TransactionPeer struct {
	AccountID AccountID
	Extra     []byte
}

func NewTransactionPeer(id AccountID, extra []byte) *TransactionPeer {
	return &TransactionPeer{AccountID: id, Extra: extra}
}

//Serialize/Deserialize
func (txpeer *TransactionPeer) Serialize() serialize.SerializeStream {
	accountID := txpeer.AccountID.Serialize().(*protobuf.AccountID)
	peer := protobuf.TransactionPeer{
		AccountID: accountID,
		Extra:     proto.NewBuffer(txpeer.Extra).Bytes(),
	}
	return &peer
}

func (txpeer *TransactionPeer) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.TransactionPeer)
	err := txpeer.AccountID.Deserialize(data.AccountID)
	if err != nil {
		return err
	}
	txpeer.Extra = data.Extra
	return nil
}

func (txpeer *TransactionPeer) GetID() account.IAccountID {
	return &txpeer.AccountID
}

type FromSign struct {
	Code []byte
}

type Transaction struct {
	// Version of the Transaction.  This is not the same as the Blocks version.
	Version uint32

	From TransactionPeer

	To TransactionPeer

	Amount Amount

	Time time.Time
	// Data used to extenion the block.

	Nounce uint32

	Extra []byte

	Signs FromSign

	txid math.Hash
}

func NewTransaction(version uint32, from TransactionPeer, to TransactionPeer, amount Amount, time time.Time, nounce uint32, extra []byte, signs FromSign) *Transaction {
	return &Transaction{
		Version: version,
		From:    from,
		To:      to,
		Amount:  amount,
		Time:    time,
		Nounce:  nounce,
		Extra:   extra,
		Signs:   signs,
	}
}

func (tx *Transaction) GetTxID() meta.DataID {
	if tx.txid.IsEmpty() {
		//TODO Deserialize
		tx.Deserialize(tx.Serialize())
	}
	return &tx.txid
}

func (tx *Transaction) SetFrom(from tx.ITxPeer) {
	txin := *from.(*TransactionPeer)
	tx.From = txin
}

func (tx *Transaction) SetTo(to tx.ITxPeer) {
	txout := *to.(*TransactionPeer)
	tx.To = txout
}

func (tx *Transaction) ChangeFromTo() tx.ITx {
	temp := tx.From
	tx.From = tx.To
	tx.To = temp
	tx.Nounce -= 1
	return tx
}

func (tx *Transaction) SetAmount(iAmount meta.IAmount) {
	amount := *iAmount.(*Amount)
	tx.Amount = amount
}

func (tx *Transaction) SetNounce(nounce uint32) {
	tx.Nounce = nounce
}

func (tx *Transaction) GetFrom() tx.ITxPeer {
	return &tx.From
}

func (tx *Transaction) GetTo() tx.ITxPeer {
	return &tx.To
}

func (tx *Transaction) GetAmount() meta.IAmount {
	return &tx.Amount
}

func (tx *Transaction) GetNounce() uint32 {
	return tx.Nounce
}

func (tx *Transaction) Sign() (math.ISignature, error) {
	//TODO sign need to finish
	return nil, nil
}

func (tx *Transaction) GetSignature() math.ISignature {
	return nil
}

func (tx *Transaction) SetSignature(code []byte) {
	tx.Signs = FromSign{Code: code}
}

func (tx *Transaction) Verify() error {
	signature, err := btcec.ParseSignature(tx.Signs.Code, btcec.S256())
	if err != nil {
		log.Error("Transaction", "VerifySign", err)
		return err
	}

	pk, err := btcec.ParsePubKey(tx.From.AccountID.ID, btcec.S256())
	if err != nil {
		return errors.New("Transaction VerifySign ParsePubKey is error")
	}

	verified := signature.Verify(tx.GetTxID().CloneBytes(), pk)
	if verified {
		return nil
	} else {
		return errors.New("Transaction VerifySign failed: Error Sign")
	}
}

//Serialize/Deserialize
func (tx *Transaction) Serialize() serialize.SerializeStream {
	from := tx.From.Serialize().(*protobuf.TransactionPeer)
	to := tx.To.Serialize().(*protobuf.TransactionPeer)
	amount := tx.Amount.Serialize().(*protobuf.Amount)

	t := protobuf.Transaction{
		Version: proto.Uint32(tx.Version),
		From:    from,
		To:      to,
		Time:    proto.Int64(tx.Time.Unix()),
		Amount:  amount,
		Nounce:  proto.Uint32(tx.Nounce),
		Extra:   proto.NewBuffer(tx.Extra).Bytes(),
		Sign:    proto.NewBuffer(tx.Signs.Code).Bytes(),
	}
	return &t
}

func (tx *Transaction) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.Transaction)
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
	tx.Extra = data.Extra
	tx.Signs = FromSign{Code: data.Sign}

	t := protobuf.Transaction{
		Version: data.Version,
		From:    data.From,
		To:      data.To,
		Time:    data.Time,
		Nounce:  data.Nounce,
		Amount:  data.Amount,
		Extra:   data.Extra,
	}
	tx.txid = math.MakeHash(&t)
	return nil
}

func (tx *Transaction) String() string {
	data, err := json.Marshal(tx)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
