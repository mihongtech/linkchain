package contract

import (
	"github.com/linkchain/core"
	"math/big"

	_ "github.com/linkchain/common"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
)

type BlockHeaderData struct {
	ReceiptHash math.Hash  `json:"receiptsRoot"     gencodec:"required"`
	Bloom       core.Bloom `json:"logsBloom"        gencodec:"required"`
	GasLimit    uint64     `json:"gasLimit"         gencodec:"required"`
	GasUsed     uint64     `json:"gasUsed"          gencodec:"required"`
}

//Serialize/Deserialize
func (a *BlockHeaderData) Serialize() serialize.SerializeStream {
	receiptHash := a.ReceiptHash.Serialize().(*protobuf.Hash)
	headerData := protobuf.BlockHeaderData{
		GasLimit:    &a.GasLimit,
		GasUsed:     &a.GasUsed,
		ReceiptHash: receiptHash,
	}

	return &headerData
}

func (a *BlockHeaderData) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.BlockHeaderData)
	a.GasUsed = *data.GasUsed
	a.GasLimit = *data.GasLimit
	if err := a.ReceiptHash.Deserialize(data.ReceiptHash); err != nil {
		return err
	}
	return nil
}

func GetHeaderData(header *meta.BlockHeader) *BlockHeaderData {
	if len(header.Data) == 0 {
		return &BlockHeaderData{GasLimit: DefaultBlockGasLimit}
	}
	buffer := make([]byte, 256)

	headerData := new(protobuf.BlockHeaderData)
	data := new(BlockHeaderData)
	err := proto.Unmarshal(header.Data, headerData)
	if err != nil {
		log.Error("Unmarshal block header data error", "err", err)
		return nil
	}

	err = data.Deserialize(headerData)
	if err != nil {
		log.Error("Deserialize block header data error", "err", err)
		return nil
	}

	data.Bloom = core.BytesToBloom(buffer)
	return data
}

func NewBlockHeaderData(receipts core.Receipts, gasUsed uint64) *BlockHeaderData {
	headerData := BlockHeaderData{GasLimit: DefaultBlockGasLimit, GasUsed: gasUsed}
	headerData.ReceiptHash, _ = core.GetReceiptHash(receipts)
	headerData.Bloom = core.CreateBloom(receipts)
	return &headerData
}

type TxData struct {
	Price    *big.Int `json:"gasPrice" gencodec:"required"`
	GasLimit uint64   `json:"gas"      gencodec:"required"`
	Payload  []byte   `json:"input"    gencodec:"required"`
}

//Serialize/Deserialize
func (a *TxData) Serialize() serialize.SerializeStream {
	price := a.Price.Uint64()
	txData := protobuf.TxData{
		GasLimit: &a.GasLimit,
		Price:    &price,
		Payload:  a.Payload,
	}

	return &txData
}

func (a *TxData) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.TxData)
	a.Price = big.NewInt(int64(*data.Price))
	a.GasLimit = *data.GasLimit
	a.Payload = data.Payload
	return nil
}

func GetTxData(tx *meta.Transaction) *TxData {
	txData := new(protobuf.TxData)
	data := new(TxData)
	err := proto.Unmarshal(tx.Data, txData)
	if err != nil {
		log.Error("Unmarshal tx contract paload error", "err", err)
		return nil
	}

	err = data.Deserialize(txData)
	if err != nil {
		log.Error("Deserialize tx data error", "err", err)
		return nil
	}

	return data
}
