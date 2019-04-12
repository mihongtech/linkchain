package vm

import (
	"math/big"

	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/core/meta"
)

// stateDB is an EVM database for full state querying.
type StateDB interface {
	GetCallTx() math.Hash

	CreateContractAccount(meta.AccountID)

	Transfer(from, to *meta.AccountID, value *big.Int, code int) error
	GetAvailableBalance(meta.AccountID) *big.Int

	GetCodeHash(meta.AccountID) math.Hash
	GetCode(meta.AccountID) []byte
	SetCode(meta.AccountID, []byte)
	GetCodeSize(meta.AccountID) int

	AddRefund(uint64)
	SubRefund(uint64)
	GetRefund() uint64

	GetCommittedState(meta.AccountID, math.Hash) math.Hash
	GetState(meta.AccountID, math.Hash) math.Hash
	SetState(meta.AccountID, math.Hash, math.Hash)

	Suicide(meta.AccountID) bool
	HasSuicided(meta.AccountID) bool

	// Exist reports whether the given account exists in state.
	// Notably this should also return true for suicided accounts.
	Exist(meta.AccountID) bool
	// Empty returns whether the given account is empty. Empty
	// is defined according to EIP161 (balance = nonce = code = 0).
	Empty(meta.AccountID) bool

	RevertToSnapshot(int)
	Snapshot() int

	AddLog(*meta.Log)
	AddPreimage(math.Hash, []byte)
}

// CallContext provides a basic interface for the EVM calling conventions. The EVM
// depends on this context being implemented for doing subcalls and initialising new EVM contracts.
type CallContext interface {
	// Call another contract
	Call(env *EVM, me ContractRef, addr meta.AccountID, data []byte, gas, value *big.Int) ([]byte, error)
	// Take another's contract code and execute within our own context
	CallCode(env *EVM, me ContractRef, addr meta.AccountID, data []byte, gas, value *big.Int) ([]byte, error)
	// Same as CallCode except sender and value is propagated from parent to child scope
	DelegateCall(env *EVM, me ContractRef, addr meta.AccountID, data []byte, gas *big.Int) ([]byte, error)
	// Create a new contract
	Create(env *EVM, me ContractRef, data []byte, gas, value *big.Int) ([]byte, meta.AccountID, error)
}
