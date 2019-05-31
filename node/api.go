package node

import (
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/util/event"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/interpreter"
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

func (a *PublicNodeAPI) ProcessBlock(block *meta.Block) error {
	return a.n.blockchain.ProcessBlock(block)
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
