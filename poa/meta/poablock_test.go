package meta

import (
	"testing"
	"time"
	"github.com/linkchain/common/math"
	"github.com/golang/protobuf/proto"
	"encoding/hex"
	"github.com/linkchain/poa/meta/protobuf"
)

func Test_Serialize_1(t *testing.T) {
	txs := []POATransaction{}
	block := &POABlock{
		Header: POABlockHeader{Version: 0, PrevBlock: math.Hash{}, MerkleRoot: math.Hash{}, Timestamp: time.Unix(1487780010, 0), Difficulty: 0x207fffff, Nonce: 0, Extra: nil, Height: 0},
		TXs:    txs,
	}
	s := block.Serialize()

	buffer,err := proto.Marshal(s)
	if err != nil {
		t.Error("序列化不通过 marshaling error",err)
	}
	t.Log("序列化","buffer",hex.EncodeToString(buffer),"block hash",block.GetBlockID().GetString())
}

func Test_Deserialize_1(t *testing.T) {
	blockhash,_ := math.NewHashFromStr("57babc24019b8528b7a59f23af3bc0b18564ed0dcc51da9597bad150cc192ecd")
	str := "0a5a080012220a2000000000000000000000000000000000000000000000000000000000000000001a220a20000000000000000000000000000000000000000000000000000000000000000020aaf1b6c50528ffffff830230003800"
	buffer, _ := hex.DecodeString(str)
	block := &protobuf.POABlock{}

	err := proto.Unmarshal(buffer, block)
	if err != nil {
		t.Error("反序列化不通过 unmarshaling error: ", err)
	}


	newBlock := POABlock{}
	newBlock.Deserialize(block)
	newBlockHash := newBlock.GetBlockID().(math.Hash)

	if blockhash.IsEqual(&newBlockHash) {
		t.Log("反序列化通过")
	} else {
		t.Error("反序列化不通过")
	}
}

func Test_Serialize_3(t *testing.T) {
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