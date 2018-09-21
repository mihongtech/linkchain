package poamanager

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

type POABlockManager struct {
	blockMtx            sync.RWMutex
	mapBlockIndexByHash map[math.Hash]poameta.POABlock
}

func (m *POABlockManager) readBlock(key math.Hash) (poameta.POABlock, bool) {
	m.blockMtx.RLock()
	defer m.blockMtx.RUnlock()
	value, ok := m.mapBlockIndexByHash[key]
	return value, ok
}

func (m *POABlockManager) writeBlock(key math.Hash, value poameta.POABlock) {
	m.blockMtx.Lock()
	defer m.blockMtx.Unlock()

	m.mapBlockIndexByHash[key] = value
}

func (m *POABlockManager) removeBlock(key math.Hash) {
	m.blockMtx.Lock()
	defer m.blockMtx.Unlock()

	delete(m.mapBlockIndexByHash, key)
}

/** interface: common.IService **/
func (m *POABlockManager) Init(i interface{}) bool {
	log.Info("POABlockManager init...")
	m.mapBlockIndexByHash = make(map[math.Hash]poameta.POABlock)
	//load gensis block
	gensisBlock := GetManager().BlockManager.GetGensisBlock()
	m.AddBlock(gensisBlock)
	//load block by chainmanager

	return true
}

func (m *POABlockManager) Start() bool {
	log.Info("POABlockManager start...")
	return true
}

func (m *POABlockManager) Stop() {
	log.Info("POABlockManager stop...")
}

/** interface: BlockBaseManager **/
func (m *POABlockManager) NewBlock() (block.IBlock, error) {
	bestBlock := GetManager().ChainManager.GetBestBlock()
	if bestBlock != nil {
		bestHash := bestBlock.GetBlockID()
		txs := []poameta.POATransaction{}
		b := &poameta.POABlock{
			Header: poameta.POABlockHeader{Version: 0, PrevBlock: *bestHash.(*math.Hash), MerkleRoot: math.Hash{}, Timestamp: time.Now(), Difficulty: 0x207fffff, Nonce: 0, Extra: nil, Height: bestBlock.GetHeight() + 1},
			TXs:    txs,
		}
		return m.RebuildBlock(b)
	} else {
		return m.GetGensisBlock(), nil
	}
}

