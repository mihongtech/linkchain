package node

import (
	"errors"

	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/util/event"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/consensus"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/storage/state"
)

type PublicNodeAPI struct {
	n *Node
}

func (n *PublicNodeAPI) Setup(i interface{}) bool {
	return true
}

func (n *PublicNodeAPI) Start() bool {
	return true
}

func (n *PublicNodeAPI) Stop() {
}

func NewPublicNodeAPI(n *Node) *PublicNodeAPI {
	return &PublicNodeAPI{n}
}

/*
	APIs
*/

//event
func (a *PublicNodeAPI) GetBlockEvent() *event.TypeMux {
	return a.n.newBlockEvent
}

func (a *PublicNodeAPI) GetAccountEvent() *event.TypeMux {
	return a.n.newAccountEvent
}

func (a *PublicNodeAPI) GetTxPoolEvent() *event.TypeMux {
	return a.n.txPoolEvent
}

func (a *PublicNodeAPI) GetTxEvent() *event.Feed {
	return a.n.newTxEvent
}

//block
func (a *PublicNodeAPI) GetBestBlock() *meta.Block {
	return a.n.blockchain.CurrentBlock()
}

func (a *PublicNodeAPI) HasBlock(hash meta.BlockID) bool {
	return a.n.blockchain.HasBlock(hash)
}

func (a *PublicNodeAPI) GetBlockByID(hash meta.BlockID) (*meta.Block, error) {
	return a.n.blockchain.GetBlockByID(hash)
}

func (a *PublicNodeAPI) GetHeader(hash math.Hash, height uint64) *meta.BlockHeader {
	block, err := a.n.blockchain.GetBlockByID(hash)
	if err != nil {
		return nil
	}
	return &block.Header
}

func (a *PublicNodeAPI) GetBlockByHeight(height uint32) (*meta.Block, error) {
	return a.n.blockchain.GetBlockByHeight(height)
}

func (a *PublicNodeAPI) GetChainConfig() *config.ChainConfig {
	return a.n.blockchain.Config()
}

func (a *PublicNodeAPI) ProcessBlock(block *meta.Block) error {
	return a.n.blockchain.ProcessBlock(block)
}

func (a *PublicNodeAPI) CheckBlock(block *meta.Block) error {
	return a.n.checkBlock(block)
}

//account
func (a *PublicNodeAPI) GetAccount(id meta.AccountID) (meta.Account, error) {
	return a.n.getAccount(id)
}

// tx
func (a *PublicNodeAPI) GetTXByID(hash meta.TxID) (*meta.Transaction, math.Hash, uint64, uint64) {
	return a.n.getTxByID(hash)
}

//chain
func (a *PublicNodeAPI) GetBlockChainInfo() interface{} {
	// TODO: implement me
	block := a.n.blockchain.CurrentBlock()
	info := &meta.ChainInfo{
		BestHeight: block.GetHeight(),
		BestHash:   block.GetBlockID().GetString(),
		ChainId:    int(a.n.blockchain.GetChainID().Int64())}
	return info
}

//offchain
func (a *PublicNodeAPI) GetOffChain() interpreter.OffChain {
	return a.n.offchain
}

//consensus

func (a *PublicNodeAPI) GetEngine() consensus.Engine {
	return a.n.engine
}

func (a *PublicNodeAPI) Engine() consensus.Engine {
	return a.n.engine
}

func (a *PublicNodeAPI) GetCode(id meta.AccountID) ([]byte, error) {
	state, err := a.n.blockchain.State()
	if err != nil {
		return nil, err
	}
	obj := state.GetObject(meta.GetAccountHash(id))
	if obj == nil {
		log.Error("Get code failed", "id", id)
		return nil, errors.New("can not find account in GetCode()")
	}
	return obj.Code(), nil
}

func (a *PublicNodeAPI) GetState(id meta.AccountID, key math.Hash) ([]byte, error) {
	state, err := a.n.blockchain.State()
	if err != nil {
		return nil, err
	}
	obj := state.GetObject(meta.GetAccountHash(id))
	if obj == nil {
		return nil, errors.New("can not find account in GetState()")
	}
	value := obj.GetState(state.DataBase(), key)
	return value.CloneBytes(), nil
}

func (a *PublicNodeAPI) StateAt(root math.Hash) (*state.StateDB, error) {
	return a.n.blockchain.StateAt(root)
}

func (a *PublicNodeAPI) GetReceiptsByHash(hash math.Hash) core.Receipts {
	return a.n.blockchain.GetReceiptsByHash(hash)
}

//Miner
func (a *PublicNodeAPI) ExecuteBlock(block *meta.Block) (error, []interpreter.Result, math.Hash, *meta.Amount) {
	return a.n.blockchain.executeBlock(block)
}

func (a *PublicNodeAPI) CalcNextRequiredDifficulty() (uint32, error) {
	return a.n.blockchain.CalcNextRequiredDifficulty()
}
