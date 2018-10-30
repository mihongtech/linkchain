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

func Test_Serialize_block(t *testing.T) {
	txs := []Transaction{}
	block := Block{
		Header: BlockHeader{Version: 0, Prev: math.Hash{}, TxRoot: math.Hash{}, Time: time.Unix(1487780010, 0), Difficulty: 0x207fffff, Nonce: 0, Data: nil, Height: 0},
		TXs:    txs,
	}

	t.Log("createblock", "data", block)

	s := block.Serialize()

	buffer, err := proto.Marshal(s)
	if err != nil {
		t.Error("block 序列化不通过 marshaling error", err)
	}
	t.Log("block 序列化", "buffer->", hex.EncodeToString(buffer))
	t.Log("block 序列化", "block hash->", block.GetBlockID().GetString())
}

func Test_Deserialize_block(t *testing.T) {
	blockhash, _ := math.NewHashFromStr("57babc24019b8528b7a59f23af3bc0b18564ed0dcc51da9597bad150cc192ecd")
	str := "0a5a080012220a2000000000000000000000000000000000000000000000000000000000000000001a220a20000000000000000000000000000000000000000000000000000000000000000020aaf1b6c50528ffffff8302300038001200"
	buffer, _ := hex.DecodeString(str)
	block := &protobuf.Block{}

	err := proto.Unmarshal(buffer, block)
	if err != nil {
		t.Error("block 反序列化不通过 unmarshaling error: ", err)
	}

	newBlock := Block{}
	newBlock.Deserialize(block)
	newBlockHash := newBlock.GetBlockID()

	if blockhash.IsEqual(newBlockHash) {
		t.Log("block 反序列化通过")
	} else {
		t.Error("block 反序列化不通过")
	}
}

func Test_Serialize_block_with_tx(t *testing.T) {
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

	txs := []Transaction{}
	txs = append(txs, tx)

	block := &Block{
		Header: BlockHeader{Version: 0, Prev: math.Hash{}, TxRoot: math.Hash{}, Time: time.Unix(1487780010, 0), Difficulty: 0x207fffff, Nonce: 0, Data: nil, Height: 0},
		TXs:    txs,
	}

	blockHash := block.GetBlockID()
	s := block.Serialize()

	newBlock := Block{}
	newBlock.Deserialize(s)
	newBlockHash := newBlock.GetBlockID()
	if blockHash.IsEqual(newBlockHash) {
		t.Log("block with tx 反/序列化通过")
	} else {
		t.Error("block with tx 反/序列化不通过")
	}
}
