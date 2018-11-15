package meta

import (
	"encoding/hex"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/math"
	"github.com/linkchain/meta/amount"
	"github.com/linkchain/poa/config"
	"github.com/linkchain/protobuf"
)

func Test_Serialize_tx(t *testing.T) {
	ex, _ := btcec.NewPrivateKey(btcec.S256())
	id := NewAccountId(ex.PubKey())
	utxos := make([]UTXO, 0)
	formAccount := NewAccount(*id, 0, utxos, 0, *id)
	toAccount := NewAccount(*id, 0, utxos, 0, *id)
	fromId := *formAccount.GetAccountID().(*AccountID)
	toId := *toAccount.GetAccountID().(*AccountID)

	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(fromId, tickets)
	fcs := make([]FromCoin, 0)
	fcs = append(fcs, *fc)
	tf := NewTransactionFrom(fcs)

	tc := NewToCoin(toId, amount.NewAmount(10))
	tcs := make([]ToCoin, 0)
	tcs = append(tcs, *tc)
	tt := NewTransactionTo(tcs)

	tx := NewTransaction(config.TransactionVersion, config.TransactionVersion, *tf, *tt, nil, nil)

	t.Log("createtx", "data", tx)

	s := tx.Serialize()

	buffer, err := proto.Marshal(s)
	if err != nil {
		t.Error("tx 序列化不通过 marshaling error", err)
	}
	t.Log("tx 序列化", "buffer->", hex.EncodeToString(buffer))
	t.Log("tx 序列化", "Txid hash->", tx.GetTxID().GetString())
}

func Test_DeSerialize_tx(t *testing.T) {
	txid, _ := math.NewHashFromStr("15e05dfc0f7d38be1215418dfcd8ed49c84c0668dd85fc1a10e0c653eb89c601")
	str := "080110011a4f0a4d0a230a2102ca039b16811c50b2bc015bf36271141742d6419c983fddcf838b3a4de68672f612260a220a20640fe47bb4898f552ef35ea64026fd7304960254269b9ac3dabcdd6cfc126e5e1000222a0a280a230a2102ca039b16811c50b2bc015bf36271141742d6419c983fddcf838b3a4de68672f612010a"
	buffer, _ := hex.DecodeString(str)
	tx := &protobuf.Transaction{}

	err := proto.Unmarshal(buffer, tx)
	if err != nil {
		t.Error("tx 反序列化不通过 unmarshaling error: ", err)
	}

	newTx := Transaction{}
	err = newTx.Deserialize(tx)
	if err != nil {
		t.Error("tx 反序列化不通过")
	}
	newTxHash := newTx.GetTxID()

	if txid.IsEqual(newTxHash) {
		t.Log("tx 反序列化通过")
	} else {
		t.Error("tx 反序列化不通过")
	}
}
