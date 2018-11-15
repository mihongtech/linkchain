package storage

import (
	_ "bytes"
	_ "math/big"
	"testing"

	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/math"
	"github.com/linkchain/poa/meta"
)

// Tests block storage and retrieval operations.
func TestBlockStorage(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()

	// Create a test block to move around the database and make sure it's really new
	block := poameta.NewBlock(poameta.BlockHeader{
		Data: []byte("test block"),
	}, []poameta.Transaction{})
	if entry := GetBlock(db, *block.GetBlockID(), uint64(block.GetHeight())); entry != nil {
		t.Fatalf("Non existent block returned: %v", entry)
	}
	// Write and verify the block in the database
	if err := WriteBlock(db, block); err != nil {
		t.Fatalf("Failed to write block into database: %v", err)
	}
	if entry := GetBlock(db, *block.GetBlockID(), uint64(block.GetHeight())); entry == nil {
		t.Fatalf("Stored block not found")
	} else if !entry.GetBlockID().IsEqual(block.GetBlockID()) {
		t.Fatalf("Retrieved block mismatch: have %v, want %v", entry, block)
	}
	// Delete the block and verify the execution
	DeleteBlock(db, *block.GetBlockID(), uint64(block.GetHeight()))
	if entry := GetBlock(db, *block.GetBlockID(), uint64(block.GetHeight())); entry != nil {
		t.Fatalf("Deleted block returned: %v", entry)
	}
}

// Tests that canonical numbers can be mapped to hashes and retrieved.
func TestCanonicalMappingStorage(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()

	// Create a test canonical number and assinged hash to move around
	hash, number := math.Hash{0: 0xff}, uint64(314)
	if entry := GetCanonicalHash(db, number); entry != (math.Hash{}) {
		t.Fatalf("Non existent canonical mapping returned: %v", entry)
	}
	// Write and verify the TD in the database
	if err := WriteCanonicalHash(db, hash, number); err != nil {
		t.Fatalf("Failed to write canonical mapping into database: %v", err)
	}
	if entry := GetCanonicalHash(db, number); entry == (math.Hash{}) {
		t.Fatalf("Stored canonical mapping not found")
	} else if entry != hash {
		t.Fatalf("Retrieved canonical mapping mismatch: have %v, want %v", entry, hash)
	}
	// Delete the TD and verify the execution
	DeleteCanonicalHash(db, number)
	if entry := GetCanonicalHash(db, number); entry != (math.Hash{}) {
		t.Fatalf("Deleted canonical mapping returned: %v", entry)
	}
}

// Tests that head headers and head blocks can be assigned, individually.
func TestHeadStorage(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	blockFull := poameta.NewBlock(poameta.BlockHeader{
		Data: []byte("test block full"),
	}, []poameta.Transaction{})
	blockFast := poameta.NewBlock(poameta.BlockHeader{
		Data: []byte("test block fast"),
	}, []poameta.Transaction{})

	// Check that no head entries are in a pristine database
	if entry := GetHeadBlockHash(db); entry != (math.Hash{}) {
		t.Fatalf("Non head header entry returned: %v", entry)
	}
	if entry := GetHeadFastBlockHash(db); entry != (math.Hash{}) {
		t.Fatalf("Non fast head block entry returned: %v", entry)
	}
	if err := WriteHeadBlockHash(db, *blockFull.GetBlockID()); err != nil {
		t.Fatalf("Failed to write head block hash: %v", err)
	}
	if err := WriteHeadFastBlockHash(db, *blockFast.GetBlockID()); err != nil {
		t.Fatalf("Failed to write fast head block hash: %v", err)
	}
	// Check that both heads are present, and different (i.e. two heads maintained)
	if entry := GetHeadBlockHash(db); entry != *blockFull.GetBlockID() {
		t.Fatalf("Head block hash mismatch: have %v, want %v", entry, blockFull.GetBlockID())
	}
	if entry := GetHeadFastBlockHash(db); entry != *blockFast.GetBlockID() {
		t.Fatalf("Fast head block hash mismatch: have %v, want %v", entry, blockFast.GetBlockID())
	}
}

//// Tests that positional lookup metadata can be stored and retrieved.
//func TestLookupStorage(t *testing.T) {
//	db, _ := lcdb.NewMemDatabase()
//
//	tx1 := types.NewTransaction(1, common.BytesToAddress([]byte{0x11}), big.NewInt(111), []byte{0x11, 0x11, 0x11})
//	tx2 := types.NewTransaction(2, common.BytesToAddress([]byte{0x22}), big.NewInt(222), []byte{0x22, 0x22, 0x22})
//	tx3 := types.NewTransaction(3, common.BytesToAddress([]byte{0x33}), big.NewInt(333), []byte{0x33, 0x33, 0x33})
//	txs := []*types.Transaction{tx1, tx2, tx3}
//
//	block := types.NewBlock(&types.Header{Number: big.NewInt(314)}, txs, nil, nil)
//
//	// Check that no transactions entries are in a pristine database
//	for i, tx := range txs {
//		if txn, _, _, _ := GetTransaction(db, tx.Hash()); txn != nil {
//			t.Fatalf("tx #%d [%x]: non existent transaction returned: %v", i, tx.Hash(), txn)
//		}
//	}
//	// Insert all the transactions into the database, and verify contents
//	if err := WriteBlock(db, block); err != nil {
//		t.Fatalf("failed to write block contents: %v", err)
//	}
//	if err := WriteTxLookupEntries(db, block); err != nil {
//		t.Fatalf("failed to write transactions: %v", err)
//	}
//	for i, tx := range txs {
//		if txn, hash, number, index := GetTransaction(db, tx.Hash()); txn == nil {
//			t.Fatalf("tx #%d [%x]: transaction not found", i, tx.Hash())
//		} else {
//			if hash != block.Hash() || number != block.NumberU64() || index != uint64(i) {
//				t.Fatalf("tx #%d [%x]: positional metadata mismatch: have %x/%d/%d, want %x/%v/%v", i, tx.Hash(), hash, number, index, block.Hash(), block.NumberU64(), i)
//			}
//			if tx.String() != txn.String() {
//				t.Fatalf("tx #%d [%x]: transaction mismatch: have %v, want %v", i, tx.Hash(), txn, tx)
//			}
//		}
//	}
//	// Delete the transactions and check purge
//	for i, tx := range txs {
//		DeleteTxLookupEntry(db, tx.Hash())
//		if txn, _, _, _ := GetTransaction(db, tx.Hash()); txn != nil {
//			t.Fatalf("tx #%d [%x]: deleted transaction returned: %v", i, tx.Hash(), txn)
//		}
//	}
//}