func (m *POABlockManager) RebuildBlock(block block.IBlock) (block.IBlock, error) {
	bestBlock := GetManager().ChainManager.GetBestBlock()
	if bestBlock != nil {
		pb := *block.(*poameta.POABlock)
		root := pb.CalculateTxTreeRoot()
		pb.Header.SetMerkleRoot(root)

		ls, err := bestBlock.(*poameta.POABlock).Header.GetSignerID()
		if err != nil {
			log.Error("POABlockManager", "NewBlock", err)
			return &pb, err
		}
		lf, err := hex.DecodeString(ls.GetString())
		if err != nil {
			log.Error("POABlockManager", "NewBlock", err)
			return &pb, err
		}
		pubIndex := ChooseNextSigner(lf)
		pb.Header.SetSignerPub(poameta.PubSigners[pubIndex])
		signer, err := pb.Header.GetSigner()
		if err != nil {
			log.Error("POABlockManager", "NewBlock", err)
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

func (m *POABlockManager) RebuildTestBlock(block block.IBlock) (block.IBlock, error) {
	bestBlock := GetManager().ChainManager.GetBestBlock()
	if bestBlock != nil {
		pb := *block.(*poameta.POABlock)
		root := pb.CalculateTxTreeRoot()
		pb.Header.SetMerkleRoot(root)

		ls, err := bestBlock.(*poameta.POABlock).Header.GetSignerID()
		if err != nil {
			log.Error("POABlockManager", "NewBlock", err)
			return &pb, err
		}
		lf, err := hex.DecodeString(ls.GetString())
		if err != nil {
			log.Error("POABlockManager", "NewBlock", err)
			return &pb, err
		}
		pubIndex := ChooseNextSigner(lf)
		pubIndex = 0
		pb.Header.SetSignerPub(poameta.PubSigners[pubIndex])
		signer, err := pb.Header.GetSigner()
		if err != nil {
			log.Error("POABlockManager", "NewBlock", err)
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
func (m *POABlockManager) GetGensisBlock() block.IBlock {
	txs := []poameta.POATransaction{}
	b := &poameta.POABlock{
		Header: poameta.POABlockHeader{Version: 0, PrevBlock: math.Hash{}, MerkleRoot: math.Hash{}, Timestamp: time.Unix(1487780010, 0), Difficulty: 0x207fffff, Nonce: 0, Extra: nil, Height: 0},
		TXs:    txs,
	}
	root := b.CalculateTxTreeRoot()
	b.Header.SetMerkleRoot(root)
	b.Header.SetSignerPub(poameta.PubSigners[0])
	//TODO test
	signer, err := b.Header.GetSigner()
	if err != nil {
		log.Error("POABlockManager", "NewBlock", err)
		return b
	}
	b.Deserialize(b.Serialize())
	signer.Sign(poameta.PrivSigner[0], *b.GetBlockID().(*math.Hash))
	b.Header.SetSigner(signer)
	return b
}

/** interface: BlockPoolManager **/
func (m *POABlockManager) GetBlockByID(hash meta.DataID) (block.IBlock, error) {
	index, ok := m.readBlock(*hash.(*math.Hash))
	if ok {
		return &index, nil
	}
	//TODO need to storage
	return nil, errors.New("POABlockManager can not find block by hash:" + hash.GetString())
}

func (m *POABlockManager) GetBlockByHeight(height uint32) ([]block.IBlock, error) {
	//TODO may not be need
	return nil, nil
}

func (m *POABlockManager) AddBlock(block block.IBlock) error {
	hash := *block.GetBlockID().(*math.Hash)
	m.writeBlock(hash, *(block.(*poameta.POABlock)))
	return nil
}

func (m *POABlockManager) AddBlocks(blocks []block.IBlock) error {
	for _, b := range blocks {
		m.AddBlock(b)
	}
	return nil
}

func (m *POABlockManager) RemoveBlock(hash meta.DataID) error {
	m.removeBlock(*hash.(*math.Hash))
	return nil
}

func (m *POABlockManager) HasBlock(hash meta.DataID) bool {
	_, ok := m.readBlock(*hash.(*math.Hash))
	return ok
}

/** interface: BlockValidator **/
func (m *POABlockManager) CheckBlock(block block.IBlock) bool {
	log.Info("POA CheckBlock ...")
	croot := block.CalculateTxTreeRoot()
	if !block.GetMerkleRoot().IsEqual(croot) {
		log.Error("POA CheckBlock", "check merkle root", false)
		return false
	}
	//check poa
	prevBlock, err := GetManager().BlockManager.GetBlockByID(block.GetPrevBlockID())
	if err != nil {
		log.Error("POABlockManager", "CheckBlock", err)
		return false
	}
	ls, err := prevBlock.(*poameta.POABlock).Header.GetSignerID()

	lf, err := hex.DecodeString(ls.GetString())
	if err != nil {
		log.Error("POABlockManager", "CheckBlock", err)
		return false
	}
	nextSigner := poameta.NewSigner(poameta.PubSigners[ChooseNextSigner(lf)])
	b := block.(*poameta.POABlock)
	currentS, err := b.Header.GetSigner()
	if err != nil {
		log.Error("POABlockManager", "CheckBlock", err)
		return false
	}
	if !nextSigner.IsEqual(currentS) {
		log.Error("POABlockManager", "CheckBlock", "the block is error miner")
		return false
	}
	//check block sign
	err = block.Verify()
	if err != nil {
		log.Error("POA CheckBlock", "check sign", false)
		return false
	}
	return true
}

func (s *POABlockManager) ProcessBlock(block block.IBlock) error {
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
