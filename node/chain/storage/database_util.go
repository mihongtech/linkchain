package storage

import (
	_ "bytes"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/mihongtech/linkchain/common/lcdb"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/node/config"
	"github.com/mihongtech/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
)

// DatabaseReader wraps the Get method of a backing data store.
type DatabaseReader interface {
	Get(key []byte) (value []byte, err error)
	Has(key []byte) (bool, error)
}

// DatabaseDeleter wraps the Delete method of a backing data store.
type DatabaseDeleter interface {
	Delete(key []byte) error
}

var (
	headBlockKey = []byte("LastBlock")
	headFastKey  = []byte("LastFast")

	// Data item prefixes (use single byte to avoid mixing data types, avoid `i`).
	blockPrefix = []byte("h") // blockPrefix + num (uint64 big endian) + hash -> block

	numSuffix       = []byte("n") // blockPrefix + num (uint64 big endian) + numSuffix -> hash
	blockHashPrefix = []byte("H") // blockHashPrefix + hash -> num (uint64 big endian)
	lookupPrefix    = []byte("l") // lookupPrefix + hash -> transaction/receipt lookup metadata

	configPrefix = []byte("linkchain-config-") // config prefix for the db

	ErrChainConfigNotFound = errors.New("ChainConfig not found") // general config not found error
)

// TxLookupEntry is a positional metadata to help looking up the data content of
// a transaction or receipt given only its hash.
type TxLookupEntry struct {
	BlockHash  string `json:"blockHash"`
	BlockIndex uint64 `json:"blockIndex"`
	Index      uint64 `json:"index"`
}

// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

// GetCanonicalHash retrieves a hash assigned to a canonical block number.
func GetCanonicalHash(db DatabaseReader, number uint64) math.Hash {
	data, _ := db.Get(append(append(blockPrefix, encodeBlockNumber(number)...), numSuffix...))
	if len(data) == 0 {
		return math.Hash{}
	}
	return math.BytesToHash(data)
}

// missingNumber is returned by GetBlockNumber if no header with the
// given block hash has been stored in the database
const MissingNumber = uint64(0xffffffffffffffff)

// GetBlockNumber returns the block number assigned to a block hash
// if the corresponding header is present in the database
func GetBlockNumber(db DatabaseReader, hash math.Hash) uint64 {
	data, _ := db.Get(append(blockHashPrefix, hash.Bytes()...))
	if len(data) != 8 {
		return MissingNumber
	}
	return binary.BigEndian.Uint64(data)
}

// GetHeadHeaderHash retrieves the hash of the current canonical head block's
// header. The difference between this and GetHeadBlockHash is that whereas the
// last block hash is only updated upon a full block import, the last header
// hash is updated already at header import, allowing head tracking for the
// light synchronization mechanism.
func GetHeadBlockHash(db DatabaseReader) math.Hash {
	data, _ := db.Get(headBlockKey)
	if len(data) == 0 {
		return math.Hash{}
	}
	return math.BytesToHash(data)
}

// GetHeadFastBlockHash retrieves the hash of the current canonical head block during
// fast synchronization. The difference between this and GetHeadBlockHash is that
// whereas the last block hash is only updated upon a full block import, the last
// fast hash is updated when importing pre-processed blocks.
func GetHeadFastBlockHash(db DatabaseReader) math.Hash {
	data, _ := db.Get(headFastKey)
	if len(data) == 0 {
		return math.Hash{}
	}
	return math.BytesToHash(data)
}

// GetHeaderBytes retrieves a block header in its raw database encoding, or nil
// if the header's not found.
func GetBlockBytes(db DatabaseReader, hash math.Hash, number uint64) []byte {
	data, _ := db.Get(blockKey(hash, number))
	return data
}

func HasBlock(db DatabaseReader, hash math.Hash, number uint64) bool {
	ok, _ := db.Has(blockKey(hash, number))
	return ok
}

