package node

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/genesis"
	"github.com/linkchain/storage"
)

var (
	chainMtx       sync.RWMutex
	chains         []meta.Chain     //the chain tree for storing all chains
	mainChainIndex []meta.ChainNode //the mainChain is slice for search block
	mainChain      meta.BlockChain  //the mainChain is linked list for converting chain
	db             lcdb.Database
)

func initChainManager(database lcdb.Database, path string) bool {
	log.Info("ChainManage init...")

	//load genesis from storage
	db = database
	genesisPath := path
	hash, err := InitGenesis(genesisPath)
	if err != nil {
		log.Error("init genesis failed , exit", "err", err)
		return false
	}

	//create gensis chain
	gensisBlock := storage.GetBlock(db, hash, 0)
	addBlockCache(gensisBlock)
	gensisChain := meta.NewPOAChain(gensisBlock, nil)

	chains = make([]meta.Chain, 0)
	chains = append(chains, *gensisChain)

	gensisChainNode := meta.NewPOAChainNode(gensisBlock)
	mainChainIndex = make([]meta.ChainNode, 0)

	mainChain = meta.NewBlockChain(gensisChainNode)

	err = LoadBlocks()
	if err != nil {
		log.Error("load blocks from db failed , exit", "err", err)
		return false
	}

	//TODO BlockManager need inited
	return true
}

func Start() bool {
	log.Info("ChainManage start...")
	//TODO need to updateMainChain
	return true
}

func Stop() {
	log.Info("ChainManage stop...")
}

func InitGenesis(genesisPath string) (math.Hash, error) {
	if len(genesisPath) == 0 {
		return math.Hash{}, errors.New("genesis file is nil")
	}
	file, err := os.Open(genesisPath)
	if err != nil {
		log.Info("genesis file not found, use default Genesis")
	}
	defer file.Close()

	genesisBlock := new(genesis.Genesis)
	if err == nil {
		if err := json.NewDecoder(file).Decode(genesisBlock); err != nil {
			log.Error("invalid genesis file")
			return math.Hash{}, errors.New("invalid genesis file")
		}
	} else {
		genesisBlock = nil
	}

	_, hash, err := genesis.SetupGenesisBlock(db, genesisBlock)
	if err != nil {
		log.Error("Setup genesis failed", err)
		return math.Hash{}, errors.New("Setup genesis failed")
	}

	return hash, nil
}

func LoadBlocks() error {
	hash := storage.GetHeadBlockHash(db)
	number := storage.GetBlockNumber(db, hash)
	var blocks []meta.Block
	log.Info("Best Node is", "hash", hash, "number", number)
	for i := number; i > 0; i-- {
		block := storage.GetBlock(db, hash, i)
		if block == nil {
			log.Error("get block failed", "hash", hash, "number", i)
			return errors.New("get block failed")
		}

		if err := addBlockCache(block); err != nil {
			log.Error("Add block failed", "hash", hash, "block", block)
			return err
		}
		log.Debug("Load block is", "hash", hash, "number", number, "block", block)

		blocks = append(blocks, *(block))

		hash = *block.GetPrevBlockID()
	}

	updateChainAndIndex()

	if _, err := getBestNode(); err != nil {
		return err
	}

	for i := len(blocks) - 1; i >= 0; i-- {
		log.Debug("sort block is", "block", blocks[i], "hash", blocks[i].GetBlockID())
		sortChains(&blocks[i])
		updateChainAndIndex()
	}

	return nil
}

func GetBestBlock() *meta.Block {
	bestHeight, err := getBestHeight()
	if err != nil {
		return nil
	}

	chainMtx.RLock()
	hash := mainChainIndex[bestHeight].GetNodeHash()
	chainMtx.RUnlock()
	bestBlock, _ := GetBlockByID(hash)
	return bestBlock
}

func getBestNode() (meta.ChainNode, error) {
	bestHeight, err := getBestHeight()
	if err != nil {
		return meta.ChainNode{}, errors.New("the chain is not init")
	}

	chainMtx.RLock()
	defer chainMtx.RUnlock()
	node := mainChainIndex[bestHeight]
	return node, nil
}

