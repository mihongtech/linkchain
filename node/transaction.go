package node

import (
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/storage"
)

func (n *Node) getTxByID(hash meta.TxID) (*meta.Transaction, math.Hash, uint64, uint64) {
	tx, hash, number, index := storage.GetTransaction(n.db, hash)
	if tx == nil {
		return nil, math.Hash{}, 0, 0
	} else {
		return tx, hash, number, index
	}
}
