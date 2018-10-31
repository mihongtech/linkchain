package meta

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/math"
	"github.com/linkchain/poa/config"
	"github.com/linkchain/protobuf"
)

func Test_Serialize_tx(t *testing.T) {
	ex, _ := btcec.NewPrivateKey(btcec.S256())
	formAccount := NewAccount(*NewAccountId(ex.PubKey()), *NewAmout(10), 0)
	toAccount := NewAccount(*NewAccountId(ex.PubKey()), *NewAmout(10), 0)
	amount := Amount{Value: 10}
	fromId := *formAccount.GetAccountID().(*AccountID)
	toId := *toAccount.GetAccountID().(*AccountID)
	fp := *NewTransactionPeer(fromId, nil)
	tp := *NewTransactionPeer(toId, nil)
	tx := NewTransaction(config.TransactionVersion, fp, tp, amount, time.Now(), (formAccount.GetNounce() + 1), nil, FromSign{})

	t.Log("createtx", "data", tx)

	s := tx.Serialize()

	buffer, err := proto.Marshal(s)
	if err != nil {
		t.Error("tx 序列化不通过 marshaling error", err)
	}
	t.Log("tx 序列化", "buffer->", hex.EncodeToString(buffer))
	t.Log("tx 序列化", "txid hash->", tx.GetTxID().GetString())
}

func Test_DeSerialize_tx(t *testing.T) {
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	str := "080112250a230a2102a2a01b4c92af013d703e76838a3d1482f0b93c177761cec9549921bf6ac51e781a250a230a2102a2a01b4c92af013d703e76838a3d1482f0b93c177761cec9549921bf6ac51e782202080a28afbbe0de053001"
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
