package bind

import (
	"context"
	"errors"
	_ "fmt"
	"math/big"

	"github.com/linkchain/accounts/abi"
	_ "github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/contract"
	"github.com/linkchain/contract/vm"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/helper"

	"github.com/golang/protobuf/proto"
)

// SignerFn is a signer function callback when a contract requires a method to
// sign the transaction before submission.
type SignerFn func(meta.AccountID, *meta.Transaction) (*meta.Transaction, error)

// CallOpts is the collection of options to fine tune a contract call request.
type CallOpts struct {
	Pending bool           // Whether to operate on the pending state or the last known one
	From    meta.AccountID // Optional the sender address, otherwise the first account is used

	Context context.Context // Network context to support cancellation and timeouts (nil = no timeout)
}

// TransactOpts is the collection of authorization data required to create a
// valid Ethereum transaction.
type TransactOpts struct {
	From     meta.AccountID // Ethereum account to send the transaction from
	FromCoin *meta.FromCoin
	Nonce    *big.Int // Nonce to use for the transaction execution (nil = use pending state)
	Signer   SignerFn // Method to use for signing the transaction (mandatory)

	Value    *big.Int // Funds to transfer along along the transaction (nil = 0 = no funds)
	GasPrice *big.Int // Gas price to use for the transaction execution (nil = gas price oracle)
	GasLimit uint64   // Gas limit to set for the transaction execution (0 = estimate)

	Context context.Context // Network context to support cancellation and timeouts (nil = no timeout)
}

// FilterOpts is the collection of options to fine tune filtering for events
// within a bound contract.
type FilterOpts struct {
	Start uint64  // Start of the queried range
	End   *uint64 // End of the range (nil = latest)

	Context context.Context // Network context to support cancellation and timeouts (nil = no timeout)
}

// WatchOpts is the collection of options to fine tune subscribing for events
// within a bound contract.
type WatchOpts struct {
	Start   *uint64         // Start of the queried range (nil = latest)
	Context context.Context // Network context to support cancellation and timeouts (nil = no timeout)
}

// BoundContract is the base wrapper object that reflects a contract on the
// Ethereum network. It contains a collection of methods that are used by the
// higher level contract bindings to operate.
type BoundContract struct {
	address    meta.AccountID     // Deployment address of the contract on the Ethereum blockchain
	abi        abi.ABI            // Reflect based ABI to access the correct Ethereum methods
	caller     ContractCaller     // Read interface to interact with the blockchain
	transactor ContractTransactor // Write interface to interact with the blockchain
	filterer   ContractFilterer   // Event filtering to interact with the blockchain
}

// NewBoundContract creates a low level contract interface through which calls
// and transactions may be made through.
func NewBoundContract(address meta.AccountID, abi abi.ABI, caller ContractCaller, transactor ContractTransactor, filterer ContractFilterer) *BoundContract {
	return &BoundContract{
		address:    address,
		abi:        abi,
		caller:     caller,
		transactor: transactor,
		filterer:   filterer,
	}
}