func getBestBlockHash() meta.BlockID {
	bestHeight, err := getBestHeight()
	if err != nil {
		return meta.BlockID{}
	}

	chainMtx.RLock()
	defer chainMtx.RUnlock()
	hash := mainChainIndex[bestHeight].GetNodeHash()
	return hash
}

func getBestHeight() (uint32, error) {
	chainMtx.RLock()
	defer chainMtx.RUnlock()
	bestHeight := len(mainChainIndex) - 1
	if bestHeight < 0 {
		return uint32(0), errors.New("the chain is not init")
	}
	return uint32(bestHeight), nil
}

func getBlockByHash(hash math.Hash) (*meta.Block, error) {
	//TODO need to lock chain
	b, err := GetBlockByID(hash)
	if err != nil {
		return b, err
	}
	return b, nil
}

func GetBlockByHeight(height uint32) (*meta.Block, error) {
	if height < 0 || height > uint32(len(mainChainIndex)-1) {
		return nil, errors.New("ChainManage: GetBlockByHeight->height is error")
	}

	chainMtx.RLock()
	hash := mainChainIndex[height].GetNodeHash()
	chainMtx.RUnlock()

	b, err := GetBlockByID(hash)
	if err != nil {
		return b, err
	}
	return b, nil
}

func getBlockNodeByHeight(height uint32) (meta.ChainNode, error) {
	if height > uint32(len(mainChainIndex)-1) {
		return meta.ChainNode{}, errors.New("the height is too large")
	}
	chainMtx.RLock()
	defer chainMtx.RUnlock()
	node := mainChainIndex[height]

	return node, nil
}

func GetBlockChainInfo() string {
	chainMtx.RLock()
	defer chainMtx.RUnlock()

	log.Info("ChainManage mainchain", "chainHeight", mainChain.GetHeight(), "bestHash", mainChain.GetLastNode().GetNodeHash())

	//log.Info("ChainManage chains", "chains", len(chains))
	for i, chain := range chains {
		log.Info("ChainManage chains", "chainId", i, "chainHeight", chain.GetHeight(), "bestHash", chain.GetLastBlock().GetBlockID().GetString())
	}

	for e := mainChain.GetLastElement(); e != nil; e = e.Prev() {
		currentNode := e.Value.(meta.ChainNode)
		log.Info("ChainManage mainchain", "Height", currentNode.GetNodeHeight(), "current hash", currentNode.GetNodeHash(), "prev hash", currentNode.GetPrevHash())
	}

	for _, b := range mainChainIndex {
		log.Info("ChainManage mainchainIndex", "chainHeight", b.GetNodeHeight(), "bestHash", b.GetNodeHash())
	}

	return "this is poa chain"
}

func addBlock(block *meta.Block) {
	newblock := block

	addBlockCache(newblock)

	if err := storage.WriteBlock(db, block); err != nil {
		log.Error("WriteBlock to db failed", "error", err)
		return
	}

	_, err := getBestNode()
	if err != nil {
		log.Error("ChainManage", "error", err)
		return
	}
	sortChains(newblock)
	//longest, _ := GetLongestChain()
	//log.Info("AddBlock", "Longest Chain height", len(longest.Blocks), "Longest Chain bestHash", longest.GetLastBlock().GetBlockID().String())
}

func getLongestChain() (*meta.Chain, int) {
	var lc meta.Chain
	bestHeight := uint32(0)
	position := 0
	for i, chain := range chains {
		if bestHeight <= chain.GetHeight() {
			bestHeight = chain.GetHeight()
			lc = chain
			position = i
		}
	}
	return &lc, position
}

func updateChainAndIndex() bool {
	return updateChain() && updateChainIndex()
}

