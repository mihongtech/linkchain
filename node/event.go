package node

import "github.com/mihongtech/linkchain/core/meta"

type NewMinedBlockEvent struct {
	Block *meta.Block
}

type TxEvent struct {
	Tx *meta.Transaction
}

type AccountEvent struct {
	IsUpdate bool
}
