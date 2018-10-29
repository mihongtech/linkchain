package meta

import (
	"encoding/json"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/merkle"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/protobuf"
)

type Block struct {
	Header BlockHeader
	TXs    []Transaction
}

func NewBlock(header BlockHeader, txs []Transaction) *Block {
	return &Block{
		Header: header,
		TXs:    txs,
	}
}

func (b *Block) SetTx(newTXs []tx.ITx) error {
	for _, iTx := range newTXs {
		b.TXs = append(b.TXs, *iTx.(*Transaction))
	}
	b.Header.SetMerkleRoot(b.CalculateTxTreeRoot()) //calculate merkle root
	return nil
}

func (b *Block) GetHeight() uint32 {
	return b.Header.Height
}

func (b *Block) GetBlockID() meta.DataID {
	return b.Header.GetBlockID()
}

func (b *Block) GetPrevBlockID() meta.DataID {
	return &b.Header.Prev
}
func (b *Block) GetMerkleRoot() meta.DataID {
	return &b.Header.TxRoot
}
func (b *Block) Verify() error {
	return b.Header.Verify()
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

func (b *Block) GetTxs() []tx.ITx {
	txs := make([]tx.ITx, 0)
	for _, transaction := range b.TXs {
		txs = append(txs, &transaction)
	}
	return txs
}

func (b *Block) CalculateTxTreeRoot() meta.DataID {
	var transactions [][]byte
	for _, transaction := range b.TXs {
		txbuff, _ := proto.Marshal(transaction.Serialize())
		transactions = append(transactions, txbuff)
	}
	mTree := merkle.NewMerkleTree(transactions)
	hash, _ := math.NewHash(mTree.RootNode.Data)
	return hash
}

func (b *Block) IsGensis() bool {
	return b.Header.IsGensis()
}

type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version uint32

	//the height of block
	Height uint32

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Time time.Time

	// Nonce used to generate the block.
	Nonce uint32

	// Difficulty target for the block.
	Difficulty uint32

	// Hash of the previous block header in the block chain.
	Prev math.Hash

	// Merkle tree reference to hash of all transactions for the block.
	TxRoot math.Hash

	// The status of the whole system
	Status math.Hash

	// The sign of miner
	Sign []byte

	// Data used to extenion the block.
	Data []byte

	//The Hash of this block
	hash math.Hash

	signer Signer
}

func NewBlockHeader(version uint32, prev math.Hash, root math.Hash, time time.Time, difficulty uint32, nounce uint32, height uint32, extra []byte) *BlockHeader {
	return &BlockHeader{
		Version: version,
		Height:  height,
		Time:    time,
		Nonce:   nounce,

		Difficulty: difficulty,
		Prev:       prev,
		TxRoot:     root,
		Status:     prev,  //TODO
		Sign:       extra, //TODO
		Data:       extra,
	}
}

func (bh *BlockHeader) GetBlockID() meta.DataID {
	if bh.hash.IsEmpty() {
		//TODO Deserialize
		bh.Deserialize(bh.Serialize())
	}
	return &bh.hash
}

func (bh *BlockHeader) GetSignerID() (account.IAccountID, error) {
	signer := Signer{}
	err := signer.Encode(bh.Sign)
	if err != nil {
		log.Error("BlockHeader", "Encode Signer failed", err)
		return nil, err
	}
	return &signer.AccountID, nil
}

func (bh *BlockHeader) GetSigner() (Signer, error) {
	return bh.signer, nil
}

func (bh *BlockHeader) SetSigner(signer Signer) error {
	buf, err := signer.Decode()
	if err != nil {
		return err
	}
	bh.Sign = buf
	bh.signer = signer
	return nil
}

func (bh *BlockHeader) GetMerkleRoot() meta.DataID {
	return &bh.TxRoot
}

func (bh *BlockHeader) SetMerkleRoot(root meta.DataID) {
	bh.TxRoot = *root.(*math.Hash)
}

//Serialize/Deserialize
func (bh *BlockHeader) Serialize() serialize.SerializeStream {
	prevHash := bh.Prev.Serialize().(*protobuf.Hash)
	merkleRoot := bh.TxRoot.Serialize().(*protobuf.Hash)
	status := bh.Status.Serialize().(*protobuf.Hash)
	header := protobuf.BlockHeader{
		Version:    proto.Uint32(bh.Version),
		Height:     proto.Uint32(bh.Height),
		Time:       proto.Int64(bh.Time.Unix()),
		Nounce:     proto.Uint32(bh.Nonce),
		Difficulty: proto.Uint32(bh.Difficulty),
		Prev:       prevHash,
		TxRoot:     merkleRoot,
		Status:     status,
		Sign:       proto.NewBuffer(bh.Sign).Bytes(),
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
	err := bh.Prev.Deserialize(data.Prev)
	if err != nil {
		return err
	}
	err = bh.TxRoot.Deserialize(data.TxRoot)
	if err != nil {
		return err
	}
	err = bh.Status.Deserialize(data.Status)
	if err != nil {
		return err
	}
	bh.Sign = data.Sign
	bh.Data = data.Data

	signer := Signer{}
	err = signer.Encode(bh.Sign)
	if err != nil {
		log.Error("BlockHeader", "Deserialize Signer failed", err)
		return err
	}
	bh.signer = signer

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
	bh.hash = math.MakeHash(&t)
	return nil
}

func (bh *BlockHeader) String() string {
	data, err := json.Marshal(bh)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (b *BlockHeader) IsGensis() bool {
	return b.Height == 0 && b.Prev.IsEmpty()
}

func (b *BlockHeader) Verify() error {
	signer, err := b.GetSigner()
	if err != nil {
		return err
	}
	err = signer.Verify(*b.GetBlockID().(*math.Hash))
	if err != nil {
		return err
	}
	return nil
}
