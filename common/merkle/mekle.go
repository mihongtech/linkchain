package merkle

import (
	"github.com/linkchain/common/math"
)

type MerkleNode struct {
	Left  *MerkleNode //left tree node
	Right *MerkleNode //right tree node
	Data  []byte      //Hash
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil {
		hash := math.DoubleHashB(data)
		mNode.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := math.DoubleHashB(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}

type MerkleTree struct {
	RootNode *MerkleNode
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, dataitem := range data {
		node := NewMerkleNode(nil, nil, dataitem)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newNodes []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newNodes = append(newNodes, *node)
		}

		nodes = newNodes
	}

	if len(data) <= 0 {
		node := NewMerkleNode(nil, nil, nil)
		nodes = append(nodes, *node)
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}
