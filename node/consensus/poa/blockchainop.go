package poa

import (
	"time"

	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/node/config"
)

/*

	Block
*/
func CreateBlock(prevHeight uint32, prevHash meta.BlockID) (*meta.Block, error) {
	var txs []meta.Transaction
	header := meta.NewBlockHeader(config.DefaultBlockVersion, prevHeight+1, time.Now(),
		config.DefaultNounce, config.DefaultDifficulty, prevHash,
		math.Hash{}, math.Hash{}, meta.Signature{}, nil)
	b := meta.NewBlock(*header, txs)
	return RebuildBlock(b)

}

func RebuildBlock(block *meta.Block) (*meta.Block, error) {
	pb := block
	root := pb.CalculateTxTreeRoot()
	pb.Header.SetMerkleRoot(root)
	return pb, nil
}
