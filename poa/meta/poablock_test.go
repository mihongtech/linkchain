package meta

import (
	"testing"
	"time"
	"github.com/linkchain/common/math"
)

func Test_Serialize_1(t *testing.T) {
	txs := []POATransaction{}
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
		t.Log("序列化通过")
	} else {
		t.Error("序列化不通过")
	}
}

func Test_Deserialize_1(t *testing.T) {
	txs := []POATransaction{}
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
		t.Log("反序列化通过")
	} else {
		t.Error("反序列化不通过")
	}
}
