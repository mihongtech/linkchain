package node

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/genesis"
	"github.com/linkchain/storage"
)

func initChainManager(n *Node, database lcdb.Database, path string) bool {
	log.Info("ChainManage init...")

	//load genesis from storage
	n.db = database
	genesisPath := path
	hash, err := n.initGenesis(genesisPath)
	if err != nil {
		log.Error("init genesis failed , exit", "err", err)
		return false
	}

	//create gensis chain
	gensisBlock := storage.GetBlock(n.db, hash, 0)
	n.addBlockCache(gensisBlock)
	gensisChain := meta.NewPOAChain(gensisBlock, nil)

	n.chains = make([]meta.ChainSketch, 0)
	n.chains = append(n.chains, *gensisChain)

	gensisChainNode := meta.NewPOAChainNode(gensisBlock)
	n.mainChainIndex = make([]meta.ChainNode, 0)

	n.mainChain = meta.NewBlockChain(gensisChainNode)

	err = n.loadBlocks()
	if err != nil {
		log.Error("load blocks from db failed , exit", "err", err)
		return false
	}

	return true
}

func (n *Node) initGenesis(genesisPath string) (math.Hash, error) {
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

	_, hash, err := genesis.SetupGenesisBlock(n.db, genesisBlock)
	if err != nil {
		log.Error("Setup genesis failed", "err", err)
		return math.Hash{}, errors.New("Setup genesis failed")
	}

	return hash, nil
}

func (n *Node) loadBlocks() error {
	hash := storage.GetHeadBlockHash(n.db)
	number := storage.GetBlockNumber(n.db, hash)
	var blocks []meta.Block
	log.Info("Best Node is", "hash", hash, "number", number)
	for i := number; i > 0; i-- {
		block := storage.GetBlock(n.db, hash, i)
		if block == nil {
			log.Error("get block failed", "hash", hash, "number", i)
			return errors.New("get block failed")
		}

		if err := n.addBlockCache(block); err != nil {
			log.Error("Add block failed", "hash", hash, "block", block)
			return err
		}
		log.Debug("Load block is", "hash", hash, "number", number, "block", block)

		blocks = append(blocks, *(block))

		hash = *block.GetPrevBlockID()
	}

	n.updateChainAndIndex()

	if _, err := n.getBestNode(); err != nil {
		return err
	}

	for i := len(blocks) - 1; i >= 0; i-- {
		log.Debug("sort block is", "block", blocks[i], "hash", blocks[i].GetBlockID())
		n.sortChains(&blocks[i])
		n.updateChainAndIndex()
	}

	return nil
}

func (n *Node) getBestBlock() *meta.Block {
	bestHeight, err := n.getBestHeight()
	if err != nil {
		return nil
	}

	n.chainMtx.RLock()
	hash := n.mainChainIndex[bestHeight].GetNodeHash()
	n.chainMtx.RUnlock()
	bestBlock, _ := n.getBlockByID(hash)
	return bestBlock
}

func (n *Node) getBestNode() (meta.ChainNode, error) {
	bestHeight, err := n.getBestHeight()
	if err != nil {
		return meta.ChainNode{}, errors.New("the chain is not init")
	}

	n.chainMtx.RLock()
	defer n.chainMtx.RUnlock()
	node := n.mainChainIndex[bestHeight]
	return node, nil
}

func (n *Node) getBestBlockHash() meta.BlockID {
	bestHeight, err := n.getBestHeight()
	if err != nil {
		return meta.BlockID{}
	}

	n.chainMtx.RLock()
	defer n.chainMtx.RUnlock()
	hash := n.mainChainIndex[bestHeight].GetNodeHash()
	return hash
}

func (n *Node) getBestHeight() (uint32, error) {
	n.chainMtx.RLock()
	defer n.chainMtx.RUnlock()
	bestHeight := len(n.mainChainIndex) - 1
	if bestHeight < 0 {
		return uint32(0), errors.New("the chain is not init")
	}
	return uint32(bestHeight), nil
}

