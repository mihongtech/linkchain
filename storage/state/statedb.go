package state

import (
	"fmt"
	"github.com/linkchain/storage"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/protobuf"
)

//stateDB manage a MPT
type StateDB struct {
	db   Database
	lcdb lcdb.Database
	trie Trie

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects      map[string]*StateObject
	stateObjectsDirty map[string]struct{}

	// DB error
	dbErr error

	lock sync.RWMutex

	hash math.Hash
}

// New create a new state from a given trie.
func New(root math.Hash, db lcdb.Database) (*StateDB, error) {
	sdb := NewDatabase(db)
	tr, err := sdb.OpenTrie(root)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		db:                sdb,
		lcdb:              db,
		trie:              tr,
		stateObjects:      make(map[string]*StateObject),
		stateObjectsDirty: make(map[string]struct{}),
	}, nil
}

// setError remembers the first non-nil error it is called with.
func (s *StateDB) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *StateDB) Error() error {
	return s.dbErr
}

// GetObject retrieve a StateObject from stateDB
func (s *StateDB) GetObject(key math.Hash) *StateObject {
	s.lock.RLock()
	defer s.lock.RUnlock()
	addr := key.String()
	// Prefer 'live' objects.
	if obj := s.stateObjects[addr]; obj != nil {
		return obj
	}

	// Load the rpcobject from the database.
	data, err := s.trie.TryGet(key[:])
	if len(data) == 0 {
		s.setError(err)
		return nil
	}

	pa := &protobuf.Account{}
	if err = proto.Unmarshal(data, pa); err != nil {
		s.setError(err)
		log.Error("statedb", "cannot unmarshal rpcobject, key:", key.String())
		return nil
	}

	a := meta.Account{}
	if err = a.Deserialize(pa); err != nil {
		s.setError(err)
		log.Error("statedb", "cannot deserialize account,  key:", key.String())
		return nil
	}

	// Insert into the live set.
	obj := s.NewObject(key, a)
	return obj
}

func (s *StateDB) NewObject(key math.Hash, data meta.Account) *StateObject {
	obj := newObject(s, key, data)
	s.addCache(obj)
	return obj
}

//addCache add a StateObject into intermediate cache
func (s *StateDB) addCache(obj *StateObject) {
	s.stateObjects[obj.key.String()] = obj
}

//SetObject mark state as dirty
func (s *StateDB) SetObject(obj *StateObject) {
	s.stateObjectsDirty[obj.key.String()] = struct{}{}
}

//GetRootHash return the root hash of MPT
func (s *StateDB) GetRootHash() math.Hash {
	if s.hash.IsEmpty() {
		s.hash = s.trie.Hash()
	}

	return s.hash
}

// Finalise finalises the state by removing the self destructed objects
// and clears the journal as well as the refunds.
func (s *StateDB) Finalise() {
	for addr := range s.stateObjectsDirty {
		stateObject, exist := s.stateObjects[addr]
		if !exist {
			// ripeMD is 'touched' at block 1714175, in tx 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
			// That tx goes out of gas, and although the notion of 'touched' does not exist there, the
			// touch-event will still be recorded in the journal. Since ripeMD is a special snowflake,
			// it will persist in the journal even though the journal is reverted. In this special circumstance,
			// it may exist in `s.journal.dirties` but not in `s.stateObjects`.
			// Thus, we can safely ignore it here
			continue
		}
		if stateObject.suicided { //|| (stateObject.empty())
			s.deleteStateObject(stateObject)
		} else {
			stateObject.updateRoot(s.db)
			s.updateStateObject(stateObject)
		}
	}
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (s *StateDB) IntermediateRoot() math.Hash {
	s.Finalise()
	return s.trie.Hash()
}

//Commit write all StateObject to MPT, then finalize it to disk
func (s *StateDB) Commit() (math.Hash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	// Commit objects to the trie.
	for addr, _ := range s.stateObjectsDirty {
		stateObject := s.stateObjects[addr]
		// Update the rpcobject in the main account trie.
		if stateObject.suicided { //|| (stateObject.empty())
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			s.deleteStateObject(stateObject)
		} else {
			// Write any contract code associated with the state object
			if stateObject.code != nil && stateObject.dirtyCode {
				storage.WriteCode(s.lcdb, stateObject.data.CodeHash, stateObject.code)
				stateObject.dirtyCode = false
			}
			// Write any storage changes in the state object to its storage trie.
			if err := stateObject.CommitTrie(s.db); err != nil {
				return math.Hash{}, err
			}

			s.updateStateObject(stateObject)
		}

		delete(s.stateObjectsDirty, addr)
	}

	// Write trie changes.
	root, err := s.trie.Commit(nil)
	s.hash = root

	return root, err
}

//updateStateObject writes the given rpcobject to the trie.
func (s *StateDB) updateStateObject(obj *StateObject) {
	key := obj.key
	ser := obj.data.Serialize()
	data, err := proto.Marshal(ser)

	if err != nil {
		panic(fmt.Errorf("can't encode rpcobject at %x: %v", key[:], err))
	}

	s.setError(s.trie.TryUpdate(key[:], data))
}

// deleteStateObject removes the given object from the state trie.
func (self *StateDB) deleteStateObject(stateObject *StateObject) {
	stateObject.deleted = true
	addr := stateObject.key
	self.setError(self.trie.TryDelete(addr[:]))
}

func (s *StateDB) GetDB() lcdb.Database {
	return s.lcdb
}

func (s *StateDB) DataBase() Database {
	return s.db
}
