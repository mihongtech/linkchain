package contract

import (
	"github.com/linkchain/common/math"
	"github.com/linkchain/core/meta"
)

type cacheState struct {
	account  meta.Account
	code     []byte
	balance  meta.Amount
	suicided bool
	storage  map[math.Hash]math.Hash
}

func (cs *cacheState) GetAccount() *meta.Account {
	return &cs.account
}

func (cs *cacheState) SetAccount(account meta.Account) {
	cs.account = account
}

func (cs *cacheState) GetCode() []byte {
	return cs.code
}

func (cs *cacheState) SetCode(code []byte) {
	copy(cs.code, code)
}

func (cs *cacheState) SetBalance(amount meta.Amount) {
	cs.balance = amount
}

func (cs *cacheState) GetBalance() meta.Amount {
	return cs.balance
}

func (cs *cacheState) GetSuicided() bool {
	return cs.suicided
}

func (cs *cacheState) SetSuicided(suicided bool) {
	cs.suicided = suicided
}

func (cs *cacheState) SetStorage(key math.Hash, value math.Hash) {
	cs.storage[key] = value
}

func (cs *cacheState) GetStorage(key math.Hash) math.Hash {
	if value, ok := cs.storage[key]; ok {
		return value
	}
	return math.Hash{}
}
