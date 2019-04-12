package node

import (
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/core/meta"
)

/** interface: BlockValidator **/
func (n *Node) checkBlock(block *meta.Block) error {
	//log.Info("POA checkBlock ...")

	if err := n.validatorAPI.ValidateBlockHeader(n.engine, n.blockchain, block); err != nil {
		log.Error("Verify poa Block header failed")
		return err
	}

	if err := n.validatorAPI.ValidateBlockBody(n.validatorAPI, n.blockchain, block); err != nil {
		log.Error("Verify poa Block failed")
		return err
	}

	return nil
}
