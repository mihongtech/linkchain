package meta

import (
	"testing"
	"time"
	"crypto/sha256"
	"encoding/hex"

	"github.com/linkchain/common/math"
	"github.com/golang/protobuf/proto"
	"github.com/linkchain/poa/meta/protobuf"
)

func Test_Serialize_1(t *testing.T) {
	txs := []POATransaction{}
	block := POABlock{
		Header: POABlockHeader{Version: 0, PrevBlock: math.Hash{}, MerkleRoot: math.Hash{}, Timestamp: time.Unix(1487780010, 0), Difficulty: 0x207fffff, Nonce: 0, Extra: nil, Height: 0},
		TXs:    txs,
	}

	t.Log("createblock","data",block)

	s := block.Serialize()

	buffer,err := proto.Marshal(s)
	if err != nil {
		t.Error("block 序列化不通过 marshaling error",err)
	}
	t.Log("block 序列化","buffer->",hex.EncodeToString(buffer))
	t.Log("block 序列化","block hash->",block.GetBlockID().GetString())
}

func Test_Deserialize_1(t *testing.T) {
	blockhash,_ := math.NewHashFromStr("57babc24019b8528b7a59f23af3bc0b18564ed0dcc51da9597bad150cc192ecd")
	str := "0a5a080012220a2000000000000000000000000000000000000000000000000000000000000000001a220a20000000000000000000000000000000000000000000000000000000000000000020aaf1b6c50528ffffff830230003800"
	buffer, _ := hex.DecodeString(str)
	block := &protobuf.POABlock{}

	err := proto.Unmarshal(buffer, block)
	if err != nil {
		t.Error("block 反序列化不通过 unmarshaling error: ", err)
	}


	newBlock := POABlock{}
	newBlock.Deserialize(block)
	newBlockHash := newBlock.GetBlockID().(math.Hash)

	if blockhash.IsEqual(&newBlockHash) {
		t.Log("block 反序列化通过")
	} else {
		t.Error("block 反序列化不通过")
	}
}

func Test_Serialize_3(t *testing.T) {
	fromAddress := math.Hash(sha256.Sum256([]byte("lf")))
	toAddress := math.Hash(sha256.Sum256([]byte("lc")))
	formAccount := &POAAccount{AccountID:POAAccountID{ID:fromAddress}}
	toAccount := &POAAccount{AccountID:POAAccountID{ID:toAddress}}
	amount := POAAmount{Value:10}
	tx := POATransaction{Version:0,
		From:GetPOATransactionPeer(formAccount,nil),
		To:GetPOATransactionPeer(toAccount,nil),
		Amount:amount,
		Time:time.Now()}

	t.Log("createtx","data",tx)

	s := tx.Serialize()

	buffer,err := proto.Marshal(s)
	if err != nil {
		t.Error("tx 序列化不通过 marshaling error",err)
	}
	t.Log("tx 序列化","buffer->",hex.EncodeToString(buffer))
	t.Log("tx 序列化","txid hash->",tx.GetTxID().GetString())
}

func Test_Serialize_4(t *testing.T) {
	txid,_ := math.NewHashFromStr("5d22d6602532f9b19a754d5a9b70256e0d9f779c43bdeee270e62475aa45b039")
	str := "080012260a240a220a2039e741eddb03e3118da619625b9200e131832dc8cc8542e2198b1e00e02c48a31a260a240a220a20ef07b359570add31929a5422d400b16c7c84e35644cb2e84b142f0710973e2a82202080a28ca95e3dc05"
	buffer, _ := hex.DecodeString(str)
	tx := &protobuf.POATransaction{}

	err := proto.Unmarshal(buffer, tx)
	if err != nil {
		t.Error("tx 反序列化不通过 unmarshaling error: ", err)
	}

	newTx := POATransaction{}
	newTx.Deserialize(tx)
	newTxHash := newTx.GetTxID().(math.Hash)

	if txid.IsEqual(&newTxHash) {
		t.Log("tx 反序列化通过")
	} else {
		t.Error("tx 反序列化不通过")
	}
}

func Test_Serialize_5(t *testing.T) {

	fromAddress := math.Hash(sha256.Sum256([]byte("lf")))
	toAddress := math.Hash(sha256.Sum256([]byte("lc")))
	formAccount := &POAAccount{AccountID:POAAccountID{ID:fromAddress}}
	toAccount := &POAAccount{AccountID:POAAccountID{ID:toAddress}}
	amount := POAAmount{Value:10}
	tx := POATransaction{Version:0,
		From:GetPOATransactionPeer(formAccount,nil),
		To:GetPOATransactionPeer(toAccount,nil),
		Amount:amount,
		Time:time.Now()}

	txs := []POATransaction{}
	txs = append(txs,tx)

	block := &POABlock{
		Header: POABlockHeader{Version: 0, PrevBlock: math.Hash{}, MerkleRoot: math.Hash{}, Timestamp: time.Unix(1487780010, 0), Difficulty: 0x207fffff, Nonce: 0, Extra: nil, Height: 0},
		TXs:    txs,
	}


	blockHash := block.GetBlockID().(math.Hash)
	s := block.Serialize()

	newBlock := POABlock{}
	newBlock.Deserialize(s)
	newBlockHash := newBlock.GetBlockID().(math.Hash)
	if blockHash.IsEqual(&newBlockHash) {
		t.Log("block with tx 反/序列化通过")
	} else {
		t.Error("block with tx 反/序列化不通过")
	}
}

