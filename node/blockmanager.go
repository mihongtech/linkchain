package node

import (
	"errors"
	"sync"

	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/config"
	"github.com/linkchain/core/meta"
	"time"
	"encoding/hex"
)

var (
	blockMtx sync.RWMutex
	mapBlockIndexByHash = make(map[math.Hash]meta.Block)
)

func readBlockCache(key math.Hash) (*meta.Block, bool) {
	blockMtx.RLock()
	defer blockMtx.RUnlock()
	value, ok := mapBlockIndexByHash[key]
	return &value, ok
}

func writeBlockCache(key math.Hash, value *meta.Block) {
	blockMtx.Lock()
	defer blockMtx.Unlock()

	mapBlockIndexByHash[key] = *value
}

func removeBlockCache(key math.Hash) {
	blockMtx.Lock()
	defer blockMtx.Unlock()

	delete(mapBlockIndexByHash, key)
}

/** interface: BlockPoolManager **/
func GetBlockByID(hash meta.BlockID) (*meta.Block, error) {
	b, ok := readBlockCache(hash)
	if ok {
		return b, nil
	}
	//TODO need to storage
	return nil, errors.New("blockManage can not find block by hash:" +
		hash.GetString())
}


//Find block Ancestor by height.
func  getBlockAncestor(block *meta.Block, height uint32) (*meta.Block, error) {
	if height > block.GetHeight() {
		log.Error("ChainManage", "GetBlockAncestor error", "height is plus block's height")
		return nil, errors.New("chainManage :GetBlockAncestor error->height is plus block's height")
	} else {
		ancestor := block
		var e error
		for height < ancestor.GetHeight() {
			ancestor, e = GetBlockByID(*ancestor.GetPrevBlockID())
			if e != nil {
				log.Error("ChainManage", "GetBlockAncestor error", "can not find ancestor")
				return nil, errors.New("chainManage :GetBlockAncestor error->can not find ancestor")
			}
		}
		return ancestor, nil
	}
}

func addBlockCache(block *meta.Block) error {
	hash := block.GetBlockID()
	writeBlockCache(*hash, block)
	return nil
}

func removeBlockByID(hash meta.BlockID) error {
	removeBlockCache(hash)
	return nil
}

func HasBlock(hash meta.BlockID) bool {
	_, ok := readBlockCache(hash)
	return ok
}

/** interface: BlockValidator **/
func CheckBlock(block *meta.Block) bool {
	//log.Info("POA CheckBlock ...")
	croot := block.CalculateTxTreeRoot()
	if !block.GetMerkleRoot().IsEqual(&croot) {
		log.Error("POA CheckBlock", "check merkle root", false)
		return false
	}

	//check poa
	prevBlock, err := GetBlockByID(*block.GetPrevBlockID())

	if err != nil {
		log.Error("BlockManage", "CheckBlock", err)
		return false
	}

	if prevBlock.GetHeight()+1 != block.GetHeight() {
		log.Error("BlockManage", "CheckBlock", "current block height is error")
		return false
	}
	signerIndex := block.GetHeight() % uint32(len(config.SignMiners))
	//check block sign
	err = block.Verify(config.SignMiners[signerIndex])
	if err != nil {
		log.Error("POA CheckBlock", "check sign", false)
		return false
	}

	//check TXs
	for _, tx := range block.GetTxs() {
		if !checkTx(&tx) {
			log.Error("POA CheckBlock", "check tx", false)
			return false
		}
	}
	return true
}

func ProcessBlock(block *meta.Block) error {

	//log.Info("POA ProcessBlock ...")
	//1.CheckBlock
	if !CheckBlock(block) {
		log.Error("POA CheckBlock failed")
		return errors.New("POA CheckBlock failed")
	}

	//2.acceptBlock
	addBlockCache(block)
	log.Info("POA Add a Blocks", "block", block.GetBlockID(), "prev", block.GetPrevBlockID())
	//log.Info("POA Add a Blocks", "block hash", block.GetBlockID().String())
	//log.Info("POA Add a Blocks", "prev hash", block.GetPrevBlockID().String())

	//3.updateChain
	if !updateChain() {
		log.Info("POA Update chain failed")
		updateChain()
		return errors.New("Update chain failed")
	}

	//log.Info("POA ProcessBlock successed")
	//GetManager().ChainManager.GetBlockChainInfo()

	//4.updateStorage

	//5.broadcast

	return nil
}

var fristPrivMiner, _ = hex.DecodeString("55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b")

func GetGensisBlock() *meta.Block {
	txs := []meta.Transaction{}

	header := meta.NewBlockHeader(config.DefaultBlockVersion, 0, time.Unix(1487780010, 0), config.DefaultNounce, config.DefaultDifficulty, math.Hash{}, math.Hash{}, math.Hash{}, meta.Signature{Code: make([]byte, 0)}, nil)
	b := meta.NewBlock(*header, txs)
	id, _ := CreateAccountIdByPrivKey(hex.EncodeToString(fristPrivMiner))
	coinbase := CreateCoinBaseTx(*id, meta.NewAmount(50))
	b.SetTx(*coinbase)
	root := b.CalculateTxTreeRoot()
	b.Header.SetMerkleRoot(root)

	SignGensisBlock(b)
	return b
}

func SignGensisBlock(block *meta.Block) error {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), fristPrivMiner)
	log.Info("SignGensisBlock", "block hash", block.GetBlockID().String())
	signature, err := priv.Sign(block.GetBlockID().CloneBytes())
	if err != nil {
		log.Error("SignGensisBlock", "Sign", err)
		return nil
	}
	sign := meta.NewSignatrue(signature.Serialize())
	block.SetSign(sign)
	return nil
}

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
