package vm

import (
	"github.com/linkchain/storage/state"
	"math/big"
	"testing"

	"github.com/linkchain/common/math"
	"github.com/linkchain/config"
	"github.com/linkchain/core/meta"
)

var (
	//TestChainConfig = &config.ChainConfig{big.NewInt(1), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, new(EthashConfig), nil}
	TestChainConfig = &config.ChainConfig{}
)

type dummyContractRef struct {
	calledForEach bool
}

func (dummyContractRef) ReturnGas(*big.Int)        {}
func (dummyContractRef) Address() meta.AccountID   { return meta.AccountID{} }
func (dummyContractRef) Value() *big.Int           { return new(big.Int) }
func (dummyContractRef) SetCode(math.Hash, []byte) {}
func (d *dummyContractRef) ForEachStorage(callback func(key, value math.Hash) bool) {
	d.calledForEach = true
}
func (d *dummyContractRef) SubBalance(amount *big.Int) {}
func (d *dummyContractRef) AddBalance(amount *big.Int) {}
func (d *dummyContractRef) SetBalance(*big.Int)        {}
func (d *dummyContractRef) SetNonce(uint64)            {}
func (d *dummyContractRef) Balance() *big.Int          { return new(big.Int) }

type dummyStatedb struct {
	state.StateDB
}

func (*dummyStatedb) GetRefund() uint64 { return 1337 }

func TestStoreCapture(t *testing.T) {
	var (
		env      = NewEVM(Context{}, &dummyStatedb{}, TestChainConfig, Config{})
		logger   = NewStructLogger(nil)
		mem      = NewMemory()
		stack    = newstack()
		contract = NewContract(&dummyContractRef{}, &dummyContractRef{}, new(big.Int), 0)
	)
	stack.push(big.NewInt(1))
	stack.push(big.NewInt(0))
	var index math.Hash
	logger.CaptureState(env, 0, SSTORE, 0, 0, mem, stack, contract, 0, nil)
	if len(logger.changedValues[contract.Address()]) == 0 {
		t.Fatalf("expected exactly 1 changed value on address %x, got %d", contract.Address(), len(logger.changedValues[contract.Address()]))
	}
	exp := math.BigToHash(big.NewInt(1))
	if logger.changedValues[contract.Address()][index] != exp {
		t.Errorf("expected %x, got %x", exp, logger.changedValues[contract.Address()][index])
	}
}
