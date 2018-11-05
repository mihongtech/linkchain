package config

import (
	_ "bytes"
	_ "encoding/hex"
	_ "encoding/json"
	"errors"
	_ "fmt"
	"math/big"
	_ "strings"

	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/math"
)

var errGenesisNoConfig = errors.New("genesis has no chain configuration")

type ChainConfig struct {
	ChainId *big.Int `json:"chainId"` // Chain id identifies the current chain and is used for replay protection
	Period  uint64   `json:"period"`  // Number of seconds between blocks to enforce
}

type Genesis struct {
	Config     *ChainConfig `json:"config"`
	Version    uint32       `json:"version"`
	Time       uint64       `json:"time"`
	Data       []byte       `json:"data"`
	Difficulty *big.Int     `json:"difficulty" gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Height uint32    `json:"height"`
	Prev   math.Hash `json:"prev"`
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
func SetupGenesisBlock(db lcdb.Database, genesis *Genesis) (*ChainConfig, math.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return DefaultChainConfig, math.Hash{}, errGenesisNoConfig
	}

	// TODO: implement me
	return nil, math.Hash{}, nil

	// Just commit the new block if there is no stored genesis block.
	//	stored := GetCanonicalHash(db, 0)
	//	if (stored == math.Hash{}) {
	//		if genesis == nil {
	//			log.Info("Writing default main-net genesis block")
	//			genesis = DefaultGenesisBlock()
	//		} else {
	//			log.Info("Writing custom genesis block")
	//		}
	//		block, err := genesis.Commit(db)
	//		return genesis.Config, block.Hash(), err
	//	}
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
