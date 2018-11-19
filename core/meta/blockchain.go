package meta

import (
	"container/list"
)

type BlockChain struct {
	chain *list.List // the data of chain
}

func NewBlockChain(startNode ChainNode) BlockChain {
	chain := list.New()
	chain.PushBack(startNode)
	return BlockChain{chain: chain}
}

func (bc *BlockChain) AddNode(newNode ChainNode) error {
	bc.chain.PushBack(newNode)
	return nil
}

func (bc *BlockChain) GetHeight() uint32 {
	return uint32(bc.chain.Len() - 1)
}

func (bc *BlockChain) GetLastNode() *ChainNode {
	lastNode := bc.chain.Back().Value.(ChainNode)
	return &lastNode
}

func (bc *BlockChain) IsOnChain(checkNode ChainNode) bool {
	index := bc.chain.Len() - 1
	for element := bc.chain.Back(); element != nil && uint32(index) >= checkNode.height; element = element.Prev() {
		chainNode := element.Value.(ChainNode)
		if chainNode.IsEuqal(checkNode) {
			return true
		}
		index--
	}
	return false
}

//func (bc *BlockChain) FillChain(blockManager manager.BlockManager) error {

func (bc *BlockChain) CloneChainIndex(index []ChainNode) []ChainNode {
	forkNode := bc.chain.Back()
	forkPosition := len(index) - 1
	for ; forkNode != nil && forkPosition >= 0; forkNode = forkNode.Prev() {
		chainNode := forkNode.Value.(ChainNode)
		nodeHash := chainNode.GetNodeHash()
		if chainNode.GetNodeHeight() > uint32(forkPosition) {
			continue
		} else if int(chainNode.GetNodeHeight()) < forkPosition {
			forkPosition--
			continue
		}
		checkIndexHash := index[forkPosition].GetNodeHash()
		if checkIndexHash.IsEqual(&nodeHash) {
			break
		}
		forkPosition--
	}
	//delete indexs after forkpoint
	index = index[:forkPosition+1]
	//push index from the behind of forkNode which from mainChain
	for forkNode = forkNode.Next(); forkNode != nil; forkNode = forkNode.Next() {
		chainNode := forkNode.Value.(ChainNode)
		index = append(index, chainNode)
	}
	return index
}

func (bc *BlockChain) GetLastElement() *list.Element {
	return bc.chain.Back()
}

func (bc *BlockChain) GetFristElement() *list.Element {
	return bc.chain.Front()
}

func (bc *BlockChain) RemoveElement(element *list.Element) {
	bc.chain.Remove(element)
}

func (bc *BlockChain) InsertBeforeElement(insertBlock *Block, element *list.Element) {
	bc.chain.InsertBefore(NewPOAChainNode(insertBlock), element)
}

func (bc *BlockChain) IsFillChain() bool {
	return bc.GetLastNode().GetNodeHeight() == bc.GetHeight()
}

/**
checkChainElement
aim:if the currentE of prevpoint is prevE,then return true
*/
func (bc *BlockChain) CheckPrevElement(currentE *list.Element) bool {
	currentNode := currentE.Value.(ChainNode)
	prevNode := currentE.Prev().Value.(ChainNode)
	return currentNode.CheckPrev(prevNode)
}

func (bc *BlockChain) CheckPrevByHeight(currentE *list.Element) bool {
	currentNode := currentE.Value.(ChainNode)
	prevNode := currentE.Prev().Value.(ChainNode)
	return currentNode.height == (prevNode.height + 1)
}

/**
checkEqualElement
aim:if the firstE of hash is equal secondE,then return true
*/
func CheckEqualElement(firstE *list.Element, secondE *list.Element) bool {
	firstNode := firstE.Value.(ChainNode)
	secondNode := secondE.Value.(ChainNode)
	return firstNode.IsEuqal(secondNode)
}
