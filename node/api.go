package node

type PublicNodeAPI struct {
	n *Node
}

func NewPublicNodeAPI(n *Node) *PublicNodeAPI {
	return &PublicNodeAPI{n}
}
