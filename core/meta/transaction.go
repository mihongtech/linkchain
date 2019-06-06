package meta

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/serialize"
	"github.com/mihongtech/linkchain/protobuf"
)

type Transaction struct {
	Data []byte
	txid TxID
}

func (t *Transaction) Serialize() serialize.SerializeStream {
	return &protobuf.Transaction{Data: t.Data}
}

func (t *Transaction) Deserialize(s serialize.SerializeStream) error {
	protoTransaction := s.(*protobuf.Transaction)
	buffer := protoTransaction.Data
	t.Data = make([]byte, 0)
	copy(t.Data, buffer)
	return nil
}

func (t *Transaction) String() string {
	return hex.EncodeToString(t.Data)
}

func (t Transaction) GetTxID() *TxID {
	if t.txid.IsEmpty() {
		t.txid = math.DoubleHashH(t.Data)
	}

	return &t.txid
}

type Transactions struct {
	Txs []Transaction
}

func NewTransactions(txs ...Transaction) *Transactions {
	ntxs := make([]Transaction, 0)
	for _, tx := range txs {
		ntxs = append(ntxs, tx)
	}
	return &Transactions{Txs: ntxs}
}

func (ts *Transactions) Serialize() serialize.SerializeStream {
	txs := make([]*protobuf.Transaction, 0)
	for i, _ := range ts.Txs {
		txs = append(txs, ts.Txs[i].Serialize().(*protobuf.Transaction))
	}
	return &protobuf.Transactions{Txs: txs}
}

func (ts *Transactions) Deserialize(s serialize.SerializeStream) error {
	protoTransactions := s.(*protobuf.Transactions)
	ts.Txs = make([]Transaction, 0)
	for i, _ := range protoTransactions.Txs {
		tx := Transaction{}
		if err := tx.Deserialize(protoTransactions.Txs[i]); err != nil {
			return err
		}
		ts.Txs = append(ts.Txs, tx)
	}
	return nil
}

func (ts *Transactions) String() string {
	data, err := json.Marshal(ts)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (ts *Transactions) SetTx(newTXs ...Transaction) error {
	for _, tx := range newTXs {
		ts.Txs = append(ts.Txs, tx)
	}

	return nil
}

func (ts *Transactions) GetTx(id TxID) (*Transaction, error) {
	for i, t := range ts.Txs {
		if t.GetTxID().IsEqual(&id) {
			return &ts.Txs[i], nil
		}
	}
	return nil, errors.New("can not fin tx in block")
}

func TxDifference(a, b []Transaction) (keep []Transaction) {
	keep = make([]Transaction, 0, len(a))

	remove := make(map[TxID]struct{})
	for _, tx := range b {
		remove[*tx.GetTxID()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[*tx.GetTxID()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}
