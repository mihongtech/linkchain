package contract

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/contract/vm/params"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/helper"
	"github.com/linkchain/storage/state"
)

type BalanceChange struct {
	from  *meta.AccountID
	to    *meta.AccountID
	value *big.Int
	Code  int
}

type revision struct {
	id           int
	journalIndex int
}

type StateAdapter struct {
	stateDB     *state.StateDB
	coinBase    meta.AccountID
	blockHeight int64
	//TODO:can use cacheState
	cacheAccount      map[meta.AccountID]meta.Account
	cacheCode         map[meta.AccountID][]byte
	cacheBalance      map[meta.AccountID]meta.Amount
	cacheSuicided     map[meta.AccountID]bool
	cacheAccountState map[meta.AccountID]map[math.Hash]math.Hash

	transfers []BalanceChange
	refund    uint64

	thash, bhash math.Hash

	logs    map[math.Hash][]*meta.Log
	logSize uint

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionId int
}

func NewStateAdapter(stateDB *state.StateDB, callTx meta.TxID, blockHash meta.BlockID, coinBase meta.AccountID, blockHeight int64) *StateAdapter {
	return &StateAdapter{stateDB: stateDB,
		cacheAccount:      make(map[meta.AccountID]meta.Account),
		cacheCode:         make(map[meta.AccountID][]byte),
		cacheBalance:      make(map[meta.AccountID]meta.Amount),
		cacheSuicided:     make(map[meta.AccountID]bool),
		cacheAccountState: make(map[meta.AccountID]map[math.Hash]math.Hash),
		transfers:         make([]BalanceChange, 0),
		thash:             callTx,
		bhash:             blockHash,
		logs:              make(map[math.Hash][]*meta.Log),
		journal:           newJournal(),
		validRevisions:    make([]revision, 0),
		refund:            0,
		logSize:           0,
		coinBase:          coinBase,
		blockHeight:       blockHeight}
}

func (s *StateAdapter) GetCallTx() math.Hash {
	return s.thash
}

func (s *StateAdapter) CreateContractAccount(accountID meta.AccountID) {
	account := helper.CreateTemplateAccount(accountID)
	account.AccountType = ContractAccount
	s.cacheAccount[account.Id] = *account
	s.journal.append(createObjectChange{account: &account.Id})
}

func (s *StateAdapter) Transfer(from, to *meta.AccountID, value *big.Int, code int) error {
	switch code {
	case params.CoinBase:
		if from == nil && to != nil {
			s.AddBalance(*to, value)
		}
	case params.AddZero:
	case params.Refund:
		if from == nil && to != nil {
			s.AddBalance(*to, value)
		}
	case params.BuyGas:
		if from != nil && to == nil {
			s.SubBalance(*from, value)
		}
	case params.Suicide:
		if from != nil && to != nil {
			s.SubBalance(*from, value)
			s.AddBalance(*to, value)
		}
	case params.Normal:
		if from != nil && to != nil {
			s.SubBalance(*from, value)
			s.AddBalance(*to, value)
		}
	}
	transfer := BalanceChange{from: from, to: to, value: value, Code: code}
	s.transfers = append(s.transfers, transfer)
	s.journal.append(transferChange{len(s.transfers) - 1})
	return nil
}

func (s *StateAdapter) AddBalance(accountId meta.AccountID, value *big.Int) {
	addAmount := *meta.NewAmount(value.Int64())
	if v, ok := s.cacheBalance[accountId]; ok {
		amount := &v
		s.journal.append(balanceChange{
			account: &accountId,
			prev:    amount.GetBigInt(),
		})
		amount.Addition(addAmount)
		s.cacheBalance[accountId] = *amount
	} else {
		obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
		if obj != nil {
			amount := obj.GetAccount().GetAmount()
			s.journal.append(balanceChange{
				account: &accountId,
				prev:    amount.GetBigInt(),
			})
			amount.Addition(addAmount)
			s.cacheBalance[accountId] = *amount
		} else {
			amount := meta.NewAmount(0)
			s.journal.append(balanceChange{
				account: &accountId,
				prev:    amount.GetBigInt(),
			})
			amount.Addition(addAmount)
			s.cacheBalance[accountId] = *amount
		}
	}
}