func (n *Node) getBlockByHash(hash math.Hash) (*meta.Block, error) {
	//TODO need to lock chain
	b, err := n.getBlockByID(hash)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (n *Node) getBlockByHeight(height uint32) (*meta.Block, error) {
	if height < 0 || height > uint32(len(n.mainChainIndex)-1) {
		return nil, errors.New("ChainManage: getBlockByHeight->height is error")
	}

	n.chainMtx.RLock()
	hash := n.mainChainIndex[height].GetNodeHash()
	n.chainMtx.RUnlock()

	b, err := n.getBlockByID(hash)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (n *Node) getBlockNodeByHeight(height uint32) (meta.ChainNode, error) {
	if height > uint32(len(n.mainChainIndex)-1) {
		return meta.ChainNode{}, errors.New("the height is too large")
	}
	n.chainMtx.RLock()
	defer n.chainMtx.RUnlock()
	node := n.mainChainIndex[height]

	return node, nil
}

func (n *Node) getBlockChainInfo() string {
	n.chainMtx.RLock()
	defer n.chainMtx.RUnlock()

	log.Info("ChainManage mainchain",
		"chainHeight", n.mainChain.GetHeight(), "bestHash",
		n.mainChain.GetLastNode().GetNodeHash())

	//log.Info("ChainManage chains", "chains", len(chains))
	for i, chain := range n.chains {
		log.Info("ChainManage chains", "chainId", i, "chainHeight", chain.GetHeight(), "bestHash", chain.GetLastBlock().GetBlockID().GetString())
	}

	for e := n.mainChain.GetLastElement(); e != nil; e = e.Prev() {
		currentNode := e.Value.(meta.ChainNode)
		log.Info("ChainManage mainchain", "Height", currentNode.GetNodeHeight(), "current hash", currentNode.GetNodeHash(), "prev hash", currentNode.GetPrevHash())
	}

	for _, b := range n.mainChainIndex {
		log.Info("ChainManage mainchainIndex", "chainHeight", b.GetNodeHeight(), "bestHash", b.GetNodeHash())
	}

	return "this is poa chain"
}

func (n *Node) addBlock(block *meta.Block) {
	newblock := block

	n.addBlockCache(newblock)

	if err := storage.WriteBlock(n.db, block); err != nil {
		log.Error("WriteBlock to db failed", "error", err)
		return
	}

	_, err := n.getBestNode()
	if err != nil {
		log.Error("ChainManage", "error", err)
		return
	}
	n.sortChains(newblock)
	//longest, _ := GetLongestChain()
	//log.Info("AddBlock", "Longest ChainSketch height", len(longest.Blocks), "Longest ChainSketch bestHash", longest.GetLastBlock().GetBlockID().String())
}

func (n *Node) getLongestChain() (*meta.ChainSketch, int) {
	var lc meta.ChainSketch
	bestHeight := uint32(0)
	position := 0
	for i, chain := range n.chains {
		if bestHeight <= chain.GetHeight() {
			bestHeight = chain.GetHeight()
			lc = chain
			position = i
		}
	}
	return &lc, position
}

func (n *Node) updateChainAndIndex() bool {
	return n.updateChain() && n.updateChainIndex()
}

func (n *Node) sortChains(block *meta.Block) bool {
	n.chainMtx.Lock()
	defer n.chainMtx.Unlock()

	isUpdated := false
	deletIndex := make([]int, 0)
	blockNode := meta.NewPOAChainNode(block)

	prevBlock, err := n.getBlockByID(blockNode.GetPrevHash())
	//check the block's parent, if parent is not exist, the create is incomplete chain for create a chain when the parent is coming
	if err != nil {
		log.Error("ChainManage sortChains", "error", err)
		newChain := meta.NewPOAChain(nil, prevBlock)
		newChain.AddNewBlock(block)
		n.chains = append(n.chains, *newChain)
		return false
	}

	prevNode := meta.NewPOAChainNode(prevBlock)

	//1.find block Prev from mainChain
	if n.mainChain.IsOnChain(prevNode) {
		//If prevNodeInMain is bestNode: update chain; else : add new chain
		_, index := n.getLongestChain()
		err = n.chains[index].UpdateChainTop(block)
		if err != nil {
			newChain := n.chains[index].GetNewChain(prevBlock)
			newChain.AddNewBlock(block)
			n.chains = append(n.chains, *newChain)
			isUpdated = true
		}
	} else {
		//3.find block Prev from other sideChain,If cannot find then give up
		// a.update sidechain
		for index := range n.chains {
			err = n.chains[index].UpdateChainTop(block)
			if err == nil {
				// if update chain then check complete chain is the chain next
				isUpdated = true
				break
			}
		}
		if !isUpdated {
			// b.add new sidechain
			for index, chain := range n.chains {

				if !chain.IsInComplete {
					continue
				}

				ancestorBlock, err := n.getBlockAncestor(chain.GetLastBlock(), prevNode.GetNodeHeight()) //find prevheight block
				if err != nil {
					log.Error("sortChains addNewSideChain", "GetBlockAncestor", err)
					log.Info("sortChains :the chain is bad chain ,because the data of chain is imcomplete. the give up the chain")
					deletIndex = append(deletIndex, index)
					index--
					continue
				}
				ancestorNode := meta.NewPOAChainNode(ancestorBlock)

				if ancestorNode.IsEuqal(prevNode) {
					newChain := n.chains[index].GetNewChain(ancestorBlock)
					newChain.AddNewBlock(block)
					n.chains = append(n.chains, *newChain)
					isUpdated = true
					break
				}
			}
		}
	}

	//sort InCompleteChain
	for index := range n.chains {
		// if update chain then check complete chain is the chain next
		for i, imcompletchain := range n.chains {
			if imcompletchain.IsInComplete {
				if n.chains[index].CanLink(&imcompletchain) {
					n.chains[index].AddChain(&imcompletchain)
					deletIndex = append(deletIndex, i)
				}
			}
		}
	}
	//delete  imcomplete chain which have been added, or chain which need to giving up
	for _, index := range deletIndex {
		n.chains = append(n.chains[:index], n.chains[index+1:]...)
	}
	return isUpdated
}

/**
updateChainIndex
aim:update mainChainIndex from mainChain
TODO need to test
*/
func (n *Node) updateChainIndex() bool {
	n.chainMtx.Lock()
	defer n.chainMtx.Unlock()

	forkNode := n.mainChain.GetLastElement()
	forkPosition := len(n.mainChainIndex) - 1
	endNode := forkNode.Value.(meta.ChainNode)
	if forkPosition < 0 {
		//init mainchain index
		for e := n.mainChain.GetFristElement(); e != nil; e = e.Next() {
			node := e.Value.(meta.ChainNode)

			//add indexs(block status)
			b, err := n.getBlockByID(node.GetNodeHash())
			if err != nil {
				log.Error("ChainManage", "init new chain account failed. block hash", b.GetBlockID().GetString())
				return false
			}
			errorStatus := n.updateStatus(b, true)
			if errorStatus != nil {
				log.Error("ChainManage", "init new chain account failed", errorStatus)
				n.removeErrorNode(endNode)
				return false
			}

			n.mainChainIndex = append(n.mainChainIndex, node)
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
		checkIndexHash := n.mainChainIndex[forkPosition].GetNodeHash()
		if checkIndexHash.IsEqual(&nodeHash) {
			break
		}
		forkPosition--
	}

	//delete indexs after forkpoint
	//delete indexs(block status)
	for i := len(n.mainChainIndex) - 1; i >= forkPosition+1; i-- {
		b, err := n.getBlockByID(n.mainChainIndex[i].GetNodeHash())
		if err != nil {
			log.Error("ChainManage", "remove old chain account failed. block hash", b.GetBlockID().GetString())
			return false
		}
		errorStatus := n.updateStatus(b, false)
		if errorStatus != nil {
			log.Error("ChainManage", "remove old chain account failed", errorStatus)
			n.removeErrorNode(endNode)
			return false
		}
	}
	n.mainChainIndex = n.mainChainIndex[:forkPosition+1]

	//push index from the behind of forkNode which from mainChain
	for forkNode = forkNode.Next(); forkNode != nil; forkNode = forkNode.Next() {
		node := forkNode.Value.(meta.ChainNode)

		//add indexs(block status)
		b, err := n.getBlockByID(node.GetNodeHash())
		if err != nil {
			log.Error("ChainManage", "add new chain account failed. block hash", b.GetBlockID().GetString())
			return false
		}
		errorStatus := n.updateStatus(b, true)
		if errorStatus != nil {
			log.Error("ChainManage", "add new chain account failed", errorStatus)
			log.Error("ChainManage", "removeErrorNode", endNode.GetNodeHash().GetString())
			n.removeErrorNode(endNode)
			return false
		}

		n.mainChainIndex = append(n.mainChainIndex, node)
	}

	hash := n.mainChainIndex[len(n.mainChainIndex)-1].GetNodeHash()
	// TODO: add code to WriteCanonicalHash
	// storage.WriteCanonicalHash(db, math.BytesToHash(hash.CloneBytes()), uint64(mainChainIndex[len(mainChainIndex)-1].GetNodeHeight()))
	storage.WriteHeadBlockHash(n.db, math.BytesToHash(hash.CloneBytes()))
	return true
}

/**
updateChain
aim:update mainChain from chains
TODO need to test
*/
func (n *Node) updateChain() bool {
	n.chainMtx.Lock()
	defer n.chainMtx.Unlock()

	longestChain, _ := n.getLongestChain()
	bestBlock := longestChain.GetLastBlock()
	log.Info("ChainManage UpdateChain", "bestheight", bestBlock.GetHeight(), "bestblock", bestBlock.GetBlockID(), "prev", bestBlock.GetPrevBlockID())
	n.mainChain.AddNode(meta.NewPOAChainNode(bestBlock))

	err := n.mainChain.FillChain()
	if err != nil {
		log.Error("ChainManage", "updateChain failed", err)
		return false
	}
	return true
}

func (n *Node) updateStatus(block *meta.Block, isAdd bool) error {
	//GetManager().AccountManager.GetAllAccounts()
	//update mine account status

	if isAdd {
		//add block account status
		if err := n.updateAccountsByBlock(block); err != nil {
			return err
		}
	} else {
		//remove block account status
		if err := n.revertAccountsByBlock(block); err != nil {
			return err
		}
	}

	//update tx pool
	//update normal account status
	for _, tx := range block.GetTxs() {
		if isAdd {
			n.removeTransaction(*tx.GetTxID())
		} else {
			n.addTransaction(&tx)
		}
	}
	return nil
}

func (n *Node) removeErrorNode(node meta.ChainNode) {
	deleteChain := -1
	deleteNode := -1
	for chainId, chain := range n.chains {
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
		n.chains[deleteChain].Blocks = append(n.chains[deleteChain].Blocks[:deleteNode],
			n.chains[deleteChain].Blocks[deleteNode+1:]...)
		return
	}
}
