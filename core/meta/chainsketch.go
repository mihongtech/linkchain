package meta

import (
	"errors"
)

type ChainSketch struct {
	Blocks       []Block
	IsInComplete bool
}

func NewPOAChain(startNode *Block, endNode *Block) *ChainSketch {
	chainNode := make([]Block, 0)
	isInComplete := false
	if startNode != nil {
		chainNode = append(chainNode, *startNode)
		isInComplete = true
	}
	if endNode != nil {
		chainNode = append(chainNode, *endNode)
	}
	return &ChainSketch{Blocks: chainNode, IsInComplete: isInComplete}
}

func (c *ChainSketch) AddNewBlock(block *Block) {
	c.Blocks = append(c.Blocks, *block)
}

/**invalidate block by block*/
func (c *ChainSketch) Rollback(*Block) {

}

/**invalidate block by height*/
func (c *ChainSketch) RollbackAtHeight(int) {

}

func (c *ChainSketch) GetLastBlock() *Block {
	return &c.Blocks[len(c.Blocks)-1]
}

func (c *ChainSketch) GetFirstBlock() *Block {
	return &c.Blocks[0]
}

func (c *ChainSketch) GetHeight() uint32 {
	return c.GetLastBlock().GetHeight()
}

func (c *ChainSketch) UpdateChainTop(topBlock *Block) error {
	if topBlock.GetHeight() < c.GetHeight() {
		return errors.New("BlockChain the topnode is not height than current chain")
	}
	lastNode := NewPOAChainNode(c.GetLastBlock())
	topNode := NewPOAChainNode(topBlock)
	if topNode.CheckPrev(lastNode) {
		c.AddNewBlock(topBlock)
		return nil
	} else {
		return errors.New("BlockChain the topBlock is not next of lastBlock chain")
	}
}

func (c *ChainSketch) AddChain(newChain *ChainSketch) error {
	if c.CanLink(newChain) {
		for _, block := range newChain.Blocks {
			c.UpdateChainTop(&block)
		}
		return nil
	} else {
		return errors.New("BlockChain the topBlock is not next of lastBlock chain")
	}
}

func (c *ChainSketch) CanLink(newChain *ChainSketch) bool {
	if !newChain.IsInComplete {
		return false
	}
	topNode := NewPOAChainNode(&newChain.Blocks[1])
	lastNode := NewPOAChainNode(c.GetLastBlock())
	return topNode.CheckPrev(lastNode)
}

/**
GetNewChain
get a new chain from this chain
*/
func (c *ChainSketch) GetNewChain(forkBlock *Block) *ChainSketch {
	newChain := ChainSketch{IsInComplete: true}
	for _, block := range c.Blocks {
		if block.GetHeight() < forkBlock.GetHeight() {
			newChain.AddNewBlock(&block)
		}
	}
	newChain.AddNewBlock(forkBlock)
	return &newChain
}

func GetChainHeight(c *ChainSketch) uint32 {
	return c.GetHeight()
}
