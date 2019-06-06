package genesis

import (
	"errors"
	"fmt"
	"time"

	"github.com/mihongtech/linkchain/common/lcdb"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/node/chain/storage"
	"github.com/mihongtech/linkchain/node/config"
)

var errGenesisNoConfig = errors.New("genesis has no chain configuration")

type Genesis struct {
	Config     *config.ChainConfig `json:"config"`
	Version    uint32              `json:"version"`
	Time       int64               `json:"time"`
	Data       []byte              `json:"data"`
	Difficulty uint32              `json:"difficulty" gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Height uint32    `json:"height"`
	Prev   math.Hash `json:"prev"`
}

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New math.Hash
}

func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database already contains an incompatible genesis block (have %x, new %x)", e.Stored[:8], e.New[:8])
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	return &Genesis{
		Config:  config.DefaultChainConfig,
		Version: config.DefaultBlockVersion,
		Time:    1487780010,
		Height:  0,
		Data:    nil,
		// Data:  hexutil.MustDecode("0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa"),
		Difficulty: config.DefaultDifficulty,
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
func SetupGenesisBlock(db lcdb.Database, genesis *Genesis) (*config.ChainConfig, math.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		log.Error("genesis.Config can not be nil")
		return nil, math.Hash{}, errGenesisNoConfig
	}

	// Just commit the new block if there is no stored genesis block.
	stored := storage.GetCanonicalHash(db, 0)
	log.Debug("stored data is", "store", stored)
	if (stored.IsEqual(&math.Hash{})) {
		if genesis == nil {
			log.Info("Writing default main-net genesis block")
			genesis = DefaultGenesisBlock()
		} else {
			log.Info("Writing custom genesis block")
		}
		block, err := genesis.Commit(db)

		hash := math.BytesToHash(block.GetBlockID().CloneBytes())
		return genesis.Config, hash, err
	}

	newcfg := config.DefaultChainConfig

	storedcfg, err := storage.GetChainConfig(db, stored)
	if err != nil {
		if err == storage.ErrChainConfigNotFound {
			// This case happens if a genesis write was interrupted.
			log.Warn("Found genesis block without chain config")
			err = storage.WriteChainConfig(db, &stored, newcfg)
		}
		return newcfg, stored, err
	}
	if genesis == nil {
		newcfg = config.DefaultChainConfig
	}
	newcfg = storedcfg

	// Check whether the genesis block is already written.
	if genesis != nil {
		hash := math.BytesToHash(genesis.ToBlock(nil).GetBlockID().CloneBytes())
		if hash == stored {
			return genesis.Config, hash, &GenesisMismatchError{stored, hash}
		}
	}

	height := storage.GetBlockNumber(db, storage.GetHeadBlockHash(db))
	if height == storage.MissingNumber {
		return genesis.Config, stored, fmt.Errorf("missing block number for head block hash")
	}

	return newcfg, stored, storage.WriteChainConfig(db, &stored, newcfg)

}

func (g *Genesis) configOrDefault(ghash math.Hash) *config.ChainConfig {
	switch {
	case g != nil:
		return g.Config
	default:
		return config.DefaultChainConfig
	}
}

// ToBlock creates the genesis block
// to the given database (or discards it if nil).
func (g *Genesis) ToBlock(db lcdb.Database) *meta.Block {
	if db == nil {
		db, _ = lcdb.NewMemDatabase()
	}

	// TODO: add tx to account

	head := meta.BlockHeader{
		Version:    g.Version,
		Height:     g.Height,
		Time:       time.Unix(g.Time, 0),
		Prev:       math.Hash{},
		Data:       g.Data,
		Nonce:      config.DefaultNounce,
		TxRoot:     math.Hash{},
		Status:     math.Hash{},
		Difficulty: g.Difficulty,
	}

	block := meta.NewBlock(head, []meta.Transaction{})
	txRoot := block.CalculateTxTreeRoot()
	block.Header.SetMerkleRoot(txRoot)

	return block
}

// Commit writes the block to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db lcdb.Database) (*meta.Block, error) {
	block := g.ToBlock(db)
	if block.GetHeight() != 0 {
		return nil, fmt.Errorf("can't commit genesis block with number > 0")
	}

	if err := storage.WriteBlock(db, block); err != nil {
		log.Info("Commit", "err", err)
		return nil, err
	}

	if err := storage.WriteCanonicalHash(db, *block.GetBlockID(), uint64(block.GetHeight())); err != nil {
		return nil, err
	}

	if err := storage.WriteHeadBlockHash(db, *block.GetBlockID()); err != nil {
		return nil, err
	}

	globalConfig := g.Config
	if globalConfig == nil {
		globalConfig = config.DefaultChainConfig
	}
	return block, storage.WriteChainConfig(db, block.GetBlockID(), globalConfig)
}
