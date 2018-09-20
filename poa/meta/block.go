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
	"github.com/linkchain/meta/block"
	"github.com/linkchain/meta/tx"
	"github.com/linkchain/protobuf"
)

type Block struct {
	Header BlockHeader
	TXs    []Transaction
}

type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version uint32

	// Hash of the previous block header in the block chain.
	PrevBlock math.Hash

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot math.Hash

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Timestamp time.Time

	// Difficulty target for the block.
	Difficulty uint32

	// Nonce used to generate the block.
	Nonce uint32

	//the height of block
	Height uint32

	// Extra used to extenion the block.
	Extra []byte

	hash math.Hash

	signer Signer
}

func NewPOABlock() (block.IBlock, error) {
	b := &Block{}
	return b, nil
}

func (b *Block) SetTx(newTXs []tx.ITx) error {
	for _, tx := range newTXs {
		b.TXs = append(b.TXs, *tx.(*Transaction))
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
	return &b.Header.PrevBlock
}
func (b *Block) GetMerkleRoot() meta.DataID {
	return &b.Header.MerkleRoot
}
func (b *Block) Verify() error {
	return b.Header.Verify()
}

//Serialize/Deserialize
func (b *Block) Serialize() serialize.SerializeStream {
	header := b.Header.Serialize().(*protobuf.BlockHeader)

	txs := make([]*protobuf.Transaction, 0)
	for _, tx := range b.TXs {
		txs = append(txs, tx.Serialize().(*protobuf.Transaction))
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

func (b *Block) Deserialize(s serialize.SerializeStream) {
	data := *s.(*protobuf.Block)
	b.Header.Deserialize(data.Header)
	for _, tx := range data.TxList.Txs {
		newTx := Transaction{}
		newTx.Deserialize(tx)
		b.TXs = append(b.TXs, newTx)
	}
}

func (b *Block) ToString() string {
	data, err := json.Marshal(b)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (b *Block) GetTxs() []tx.ITx {
	txs := make([]tx.ITx, 0)
	for _, tx := range b.TXs {
		txs = append(txs, &tx)
	}
	return txs
}

func (b *Block) CalculateTxTreeRoot() meta.DataID {
	//var txHash [32]byte
	//var txHashes [][]byte
	var transactions [][]byte

	for _, tx := range b.TXs {
		//txHashes = append(txHashes,tx.Hash())
		txbuff, _ := proto.Marshal(tx.Serialize())
		transactions = append(transactions, txbuff)
	}
	//txHash = sha256.Sum256(bytes.Join(txHashes,[]byte{}))
	mTree := merkle.NewMerkleTree(transactions)

	//return txHash[:]
	hash, _ := math.NewHash(mTree.RootNode.Data)
	return hash
}

func (b *Block) IsGensis() bool {
	return b.Header.IsGensis()
}

func (bh *BlockHeader) GetBlockID() meta.DataID {
	if bh.hash.IsEmpty() {
		bh.Deserialize(bh.Serialize())
	}
	return &bh.hash
}

func (bh *BlockHeader) GetSignerID() (account.IAccountID, error) {
	signer := Signer{}
	err := signer.Encode(bh.Extra)
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
	bh.Extra = buf
	bh.signer = signer
	return nil
}

func (bh *BlockHeader) GetMerkleRoot() meta.DataID {
	return &bh.MerkleRoot
}

func (bh *BlockHeader) SetMerkleRoot(root meta.DataID) {
	bh.MerkleRoot = *root.(*math.Hash)
}

//Serialize/Deserialize
func (bh *BlockHeader) Serialize() serialize.SerializeStream {
	prevHash := bh.PrevBlock.Serialize().(*protobuf.Hash)
	merkleRoot := bh.MerkleRoot.Serialize().(*protobuf.Hash)
	header := protobuf.BlockHeader{
		Version:    proto.Uint32(bh.Version),
		PrevHash:   prevHash,
		MerkleRoot: merkleRoot,
		Time:       proto.Int64(bh.Timestamp.Unix()),
		Difficulty: proto.Uint32(bh.Difficulty),
		Nounce:     proto.Uint32(bh.Nonce),
		Height:     proto.Uint32(bh.Height),
		Extra:      proto.NewBuffer(bh.Extra).Bytes(),
	}
	return &header
}

func (bh *BlockHeader) Deserialize(s serialize.SerializeStream) {
	data := s.(*protobuf.BlockHeader)
	bh.Version = *data.Version
	bh.PrevBlock.Deserialize(data.PrevHash)
	bh.MerkleRoot.Deserialize(data.MerkleRoot)
	bh.Timestamp = time.Unix(*data.Time, 0)
	bh.Difficulty = *data.Difficulty
	bh.Nonce = *data.Nounce
	bh.Height = *data.Height
	bh.Extra = data.Extra

	signer := Signer{}
	err := signer.Encode(bh.Extra)
	//TODO need handle error
	if err != nil {
		log.Error("BlockHeader", "Deserialize Signer failed", err)
	}
	bh.signer = signer

	t := protobuf.BlockHeader{
		Version:    data.Version,
		PrevHash:   data.PrevHash,
		MerkleRoot: data.MerkleRoot,
		Time:       data.Time,
		Difficulty: data.Difficulty,
		Nounce:     data.Nounce,
		Height:     data.Height,
		Extra:      proto.NewBuffer(signer.AccountID.ID.SerializeCompressed()).Bytes(),
	}
	bh.hash = math.MakeHash(&t)
}

func (b *BlockHeader) IsGensis() bool {
	return b.Height == 0 && b.PrevBlock.IsEmpty()
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
