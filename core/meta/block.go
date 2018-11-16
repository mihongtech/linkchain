package meta

import (
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/merkle"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/math"
)

type Block struct {
	Header BlockHeader   `json:"header"`
	TXs    []Transaction `json:"txs"`
}

func NewBlock(header BlockHeader, txs []Transaction) *Block {
	ntxs := make([]Transaction, 0)
	for _, tx := range txs {
		ntxs = append(ntxs, tx)
	}
	return &Block{
		Header: header,
		TXs:    ntxs,
	}
}

func (b *Block) SetTx(newTXs ...Transaction) error {
	for _, tx := range newTXs {
		b.TXs = append(b.TXs, tx)
	}
	b.Header.SetMerkleRoot(b.CalculateTxTreeRoot()) //calculate merkle root

	return nil
}

func (b *Block) SetSign(signature math.ISignature) {
	b.Header.Sign = *signature.(*Signature)
}

func (b *Block) GetHeight() uint32 {
	return b.Header.Height
}

func (b *Block) GetBlockID() *BlockID {
	return b.Header.GetBlockID()
}

func (b *Block) GetPrevBlockID() *BlockID {
	return &b.Header.Prev
}
func (b *Block) GetMerkleRoot() *TreeID {
	return b.Header.GetMerkleRoot()
}
func (b *Block) Verify(minerPKStr string) error {
	return b.Header.Verify(minerPKStr)
}

//Serialize/Deserialize
func (b *Block) Serialize() serialize.SerializeStream {
	header := b.Header.Serialize().(*protobuf.BlockHeader)

	txs := make([]*protobuf.Transaction, 0)
	for _, transaction := range b.TXs {
		txs = append(txs, transaction.Serialize().(*protobuf.Transaction))
	}

	txlist := protobuf.Transactions{
		Txs: txs,
	}

	block := protobuf.Block{
		Header: header,
		TxList: &txlist,
	}

	return &block
}

func (b *Block) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.Block)
	err := b.Header.Deserialize(data.Header)
	if err != nil {
		return err
	}
	b.TXs = b.TXs[:0] // transaction clear
	for _, transaction := range data.TxList.Txs {
		newTx := Transaction{}
		err = newTx.Deserialize(transaction)
		if err != nil {
			return err
		}
		b.TXs = append(b.TXs, newTx)
	}
	return nil
}

func (b *Block) String() string {
	data, err := json.Marshal(b)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (b *Block) GetTxs() []Transaction {
	return b.TXs
}

func (b *Block) CalculateTxTreeRoot() TreeID {
	var transactions [][]byte
	for _, transaction := range b.TXs {
		txbuff, _ := proto.Marshal(transaction.Serialize())
		transactions = append(transactions, txbuff)
	}
	mTree := merkle.NewMerkleTree(transactions)

	hash, _ := MakeTreeID(mTree.RootNode.Data)
	return *hash
}

func (b *Block) IsGensis() bool {
	return b.Header.IsGensis()
}

type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version uint32 `json:"version,int"`

	//the height of block
	Height uint32 `json:"height,int"`

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Time time.Time `json:"time"`

	// Nonce used to generate the block.
	Nonce uint32 `json:"nonce"`

	// Difficulty target for the block.
	Difficulty uint32 `json:"difficulty"`

	// Hash of the previous block header in the block chain.
	Prev BlockID `json:"prev"`

	// Merkle tree reference to hash of all transactions for the block.
	TxRoot TreeID `json:"txroot"`

	// The status of the whole system
	Status TreeID `json:"status"`

	// The sign of miner
	Sign Signature `json:"sign"`

	// Data used to extenion the block.
	Data []byte `json:"data"`

	//The Hash of this block
	hash BlockID
}

func NewBlockHeader(version uint32, height uint32, time time.Time, nounce uint32, difficulty uint32, prev BlockID, root TreeID, status TreeID, sign Signature, extra []byte) *BlockHeader {

	return &BlockHeader{
		Version: version,
		Height:  height,
		Time:    time,
		Nonce:   nounce,

		Difficulty: difficulty,
		Prev:       prev,
		TxRoot:     root,
		Status:     status,
		Sign:       sign,
		Data:       extra,
	}
}

func (bh *BlockHeader) GetBlockID() *BlockID {
	if bh.hash.IsEmpty() {
		err := bh.Deserialize(bh.Serialize())
		if err != nil {
			log.Error("BlockHeader", "GetBlockID() error", err)
			return nil
		}
	}
	return &bh.hash
}

func (bh *BlockHeader) GetMerkleRoot() *TreeID {
	return &bh.TxRoot
}

func (bh *BlockHeader) SetMerkleRoot(root TreeID) {
	bh.TxRoot = root
}

//Serialize/Deserialize
func (bh *BlockHeader) Serialize() serialize.SerializeStream {
	prevHash := bh.Prev.Serialize().(*protobuf.Hash)
	merkleRoot := bh.TxRoot.Serialize().(*protobuf.Hash)
	status := bh.Status.Serialize().(*protobuf.Hash)
	sign := bh.Sign.Serialize().(*protobuf.Signature)
	header := protobuf.BlockHeader{
		Version:    proto.Uint32(bh.Version),
		Height:     proto.Uint32(bh.Height),
		Time:       proto.Int64(bh.Time.Unix()),
		Nounce:     proto.Uint32(bh.Nonce),
		Difficulty: proto.Uint32(bh.Difficulty),
		Prev:       prevHash,
		TxRoot:     merkleRoot,
		Status:     status,
		Sign:       sign,
		Data:       proto.NewBuffer(bh.Data).Bytes(),
	}
	return &header
}

func (bh *BlockHeader) Deserialize(s serialize.SerializeStream) error {
	data := s.(*protobuf.BlockHeader)
	bh.Version = *data.Version
	bh.Height = *data.Height
	bh.Time = time.Unix(*data.Time, 0)
	bh.Nonce = *data.Nounce
	bh.Difficulty = *data.Difficulty
	if err := bh.Prev.Deserialize(data.Prev); err != nil {
		return err
	}

	if err := bh.TxRoot.Deserialize(data.TxRoot); err != nil {
		return err
	}

	if err := bh.Status.Deserialize(data.Status); err != nil {
		return err
	}

	if err := bh.Sign.Deserialize(data.Sign); err != nil {
		return err
	}

	bh.Data = data.Data

	t := protobuf.BlockHeader{
		Version:    data.Version,
		Height:     data.Height,
		Time:       data.Time,
		Nounce:     data.Nounce,
		Difficulty: data.Difficulty,
		Prev:       data.Prev,
		TxRoot:     data.TxRoot,
		Status:     data.Status,
		Data:       data.Data,
	}

	buffer, err := proto.Marshal(&t)
	if err != nil {
		return err
	}

	bh.hash = *MakeBlockId(buffer)
	return nil
}

func (bh *BlockHeader) String() string {
	data, err := json.Marshal(bh)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (bh *BlockHeader) IsGensis() bool {
	return bh.Height == 0 && bh.Prev.IsEmpty()
}

func (bh *BlockHeader) Verify(minerPKStr string) error {
	signature, err := btcec.ParseSignature(bh.Sign.Code, btcec.S256())
	if err != nil {
		log.Error("Signer", "VerifySign", err)
		return err
	}

	minerPK, err := hex.DecodeString(minerPKStr)
	if err != nil {
		return err
	}
	pk, err := btcec.ParsePubKey(minerPK, btcec.S256())
	if err != nil {
		return err
	}

	verified := signature.Verify(bh.GetBlockID().CloneBytes(), pk)
	if !verified {
		return err
	}
	return nil
}
