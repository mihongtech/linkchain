package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/trie"
	"github.com/linkchain/protobuf"
	"unsafe"

	"github.com/linkchain/common"
	"github.com/linkchain/common/math"
	"github.com/linkchain/core/meta"
)

//go:generate gencodec -type Receipt -field-override receiptMarshaling -out gen_receipt_json.go

var (
	receiptStatusFailedRLP     = []byte{}
	receiptStatusSuccessfulRLP = []byte{0x01}
)

const (
	// ReceiptStatusFailed is the status Code of a transaction if execution failed.
	ReceiptStatusFailed = uint64(0)

	// ReceiptStatusSuccessful is the status Code of a transaction if execution succeeded.
	ReceiptStatusSuccessful = uint64(1)
)

// Receipt represents the results of a transaction.
type Receipt struct {
	// Consensus fields
	PostState         []byte      `json:"root"`
	Status            uint64      `json:"status"`
	CumulativeGasUsed uint64      `json:"cumulativeGasUsed" gencodec:"required"`
	Bloom             Bloom       `json:"logsBloom"         gencodec:"required"`
	Logs              []*meta.Log `json:"logs"              gencodec:"required"`

	// Implementation fields (don't reorder!)
	TxHash          math.Hash      `json:"transactionHash" gencodec:"required"`
	ContractAddress meta.AccountID `json:"contractAddress"`
	GasUsed         uint64         `json:"gasUsed" gencodec:"required"`
}

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewReceipt(root []byte, failed bool, cumulativeGasUsed uint64) *Receipt {
	r := &Receipt{PostState: common.CopyBytes(root), CumulativeGasUsed: cumulativeGasUsed}
	if failed {
		r.Status = ReceiptStatusFailed
	} else {
		r.Status = ReceiptStatusSuccessful
	}
	return r
}

func (r *Receipt) Serialize() serialize.SerializeStream {
	logs := make([]*protobuf.Log, 0)
	for i := range r.Logs {
		logs = append(logs, r.Logs[i].Serialize().(*protobuf.Log))
	}
	receipt := protobuf.Receipt{
		PostStateOrStatus: proto.NewBuffer(r.statusEncoding()).Bytes(),
		CumulativeGasUsed: proto.Uint64(r.CumulativeGasUsed),
		Bloom:             proto.NewBuffer(r.Bloom.Bytes()).Bytes(),
		Logs:              logs,
	}
	return &receipt
}

func (r *Receipt) Deserialize(s serialize.SerializeStream) error {
	data := s.(*protobuf.Receipt)
	if err := r.setStatus(data.PostStateOrStatus); err != nil {
		return err
	}

	r.CumulativeGasUsed = *data.CumulativeGasUsed
	r.Bloom.SetBytes(data.Bloom)
	r.Logs = make([]*meta.Log, 0)
	for i := range data.Logs {
		log := &meta.Log{}
		if err := log.Deserialize(data.Logs[i]); err != nil {
			return err
		}
		r.Logs = append(r.Logs, log)
	}

	return nil
}

func (r *Receipt) String() string {
	data, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r *Receipt) setStatus(postStateOrStatus []byte) error {
	switch {
	case bytes.Equal(postStateOrStatus, receiptStatusSuccessfulRLP):
		r.Status = ReceiptStatusSuccessful
	case bytes.Equal(postStateOrStatus, receiptStatusFailedRLP):
		r.Status = ReceiptStatusFailed
	case len(postStateOrStatus) == len(math.Hash{}):
		r.PostState = postStateOrStatus
	default:
		return fmt.Errorf("invalid receipt status %x", postStateOrStatus)
	}
	return nil
}

func (r *Receipt) statusEncoding() []byte {
	if len(r.PostState) == 0 {
		if r.Status == ReceiptStatusFailed {
			return receiptStatusFailedRLP
		}
		return receiptStatusSuccessfulRLP
	}
	return r.PostState
}

// Size returns the approximate memory used by all internal contents. It is used
// to approximate and limit the memory consumption of various caches.
func (r *Receipt) Size() common.StorageSize {
	size := common.StorageSize(unsafe.Sizeof(*r)) + common.StorageSize(len(r.PostState))

	size += common.StorageSize(len(r.Logs)) * common.StorageSize(unsafe.Sizeof(meta.Log{}))
	for _, log := range r.Logs {
		size += common.StorageSize(len(log.Topics)*math.HashSize + len(log.Data))
	}
	return size
}

// ReceiptForStorage is a wrapper around a Receipt that flattens and parses the
// entire content of a receipt, as opposed to only the consensus fields originally.
type ReceiptForStorage Receipt

func (r *ReceiptForStorage) Serialize() serialize.SerializeStream {
	logs := make([]*protobuf.LogForStorage, 0)
	for i := range r.Logs {
		logs = append(logs, (*meta.LogForStorage)(r.Logs[i]).Serialize().(*protobuf.LogForStorage))
	}
	receipt := protobuf.ReceiptForStorage{
		PostStateOrStatus: proto.NewBuffer((*Receipt)(r).statusEncoding()).Bytes(),
		CumulativeGasUsed: proto.Uint64(r.CumulativeGasUsed),
		Bloom:             proto.NewBuffer(r.Bloom.Bytes()).Bytes(),
		Logs:              logs,
		TxHash:            r.TxHash.Serialize().(*protobuf.Hash),
		GasUsed:           proto.Uint64(r.GasUsed),
		ContractAddress:   r.ContractAddress.Serialize().(*protobuf.AccountID),
	}
	return &receipt
}

func (r *ReceiptForStorage) Deserialize(s serialize.SerializeStream) error {
	data := s.(*protobuf.ReceiptForStorage)
	if err := (*Receipt)(r).setStatus(data.PostStateOrStatus); err != nil {
		return err
	}

	r.CumulativeGasUsed = *data.CumulativeGasUsed
	r.Bloom.SetBytes(data.Bloom)
	r.GasUsed = *data.GasUsed
	r.Logs = make([]*meta.Log, 0)
	for i := range data.Logs {
		log := meta.LogForStorage{}
		if err := log.Deserialize(data.Logs[i]); err != nil {
			return err
		}

		r.Logs = append(r.Logs, log.ConvertToLog())
	}
	if err := r.TxHash.Deserialize(data.TxHash); err != nil {
		return err
	}
	if err := r.ContractAddress.Deserialize(data.ContractAddress); err != nil {
		return err
	}

	return nil
}

// Receipts is a wrapper around a Receipt array to implement DerivableList.
type Receipts []*Receipt

func GetReceiptHash(list []*Receipt) (math.Hash, error) {
	trie := new(trie.Trie)
	for i := range list {
		data, err := proto.Marshal(list[i].Serialize())
		if err != nil {
			return math.Hash{}, err
		}
		trie.Update(math.HashB(data), data)
	}
	return trie.Hash(), nil
}

// Len returns the number of receipts in this list.
func (r Receipts) Len() int { return len(r) }