func sortChains(block *meta.Block) bool {
	chainMtx.Lock()
	defer chainMtx.Unlock()

	isUpdated := false
	deletIndex := make([]int, 0)
	blockNode := meta.NewPOAChainNode(block)

	prevBlock, err := GetBlockByID(blockNode.GetPrevHash())
	//check the block's parent, if parent is not exist, the create is incomplete chain for create a chain when the parent is coming
	if err != nil {
		log.Error("ChainManage sortChains", "error", err)
		newChain := meta.NewPOAChain(nil, prevBlock)
		newChain.AddNewBlock(block)
		chains = append(chains, *newChain)
		return false
	}

	prevNode := meta.NewPOAChainNode(prevBlock)

	//1.find block Prev from mainChain
	if mainChain.IsOnChain(prevNode) {
		//If prevNodeInMain is bestNode: update chain; else : add new chain
		_, index := getLongestChain()
		err = chains[index].UpdateChainTop(block)
		if err != nil {
			newChain := chains[index].GetNewChain(prevBlock)
			newChain.AddNewBlock(block)
			chains = append(chains, *newChain)
			isUpdated = true
		}
	} else {
		//3.find block Prev from other sideChain,If cannot find then give up
		// a.update sidechain
		for index := range chains {
			err = chains[index].UpdateChainTop(block)
			if err == nil {
				// if update chain then check complete chain is the chain next
				isUpdated = true
				break
			}
		}
		if !isUpdated {
			// b.add new sidechain
			for index, chain := range chains {

				if !chain.IsInComplete {
					continue
				}

				ancestorBlock, err := getBlockAncestor(chain.GetLastBlock(), prevNode.GetNodeHeight()) //find prevheight block
				if err != nil {
					log.Error("sortChains addNewSideChain", "GetBlockAncestor", err)
					log.Info("sortChains :the chain is bad chain ,because the data of chain is imcomplete. the give up the chain")
					deletIndex = append(deletIndex, index)
					index--
					continue
				}
				ancestorNode := meta.NewPOAChainNode(ancestorBlock)

				if ancestorNode.IsEuqal(prevNode) {
					newChain := chains[index].GetNewChain(ancestorBlock)
					newChain.AddNewBlock(block)
					chains = append(chains, *newChain)
					isUpdated = true
					break
				}
			}
		}
	}

	//sort InCompleteChain
	for index := range chains {
		// if update chain then check complete chain is the chain next
		for i, imcompletchain := range chains {
			if imcompletchain.IsInComplete {
				if chains[index].CanLink(&imcompletchain) {
					chains[index].AddChain(&imcompletchain)
					deletIndex = append(deletIndex, i)
				}
			}
		}
	}
	//delete  imcomplete chain which have been added, or chain which need to giving up
	for _, index := range deletIndex {
		chains = append(chains[:index], chains[index+1:]...)
	}
	return isUpdated
}

