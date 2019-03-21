package runtime

import (
	"github.com/linkchain/common/math"
	"github.com/linkchain/contract"
	"github.com/linkchain/contract/vm"
)

func NewEnv(cfg *Config) *vm.EVM {
	context := vm.Context{
		CanTransfer: contract.CanTransfer,
		Transfer:    contract.Transfer,
		GetHash:     func(uint64) math.Hash { return math.Hash{} },

		Origin:      cfg.Origin,
		Coinbase:    cfg.Coinbase,
		BlockNumber: cfg.BlockNumber,
		Time:        cfg.Time,
		Difficulty:  cfg.Difficulty,
		GasLimit:    cfg.GasLimit,
		GasPrice:    cfg.GasPrice,
	}
	return vm.NewEVM(context, cfg.State, cfg.ChainConfig, cfg.EVMConfig)
}
