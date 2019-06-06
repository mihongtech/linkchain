package meta

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/node/config"
	"github.com/mihongtech/linkchain/unittest"

	"github.com/golang/protobuf/proto"
)

func TestMakeBlockId(t *testing.T) {
	expBlockId, _ := math.NewHashFromStr("3ca925d511119d96a9b5fcd5874eae79cc9e9c457a23eacec7354bc8f1e29e8a")
	hash, _ := math.NewHashFromStr("ae619e6eff98ef61079be51c6e1bfb0bb6b0015737a21339d7e1cb416a06b548")
	var txs = make([]Transaction, 0)
	header := BlockHeader{
		Version:    config.DefaultBlockVersion,
		Height:     10,
		Time:       time.Unix(1487780010, 0),
		Nonce:      config.DefaultNounce,
		Difficulty: config.DefaultDifficulty,
		Prev:       *hash,
		TxRoot:     *hash,
		Status:     *hash,
		Data:       nil,
	}
	block := NewBlock(header, txs)

	newBuffer, err := proto.Marshal(block.Serialize())
	unittest.NotError(t, err)
	newBlockId := MakeBlockId(newBuffer)
	unittest.Assert(t, newBlockId.IsEqual(expBlockId), "TestMakeBlockId")
}

func TestMakeTxID(t *testing.T) {
	hash, _ := math.NewHashFromStr("6045e1be843b2d7292a7ecd512df315d81e77b7817dbd1c6cb379926f4d235e9")
	buffer, _ := hex.DecodeString("0a473045022100b3b46c98236f2760344e5c9aaec44e7463f2435858c396fd731eed0e03f28d2502205cee7691880636a5643b5cc24fc05196ec2184939fccf289753efaaefe279516")
	unittest.Assert(t, hash.IsEqual(MakeTxID(buffer)), "TestMakeTxID")
}
