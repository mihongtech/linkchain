package contract

import (
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/consensus"
	"github.com/mihongtech/linkchain/contract/vm"
	"github.com/mihongtech/linkchain/core/meta"
	"math/big"
)

// ChainContext supports retrieving headers and consensus parameters from the
// current blockchain to be used during transaction processing.
type ChainContext interface {
	// Engine retrieves the ChainReader's consensus engine.
	Engine() consensus.Engine

	// GetHeader returns the hash corresponding to their hash.
	GetHeader(math.Hash, uint64) *meta.BlockHeader
}

// NewEVMContext creates a new context for use in the EVM.
func NewEVMContext(msg Message, header *meta.BlockHeader, chain ChainContext, author *meta.AccountID) vm.Context {
	// If we don't have an explicit author (i.e. not mining), extract from the Header
	var beneficiary []byte
	if author == nil {
		beneficiary, _ = chain.Engine().Author(header) // Ignore error, we're past Header validation
	} else {
		beneficiary = author.CloneBytes()
	}
	headerData := GetHeaderData(header)
	return vm.Context{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHashFn(header, chain),
		Origin:      msg.From(),
		Coinbase:    meta.BytesToAccountID(beneficiary),
		BlockNumber: new(big.Int).SetInt64(int64(header.Height)),
		Time:        new(big.Int).SetInt64(header.Time.Unix()),
		Difficulty:  new(big.Int).SetInt64(int64(header.Difficulty)),
		GasLimit:    headerData.GasLimit,
		GasPrice:    new(big.Int).Set(msg.GasPrice()),
	}
}

// GetHashFn returns a GetHashFunc which retrieves Header hashes by number
func GetHashFn(ref *meta.BlockHeader, chain ChainContext) func(n uint64) math.Hash {
	var cache map[uint64]math.Hash

	return func(n uint64) math.Hash {
		// If there's no hash cache yet, make one
		if cache == nil {
			cache = map[uint64]math.Hash{
				uint64(ref.Height - 1): ref.Prev,
			}
		}
		// Try to fulfill the request from the cache
		if hash, ok := cache[n]; ok {
			return hash
		}
		// Not cached, iterate the blocks and cache the hashes
		for header := chain.GetHeader(ref.Prev, uint64(ref.Height-1)); header != nil; header = chain.GetHeader(header.Prev, uint64(header.Height-1)) {
			cache[uint64(header.Height-1)] = header.Prev
			if n == uint64(header.Height-1) {
				return header.Prev
			}
		}
		return math.Hash{}
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr meta.AccountID, amount *big.Int) bool {
	return db.GetAvailableBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient meta.AccountID, amount *big.Int, code int) {
	db.Transfer(&sender, &recipient, amount, code)
}
