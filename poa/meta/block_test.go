package meta

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/math"
	"github.com/linkchain/poa/config"
	"github.com/linkchain/protobuf"
)

func Test_Serialize_block(t *testing.T) {
	txs := []Transaction{}
	header := NewBlockHeader(config.BlockVersion, 10, time.Now(), config.DefaultNounce, config.Difficulty, math.Hash{}, math.Hash{}, math.Hash{}, nil, nil)
	block := NewBlock(*header, txs)

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
	blockhash, _ := math.NewHashFromStr("ae619e6eff98ef61079be51c6e1bfb0bb6b0015737a21339d7e1cb416a06b548")
	str := "0a7e0801100a18cab8e0de05200028ffffffff0f32220a2000000000000000000000000000000000000000000000000000000000000000003a220a20000000000000000000000000000000000000000000000000000000000000000042220a2000000000000000000000000000000000000000000000000000000000000000001200"
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

	header := NewBlockHeader(config.BlockVersion, 10, time.Now(), config.DefaultNounce, config.Difficulty, math.Hash{}, math.Hash{}, math.Hash{}, nil, nil)
	block := NewBlock(*header, txs)

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
