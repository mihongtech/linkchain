package node

import (
	"errors"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/core/meta"
)

func (n *Node) readBlockCache(key math.Hash) (*meta.Block, bool) {
	n.blockMtx.RLock()
	defer n.blockMtx.RUnlock()
	value, ok := n.mapBlockIndexByHash[key]
	return &value, ok
}

func (n *Node) writeBlockCache(key math.Hash, value *meta.Block) {
	n.blockMtx.Lock()
	defer n.blockMtx.Unlock()

	n.mapBlockIndexByHash[key] = *value
}

func (n *Node) removeBlockCache(key math.Hash) {
	n.blockMtx.Lock()
	defer n.blockMtx.Unlock()

	delete(n.mapBlockIndexByHash, key)
}

/** interface: BlockPoolManager **/
func (n *Node) getBlockByID(hash meta.BlockID) (*meta.Block, error) {
	b, ok := n.readBlockCache(hash)
	if ok {
		return b, nil
	}
	//TODO need to storage
	return nil, errors.New("blockManage can not find block by hash:" +
		hash.GetString())
}

//Find block Ancestor by height.
func (n *Node) getBlockAncestor(block *meta.Block, height uint32) (*meta.Block, error) {
	if height > block.GetHeight() {
		log.Error("ChainManage", "GetBlockAncestor error", "height is plus block's height")
		return nil, errors.New("chainManage :GetBlockAncestor error->height is plus block's height")
	} else {
		ancestor := block
		var e error
		for height < ancestor.GetHeight() {
			ancestor, e = n.getBlockByID(*ancestor.GetPrevBlockID())
			if e != nil {
				log.Error("ChainManage", "GetBlockAncestor error", "can not find ancestor")
				return nil, errors.New("chainManage :GetBlockAncestor error->can not find ancestor")
			}
		}
		return ancestor, nil
	}
}

func (n *Node) addBlockCache(block *meta.Block) error {
	hash := block.GetBlockID()
	n.writeBlockCache(*hash, block)
	return nil
}

func (n *Node) removeBlockByID(hash meta.BlockID) error {
	n.removeBlockCache(hash)
	return nil
}

func (n *Node) hasBlock(hash meta.BlockID) bool {
	_, ok := n.readBlockCache(hash)
	return ok
}

/** interface: BlockValidator **/
func (n *Node) checkBlock(block *meta.Block) bool {
	//log.Info("POA checkBlock ...")
	croot := block.CalculateTxTreeRoot()
	if !block.GetMerkleRoot().IsEqual(&croot) {
		log.Error("POA checkBlock", "check merkle root", false)
		return false
	}

	//check poa
	prevBlock, err := n.getBlockByID(*block.GetPrevBlockID())

	if err != nil {
		log.Error("BlockManage", "checkBlock", err)
		return false
	}

	if prevBlock.GetHeight()+1 != block.GetHeight() {
		log.Error("BlockManage", "checkBlock", "current block height is error")
		return false
	}
	signerIndex := block.GetHeight() % uint32(len(config.SignMiners))
	//check block sign
	err = block.Verify(config.SignMiners[signerIndex])
	if err != nil {
		log.Error("POA checkBlock", "check sign", false)
		return false
	}

	//check TXs
	for _, tx := range block.GetTxs() {
		if !n.checkTx(&tx) {
			log.Error("POA checkBlock", "check tx", false)
			return false
		}
	}
	return true
}

func (n *Node) processBlock(block *meta.Block) error {
	log.Info("processBlock ...")
	//1.checkBlock
	if !n.checkBlock(block) {
		log.Error("checkBlock failed")
		return errors.New("checkBlock failed")
	}

	//2.acceptBlock
	n.addBlock(block)
	n.addBlockCache(block)
	/*log.Info("POA Add a Blocks", "block", block.GetBlockID(), "prev", block.GetPrevBlockID())
	log.Info("POA Add a Blocks", "block hash", block.GetBlockID().String())
	log.Info("POA Add a Blocks", "prev hash", block.GetPrevBlockID().String())*/
	log.Info("Add a Blocks", "block", block.String())

	//3.updateChain
	if !n.updateChainAndIndex() {
		log.Info("Update chain failed")
		n.updateChainAndIndex()
		return errors.New("update chain failed")
	}

	//log.Info("POA processBlock successed")
	//n.getBlockChainInfo()

	//4.updateStorage

	//5.broadcast

	return nil
}