func (s *StateAdapter) SubBalance(accountId meta.AccountID, value *big.Int) {
	subAmount := *meta.NewAmount(value.Int64())
	if v, ok := s.cacheBalance[accountId]; ok {
		amount := &v
		s.journal.append(balanceChange{
			account: &accountId,
			prev:    amount.GetBigInt(),
		})
		amount.Subtraction(subAmount)
		s.cacheBalance[accountId] = *amount
	} else {
		obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
		if obj != nil {
			amount := obj.GetAccount().GetAmount()
			s.journal.append(balanceChange{
				account: &accountId,
				prev:    amount.GetBigInt(),
			})
			amount.Subtraction(subAmount)
			s.cacheBalance[accountId] = *amount
		} else {
			amount := meta.NewAmount(0)
			s.journal.append(balanceChange{
				account: &accountId,
				prev:    amount.GetBigInt(),
			})
			amount.Subtraction(subAmount)
			s.cacheBalance[accountId] = *amount
		}
	}
}

func (s *StateAdapter) setBalance(accountId meta.AccountID, value *big.Int) {
	if _, ok := s.cacheBalance[accountId]; ok {
		s.cacheBalance[accountId] = *meta.NewAmount(value.Int64())
	}
}

//TODO  get available balance need to block height
func (s *StateAdapter) GetAvailableBalance(accountId meta.AccountID) *big.Int {
	if v, ok := s.cacheBalance[accountId]; ok {
		return v.GetBigInt()
	} else {
		obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
		if obj != nil {
			return new(big.Int).Mul(new(big.Int).SetInt64(obj.GetAccount().GetAmount().GetInt64()), meta.NewAmount(10000).GetBigInt())

		} else {
			amount := meta.NewAmount(0)
			return amount.GetBigInt()
		}
	}
}

func (s *StateAdapter) AddRefund(gas uint64) {
	s.journal.append(refundChange{prev: s.refund})
	s.refund += gas
}

func (s *StateAdapter) SubRefund(gas uint64) {
	s.journal.append(refundChange{prev: s.refund})
	s.refund -= gas
}

func (s *StateAdapter) GetRefund() uint64 {
	return s.refund
}

func (s *StateAdapter) GetCodeHash(accountId meta.AccountID) math.Hash {
	if v, ok := s.cacheAccount[accountId]; ok {
		return v.CodeHash
	} else {
		obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
		if obj != nil {
			return obj.GetAccount().CodeHash
		}
	}

	return math.Hash{}
}

func (s *StateAdapter) GetCode(accountId meta.AccountID) []byte {
	if v, ok := s.cacheCode[accountId]; ok {
		return v
	} else {
		obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
		if obj != nil {
			s.cacheCode[accountId] = obj.Code()
			return s.cacheCode[accountId]
		}
	}
	return nil
}

func (s *StateAdapter) SetCode(accountId meta.AccountID, code []byte) {
	prevCode := s.GetCode(accountId)
	s.journal.append(codeChange{
		account:  &accountId,
		prevhash: math.HashB(prevCode),
		prevcode: prevCode,
	})
	s.cacheCode[accountId] = code
}

func (s *StateAdapter) GetCodeSize(accountId meta.AccountID) int {
	code := s.GetCode(accountId)
	return len(code)
}

//get contract storage state
func (s *StateAdapter) GetState(accountId meta.AccountID, hash math.Hash) math.Hash {
	if v, ok := s.cacheAccountState[accountId]; ok {
		if state, ok := v[hash]; ok {
			return state
		} else {
			obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
			if obj != nil {
				state := obj.GetState(s.stateDB.DataBase(), hash)
				v[hash] = state
				s.cacheAccountState[accountId] = v
				return state
			}
		}
	}

	obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
	if obj != nil {
		state := obj.GetState(s.stateDB.DataBase(), hash)
		accountState := make(map[math.Hash]math.Hash)
		accountState[hash] = state
		s.cacheAccountState[accountId] = accountState
		return state
	}

	return math.Hash{}
}

