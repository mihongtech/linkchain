package node

type CoreAPI struct {
	node *Node
}

func NewPublicCoreAPI(node *Node) *CoreAPI {
	return &CoreAPI{node: node}
}
