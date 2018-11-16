package storage

import (
	"errors"
	"sync"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"

	poameta "github.com/linkchain/core/meta"
	//"github.com/linkchain/node"
	"github.com/linkchain/config"
)

type StateDB struct {
	accountMtx sync.RWMutex
	accounts   map[string]poameta.Account
	blockCache map[string]map[string]poameta.Account
}

func (s *StateDB) Setup(i interface{}) bool {
	log.Info("StateDB init...")
	s.accounts = make(map[string]poameta.Account)
	s.blockCache = make(map[string]map[string]poameta.Account)
	return true
}

func (s *StateDB) Start() bool {
	log.Info("StateDB Start...")
	return true
}

func (s *StateDB) Stop() {
	log.Info("StateDB Stop...")
}

// get account by key with R/W lock.
func (s *StateDB) GetAccount(id meta.AccountID) (*meta.Account, bool) {
	s.accountMtx.RLock()
	defer s.accountMtx.RUnlock()
	value, ok := s.accounts[id.String()]
	return &value, ok
}

// set account by key with R/W lock.
func (s *StateDB) SetAccount(account meta.Account) error {
	s.accountMtx.Lock()
	defer s.accountMtx.Unlock()
	_, ok := s.accounts[account.GetAccountID().String()]
	if ok {
		return errors.New("account has already exist")
	}
	s.accounts[account.GetAccountID().String()] = account
	return nil
}

func (s *StateDB) RemoveAccount(id meta.AccountID) {
	s.accountMtx.Lock()
	defer s.accountMtx.Unlock()
	delete(s.accounts, id.String())
}

func (s *StateDB) GetAllAccount() {
	s.accountMtx.Lock()
	defer s.accountMtx.Unlock()
	for _, v := range s.accounts {
		log.Info("AccountManage", "account", v.GetAccountID().String(), "amount", v.GetAmount().GetInt64())
		for _, u := range v.UTXOs {
			log.Info("AccountManage", "Tickets", u.String())
		}
	}

}

func (s *StateDB) UpdateAccountsByBlock(block *meta.Block) error {
	s.accountMtx.Lock()
	defer s.accountMtx.Unlock()

	txs := block.GetTxs()
	processCache := make(map[string]poameta.Account, 0)

	//get related account in from/to
	for _, tx := range txs {

		tcs := tx.GetToCoins()
		for i, _ := range tcs {
			if !tcs[i].CheckValue() {
				return errors.New("transaction toCoin-Value need plus 0")
			}
			if cacheA, ok := s.accounts[tcs[i].GetId().String()]; ok {
				processCache[cacheA.GetAccountID().String()] = cacheA
			}
		}

		if tx.GetType() == config.CoinBaseTx {
			continue
		}

		fcs := tx.GetFromCoins()
		for _, fc := range fcs {
			cacheA, ok := s.accounts[fc.GetId().String()]
			if !ok {
				return errors.New("acccount can not find fromCoin-Account")
			}
			if ok = cacheA.CheckFromCoin(&fc); !ok {
				return errors.New("acccount can not contain fromCoin")
			}
			processCache[cacheA.GetAccountID().String()] = cacheA
		}

	}
	coinBase := meta.NewAmount(0)
	txFee := meta.NewAmount(0)
	for _, tx := range txs {
		fcs := tx.GetFromCoins()
		tcs := tx.GetToCoins()

		if tx.GetType() != config.CoinBaseTx {
			fcValue := meta.NewAmount(0)
			tcValue := tx.GetToValue()
			for _, fc := range fcs {
				cacheA, _ := processCache[fc.GetId().String()]

				if ok := cacheA.CheckFromCoin(&fc); !ok {
					return errors.New("cache acccount can not contain fromCoin")
				}
				value, err := cacheA.GetFromCoinValue(&fc)
				if err != nil {
					return err
				}
				fcValue.Addition(*value)

				if err = cacheA.RemoveUTXOByFromCoin(&fc); err != nil {
					return err
				}
				processCache[cacheA.GetAccountID().String()] = cacheA
			}

			if fcValue.IsLessThan(*tcValue) {
				return errors.New("the tx from value < to value")
			}

			txFee.Addition(*fcValue.Subtraction(*tcValue))
		} else {
			coinBase.Addition(*tx.GetToValue())
		}

		for index := range tcs {
			cacheA, err := processCache[tcs[index].GetId().String()]
			if !err {
				//cacheA = *node.CreateTempleteAccount(tcs[index].GetId())
			}
			nTicket := poameta.NewTicket(*tx.GetTxID(), uint32(index))
			nUTXO := poameta.NewUTXO(nTicket, block.GetHeight(), block.GetHeight(), *tcs[index].GetValue())
			cacheA.UTXOs = append(cacheA.UTXOs, *nUTXO)
			processCache[cacheA.GetAccountID().String()] = cacheA
		}
	}

	//Check coinbase money
	if coinBase.Subtraction(*meta.NewAmount(config.DefaultBlockReward)).GetInt64() != txFee.GetInt64() && len(txs) > 0 {
		return errors.New("coinbase reward is error")
	}

	//storage blockCache
	blockCache := make(map[string]poameta.Account, 0)
	for k, v := range s.accounts {
		blockCache[k] = v
	}
	s.blockCache[block.GetBlockID().String()] = blockCache

	//update Account
	for k, v := range processCache {
		s.accounts[k] = v
	}

	return nil
}

func (s *StateDB) RollBack(block *meta.Block) error {
	s.accountMtx.Lock()
	defer s.accountMtx.Unlock()
	rollbackCache, ok := s.blockCache[block.GetBlockID().String()]
	if !ok {
		return errors.New("stateDB can not find this block cache")
	}

	for k := range s.accounts {
		delete(s.accounts, k)
	}

	for k, v := range rollbackCache {
		s.accounts[k] = v
	}

	delete(s.blockCache, block.GetBlockID().String())
	return nil
}
