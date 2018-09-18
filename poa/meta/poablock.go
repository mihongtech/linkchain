package meta

import (
	"encoding/json"
	"time"

	"github.com/linkchain/meta/tx"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/math"
	"github.com/linkchain/meta/block"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/poa/meta/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/linkchain/meta"
)

type POABlock struct{
	Header POABlockHeader
	TXs []POATransaction
}

type POABlockHeader struct {
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
}


func NewPOABlock()(block.IBlock, error){
	block := &POABlock{}
	return block, nil
}


func (b *POABlock)SetTx(newTXs []tx.ITx)(error){
	for _,tx := range newTXs{
		b.TXs = append(b.TXs,*tx.(*POATransaction))
	}
	return nil
}

func (b *POABlock)GetHeight() uint32{
	return b.Header.Height
}

func (b *POABlock)GetBlockID() meta.DataID{
	return b.Header.GetBlockID()
}

func (b *POABlock) GetPrevBlockID() meta.DataID{
	return &b.Header.PrevBlock
}

func (b *POABlock)Verify()(error){
	return nil
}

//Serialize/Deserialize
func (b *POABlock)Serialize()(serialize.SerializeStream){
	header := b.Header.Serialize().(*protobuf.POABlockHeader)

	txs := make([]*protobuf.POATransaction,0)
	for _,tx := range b.TXs {
		txs = append(txs,tx.Serialize().(*protobuf.POATransaction))
	}

	txlist := protobuf.POATransactions{
		Txs:txs,
	}

	block := protobuf.POABlock{
		Header:header,
		TxList:&txlist,
	}

	return &block
}

func (b *POABlock)Deserialize(s serialize.SerializeStream){
	data := *s.(*protobuf.POABlock)
	b.Header.Deserialize(data.Header)
	for _,tx := range data.TxList.Txs {
		newTx := POATransaction{}
		newTx.Deserialize(tx)
		b.TXs = append(b.TXs,newTx)
	}
}

func (b *POABlock) GetTxs() []tx.ITx {
	txs := make([]tx.ITx,0)
	for _,tx := range b.TXs {
		txs = append(txs,&tx)
	}
	return txs
}

//
func (b *POABlock)ToString()(string){
	data, err := json.Marshal(b);
	if  err != nil {
		return err.Error()
	}
	return string(data)
}


func (b *POABlock) IsGensis() bool {
	return b.Header.IsGensis()
}

func (bh *POABlockHeader)GetBlockID() meta.DataID{
	return &bh.hash
}

func (bh *POABlockHeader) GetMineAccount() account.IAccountID {
	return NewAccountId(bh.Extra)
}

func (bh *POABlockHeader) SetMineAccount(id account.IAccountID)  {
	bh.Extra = append(bh.Extra,id.(*POAAccountID).ID.SerializeUncompressed()...)
}

//Serialize/Deserialize
func (bh *POABlockHeader) Serialize()(serialize.SerializeStream){
	prevHash := bh.PrevBlock.Serialize().(*protobuf.Hash)
	merkleRoot := bh.MerkleRoot.Serialize().(*protobuf.Hash)
	header := protobuf.POABlockHeader{
		Version:proto.Uint32(bh.Version),
		PrevHash:prevHash,
		MerkleRoot:merkleRoot,
		Time:proto.Int64(bh.Timestamp.Unix()),
		Difficulty:proto.Uint32(bh.Difficulty),
		Nounce:proto.Uint32(bh.Nonce),
		Height:proto.Uint32(bh.Height),
		Extra:proto.NewBuffer(bh.Extra).Bytes(),
	}
	return &header
}

func (bh *POABlockHeader) Deserialize(s serialize.SerializeStream){
	data := s.(*protobuf.POABlockHeader)
	bh.Version = *data.Version
	bh.PrevBlock.Deserialize(data.PrevHash)
	bh.MerkleRoot.Deserialize(data.MerkleRoot)
	bh.Timestamp = time.Unix(*data.Time,0)
	bh.Difficulty = *data.Difficulty
	bh.Nonce = *data.Nounce
	bh.Height = *data.Height
	bh.Extra = data.Extra

	bh.hash = math.MakeHash(s)
}

func (b *POABlockHeader) IsGensis() bool {
	return b.Height == 0 && b.PrevBlock.IsEmpty()
}

