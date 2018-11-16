package node

import (
	"github.com/linkchain/common/util/event"
	"github.com/linkchain/core/meta"
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

func (a *PublicNodeAPI) GetTxEvent() *event.Feed {
	return a.n.newTxEvent
}

//block
func (a *PublicNodeAPI) GetBestBlock() *meta.Block {
	return a.n.getBestBlock()
}

func (a *PublicNodeAPI) HasBlock(hash meta.BlockID) bool {
	return a.n.hasBlock(hash)
}

func (a *PublicNodeAPI) GetBlockByID(hash meta.BlockID) (*meta.Block, error) {
	return a.n.getBlockByID(hash)
}

func (a *PublicNodeAPI) GetBlockByHeight(height uint32) (*meta.Block, error) {
	return a.n.getBlockByHeight(height)
}

func (a *PublicNodeAPI) ProcessBlock(block *meta.Block) error {
	return a.n.processBlock(block)
}

func (a *PublicNodeAPI) CheckBlock(block *meta.Block) bool {
	return a.n.checkBlock(block)
}

//account
func (a *PublicNodeAPI) GetAccount(id meta.AccountID) (meta.Account, error) {
	return a.n.getAccount(id)
}

func (a *PublicNodeAPI) GetAccountInfo() {
	a.n.getAccountInfo()
}

//chain
func (a *PublicNodeAPI) GetBlockChainInfo() string {
	return a.n.getBlockChainInfo()
}

//tx

func (a *PublicNodeAPI) GetAllTransaction() []meta.Transaction {
	return a.n.getAllTransaction()
}

func (a *PublicNodeAPI) AddTransaction(tx *meta.Transaction) error {
	return a.n.addTransaction(tx)
}

func (a *PublicNodeAPI) ProcessTx(tx *meta.Transaction) error {
	return a.n.processTx(tx)
}

func (a *PublicNodeAPI) CheckTx(tx *meta.Transaction) bool {
	return a.n.checkTx(tx)
}
