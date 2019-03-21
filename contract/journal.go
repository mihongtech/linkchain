package contract

import (
	"github.com/linkchain/common/math"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/storage/state"
	"math/big"
)

// journalEntry is a modification entry in the state change journal that can be
// reverted on demand.
type journalEntry interface {
	// revert undoes the changes introduced by this journal entry.
	revert(db *StateAdapter)

	// dirtied returns the Ethereum address modified by this journal entry.
	dirtied() *meta.AccountID
}

// journal contains the list of state modifications applied since the last state
// commit. These are tracked to be able to be reverted in case of an execution
// exception or revertal request.
type journal struct {
	entries []journalEntry         // Current changes tracked by the journal
	dirties map[meta.AccountID]int // Dirty accounts and the number of changes
}

// newJournal create a new initialized journal.
func newJournal() *journal {
	return &journal{
		dirties: make(map[meta.AccountID]int),
	}
}

// append inserts a new modification entry to the end of the change journal.
func (j *journal) append(entry journalEntry) {
	j.entries = append(j.entries, entry)
	if addr := entry.dirtied(); addr != nil {
		j.dirties[*addr]++
	}
}

// revert undoes a batch of journalled modifications along with any reverted
// dirty handling too.
func (j *journal) revert(statedb *StateAdapter, snapshot int) {
	for i := len(j.entries) - 1; i >= snapshot; i-- {
		// Undo the changes made by the operation
		j.entries[i].revert(statedb)

		// Drop any dirty tracking induced by the change
		if addr := j.entries[i].dirtied(); addr != nil {
			if j.dirties[*addr]--; j.dirties[*addr] == 0 {
				delete(j.dirties, *addr)
			}
		}
	}
	j.entries = j.entries[:snapshot]
}

// dirty explicitly sets an address to dirty, even if the change entries would
// otherwise suggest it as clean. This method is an ugly hack to handle the RIPEMD
// precompile consensus exception.
func (j *journal) dirty(addr meta.AccountID) {
	j.dirties[addr]++
}

// length returns the current number of entries in the journal.
func (j *journal) length() int {
	return len(j.entries)
}

type (
	// Changes to the account trie.
	createObjectChange struct {
		account *meta.AccountID
	}
	resetObjectChange struct {
		prev *state.StateObject
	}
	suicideChange struct {
		account *meta.AccountID
		prev    bool // whether account had already suicided
	}

	// Changes to individual accounts.
	balanceChange struct {
		account *meta.AccountID
		prev    *big.Int
	}

	storageChange struct {
		account       *meta.AccountID
		key, prevalue math.Hash
	}
	codeChange struct {
		account            *meta.AccountID
		prevcode, prevhash []byte
	}

	// Changes to other state values.
	refundChange struct {
		prev uint64
	}
	addLogChange struct {
		txhash math.Hash
	}
	addPreimageChange struct {
		hash math.Hash
	}
	transferChange struct {
		transferIndex int
	}
)

func (ch transferChange) revert(s *StateAdapter) {
	s.transfers = append(s.transfers[:ch.transferIndex], s.transfers[ch.transferIndex+1:]...)
}

func (ch transferChange) dirtied() *meta.AccountID {
	return nil
}

func (ch createObjectChange) revert(s *StateAdapter) {
	delete(s.cacheAccount, *ch.account)
}

func (ch createObjectChange) dirtied() *meta.AccountID {
	return ch.account
}

func (ch resetObjectChange) revert(s *StateAdapter) {
	//TODO
	//s.setStateObject(ch.prev)
}

func (ch resetObjectChange) dirtied() *meta.AccountID {
	return nil
}

func (ch suicideChange) revert(s *StateAdapter) {
	s.cacheSuicided[*ch.account] = ch.prev
}

func (ch suicideChange) dirtied() *meta.AccountID {
	return ch.account
}

var ripemd, _ = meta.HexToAccountID("0000000000000000000000000000000000000003")

func (ch balanceChange) revert(s *StateAdapter) {
	s.setBalance(*ch.account, ch.prev)
}

func (ch balanceChange) dirtied() *meta.AccountID {
	return ch.account
}

func (ch codeChange) revert(s *StateAdapter) {
	s.SetCode(*ch.account, ch.prevcode)
}

func (ch codeChange) dirtied() *meta.AccountID {
	return ch.account
}

func (ch storageChange) revert(s *StateAdapter) {
	accountStates, _ := s.cacheAccountState[*ch.account]
	accountStates[ch.key] = ch.prevalue
	s.cacheAccountState[*ch.account] = accountStates
}

func (ch storageChange) dirtied() *meta.AccountID {
	return ch.account
}

func (ch refundChange) revert(s *StateAdapter) {
	s.refund = ch.prev
}

func (ch refundChange) dirtied() *meta.AccountID {
	return nil
}

func (ch addLogChange) revert(s *StateAdapter) {
	logs := s.logs[ch.txhash]
	if len(logs) == 1 {
		delete(s.logs, ch.txhash)
	} else {
		s.logs[ch.txhash] = logs[:len(logs)-1]
	}
	s.logSize--
}

func (ch addLogChange) dirtied() *meta.AccountID {
	return nil
}

func (ch addPreimageChange) revert(s *StateAdapter) {
	//TODO:
	//delete(s.preimages, ch.hash)
}

func (ch addPreimageChange) dirtied() *meta.AccountID {
	return nil
}
