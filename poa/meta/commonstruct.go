package meta

import (
	"github.com/linkchain/common/math"
	"time"
)

type CBlock struct {
	Header CBlockHeader
	TXs    []CTransaction
}

type CBlockHeader struct {
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
}

type CTransaction struct {
	// Version of the Transaction.  This is not the same as the Blocks version.
	Version uint32

	Type uint32

	From [][]byte

	To [][]byte

	Value []uint64

	Signs [][]byte

	Data []byte
}
