package manage

import (
	"errors"
	"sync"

	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	globalconfig "github.com/linkchain/config"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
	poameta "github.com/linkchain/poa/meta"
)

type BlockManage struct {
	blockMtx            sync.RWMutex
	mapBlockIndexByHash map[math.Hash]poameta.Block
}

func (m *BlockManage) readBlock(key math.Hash) (poameta.Block, bool) {
	m.blockMtx.RLock()
	defer m.blockMtx.RUnlock()
	value, ok := m.mapBlockIndexByHash[key]
	return value, ok
}

func (m *BlockManage) writeBlock(key math.Hash, value poameta.Block) {
	m.blockMtx.Lock()
	defer m.blockMtx.Unlock()

	m.mapBlockIndexByHash[key] = value
}

func (m *BlockManage) removeBlock(key math.Hash) {
	m.blockMtx.Lock()
	defer m.blockMtx.Unlock()

	delete(m.mapBlockIndexByHash, key)
}

/** interface: common.IService **/
func (m *BlockManage) Init(i interface{}) bool {
	log.Info("BlockManage init...")
	m.mapBlockIndexByHash = make(map[math.Hash]poameta.Block)
	//load gensis block
	//load block by chainmanager

	return true
}

func (m *BlockManage) Start() bool {
	log.Info("BlockManage start...")
	return true
}

func (m *BlockManage) Stop() {
	log.Info("BlockManage stop...")
}

/** interface: BlockPoolManager **/
func (m *BlockManage) GetBlockByID(hash meta.BlockID) (block.IBlock, error) {
	index, ok := m.readBlock(hash)
	if ok {
		return &index, nil
	}
	//TODO need to storage
	return nil, errors.New("BlockManage can not find block by hash:" + hash.GetString())
}

func (m *BlockManage) GetBlockByHeight(height uint32) ([]block.IBlock, error) {
	//TODO may not be need
	return nil, nil
}

//Find block Ancestor by height.
func (m *BlockManage) GetBlockAncestor(block block.IBlock, height uint32) (block.IBlock, error) {
	if height > block.GetHeight() {
		log.Error("ChainManage", "GetBlockAncestor error", "height is plus block's height")
		return nil, errors.New("ChainManage :GetBlockAncestor error->height is plus block's height")
	} else {
		ancestor := block
		var e error
		for height < ancestor.GetHeight() {
			ancestor, e = m.GetBlockByID(*ancestor.GetPrevBlockID())
			if e != nil {
				log.Error("ChainManage", "GetBlockAncestor error", "can not find ancestor")
				return nil, errors.New("ChainManage :GetBlockAncestor error->can not find ancestor")
			}
		}
		return ancestor, nil
	}
}

func (m *BlockManage) AddBlock(block block.IBlock) error {
	hash := *block.GetBlockID()
	m.writeBlock(hash, *(block.(*poameta.Block)))
	return nil
}

func (m *BlockManage) AddBlocks(blocks []block.IBlock) error {
	for _, b := range blocks {
		m.AddBlock(b)
	}
	return nil
}

func (m *BlockManage) RemoveBlock(hash meta.BlockID) error {
	m.removeBlock(hash)
	return nil
}

func (m *BlockManage) HasBlock(hash meta.BlockID) bool {
	_, ok := m.readBlock(hash)
	return ok
}

/** interface: BlockValidator **/
func (m *BlockManage) CheckBlock(block block.IBlock) bool {
	//log.Info("POA CheckBlock ...")
	croot := block.CalculateTxTreeRoot()
	if !block.GetMerkleRoot().IsEqual(&croot) {
		log.Error("POA CheckBlock", "check merkle root", false)
		return false
	}

	//check poa
	prevBlock, err := GetManager().BlockManager.GetBlockByID(*block.GetPrevBlockID())

	if err != nil {
		log.Error("BlockManage", "CheckBlock", err)
		return false
	}

	if prevBlock.GetHeight()+1 != block.GetHeight() {
		log.Error("BlockManage", "CheckBlock", "current block height is error")
		return false
	}
	signerIndex := block.GetHeight() % uint32(len(globalconfig.SignMiners))
	//check block sign
	err = block.Verify(globalconfig.SignMiners[signerIndex])
	if err != nil {
		log.Error("POA CheckBlock", "check sign", false)
		return false
	}

	//check TXs
	for _, tx := range block.GetTxs() {
		if !GetManager().TransactionManager.CheckTx(tx) {
			log.Error("POA CheckBlock", "check tx", false)
			return false
		}
	}
	return true
}

func (s *BlockManage) ProcessBlock(block block.IBlock) error {

	//log.Info("POA ProcessBlock ...")
	//1.checkBlock
	if !s.CheckBlock(block) {
		log.Error("POA checkBlock failed")
		return errors.New("POA checkBlock failed")
	}

	//2.acceptBlock
	GetManager().ChainManager.AddBlock(block)
	log.Info("POA Add a Blocks", "block", block.GetBlockID(), "prev", block.GetPrevBlockID())
	//log.Info("POA Add a Blocks", "block hash", block.GetBlockID().String())
	//log.Info("POA Add a Blocks", "prev hash", block.GetPrevBlockID().String())

	//3.updateChain
	if !GetManager().ChainManager.UpdateChain() {
		log.Info("POA Update chain failed")
		GetManager().ChainManager.UpdateChain()
		return errors.New("Update chain failed")
	}

	//log.Info("POA ProcessBlock successed")
	//GetManager().ChainManager.GetBlockChainInfo()

	//4.updateStorage

	//5.broadcast

	return nil
}
