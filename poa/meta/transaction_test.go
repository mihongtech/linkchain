package meta

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/math"
	"github.com/linkchain/protobuf"
)

func Test_Serialize_tx(t *testing.T) {
	fromAddress := math.Hash(sha256.Sum256([]byte("lf")))
	toAddress := math.Hash(sha256.Sum256([]byte("lc")))
	formAccount := &Account{AccountID: AccountID{ID: fromAddress.CloneBytes()}}
	toAccount := &Account{AccountID: AccountID{ID: toAddress.CloneBytes()}}
	amount := Amount{Value: 10}
	tx := Transaction{Version: 0,
		From:   *NewTransactionPeer(formAccount.AccountID, nil),
		To:     *NewTransactionPeer(toAccount.AccountID, nil),
		Amount: amount,
		Time:   time.Now()}

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
	txid, _ := math.NewHashFromStr("6208c0bca642dfadb90fd10f949a5dd0a5d1afc75f306fd3f48c68c91db2748a")
	str := "080012240a220a2039e741eddb03e3118da619625b9200e131832dc8cc8542e2198b1e00e02c48a31a240a220a20ef07b359570add31929a5422d400b16c7c84e35644cb2e84b142f0710973e2a82202080a28a58af0dd053000"
	buffer, _ := hex.DecodeString(str)
	tx := &protobuf.Transaction{}

	err := proto.Unmarshal(buffer, tx)
	if err != nil {
		t.Error("tx 反序列化不通过 unmarshaling error: ", err)
	}

	newTx := Transaction{}
	newTx.Deserialize(tx)
	newTxHash := newTx.GetTxID()

	if txid.IsEqual(newTxHash) {
		t.Log("tx 反序列化通过")
	} else {
		t.Error("tx 反序列化不通过")
	}
}
