package meta

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/protobuf"
	"github.com/mihongtech/linkchain/unittest"

	"github.com/golang/protobuf/proto"
)

func getTestBlock() *Block {
	hash, _ := math.NewHashFromStr("ae619e6eff98ef61079be51c6e1bfb0bb6b0015737a21339d7e1cb416a06b548")
	var txs = make([]Transaction, 0)
	header := NewBlockHeader(config.DefaultBlockVersion, 10, time.Unix(1487780010, 0), config.DefaultNounce, config.DefaultDifficulty, *hash, *hash, *hash, Signature{Code: make([]byte, 0)}, nil)
	return NewBlock(*header, txs)
}

func TestBlock_Serialize_Without_Tx(t *testing.T) {
	block := getTestBlock()
	s := block.Serialize()

	_, err := proto.Marshal(s)
	unittest.NotError(t, err)
	//t.Log("block Serialize", "block hash->", block.GetBlockID().String(), "buffer->", hex.EncodeToString(buffer))
}

func TestBlock_Deserialize_Without_Tx(t *testing.T) {
	hash, _ := math.NewHashFromStr("bfbc84fe6d6aae6552ad50a4f89c5b73f0f315681ec3d52dd0ea08265c90117c")
	str := "0a82010801100a18b7e7c1e005200028ffffffff0f32220a2048b5066a41cbe1d73913a2375701b0b60bfb1b6e1ce59b0761ef98ff6e9e61ae3a220a2048b5066a41cbe1d73913a2375701b0b60bfb1b6e1ce59b0761ef98ff6e9e61ae42220a2048b5066a41cbe1d73913a2375701b0b60bfb1b6e1ce59b0761ef98ff6e9e61ae4a020a001200"
	buffer, _ := hex.DecodeString(str)
	block := &protobuf.Block{}

	err := proto.Unmarshal(buffer, block)
	unittest.NotError(t, err)

	newBlock := Block{}
	newBlock.Deserialize(block)
	newBlockHash := newBlock.GetBlockID()
	unittest.Assert(t, hash.IsEqual(newBlockHash), "TestBlock_Deserialize_Without_Tx")
}

func TestBlock_Serialize(t *testing.T) {
	block := getTestBlock()
	tx := getTestTransaction()
	block.SetTx(*tx)
	s := block.Serialize()

	_, err := proto.Marshal(s)
	unittest.NotError(t, err)
	//t.Log("block Serialize", "block hash->", block.GetBlockID().GetString(), "buffer->", hex.EncodeToString(buffer))
}

func TestBlock_Deserialize(t *testing.T) {
	hash, _ := math.NewHashFromStr("86909d3cc0598cd2ce02c605303ef0db66c8e09c18c18b24d88c110794365581")
	str := "0a82010801100a18aaf1b6c505200028ffffffff0f32220a2048b5066a41cbe1d73913a2375701b0b60bfb1b6e1ce59b0761ef98ff6e9e61ae3a220a2048b5066a41cbe1d73913a2375701b0b60bfb1b6e1ce59b0761ef98ff6e9e61ae42220a2048b5066a41cbe1d73913a2375701b0b60bfb1b6e1ce59b0761ef98ff6e9e61ae4a020a001200"
	buffer, _ := hex.DecodeString(str)
	block := &protobuf.Block{}

	err := proto.Unmarshal(buffer, block)
	unittest.NotError(t, err)

	newBlock := Block{}
	newBlock.Deserialize(block)
	newBlockHash := newBlock.GetBlockID()
	unittest.Assert(t, hash.IsEqual(newBlockHash), "TestBlock_Deserialize_Without_Tx")
}

func TestBlock_CalculateTxTreeRoot(t *testing.T) {
	block := getTestBlock()
	tx := getTestTransaction()
	block.SetTx(*tx)
	txid, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")

	root := block.CalculateTxTreeRoot()
	//TODO need change merkle root class
	unittest.Assert(t, !root.IsEqual(txid), "TestBlock_CalculateTxTreeRoot")
}

func TestBlock_GetBlockID(t *testing.T) {
	block := getTestBlock()

	//when hash is empty
	block.Header.hash = math.Hash{}
	blockId := block.GetBlockID()
	unittest.NotEqual(t, blockId, nil)

	//when hash is not empty
	hash, _ := math.NewHashFromStr("86909d3cc0598cd2ce02c605303ef0db66c8e09c18c18b24d88c110794365581")
	block = getTestBlock()
	blockId = block.GetBlockID()
	unittest.Assert(t, blockId.IsEqual(hash), "TestBlock_GetBlockID")

}

func TestBlock_GetTxs(t *testing.T) {
	block := getTestBlock()
	txs := block.GetTxs()
	unittest.Equal(t, len(txs), 0)
}

func TestBlock_IsGensis(t *testing.T) {
	block := getTestBlock()
	unittest.Assert(t, !block.IsGensis(), "TestBlock_IsGensis")
}

func TestBlock_GetTx(t *testing.T) {
	block := getTestBlock()
	tx := getTestTransaction()
	block.SetTx(*tx)
	txid, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")

	_, err := block.GetTx(*txid)
	unittest.NotError(t, err)
}

func TestBlock_GetTx_Error(t *testing.T) {
	block := getTestBlock()
	tx := getTestTransaction()
	block.SetTx(*tx)
	txid, _ := math.NewHashFromStr("86909d3cc0598cd2ce02c605303ef0db66c8e09c18c18b24d88c110794365581")

	_, err := block.GetTx(*txid)
	unittest.Error(t, err)
}

func TestMakeTreeID(t *testing.T) {
	//TODO TreeID need to change
	treeID, _ := math.NewHashFromStr("890ab5fb96f4f5158177b76f026fd136bb58e2ad6150fd3074de9c16a1a5961a")
	hash, _ := math.NewHashFromStr("bfbc84fe6d6aae6552ad50a4f89c5b73f0f315681ec3d52dd0ea08265c90117c")
	buffer, _ := hex.DecodeString("0a473045022100b3b46c98236f2760344e5c9aaec44e7463f2435858c396fd731eed0e03f28d2502205cee7691880636a5643b5cc24fc05196ec2184939fccf289753efaaefe279516")
	transactions := make(map[math.Hash][]byte)
	transactions[*hash] = buffer
	id, err := GetMakeTreeID(transactions)
	unittest.Equal(t, err, nil)
	unittest.Equal(t, id, *treeID)
}
