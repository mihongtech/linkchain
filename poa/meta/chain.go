package poameta

import (
	"errors"

	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
)

type Chain struct {
	Blocks       []Block
	IsInComplete bool
}

func NewPOAChain(startNode block.IBlock, endNode block.IBlock) Chain {
	chainNode := make([]Block, 0)
	isInComplete := false
	if startNode != nil {
		chainNode = append(chainNode, *(startNode.(*Block)))
		isInComplete = true
	}
	if endNode != nil {
		chainNode = append(chainNode, *(endNode.(*Block)))
	}
	return Chain{Blocks: chainNode, IsInComplete: isInComplete}
}

func (bc *Chain) AddNewBlock(block block.IBlock) {
	bc.Blocks = append(bc.Blocks, *(block.(*Block)))
}

/**invalidate block by block*/
func (bc *Chain) Rollback(block.IBlock) {

}

/**invalidate block by height*/
func (bc *Chain) RollbackAtHeight(int) {

}

func (bc *Chain) GetLastBlock() block.IBlock {
	return &bc.Blocks[len(bc.Blocks)-1]
}

func (bc *Chain) GetFirstBlock() block.IBlock {
	return &bc.Blocks[0]
}

func (bc *Chain) GetHeight() uint32 {
	return bc.GetLastBlock().GetHeight()
}

func (bc *Chain) GetBlockByID(id meta.BlockID) block.IBlock {
	//TODO need to sorage
	return nil
}

func (bc *Chain) GetBlockByHeight(int) block.IBlock {
	//TODO need to sorage
	return nil
}

func (bc *Chain) UpdateChainTop(topBlock block.IBlock) error {
	if topBlock.GetHeight() < bc.GetHeight() {
		return errors.New("BlockChain the topnode is not height than current chain")
	}
	lastNode := NewPOAChainNode(bc.GetLastBlock())
	topNode := NewPOAChainNode(topBlock)
	if topNode.CheckPrev(lastNode) {
		bc.AddNewBlock(topBlock)
		return nil
	} else {
		return errors.New("BlockChain the topBlock is not next of lastBlock chain")
	}
}

func (bc *Chain) AddChain(newChain Chain) error {
	if bc.CanLink(newChain) {
		for _, block := range newChain.Blocks {
			bc.UpdateChainTop(&block)
		}
		return nil
	} else {
		return errors.New("BlockChain the topBlock is not next of lastBlock chain")
	}
}

func (bc *Chain) CanLink(newChain Chain) bool {
	if !newChain.IsInComplete {
		return false
	}
	topNode := NewPOAChainNode(&newChain.Blocks[1])
	lastNode := NewPOAChainNode(bc.GetLastBlock())
	return topNode.CheckPrev(lastNode)
}

/**
GetNewChain
get a new chain from this chain
*/
func (bc *Chain) GetNewChain(forkBlock block.IBlock) Chain {
	newChain := Chain{IsInComplete: true}
	for _, block := range bc.Blocks {
		if block.GetHeight() < forkBlock.GetHeight() {
			newChain.AddNewBlock(&block)
		}
	}
	newChain.AddNewBlock(forkBlock)
	return newChain
}

func GetChainHeight(bc *Chain) uint32 {
	return bc.GetHeight()
}
