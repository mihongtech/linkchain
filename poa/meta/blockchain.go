package meta

import (
	"container/list"
	"github.com/linkchain/consensus/manager"
	"github.com/linkchain/meta/block"
)

type BlockChain struct {
	chain *list.List // the data of chain
}

func NewBlockChain(startNode POAChainNode) BlockChain {
	chain := list.New()
	chain.PushBack(startNode)
	return BlockChain{chain: chain}
}

func (bc *BlockChain) AddNode(newNode POAChainNode) error {
	bc.chain.PushBack(newNode)
	return nil
}

func (bc *BlockChain) GetHeight() uint32 {
	return uint32(bc.chain.Len() - 1)
}

func (bc *BlockChain) GetLastNode() *POAChainNode {
	lastNode := bc.chain.Back().Value.(POAChainNode)
	return &lastNode
}

func (bc *BlockChain) IsOnChain(checkNode POAChainNode) bool {
	index := bc.chain.Len() - 1
	for element := bc.chain.Back(); element != nil && uint32(index) >= checkNode.height; element = element.Prev() {
		node := element.Value.(POAChainNode)
		if node.IsEuqal(checkNode) {
			return true
		}
		index--
	}
	return false
}

func (bc *BlockChain) FillChain(blockManager manager.BlockManager) error {
	currentE := bc.GetLastElement()
	prevE := currentE.Prev()

	for !bc.IsFillChain() && currentE != nil && prevE != nil {
		currentNode := currentE.Value.(POAChainNode)
		prevNode := currentE.Prev().Value.(POAChainNode)
		if !bc.CheckPrevElement(currentE) {
			if currentNode.GetNodeHeight() <= prevNode.GetNodeHeight() {
				prevE = currentE.Prev().Prev()
				bc.chain.Remove(currentE.Prev())
				continue
			}

			if !currentNode.IsGensis() {
				insertBlock, error := blockManager.GetBlockByID(currentNode.GetPrevHash())
				if error != nil {
					return error
				}
				bc.chain.InsertBefore(NewPOAChainNode(insertBlock), currentE)
			}
		}

		currentE = prevE
		if currentE == nil {
			break
		}
		prevE = currentE.Prev()
	}
	return nil
}

func (bc *BlockChain) CloneChainIndex(index []POAChainNode) []POAChainNode {
	forkNode := bc.chain.Back()
	forkPosition := len(index) - 1
	for ; forkNode != nil && forkPosition >= 0; forkNode = forkNode.Prev() {
		node := forkNode.Value.(POAChainNode)
		nodeHash := node.GetNodeHash()
		if node.GetNodeHeight() > uint32(forkPosition) {
			continue
		} else if int(node.GetNodeHeight()) < forkPosition {
			forkPosition--
			continue
		}
		checkIndexHash := index[forkPosition].GetNodeHash()
		if checkIndexHash.IsEqual(nodeHash) {
			break
		}
		forkPosition--
	}
	//delete indexs after forkpoint
	index = index[:forkPosition+1]
	//push index from the behind of forkNode which from mainChain
	for forkNode = forkNode.Next(); forkNode != nil; forkNode = forkNode.Next() {
		node := forkNode.Value.(POAChainNode)
		index = append(index, node)
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

func (bc *BlockChain) InsertBeforeElement(insertBlock block.IBlock, element *list.Element) {
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
	currentNode := currentE.Value.(POAChainNode)
	prevNode := currentE.Prev().Value.(POAChainNode)
	return currentNode.CheckPrev(prevNode)
}

func (bc *BlockChain) CheckPrevByHeight(currentE *list.Element) bool {
	currentNode := currentE.Value.(POAChainNode)
	prevNode := currentE.Prev().Value.(POAChainNode)
	return currentNode.height == (prevNode.height + 1)
}

/**
checkEqualElement
aim:if the firstE of hash is equal secondE,then return true
*/
func CheckEqualElement(firstE *list.Element, secondE *list.Element) bool {
	firstNode := firstE.Value.(POAChainNode)
	secondNode := secondE.Value.(POAChainNode)
	return firstNode.IsEuqal(secondNode)
}
