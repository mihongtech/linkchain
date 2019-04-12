// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package test

import (
	"math/big"
	"strings"

	"github.com/mihongtech/linkchain/accounts/abi"
	"github.com/mihongtech/linkchain/accounts/abi/bind"
	"github.com/mihongtech/linkchain/common"
	"github.com/mihongtech/linkchain/core/meta"
)

// MyTokenABI is the input ABI used to generate the binding from.
const MyTokenABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"initialSupply\",\"type\":\"uint256\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"constructor\"}]"

// MyTokenBin is the compiled bytecode used for deploying new contracts.
const MyTokenBin = `0x608060405260405160208061017e8339810160409081529051600160a060020a03331660009081526020819052919091205561013e806100406000396000f30060806040526004361061004b5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416635b46f8f68114610050578063a9059cbb1461008e575b600080fd5b341561005b57600080fd5b61007c73ffffffffffffffffffffffffffffffffffffffff600435166100bf565b60408051918252519081900360200190f35b341561009957600080fd5b6100bd73ffffffffffffffffffffffffffffffffffffffff600435166024356100d1565b005b60006020819052908152604090205481565b73ffffffffffffffffffffffffffffffffffffffff3381166000908152602081905260408082208054859003905593909116815291909120805490910190555600a165627a7a7230582042b48bd45637dde9900f8960c77a0ff1b39c8814c5f4a293aa9c4c1530a885a20029`

// DeployMyToken deploys a new Ethereum contract, binding an instance of MyToken to it.
func DeployMyToken(auth *bind.TransactOpts, backend bind.ContractBackend, initialSupply *big.Int) (meta.AccountID, *meta.Transaction, *MyToken, error) {
	parsed, err := abi.JSON(strings.NewReader(MyTokenABI))
	if err != nil {
		return meta.AccountID{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MyTokenBin), backend, initialSupply)
	if err != nil {
		return meta.AccountID{}, nil, nil, err
	}
	return address, tx, &MyToken{MyTokenCaller: MyTokenCaller{contract: contract}, MyTokenTransactor: MyTokenTransactor{contract: contract}, MyTokenFilterer: MyTokenFilterer{contract: contract}}, nil
}

// MyToken is an auto generated Go binding around an Ethereum contract.
type MyToken struct {
	MyTokenCaller     // Read-only binding to the contract
	MyTokenTransactor // Write-only binding to the contract
	MyTokenFilterer   // Log filterer for contract events
}

// MyTokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type MyTokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MyTokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MyTokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MyTokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MyTokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MyTokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MyTokenSession struct {
	Contract     *MyToken          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MyTokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MyTokenCallerSession struct {
	Contract *MyTokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// MyTokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MyTokenTransactorSession struct {
	Contract     *MyTokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// MyTokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type MyTokenRaw struct {
	Contract *MyToken // Generic contract binding to access the raw methods on
}

// MyTokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MyTokenCallerRaw struct {
	Contract *MyTokenCaller // Generic read-only contract binding to access the raw methods on
}

// MyTokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MyTokenTransactorRaw struct {
	Contract *MyTokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMyToken creates a new instance of MyToken, bound to a specific deployed contract.
func NewMyToken(address meta.AccountID, backend bind.ContractBackend) (*MyToken, error) {
	contract, err := bindMyToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MyToken{MyTokenCaller: MyTokenCaller{contract: contract}, MyTokenTransactor: MyTokenTransactor{contract: contract}, MyTokenFilterer: MyTokenFilterer{contract: contract}}, nil
}

// NewMyTokenCaller creates a new read-only instance of MyToken, bound to a specific deployed contract.
func NewMyTokenCaller(address meta.AccountID, caller bind.ContractCaller) (*MyTokenCaller, error) {
	contract, err := bindMyToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MyTokenCaller{contract: contract}, nil
}

// NewMyTokenTransactor creates a new write-only instance of MyToken, bound to a specific deployed contract.
func NewMyTokenTransactor(address meta.AccountID, transactor bind.ContractTransactor) (*MyTokenTransactor, error) {
	contract, err := bindMyToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MyTokenTransactor{contract: contract}, nil
}

// NewMyTokenFilterer creates a new log filterer instance of MyToken, bound to a specific deployed contract.
func NewMyTokenFilterer(address meta.AccountID, filterer bind.ContractFilterer) (*MyTokenFilterer, error) {
	contract, err := bindMyToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MyTokenFilterer{contract: contract}, nil
}

// bindMyToken binds a generic wrapper to an already deployed contract.
func bindMyToken(address meta.AccountID, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MyTokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MyToken *MyTokenRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MyToken.Contract.MyTokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MyToken *MyTokenRaw) Transfer(opts *bind.TransactOpts) (*meta.Transaction, error) {
	return _MyToken.Contract.MyTokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MyToken *MyTokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*meta.Transaction, error) {
	return _MyToken.Contract.MyTokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MyToken *MyTokenCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MyToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MyToken *MyTokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*meta.Transaction, error) {
	return _MyToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MyToken *MyTokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*meta.Transaction, error) {
	return _MyToken.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_MyToken *MyTokenCaller) BalanceOf(opts *bind.CallOpts, arg0 meta.AccountID) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MyToken.contract.Call(opts, out, "balanceOf", arg0)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_MyToken *MyTokenSession) BalanceOf(arg0 meta.AccountID) (*big.Int, error) {
	return _MyToken.Contract.BalanceOf(&_MyToken.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_MyToken *MyTokenCallerSession) BalanceOf(arg0 meta.AccountID) (*big.Int, error) {
	return _MyToken.Contract.BalanceOf(&_MyToken.CallOpts, arg0)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_MyToken *MyTokenTransactor) Transfer(opts *bind.TransactOpts, _to meta.AccountID, _value *big.Int) (*meta.Transaction, error) {
	return _MyToken.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_MyToken *MyTokenSession) Transfer(_to meta.AccountID, _value *big.Int) (*meta.Transaction, error) {
	return _MyToken.Contract.Transfer(&_MyToken.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns()
func (_MyToken *MyTokenTransactorSession) Transfer(_to meta.AccountID, _value *big.Int) (*meta.Transaction, error) {
	return _MyToken.Contract.Transfer(&_MyToken.TransactOpts, _to, _value)
}