/**
updateChainIndex
aim:update mainChainIndex from mainChain
TODO need to test
*/
func updateChainIndex() bool {
	chainMtx.Lock()
	defer chainMtx.Unlock()

	forkNode := mainChain.GetLastElement()
	forkPosition := len(mainChainIndex) - 1
	endNode := forkNode.Value.(meta.ChainNode)
	if forkPosition < 0 {
		//init mainchain index
		for e := mainChain.GetFristElement(); e != nil; e = e.Next() {
			node := e.Value.(meta.ChainNode)

			//add indexs(block status)
			b, err := GetBlockByID(node.GetNodeHash())
			if err != nil {
				log.Error("ChainManage", "init new chain account failed. block hash", b.GetBlockID().GetString())
				return false
			}
			errorStatus := updateStatus(b, true)
			if errorStatus != nil {
				log.Error("ChainManage", "init new chain account failed", errorStatus)
				removeErrorNode(endNode)
				return false
			}

			mainChainIndex = append(mainChainIndex, node)
		}
		return true
	}

	for ; forkNode != nil && forkPosition >= 0; forkNode = forkNode.Prev() {
		node := forkNode.Value.(meta.ChainNode)
		nodeHash := node.GetNodeHash()
		if node.GetNodeHeight() > uint32(forkPosition) {
			continue
		} else if int(node.GetNodeHeight()) < forkPosition {
			forkPosition--
			continue
		}
		checkIndexHash := mainChainIndex[forkPosition].GetNodeHash()
		if checkIndexHash.IsEqual(&nodeHash) {
			break
		}
		forkPosition--
	}

	//delete indexs after forkpoint
	//delete indexs(block status)
	for i := len(mainChainIndex) - 1; i >= forkPosition+1; i-- {
		b, err := GetBlockByID(mainChainIndex[i].GetNodeHash())
		if err != nil {
			log.Error("ChainManage", "remove old chain account failed. block hash", b.GetBlockID().GetString())
			return false
		}
		errorStatus := updateStatus(b, false)
		if errorStatus != nil {
			log.Error("ChainManage", "remove old chain account failed", errorStatus)
			removeErrorNode(endNode)
			return false
		}
	}
	mainChainIndex = mainChainIndex[:forkPosition+1]

	//push index from the behind of forkNode which from mainChain
	for forkNode = forkNode.Next(); forkNode != nil; forkNode = forkNode.Next() {
		node := forkNode.Value.(meta.ChainNode)

		//add indexs(block status)
		b, err := GetBlockByID(node.GetNodeHash())
		if err != nil {
			log.Error("ChainManage", "add new chain account failed. block hash", b.GetBlockID().GetString())
			return false
		}
		errorStatus := updateStatus(b, true)
		if errorStatus != nil {
			log.Error("ChainManage", "add new chain account failed", errorStatus)
			log.Error("ChainManage", "removeErrorNode", endNode.GetNodeHash().GetString())
			removeErrorNode(endNode)
			return false
		}

		mainChainIndex = append(mainChainIndex, node)
	}

	hash := mainChainIndex[len(mainChainIndex)-1].GetNodeHash()
	// TODO: add code to WriteCanonicalHash
	// storage.WriteCanonicalHash(db, math.BytesToHash(hash.CloneBytes()), uint64(mainChainIndex[len(mainChainIndex)-1].GetNodeHeight()))
	storage.WriteHeadBlockHash(db, math.BytesToHash(hash.CloneBytes()))
	return true
}

/**
updateChain
aim:update mainChain from chains
TODO need to test
*/
func updateChain() bool {
	chainMtx.Lock()
	defer chainMtx.Unlock()

	longestChain, _ := getLongestChain()
	bestBlock := longestChain.GetLastBlock()
	log.Info("ChainManage UpdateChain", "bestheight", bestBlock.GetHeight(), "bestblock", bestBlock.GetBlockID(), "prev", bestBlock.GetPrevBlockID())
	mainChain.AddNode(meta.NewPOAChainNode(bestBlock))

	err := mainChain.FillChain()
	if err != nil {
		log.Error("ChainManage", "updateChain failed", err)
		return false
	}
	return true
}

func updateStatus(block *meta.Block, isAdd bool) error {
	//GetManager().AccountManager.GetAllAccounts()
	//update mine account status

	if isAdd {
		//add block account status
		if err := updateAccountsByBlock(block); err != nil {
			return err
		}
	} else {
		//remove block account status
		if err := revertAccountsByBlock(block); err != nil {
			return err
		}
	}

	//update tx pool
	//update normal account status
	for _, tx := range block.GetTxs() {
		if isAdd {
			removeTransaction(*tx.GetTxID())
		} else {
			AddTransaction(&tx)
		}
	}
	return nil
}

func removeErrorNode(node meta.ChainNode) {
	deleteChain := -1
	deleteNode := -1
	for chainId, chain := range chains {
		for index, checkNode := range chain.Blocks {
			if node.IsEuqal(meta.NewPOAChainNode(&checkNode)) {
				deleteChain = chainId
				deleteNode = index
				break
			}
		}
		if deleteChain >= 0 && deleteNode >= 0 {
			break
		}
	}

	if deleteChain >= 0 && deleteNode >= 0 {
		chains[deleteChain].Blocks = append(chains[deleteChain].Blocks[:deleteNode], chains[deleteChain].Blocks[deleteNode+1:]...)
		return
	}
}