func blockKey(hash math.Hash, number uint64) []byte {
	return append(append(blockPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

// GetBlock retrieves an entire block corresponding to the hash, assembling it
// back from the stored header and body. If either the header or body could not
// be retrieved nil is returned.
//
// Note, due to concurrent download of header and block body the header and thus
// canonical hash can be stored in the database but the body data not (yet).
func GetBlock(db DatabaseReader, hash math.Hash, number uint64) *meta.Block {
	data := GetBlockBytes(db, hash, number)
	if len(data) == 0 {
		return nil
	}
	var b protobuf.Block
	if err := proto.Unmarshal(data, &b); err != nil {
		log.Error("decode block failed")
		return nil
	}
	block := &meta.Block{}
	block.Deserialize(&b)
	return block
}

// GetTxLookupEntry retrieves the positional metadata associated with a transaction
// hash to allow retrieving the transaction or receipt by hash.
func GetTxLookupEntry(db DatabaseReader, hash math.Hash) (math.Hash, uint64, uint64) {
	// Load the positional metadata from disk and bail if it fails
	data, _ := db.Get(append(lookupPrefix, hash.CloneBytes()...))
	if len(data) == 0 {
		return math.Hash{}, 0, 0
	}

	// Parse and return the contents of the lookup entry
	var entry TxLookupEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		log.Error("Invalid lookup entry json data", "hash", hash, "err", err)
		return math.Hash{}, 0, 0
	}

	blockHash, _ := math.NewHashFromStr(entry.BlockHash)
	return *blockHash, entry.BlockIndex, entry.Index
}

// GetTransaction retrieves a specific transaction from the database, along with
// its added positional metadata.
func GetTransaction(db DatabaseReader, hash math.Hash) (*meta.Transaction, math.Hash, uint64, uint64) {
	// Retrieve the lookup metadata and resolve the transaction from the body
	blockHash, blockNumber, txIndex := GetTxLookupEntry(db, hash)
	// log.Info("get tx id", "blockHash", blockHash)
	if !blockHash.IsEmpty() {
		block := GetBlock(db, blockHash, blockNumber)
		if block == nil || len(block.TXs.Txs) <= int(txIndex) {
			log.Error("Transaction referenced missing", "number", blockNumber, "hash", blockHash, "index", txIndex)
			return nil, math.Hash{}, 0, 0
		}
		return &block.TXs.Txs[txIndex], blockHash, blockNumber, txIndex
	} else {
		log.Error("Transaction not found", "hash", hash)
		return nil, math.Hash{}, 0, 0
	}

}

// WriteCanonicalHash stores the canonical hash for the given block number.
func WriteCanonicalHash(db lcdb.Putter, hash math.Hash, number uint64) error {
	key := append(append(blockPrefix, encodeBlockNumber(number)...), numSuffix...)
	if err := db.Put(key, hash.Bytes()); err != nil {
		log.Crit("Failed to store number to hash mapping", "err", err)
	}
	return nil
}

// WriteHeadBlockHash stores the head block's hash.
func WriteHeadBlockHash(db lcdb.Putter, hash math.Hash) error {
	if err := db.Put(headBlockKey, hash.Bytes()); err != nil {
		log.Crit("Failed to store last block's hash", "err", err)
	}
	return nil
}

// WriteHeadFastBlockHash stores the fast head block's hash.
func WriteHeadFastBlockHash(db lcdb.Putter, hash math.Hash) error {
	if err := db.Put(headFastKey, hash.Bytes()); err != nil {
		log.Crit("Failed to store last fast block's hash", "err", err)
	}
	return nil
}

// WriteBlock serializes a block into the database, header and body separately.
func WriteBlock(db lcdb.Putter, block *meta.Block) error {

	data := block.Serialize()
	bytesData, err := proto.Marshal(data)
	if err != nil {
		return err
	}

	hash := block.GetBlockID().CloneBytes()
	num := block.GetHeight()
	encNum := encodeBlockNumber(uint64(num))
	key := append(blockHashPrefix, hash...)
	if err := db.Put(key, encNum); err != nil {
		log.Crit("Failed to store hash to number mapping", "err", err)
	}
	key = append(append(blockPrefix, encNum...), hash...)

	if err := db.Put(key, bytesData); err != nil {
		log.Crit("Failed to store block", "err", err)
	}
	return nil
}

// WriteTxLookupEntries stores a positional metadata for every transaction from
// a block, enabling hash based transaction and receipt lookups.
func WriteTxLookupEntries(db lcdb.Putter, block *meta.Block) error {

	// Iterate over each transaction and encode its metadata
	for i, tx := range block.TXs.Txs {
		entry := TxLookupEntry{
			BlockHash:  block.GetBlockID().String(),
			BlockIndex: uint64(block.GetHeight()),
			Index:      uint64(i),
		}
		data, err := json.Marshal(entry)

		if err != nil {
			return err
		}
		// log.Info("write tx id", "id", tx.GetTxID(), "entry", entry, "data", data)
		if err := db.Put(append(lookupPrefix, tx.GetTxID().CloneBytes()...), data); err != nil {
			return err
		}
	}
	return nil
}

// DeleteCanonicalHash removes the number to hash canonical mapping.
func DeleteCanonicalHash(db DatabaseDeleter, number uint64) {
	db.Delete(append(append(blockPrefix, encodeBlockNumber(number)...), numSuffix...))
}

// DeleteHeader removes all block header data associated with a hash.
func DeleteBlockData(db DatabaseDeleter, hash math.Hash, number uint64) {
	db.Delete(append(blockHashPrefix, hash.Bytes()...))
	db.Delete(append(append(blockPrefix, encodeBlockNumber(number)...), hash.Bytes()...))
}

// DeleteBlock removes all block data associated with a hash.
func DeleteBlock(db DatabaseDeleter, hash math.Hash, number uint64) {
	DeleteBlockData(db, hash, number)
}

// DeleteTxLookupEntry removes all transaction data associated with a hash.
func DeleteTxLookupEntry(db DatabaseDeleter, hash math.Hash) {
	db.Delete(append(lookupPrefix, hash.Bytes()...))
}

// WriteChainConfig writes the chain config settings to the database.
func WriteChainConfig(db lcdb.Putter, hash *math.Hash, cfg *config.ChainConfig) error {
	// short circuit and ignore if nil config. GetChainConfig
	// will return a default.
	if cfg == nil {
		return nil
	}

	jsonChainConfig, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return db.Put(append(configPrefix, hash.Bytes()[:]...), jsonChainConfig)
}

// GetChainConfig will fetch the network settings based on the given hash.
func GetChainConfig(db DatabaseReader, hash math.Hash) (*config.ChainConfig, error) {
	jsonChainConfig, _ := db.Get(append(configPrefix, hash[:]...))
	if len(jsonChainConfig) == 0 {
		return nil, ErrChainConfigNotFound
	}

	var chainConfig config.ChainConfig
	if err := json.Unmarshal(jsonChainConfig, &chainConfig); err != nil {
		return nil, err
	}

	return &chainConfig, nil
}
