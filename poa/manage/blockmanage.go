package manage

import (
	"errors"
	"strings"
	"sync"
	"time"

	"encoding/hex"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
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
func (m *BlockManage) NewBlock() (block.IBlock, error) {
	bestBlock := GetManager().ChainManager.GetBestBlock()
	if bestBlock != nil {
		bestHash := bestBlock.GetBlockID()
		txs := []poameta.Transaction{}
		b := &poameta.Block{
			Header: poameta.BlockHeader{Version: 0, PrevBlock: *bestHash.(*math.Hash), MerkleRoot: math.Hash{}, Timestamp: time.Now(), Difficulty: 0x207fffff, Nonce: 0, Extra: nil, Height: bestBlock.GetHeight() + 1},
			TXs:    txs,
		}
		return m.RebuildBlock(b)
	} else {
		return m.GetGensisBlock(), nil
	}
}

func (m *BlockManage) RebuildBlock(block block.IBlock) (block.IBlock, error) {
	bestBlock := GetManager().ChainManager.GetBestBlock()
	if bestBlock != nil {
		pb := *block.(*poameta.Block)
		root := pb.CalculateTxTreeRoot()
		pb.Header.SetMerkleRoot(root)

		ls, err := bestBlock.(*poameta.Block).Header.GetSignerID()
		if err != nil {
			log.Error("BlockManage", "NewBlock", err)
			return &pb, err
		}
		lf, err := hex.DecodeString(ls.GetString())
		if err != nil {
			log.Error("BlockManage", "NewBlock", err)
			return &pb, err
		}
		pubIndex := ChooseNextSigner(lf)
		s, err := poameta.CreateSignerIdByPubKey(poameta.PubSigners[pubIndex])
		if err != nil {
			log.Error("BlockManage", "NewBlock Create Signer", err)
			return nil, err
		}
		pb.Header.SetSigner(*s)
		signer, err := pb.Header.GetSigner()
		if err != nil {
			log.Error("BlockManage", "NewBlock", err)
			return &pb, err
		}
		pb.Deserialize(pb.Serialize())
		signer.Sign(poameta.PrivSigner[pubIndex], *pb.GetBlockID().(*math.Hash))
		pb.Header.SetSigner(signer)
		return &pb, nil
	} else {
		return m.GetGensisBlock(), nil
	}
}

func (m *BlockManage) RebuildTestBlock(block block.IBlock) (block.IBlock, error) {
	bestBlock := GetManager().ChainManager.GetBestBlock()
	if bestBlock != nil {
		pb := *block.(*poameta.Block)
		root := pb.CalculateTxTreeRoot()
		pb.Header.SetMerkleRoot(root)

		ls, err := bestBlock.(*poameta.Block).Header.GetSignerID()
		if err != nil {
			log.Error("BlockManage", "NewBlock", err)
			return &pb, err
		}
		lf, err := hex.DecodeString(ls.GetString())
		if err != nil {
			log.Error("BlockManage", "NewBlock", err)
			return &pb, err
		}
		pubIndex := ChooseNextSigner(lf)
		pubIndex = 0
		s, err := poameta.CreateSignerIdByPubKey(poameta.PubSigners[pubIndex])
		if err != nil {
			log.Error("BlockManage", "NewBlock Create Signer", err)
			return nil, err
		}
		pb.Header.SetSigner(*s)
		signer, err := pb.Header.GetSigner()
		if err != nil {
			log.Error("BlockManage", "NewBlock", err)
			return &pb, err
		}
		pb.Deserialize(pb.Serialize())
		signer.Sign(poameta.PrivSigner[pubIndex], *pb.GetBlockID().(*math.Hash))
		pb.Header.SetSigner(signer)
		return &pb, nil
	} else {
		return m.GetGensisBlock(), nil
	}
}

/** interface: BlockBaseManager **/
func (m *BlockManage) GetGensisBlock() block.IBlock {
	txs := []poameta.Transaction{}
	b := &poameta.Block{
		Header: poameta.BlockHeader{Version: 0, PrevBlock: math.Hash{}, MerkleRoot: math.Hash{}, Timestamp: time.Unix(1487780010, 0), Difficulty: 0x207fffff, Nonce: 0, Extra: nil, Height: 0},
		TXs:    txs,
	}
	root := b.CalculateTxTreeRoot()
	b.Header.SetMerkleRoot(root)
	s, err := poameta.CreateSignerIdByPubKey(poameta.PubSigners[0])
	if err != nil {
		log.Error("BlockManage", "NewBlock Create Signer", err)
		return nil
	}
	b.Header.SetSigner(*s)
	//TODO test
	signer, err := b.Header.GetSigner()
	if err != nil {
		log.Error("BlockManage", "NewBlock", err)
		return b
	}
	b.Deserialize(b.Serialize())
	signer.Sign(poameta.PrivSigner[0], *b.GetBlockID().(*math.Hash))
	b.Header.SetSigner(signer)
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
	log.Info("POA CheckBlock ...")
	croot := block.CalculateTxTreeRoot()
	if !block.GetMerkleRoot().IsEqual(croot) {
		log.Error("POA CheckBlock", "check merkle root", false)
		return false
	}

	//check poa
	ls, err := GetManager().ChainManager.GetBestBlock().(*poameta.Block).Header.GetSignerID()
	if err != nil {
		log.Error("BlockManage", "CheckBlock", err)
		return false
	}
	lf, err := hex.DecodeString(ls.GetString())
	if err != nil {
		log.Error("BlockManage", "CheckBlock", err)
		return false
	}
	nextSigner, err := poameta.CreateSignerIdByPubKey(poameta.PubSigners[ChooseNextSigner(lf)])
	if err != nil {
		log.Error("BlockManage", "NewBlock Create Signer", err)
		return false
	}
	b := block.(*poameta.Block)
	currentS, err := b.Header.GetSigner()
	if err != nil {
		log.Error("BlockManage", "CheckBlock", err)
		return false
	}
	if !nextSigner.IsEqual(currentS) {
		log.Error("BlockManage", "CheckBlock", "the block is error miner")
		return false
	}

	//check block sign
	err = block.Verify()
	if err != nil {
		log.Error("POA CheckBlock", "check sign", false)
		return false
	}

	//check TXs
	for _, tx := range block.GetTxs() {
		if !GetManager().TransactionManager.CheckTx(tx) {
			log.Error("POA CheckBlock", "check tx", false)
		}
	}
	return true
}

func (s *BlockManage) ProcessBlock(block block.IBlock) error {
	log.Info("POA ProcessBlock ...")
	//1.checkBlock
	if !GetManager().BlockManager.CheckBlock(block) {
		log.Error("POA checkBlock failed")
		return errors.New("POA checkBlock failed")
	}

	//2.acceptBlock
	GetManager().ChainManager.AddBlock(block)
	//log.Info("POA Add a Blocks", "block hash", block.GetBlockID().GetString())
	//log.Info("POA Add a Blocks", "prev hash", block.GetPrevBlockID().GetString())

	//3.updateChain
	if !GetManager().ChainManager.UpdateChain() {
		log.Info("POA Update chain failed")
		GetManager().ChainManager.UpdateChain()
		return errors.New("POA Update chain failed")
	}
	log.Info("POA ProcessBlock successed", "blockchaininfo", GetManager().ChainManager.GetBlockChainInfo())

	return nil
	//4.updateStorage

	//5.broadcast
}

func ChooseNextSigner(lastSigner []byte) int {
	index := 0
	for i, signer := range poameta.PubSigners {
		if strings.Compare(signer, hex.EncodeToString(lastSigner)) == 0 {
			index = (i + 1) % 3
		}
	}
	return index
}

func IsCorrectSigner(lastSigner []byte, currentSigner []byte) bool {
	i := ChooseNextSigner(lastSigner)
	if strings.Compare(poameta.PubSigners[i], hex.EncodeToString(currentSigner)) == 0 {
		return true
	}
	return false
}