// DeployContract deploys a contract onto the Ethereum blockchain and binds the
// deployment address with a Go wrapper.
func DeployContract(opts *TransactOpts, abi abi.ABI, bytecode []byte, backend ContractBackend, params ...interface{}) (meta.AccountID, *meta.Transaction, *BoundContract, error) {
	// Otherwise try to deploy the contract
	c := NewBoundContract(meta.AccountID{}, abi, backend, backend, backend)

	input, err := c.abi.Pack("", params...)
	if err != nil {
		return meta.AccountID{}, nil, nil, err
	}
	tx, err := c.transact(opts, nil, append(bytecode, input...))
	if err != nil {
		return meta.AccountID{}, nil, nil, err
	}

	c.address = vm.CreateContractAccountID(opts.From, *tx.GetTxID())
	return c.address, tx, c, nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (c *BoundContract) Call(opts *CallOpts, result interface{}, method string, params ...interface{}) error {
	// Don't crash on a lazy user
	if opts == nil {
		opts = new(CallOpts)
	}
	// Pack the input, call and unpack the results
	input, err := c.abi.Pack(method, params...)
	if err != nil {
		return err
	}
	var (
		msg    = contract.NewMessage(opts.From, &c.address, nil, 0, nil, input, false)
		ctx    = ensureContext(opts.Context)
		code   []byte
		output []byte
	)
	if opts.Pending {
		pb, ok := c.caller.(PendingContractCaller)
		if !ok {
			return ErrNoPendingState
		}
		output, err = pb.PendingCallContract(ctx, msg)
		if err == nil && len(output) == 0 {
			// Make sure we have a contract to operate on, and bail out otherwise.
			if code, err = pb.PendingCodeAt(ctx, c.address); err != nil {
				return err
			} else if len(code) == 0 {
				return ErrNoCode
			}
		}
	} else {
		output, err = c.caller.CallContract(ctx, msg, nil)
		if err == nil && len(output) == 0 {
			// Make sure we have a contract to operate on, and bail out otherwise.
			if code, err = c.caller.CodeAt(ctx, c.address, nil); err != nil {
				return err
			} else if len(code) == 0 {
				return ErrNoCode
			}
		}
	}
	if err != nil {
		return err
	}
	return c.abi.Unpack(result, method, output)
}

// Transact invokes the (paid) contract method with params as input values.
func (c *BoundContract) Transact(opts *TransactOpts, method string, params ...interface{}) (*meta.Transaction, error) {
	// Otherwise pack up the parameters and invoke the contract
	input, err := c.abi.Pack(method, params...)
	if err != nil {
		return nil, err
	}
	return c.transact(opts, &c.address, input)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (c *BoundContract) Transfer(opts *TransactOpts) (*meta.Transaction, error) {
	return c.transact(opts, &c.address, nil)
}

// transact executes an actual transaction invocation, first deriving any missing
// authorization fields, and then scheduling the transaction for execution.
func (c *BoundContract) transact(opts *TransactOpts, address *meta.AccountID, input []byte) (*meta.Transaction, error) {
	var err error

	// Ensure a valid value field and resolve the account nonce
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}

	// Create the transaction, sign it and schedule it for execution
	var rawTx *meta.Transaction
	//	bestHeight := nodeAPI.GetBestBlock().GetHeight()
	//
	amount := meta.NewAmount(value.Int64())
	//
	//	//make from
	//	from, err := walletAPI.GetAccount(opts.From.String())
	//	if err != nil {
	//		return nil, err
	//	}
	//	fromCoin, _, err := from.MakeFromCoin(amount, bestHeight+1)
	if err != nil {
		return nil, err
	}

	if address == nil {
		contractId, _ := meta.NewAccountIdFromStr("0000000000000000000000000000000000000000")

		toCoin := helper.CreateToCoin(*contractId, amount)

		rawTx = helper.CreateTransaction(*opts.FromCoin, *toCoin)
	} else {
		toCoin := helper.CreateToCoin(*address, amount)

		rawTx = helper.CreateTransaction(*opts.FromCoin, *toCoin)
	}
	if opts.Signer == nil {
		return nil, errors.New("no signer to authorize the transaction with")
	}

	rawTx.Type = contract.ContractTx
	data := contract.TxData{Price: opts.GasPrice, GasLimit: opts.GasLimit, Payload: input}
	protobufMsg := data.Serialize()
	rawTx.Data, err = proto.Marshal(protobufMsg)
	if err != nil {
		log.Error("Marshal tx contract paload error", "err", err)
		return nil, err
	}
	signedTx := new(meta.Transaction)

	if signedTx, err = c.transactor.SendTransaction(ensureContext(opts.Context), rawTx); err != nil {
		return nil, err
	}
	signedTx.GetTxID()

	return signedTx, nil
}

// ensureContext is a helper method to ensure a context is not nil, even if the
// user specified it as such.
func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.TODO()
	}
	return ctx
}