func (s *StateAdapter) SetState(accountId meta.AccountID, key math.Hash, value math.Hash) {
	// New value is different, update and journal the change
	s.journal.append(storageChange{
		account:  &accountId,
		key:      key,
		prevalue: s.GetState(accountId, key),
	})
	if v, ok := s.cacheAccountState[accountId]; ok {
		v[key] = value
		s.cacheAccountState[accountId] = v
	} else {
		accountState := make(map[math.Hash]math.Hash)
		accountState[key] = value
		s.cacheAccountState[accountId] = accountState
	}
}

//TODO:useless
func (s *StateAdapter) GetCommittedState(accountId meta.AccountID, hash math.Hash) math.Hash {
	obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
	if obj != nil {
		return obj.GetCommittedState(s.stateDB.DataBase(), hash)
	}
	return math.Hash{}
}

func (s *StateAdapter) Suicide(accountId meta.AccountID) bool {
	s.cacheSuicided[accountId] = true
	return false
}

func (s *StateAdapter) HasSuicided(accountId meta.AccountID) bool {
	if v, ok := s.cacheSuicided[accountId]; ok {
		return v
	} else {
		obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
		if obj != nil {
			s.cacheSuicided[accountId] = obj.IsSuicided()
			return s.cacheSuicided[accountId]
		}
	}
	return false
}

func (s *StateAdapter) Exist(accountId meta.AccountID) bool {
	if _, ok := s.cacheAccount[accountId]; ok {
		return ok
	} else {
		obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
		if obj != nil {
			s.cacheAccount[accountId] = *obj.GetAccount()
		}
		return obj != nil
	}

}

func (s *StateAdapter) Empty(accountId meta.AccountID) bool {
	if _, ok := s.cacheAccount[accountId]; ok {
		return false
	} else {
		obj := s.stateDB.GetObject(meta.GetAccountHash(accountId))
		return obj == nil
	}
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (s *StateAdapter) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(s.validRevisions), func(i int) bool {
		return s.validRevisions[i].id >= revid
	})
	if idx == len(s.validRevisions) || s.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := s.validRevisions[idx].journalIndex

	// Replay the journal to undo changes and remove invalidated snapshots
	s.journal.revert(s, snapshot)
	s.validRevisions = s.validRevisions[:idx]
}

func (s *StateAdapter) Snapshot() int {
	id := s.nextRevisionId
	s.nextRevisionId++
	s.validRevisions = append(s.validRevisions, revision{id, s.journal.length()})
	return id
}

func (s *StateAdapter) AddLog(log *meta.Log) {
	s.journal.append(addLogChange{txhash: s.thash})

	log.TxHash = s.thash
	log.BlockHash = s.bhash
	log.Index = s.logSize
	s.logs[s.thash] = append(s.logs[s.thash], log)
	s.logSize++
}

func (s *StateAdapter) GetLogs(hash math.Hash) []*meta.Log {
	return s.logs[hash]
}

func (s *StateAdapter) AddPreimage(math.Hash, []byte) {
	//TODO:may be is useless
}

func (s *StateAdapter) Commit() error {

	//Account
	for k, v := range s.cacheAccount {
		if s.stateDB.GetObject(meta.GetAccountHash(k)) == nil {
			obj := s.stateDB.NewObject(meta.GetAccountHash(k), v)
			s.stateDB.SetObject(obj)
		}
	}

	//storage Tree
	for k, v := range s.cacheAccountState {
		obj := s.stateDB.GetObject(meta.GetAccountHash(k))
		if obj == nil {
			return errors.New("can not find account with cache state in StateDB")
		}
		for key, state := range v {
			obj.SetState(s.stateDB.DataBase(), key, state)
		}
		s.stateDB.SetObject(obj)
	}
	//Code
	for k, v := range s.cacheCode {
		obj := s.stateDB.GetObject(meta.GetAccountHash(k))
		if obj == nil {
			return errors.New("can not find account with cache state in StateDB")
		}
		codeHash := math.HashH(v)
		obj.SetCode(codeHash, v)
		s.stateDB.SetObject(obj)
	}
	//Suicided
	for k, v := range s.cacheSuicided {
		obj := s.stateDB.GetObject(meta.GetAccountHash(k))
		if obj == nil {
			return errors.New("can not find account with cache state in StateDB")
		}
		if v {
			obj.MarkSuicided()
			s.stateDB.SetObject(obj)
		}
	}
	return nil
}

