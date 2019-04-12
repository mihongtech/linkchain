package state

import (
	"bytes"
	"fmt"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/storage"
)

type Code []byte

func (self Code) String() string {
	return string(self) //strings.Join(Disassemble(self), " ")
}

type Storage map[math.Hash]math.Hash

func (self Storage) String() (str string) {
	for key, value := range self {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

func (self Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range self {
		cpy[key] = value
	}

	return cpy
}

type StateObject struct {
	key  math.Hash    //Key is the key of the StateObject, imaging statedb as a k,v database
	data meta.Account //Data is the value of StateObject

	db *StateDB

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by stateDB.Commit.
	dbErr error

	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access
	code Code // contract bytecode, which gets set when code is loaded

	originStorage Storage // Storage cache of original entries to dedup rewrites
	dirtyStorage  Storage // Storage entries that need to be flushed to disk

	// Cache flags.
	// When an object is marked suicided it will be delete from the trie
	// during the "update" phase of the state transition.
	dirtyCode bool // true if the code was updated
	suicided  bool
	deleted   bool
}

// newObject creates a state rpcobject.
func newObject(db *StateDB, key math.Hash, data meta.Account) *StateObject {
	return &StateObject{
		db:            db,
		key:           key,
		data:          data,
		originStorage: make(Storage),
		dirtyStorage:  make(Storage),
	}
}

// empty returns whether the account is considered empty.
func (s *StateObject) empty() bool {
	return len(s.data.UTXOs) <= 0
}

// Returns the address of the contract/account
func (obj *StateObject) Address() meta.AccountID {
	return obj.data.Id
}

// setError remembers the first non-nil error it is called with.
func (obj *StateObject) setError(err error) {
	if obj.dbErr == nil {
		obj.dbErr = err
	}
}

func (obj *StateObject) MarkSuicided() {
	obj.suicided = true
}

func (obj *StateObject) IsSuicided() bool {
	return obj.suicided
}

func (obj *StateObject) IsDeleted() bool {
	return obj.deleted
}

func (obj *StateObject) getTrie(db Database) Trie {
	if obj.trie == nil {
		var err error
		obj.trie, err = db.OpenStorageTrie(meta.GetAccountHash(obj.GetAccount().Id), obj.data.StorageRoot)
		if err != nil {
			obj.trie, _ = db.OpenStorageTrie(meta.GetAccountHash(obj.GetAccount().Id), math.Hash{})
			obj.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
	}
	return obj.trie
}

// GetState retrieves a value from the account storage trie.
func (obj *StateObject) GetState(db Database, key math.Hash) math.Hash {
	// If we have a dirty value for this state entry, return it
	value, dirty := obj.dirtyStorage[key]
	if dirty {
		return value
	}
	// Otherwise return the entry's original value
	return obj.GetCommittedState(db, key)
}

// GetCommittedState retrieves a value from the committed account storage trie.
func (obj *StateObject) GetCommittedState(db Database, key math.Hash) math.Hash {
	// If we have the original value cached, return that
	value, cached := obj.originStorage[key]
	if cached {
		return value
	}
	// Otherwise load the value from the database
	enc, err := obj.getTrie(db).TryGet(key[:])
	if err != nil {
		obj.setError(err)
		return math.Hash{}
	}
	if len(enc) > 0 {
		/*_, content, _, err := rlp.Split(enc)
		if err != nil {
			obj.setError(err)
		}*/
		value.SetBytes(enc)
	}
	obj.originStorage[key] = value
	return value
}

// SetState updates a value in account storage.
func (obj *StateObject) SetState(db Database, key, value math.Hash) {
	// If the new value is the same as old, don't set
	prev := obj.GetState(db, key)
	if prev == value {
		return
	}

	obj.setState(key, value)
}

func (obj *StateObject) setState(key, value math.Hash) {
	obj.dirtyStorage[key] = value
}

// updateTrie writes cached storage modifications into the object's storage trie.
func (obj *StateObject) updateTrie(db Database) Trie {
	tr := obj.getTrie(db)
	for key, value := range obj.dirtyStorage {
		delete(obj.dirtyStorage, key)

		// Skip noop changes, persist actual changes
		if value == obj.originStorage[key] {
			continue
		}
		obj.originStorage[key] = value

		if (value == math.Hash{}) {
			obj.setError(tr.TryDelete(key[:]))
			continue
		}
		// Encoding []byte cannot fail, ok to ignore the error.
		//v, _ := rlp.EncodeToBytes(bytes.TrimLeft(value[:], "\x00"))
		v := bytes.TrimLeft(value[:], "\x00")
		obj.setError(tr.TryUpdate(key[:], v))
	}
	return tr
}

// UpdateRoot sets the trie root to the current root hash of
func (obj *StateObject) updateRoot(db Database) {
	obj.updateTrie(db)
	obj.data.StorageRoot = obj.trie.Hash()
}

// CommitTrie the storage trie of the object to db.
// This updates the trie root.
func (obj *StateObject) CommitTrie(db Database) error {
	obj.updateTrie(db)
	if obj.dbErr != nil {
		return obj.dbErr
	}
	root, err := obj.trie.Commit(nil)
	if err == nil {
		obj.data.StorageRoot = root
	}
	return err
}

// Code returns the contract code associated with this object, if any.
func (obj *StateObject) Code() []byte {
	if obj.code != nil {
		return obj.code
	}
	if obj.GetAccount().CodeHash.IsEmpty() {
		return nil
	}

	codeHash := obj.GetAccount().CodeHash
	code, err := storage.GetCode(obj.db.GetDB(), codeHash)
	if err != nil {
		obj.setError(fmt.Errorf("can't load code hash %x: %v", obj.GetAccount().CodeHash.String(), err))
		return code
	}
	/*code, err := db.ContractCode(obj.addrHash, common.BytesToHash(obj.CodeHash()))
	if err != nil {
		obj.setError(fmt.Errorf("can't load code hash %x: %v", obj.CodeHash(), err))
	}*/
	obj.code = code
	return code
}

func (obj *StateObject) SetCode(codeHash math.Hash, code []byte) {
	obj.setCode(codeHash, code)
}

func (obj *StateObject) setCode(codeHash math.Hash, code []byte) {
	obj.code = code
	obj.data.CodeHash = codeHash
	obj.dirtyCode = true
}

func (obj *StateObject) GetAccount() *meta.Account {
	return &obj.data
}
