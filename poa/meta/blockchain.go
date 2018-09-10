package meta

import (
	"container/list"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/common/math"
	"github.com/linkchain/consensus/manager"
	"github.com/linkchain/meta/block"
)


type BlockChain struct {
	chain *list.List		// the data of chain
}

func NewBlockChain(startNode POAChainNode) BlockChain {
	chain := list.New()
	chain.PushBack(startNode)
	return BlockChain{chain:chain}
}

func (bc *BlockChain)AddNode(newNode POAChainNode) error {
	if !bc.IsOnChain(newNode) {
		bc.chain.PushBack(newNode)
	}
	return nil
}

func (bc *BlockChain)GetHeight() uint32 {
	return uint32(bc.chain.Len() - 1)
}

func (bc *BlockChain)GetLastNode() POAChainNode {
	return bc.chain.Back().Value.(POAChainNode)
}

func (bc *BlockChain)IsOnChain(checkNode POAChainNode) bool {
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

//TODO maybe give up
func (bc *BlockChain) FillChain(blockManager manager.BlockManager) error  {
	currentHeight := uint32(bc.chain.Len()) - 1
	if currentHeight == 0 {
		log.Info("BlockChain","fillChain",string("the chain only have gensisblock"))
		return nil
	}

	currentElement := bc.chain.Back()
	prevElement := currentElement.Prev()
	for !bc.CheckPrevElement(currentElement, prevElement) {

		if bc.CheckPrevByHeight(currentElement, prevElement) {
			bc.chain.Remove(prevElement)
		}

		currentNode := currentElement.Value.(POAChainNode)
		insertBlock,error := blockManager.GetBlockByID(currentNode.GetPrevHash())
		if error != nil {
			return error
		}
		//add corret element
		bc.chain.InsertBefore(NewPOAChainNode(insertBlock),currentElement)

		currentElement = currentElement.Prev()
		prevElement := currentElement.Prev()
		if prevElement == nil {
			break
		}
	}
	return nil
}

func (bc *BlockChain) CloneChainIndex(index []POAChainNode) []POAChainNode  {
	forkNode := bc.chain.Back()
	forkPosition := len(index) - 1
	for ; forkNode != nil && forkPosition >= 0 ; forkNode = forkNode.Prev() {
		node := forkNode.Value.(POAChainNode)
		nodeHash := node.GetNodeHash().(math.Hash)
		if node.GetNodeHeight() > uint32(forkPosition) {
			continue
		} else if int(node.GetNodeHeight()) < forkPosition{
			forkPosition--
			continue
		}
		checkIndexHash := index[forkPosition].GetNodeHash().(math.Hash)
		if checkIndexHash.IsEqual(&nodeHash) {
			break
		}
		forkPosition--
	}

	//delete indexs after forkpoint
	index = index[:forkPosition+1]
	//push index from the behind of forkNode which from mainChain
	for forkNode = forkNode.Next(); forkNode != nil; forkNode = forkNode.Next() {
		node := forkNode.Value.(POAChainNode)
		index = append(index,node)
	}
	return index
}


func (bc *BlockChain) GetLastElement() *list.Element{
	return bc.chain.Back()
}

func (bc *BlockChain) GetFristElement() *list.Element{
	return bc.chain.Front()
}

func (bc *BlockChain) RemoveElement(element *list.Element) {
	bc.chain.Remove(element)
}

func (bc *BlockChain) InsertBeforeElement(insertBlock block.IBlock,element *list.Element) {
	bc.chain.InsertBefore(NewPOAChainNode(insertBlock),element)
}
/**
	checkChainElement
	aim:if the currentE of prevpoint is prevE,then return true
 */
func (bc *BlockChain) CheckPrevElement(currentE *list.Element, prevE *list.Element) bool {
	currentNode := currentE.Value.(POAChainNode)
	prevNode := prevE.Value.(POAChainNode)
	return currentNode.CheckPrev(prevNode)
}

func (bc *BlockChain) CheckPrevByHeight(currentE *list.Element, prevE *list.Element) bool {
	currentNode := currentE.Value.(POAChainNode)
	prevNode := prevE.Value.(POAChainNode)
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