func (s *StateAdapter) GetStateDB() *state.StateDB {
	return s.stateDB
}

func (s *StateAdapter) GetResultTransaction() (meta.Transaction, error) {
	//CoinBase
	contractTx := helper.CreateTempleteTx(config.DefaultTransactionVersion, ContractResultTx)

	subBalance := make(map[meta.AccountID]int64)
	subSum := int64(0)
	addBalance := make(map[meta.AccountID]int64)
	addSum := int64(0)
	for _, transfer := range s.transfers {
		switch transfer.Code {
		case params.CoinBase:
			addBalance[*transfer.to] += transfer.value.Int64()
			addSum += transfer.value.Int64()
		case params.AddZero:
		case params.Refund:
			addBalance[*transfer.to] += transfer.value.Int64()
			addSum += transfer.value.Int64()
		case params.BuyGas:
			subBalance[*transfer.from] += transfer.value.Int64()
			subSum += transfer.value.Int64()
		case params.Suicide:
			subBalance[*transfer.from] += transfer.value.Int64()
			subSum += transfer.value.Int64()
			addBalance[*transfer.to] += transfer.value.Int64()
			addSum += transfer.value.Int64()
		case params.Normal:
			subBalance[*transfer.from] += transfer.value.Int64()
			subSum += transfer.value.Int64()
			addBalance[*transfer.to] += transfer.value.Int64()
			addSum += transfer.value.Int64()
		}
	}

	//make from
	for k, v := range subBalance {
		obj := s.stateDB.GetObject(meta.GetAccountHash(k))
		if obj != nil {
			//make from
			account := obj.GetAccount()
			amount := meta.NewAmount(v)
			fc, fAmount, err := account.MakeFromCoin(amount, uint32(s.blockHeight))
			if err != nil {
				return *contractTx, err
			}
			contractTx.AddFromCoin(*fc)
			backChangeAmount := fAmount.Subtraction(*amount)

			if backChangeAmount.GetInt64() > 0 {
				contractTx.SetTo(account.Id, *backChangeAmount)
			}
		}
	}

	//make to
	for k, v := range addBalance {
		contractTx.SetTo(k, *meta.NewAmount(v))
	}
	helper.SortTransaction(contractTx)
	log.Debug("state_adapter", "contract result tx", contractTx.String())
	log.Debug("state_adapter", "addSum", addSum, "subSum", subSum)
	return *contractTx, nil
}

// Retrieve a state object or create a new state object if nil.
func (s *StateAdapter) GetOrNewStateObject(addr meta.AccountID) *state.StateObject {
	stateObject := s.stateDB.GetObject(meta.GetAccountHash(addr))
	if stateObject == nil || stateObject.IsDeleted() {
		stateObject = s.stateDB.NewObject(meta.GetAccountHash(addr), meta.Account{})
	}
	return stateObject
}

func BalanceChangeToTransaction(account meta.Account, transfer BalanceChange, tx *meta.Transaction, blockHeight int64) error {
	amount := meta.NewAmount(transfer.value.Int64())
	fc, fAmount, err := account.MakeFromCoin(amount, uint32(blockHeight))
	if err != nil {
		return err
	}
	tx.AddFromCoin(*fc)

	//make to
	backChange := helper.CreateToCoin(account.Id, fAmount.Subtraction(*amount))
	if backChange.Value.GetInt64() > 0 {
		tx.AddToCoin(*backChange)
	}
	if transfer.to != nil {
		tx.SetTo(*transfer.to, *meta.NewAmount(transfer.value.Int64()))
	}

	return nil
}
