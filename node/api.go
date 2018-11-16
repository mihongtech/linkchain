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

func (a *PublicNodeAPI) ProcessBlock(block *meta.Block) error {
	return a.n.processBlock(block)
}

func (a *PublicNodeAPI) CheckBlock(block *meta.Block) bool {
	return a.n.checkBlock(block)
}

func (a *PublicNodeAPI) GetBlockByID(hash meta.BlockID) (*meta.Block, error) {
	return a.n.getBlockByID(hash)
}

func (a *PublicNodeAPI) GetBlockByHeight(height uint32) (*meta.Block, error) {
	return a.n.getBlockByHeight(height)
}

func (a *PublicNodeAPI) AddTransaction(tx *meta.Transaction) error {
	return a.n.addTransaction(tx)
}

//account
func (a *PublicNodeAPI) GetAccount(id meta.AccountID) (meta.Account, error) {
	return a.n.getAccount(id)
}

//chain
func (a *PublicNodeAPI) GetBlockChainInfo() string {
	return a.n.getBlockChainInfo()
}
