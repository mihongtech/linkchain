package poameta

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/math"
	"github.com/linkchain/poa/config"
	"github.com/linkchain/protobuf"
)

func Test_Serialize_block(t *testing.T) {

	blockhash, _ := math.NewHashFromStr("ae619e6eff98ef61079be51c6e1bfb0bb6b0015737a21339d7e1cb416a06b548")
	txs := []Transaction{}
	header := NewBlockHeader(config.BlockVersion, 10, time.Now(), config.DefaultNounce, config.Difficulty, *blockhash, *blockhash, *blockhash, Signature{Code: make([]byte, 0)}, nil)
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
	blockhash, _ := math.NewHashFromStr("6248f821a4612457d9888423cc7c213d44a9f710090dabe96087cd86d5d44a33")
	str := "0a82010801100a18d5ac90df05200028ffffffff0f32220a2048b5066a41cbe1d73913a2375701b0b60bfb1b6e1ce59b0761ef98ff6e9e61ae3a220a2048b5066a41cbe1d73913a2375701b0b60bfb1b6e1ce59b0761ef98ff6e9e61ae42220a2048b5066a41cbe1d73913a2375701b0b60bfb1b6e1ce59b0761ef98ff6e9e61ae4a020a001200"
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
	/*fromAddress := math.Hash(sha256.Sum256([]byte("lf")))
	toAddress := math.Hash(sha256.Sum256([]byte("lc")))
	formAccount := &Account{Id: Id{ID: fromAddress.CloneBytes()}}
	toAccount := &Account{Id: Id{ID: toAddress.CloneBytes()}}
	amount := Amount{Value: 10}
	tx := Transaction{Version: 0,
		From:   *NewTransactionPeer(formAccount.Id, nil),
		To:     *NewTransactionPeer(toAccount.Id, nil),
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
	}*/
}
