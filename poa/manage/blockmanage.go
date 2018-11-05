package manage

import (
	"errors"
	"sync"
	"time"

	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	globalconfig "github.com/linkchain/config"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
	"github.com/linkchain/poa/config"
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
	gensisBlock := GetManager().BlockManager.GetGensisBlock()
	m.AddBlock(gensisBlock)
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

/** interface: BlockBaseManager **/
func (m *BlockManage) CreateBlock() (block.IBlock, error) {
	bestBlock := GetManager().ChainManager.GetBestBlock()
	if bestBlock != nil {
		bestHash := bestBlock.GetBlockID()
		txs := []poameta.Transaction{}
		header := poameta.NewBlockHeader(config.DefaultBlockVersion, bestBlock.GetHeight()+1, time.Now(), config.DefaultNounce, config.DefaultDifficulty, *bestHash.(*math.Hash), math.Hash{}, math.Hash{}, nil, nil)
		b := poameta.NewBlock(*header, txs)
		return m.RebuildBlock(b)
	} else {
		return m.GetGensisBlock(), nil
	}
}

func (m *BlockManage) RebuildBlock(block block.IBlock) (block.IBlock, error) {
	pb := *block.(*poameta.Block)
	root := pb.CalculateTxTreeRoot()
	pb.Header.SetMerkleRoot(root)
	return &pb, nil
}

func (m *BlockManage) SignBlock(block block.IBlock, sign []byte) (block.IBlock, error) {
	pb := *block.(*poameta.Block)
	pb.Header.Sign = sign
	return &pb, nil
}

/** interface: BlockBaseManager **/
func (m *BlockManage) GetGensisBlock() block.IBlock {
	txs := []poameta.Transaction{}
	header := poameta.NewBlockHeader(config.DefaultBlockVersion, 0, time.Unix(1487780010, 0), config.DefaultNounce, config.DefaultDifficulty, math.Hash{}, math.Hash{}, math.Hash{}, nil, nil)
	b := poameta.NewBlock(*header, txs)
	root := b.CalculateTxTreeRoot()
	b.Header.SetMerkleRoot(root)

	return b
}

/** interface: BlockPoolManager **/
func (m *BlockManage) GetBlockByID(hash meta.DataID) (block.IBlock, error) {
	index, ok := m.readBlock(*hash.(*math.Hash))
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
			ancestor, e = m.GetBlockByID(ancestor.GetPrevBlockID())
			if e != nil {
				log.Error("ChainManage", "GetBlockAncestor error", "can not find ancestor")
				return nil, errors.New("ChainManage :GetBlockAncestor error->can not find ancestor")
			}
		}
		return ancestor, nil
	}
}

func (m *BlockManage) AddBlock(block block.IBlock) error {
	hash := *block.GetBlockID().(*math.Hash)
	m.writeBlock(hash, *(block.(*poameta.Block)))
	return nil
}

func (m *BlockManage) AddBlocks(blocks []block.IBlock) error {
	for _, b := range blocks {
		m.AddBlock(b)
	}
	return nil
}

func (m *BlockManage) RemoveBlock(hash meta.DataID) error {
	m.removeBlock(*hash.(*math.Hash))
	return nil
}

func (m *BlockManage) HasBlock(hash meta.DataID) bool {
	_, ok := m.readBlock(*hash.(*math.Hash))
	return ok
}

/** interface: BlockValidator **/
func (m *BlockManage) CheckBlock(block block.IBlock) bool {
	//log.Info("POA CheckBlock ...")
	croot := block.CalculateTxTreeRoot()
	if !block.GetMerkleRoot().IsEqual(croot) {
		log.Error("POA CheckBlock", "check merkle root", false)
		return false
	}

	//check poa
	prevBlock, err := GetManager().BlockManager.GetBlockByID(block.GetPrevBlockID())

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

	return nil
	//4.updateStorage

	//5.broadcast
}
