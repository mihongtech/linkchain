package config

import (
	_ "bytes"
	_ "encoding/hex"
	_ "encoding/json"
	"errors"
	"fmt"
	_ "math/big"
	_ "strings"
	"time"

	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	global_config "github.com/linkchain/config"
	_ "github.com/linkchain/meta"
	poa_meta "github.com/linkchain/poa/meta"
	"github.com/linkchain/storage"
)

var errGenesisNoConfig = errors.New("genesis has no chain configuration")

type Genesis struct {
	Config     *global_config.ChainConfig `json:"config"`
	Version    uint32                     `json:"version"`
	Time       int64                      `json:"time"`
	Data       []byte                     `json:"data"`
	Difficulty uint32                     `json:"difficulty" gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Height uint32    `json:"height"`
	Prev   math.Hash `json:"prev"`
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	return &Genesis{
		Config:  global_config.DefaultChainConfig,
		Version: DefaultBlockVersion,
		// Data:  hexutil.MustDecode("0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa"),
		Difficulty: DefaultDifficulty,
	}
}

// SetupGenesisBlock writes or updates the genesis block in db.
// The block that will be used is:
//
//                          genesis == nil       genesis != nil
//                       +------------------------------------------
//     db has no genesis |  main-net default  |  genesis
//     db has genesis    |  from DB           |  genesis (if compatible)
//
// The stored chain configuration will be updated if it is compatible (i.e. does not
// specify a fork block below the local head block). In case of a conflict, the
// error is a *params.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func SetupGenesisBlock(db lcdb.Database, genesis *Genesis) (*global_config.ChainConfig, math.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return global_config.DefaultChainConfig, math.Hash{}, errGenesisNoConfig
	}

	// Just commit the new block if there is no stored genesis block.
	stored := storage.GetCanonicalHash(db, 0)
	if (stored == math.Hash{}) {
		if genesis == nil {
			log.Info("Writing default main-net genesis block")
			genesis = DefaultGenesisBlock()
		} else {
			log.Info("Writing custom genesis block")
		}
		block, err := genesis.Commit(db)

		hash := math.BytesToHash(block.GetBlockID().(*math.Hash).CloneBytes())
		return genesis.Config, hash, err
	}

	// TODO: implement me
	return nil, math.Hash{}, nil
	//
	//	// Check whether the genesis block is already written.
	//	if genesis != nil {
	//		hash := genesis.ToBlock(nil).Hash()
	//		if hash != stored {
	//			return genesis.Config, hash, &GenesisMismatchError{stored, hash}
	//		}
	//	}
	//
	//	// Get the existing chain configuration.
	//	newcfg := genesis.configOrDefault(stored)
	//	storedcfg, err := GetChainConfig(db, stored)
	//	if err != nil {
	//		if err == ErrChainConfigNotFound {
	//			// This case happens if a genesis write was interrupted.
	//			log.Warn("Found genesis block without chain config")
	//			err = WriteChainConfig(db, stored, newcfg)
	//		}
	//		return newcfg, stored, err
	//	}
	//	// Special case: don't change the existing config of a non-mainnet chain if no new
	//	// config is supplied. These chains would get AllProtocolChanges (and a compat error)
	//	// if we just continued here.
	//	if genesis == nil && stored != params.MainnetGenesisHash {
	//		return storedcfg, stored, nil
	//	}
	//
	//	// Check config compatibility and write the config. Compatibility errors
	//	// are returned to the caller unless we're already at block zero.
	//	height := GetBlockNumber(db, GetHeadHeaderHash(db))
	//	if height == missingNumber {
	//		return newcfg, stored, fmt.Errorf("missing block number for head header hash")
	//	}
	//	compatErr := storedcfg.CheckCompatible(newcfg, height)
	//	if compatErr != nil && height != 0 && compatErr.RewindTo != 0 {
	//		return newcfg, stored, compatErr
	//	}
	//	return newcfg, stored, WriteChainConfig(db, stored, newcfg)

}

// ToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func (g *Genesis) ToBlock(db lcdb.Database) *poa_meta.Block {
	if db == nil {
		db, _ = lcdb.NewMemDatabase()
	}
	//	statedb, _ := state.New(math.Hash{}, state.NewDatabase(db))
	//	for addr, account := range g.Alloc {
	//		statedb.AddBalance(addr, account.Balance)
	//		statedb.SetCode(addr, account.Code)
	//		statedb.SetNonce(addr, account.Nonce)
	//		for key, value := range account.Storage {
	//			statedb.SetState(addr, key, value)
	//		}
	//	}
	//	root := statedb.IntermediateRoot(false)
	head := poa_meta.BlockHeader{
		Version:    g.Version,
		Height:     g.Height,
		Time:       time.Unix(g.Time, 0),
		Prev:       g.Prev,
		Data:       g.Data,
		Difficulty: g.Difficulty,
	}
	//	if g.Difficulty == nil {
	//		head.Difficulty = params.GenesisDifficulty
	//	}
	//	statedb.Commit(false)
	//	statedb.Database().TrieDB().Commit(root, true)

	return poa_meta.NewBlock(head, nil)
}

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db lcdb.Database) (*poa_meta.Block, error) {
	block := g.ToBlock(db)
	if block.GetHeight() != 0 {
		return nil, fmt.Errorf("can't commit genesis block with number > 0")
	}
	//	if err := WriteTd(db, block.Hash(), block.NumberU64(), g.Difficulty); err != nil {
	//		return nil, err
	//	}
	if err := storage.WriteBlock(db, block); err != nil {
		return nil, err
	}
	//	if err := WriteBlockReceipts(db, block.Hash(), block.NumberU64(), nil); err != nil {
	//		return nil, err
	//	}
	//	if err := WriteCanonicalHash(db, block.Hash(), block.NumberU64()); err != nil {
	//		return nil, err
	//	}
	//	if err := WriteHeadBlockHash(db, block.Hash()); err != nil {
	//		return nil, err
	//	}
	//	if err := WriteHeadHeaderHash(db, block.Hash()); err != nil {
	//		return nil, err
	//	}
	config := g.Config
	if config == nil {
		config = global_config.DefaultChainConfig
	}
	return block, storage.WriteChainConfig(db, block.GetBlockID().(*math.Hash), config)
}
