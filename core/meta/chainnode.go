package meta

import (
	"strconv"
)

type ChainNode struct {
	curentHash BlockID
	prevHash   BlockID
	height     uint32
}

func NewPOAChainNode(block *Block) ChainNode {
	return ChainNode{
		curentHash: *block.GetBlockID(),
		prevHash:   *block.GetPrevBlockID(),
		height:     block.GetHeight()}
}

func (bn *ChainNode) GetNodeHeight() uint32 {
	return bn.height
}

func (bn *ChainNode) GetNodeHash() BlockID {
	return bn.curentHash
}

func (bn *ChainNode) GetPrevHash() BlockID {
	return bn.prevHash
}

func (bn *ChainNode) CheckPrev(prevNode ChainNode) bool {
	return bn.prevHash.IsEqual(&prevNode.curentHash)
}

func (bn *ChainNode) IsEuqal(checkNode ChainNode) bool {
	return bn.curentHash.IsEqual(&checkNode.curentHash)
}

func (bn *ChainNode) IsGensis() bool {
	return bn.height == 0 && bn.prevHash.IsEmpty()
}

func (bn *ChainNode) GetString() string {
	str := string("height:") + strconv.Itoa(int(bn.height)) + " "
	str += string("currentHash:") + bn.curentHash.GetString() + " "
	str += string("prev:") + bn.prevHash.GetString()

	return str
}
