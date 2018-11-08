package manage

import (
	"encoding/hex"
	"errors"
	"sync"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	globalconfig "github.com/linkchain/config"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
	poameta "github.com/linkchain/poa/meta"
	"github.com/linkchain/storage"
)

type ChainManage struct {
	chainMtx       sync.RWMutex
	chains         []poameta.Chain     //the chain tree for storing all chains
	mainChainIndex []poameta.ChainNode //the mainChain is slice for search block
	mainChain      poameta.BlockChain  //the mainChain is linked list for converting chain
	db             lcdb.Database
}

func (m *ChainManage) Init(i interface{}) bool {
	log.Info("ChainManage init...")

	//create gensis chain
	gensisBlock := GetManager().BlockManager.GetGensisBlock()

	gensisChain := poameta.NewPOAChain(gensisBlock, nil)
	m.chains = make([]poameta.Chain, 0)
	m.chains = append(m.chains, gensisChain)

	gensisChainNode := poameta.NewPOAChainNode(gensisBlock)
	m.mainChainIndex = make([]poameta.ChainNode, 0)

	m.mainChain = poameta.NewBlockChain(gensisChainNode)

	//TODO need to load storage
	var err error
	m.db, err = lcdb.NewLDBDatabase("data", 1024, 256)
	if err != nil {
		return false
	}

	//TODO BlockManager need inited
	return m.UpdateChain()
}

func (m *ChainManage) Start() bool {
	log.Info("ChainManage start...")
	//TODO need to updateMainChain
	return true
}

func (m *ChainManage) Stop() {
	log.Info("ChainManage stop...")
}

func (m *ChainManage) GetBestBlock() block.IBlock {
	bestHeight, err := m.GetBestHeight()
	if err != nil {
		return nil
	}

	m.chainMtx.RLock()
	hash := m.mainChainIndex[bestHeight].GetNodeHash()
	m.chainMtx.RUnlock()
	bestBlock, _ := GetManager().BlockManager.GetBlockByID(hash)
	return bestBlock
}

func (m *ChainManage) GetBestNode() (poameta.ChainNode, error) {
	bestHeight, err := m.GetBestHeight()
	if err != nil {
		return poameta.ChainNode{}, errors.New("the chain is not init")
	}

	m.chainMtx.RLock()
	defer m.chainMtx.RUnlock()
	node := m.mainChainIndex[bestHeight]
	return node, nil
}

func (m *ChainManage) GetBestBlockHash() meta.DataID {
	bestHeight, err := m.GetBestHeight()
	if err != nil {
		return nil
	}

	m.chainMtx.RLock()
	defer m.chainMtx.RUnlock()
	hash := m.mainChainIndex[bestHeight].GetNodeHash()
	return hash
}

func (m *ChainManage) GetBestHeight() (uint32, error) {
	m.chainMtx.RLock()
	defer m.chainMtx.RUnlock()
	bestHeight := len(m.mainChainIndex) - 1
	if bestHeight < 0 {
		return uint32(0), errors.New("thechain is not Init")
	}
	return uint32(bestHeight), nil
}

func (m *ChainManage) GetBlockByHash(hash math.Hash) (block.IBlock, error) {
	//TODO need to lock chain
	b, err := GetManager().BlockManager.GetBlockByID(&hash)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (m *ChainManage) GetBlockByHeight(height uint32) (block.IBlock, error) {
	if height < 0 || height > uint32(len(m.mainChainIndex)-1) {
		return nil, errors.New("ChainManage: GetBlockByHeight->height is error")
	}

	m.chainMtx.RLock()
	hash := m.mainChainIndex[height].GetNodeHash()
	m.chainMtx.RUnlock()

	b, err := GetManager().BlockManager.GetBlockByID(hash)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (m *ChainManage) GetBlockNodeByHeight(height uint32) (poameta.ChainNode, error) {
	if height > uint32(len(m.mainChainIndex)-1) {
		return poameta.ChainNode{}, errors.New("the height is too large")
	}
	m.chainMtx.RLock()
	defer m.chainMtx.RUnlock()
	node := m.mainChainIndex[height]

	return node, nil
}

func (m *ChainManage) GetBlockChainInfo() string {
	m.chainMtx.RLock()
	defer m.chainMtx.RUnlock()

	log.Info("ChainManage mainchain", "chainHeight", m.mainChain.GetHeight(), "bestHash", m.mainChain.GetLastNode().GetNodeHash())

	//log.Info("ChainManage chains", "chains", len(m.chains))
	for i, chain := range m.chains {
		log.Info("ChainManage chains", "chainId", i, "chainHeight", chain.GetHeight(), "bestHash", chain.GetLastBlock().GetBlockID().GetString())
	}

	/*for e := m.mainChain.GetLastElement(); e != nil; e = e.Prev() {
		currentNode := e.Value.(poameta.ChainNode)
		log.Info("ChainManage mainchain", "Height", currentNode.GetNodeHeight(), "current hash", currentNode.GetNodeHash(), "prev hash", currentNode.GetPrevHash())
	}*/

	/*for _, b := range m.mainChainIndex {
		log.Info("ChainManage mainchainIndex", "chainHeight", b.GetNodeHeight(), "bestHash", b.GetNodeHash())
	}*/

	return "this is poa chain"
}

func (m *ChainManage) AddBlock(block block.IBlock) {
	newblock := *block.(*poameta.Block)

	GetManager().BlockManager.AddBlock(&newblock)

	if err := storage.WriteBlock(m.db, block); err != nil {
		log.Error("WriteBlock to db failed", "error", err)
		return
	}

	_, err := m.GetBestNode()
	if err != nil {
		log.Error("ChainManage", "error", err)
		return
	}
	m.sortChains(newblock)
	//longest, _ := m.GetLongestChain()
	//log.Info("AddBlock", "Longest Chain height", len(longest.Blocks), "Longest Chain bestHash", longest.GetLastBlock().GetBlockID().String())
}

func (m *ChainManage) GetLongestChain() (poameta.Chain, int) {
	var lc poameta.Chain
	bestHeight := uint32(0)
	position := 0
	for i, chain := range m.chains {
		if bestHeight <= chain.GetHeight() {
			bestHeight = chain.GetHeight()
			lc = chain
			position = i
		}
	}
	return lc, position
}

func (m *ChainManage) UpdateChain() bool {
	return m.updateChain() && m.updateChainIndex()
}

func (m *ChainManage) sortChains(block poameta.Block) bool {
	m.chainMtx.Lock()
	defer m.chainMtx.Unlock()

	isUpdated := false
	deletIndex := make([]int, 0)
	blockNode := poameta.NewPOAChainNode(&block)

	prevBlock, err := GetManager().BlockManager.GetBlockByID(blockNode.GetPrevHash())
	//check the block's parent, if parent is not exist, the create is incomplete chain for create a chain when the parent is coming
	if err != nil {
		log.Error("ChainManage sortChains", "error", err)
		newChain := poameta.NewPOAChain(nil, prevBlock)
		newChain.AddNewBlock(&block)
		m.chains = append(m.chains, newChain)
		return false
	}

	prevNode := poameta.NewPOAChainNode(prevBlock)

	//1.find block Prev from mainChain
	if m.mainChain.IsOnChain(prevNode) {
		//If prevNodeInMain is bestNode: update chain; else : add new chain
		_, index := m.GetLongestChain()
		err = m.chains[index].UpdateChainTop(&block)
		if err != nil {
			newChain := m.chains[index].GetNewChain(prevBlock)
			newChain.AddNewBlock(&block)
			m.chains = append(m.chains, newChain)
			isUpdated = true
		}
	} else {
		//3.find block Prev from other sideChain,If cannot find then give up
		// a.update sidechain
		for index := range m.chains {
			err = m.chains[index].UpdateChainTop(&block)
			if err == nil {
				// if update chain then check complete chain is the chain next
				isUpdated = true
				break
			}
		}
		if !isUpdated {
			// b.add new sidechain
			for index, chain := range m.chains {

				if !chain.IsInComplete {
					continue
				}

				ancestorBlock, err := GetManager().BlockManager.GetBlockAncestor(chain.GetLastBlock(), prevNode.GetNodeHeight()) //find prevheight block
				if err != nil {
					log.Error("sortChains addNewSideChain", "GetBlockAncestor", err)
					log.Info("sortChains :the chain is bad chain ,because the data of chain is imcomplete. the give up the chain")
					deletIndex = append(deletIndex, index)
					index--
					continue
				}
				ancestorNode := poameta.NewPOAChainNode(ancestorBlock)

				if ancestorNode.IsEuqal(prevNode) {
					newChain := m.chains[index].GetNewChain(ancestorBlock)
					newChain.AddNewBlock(&block)
					m.chains = append(m.chains, newChain)
					isUpdated = true
					break
				}
			}
		}
	}

	//sort InCompleteChain
	for index := range m.chains {
		// if update chain then check complete chain is the chain next
		for i, imcompletchain := range m.chains {
			if imcompletchain.IsInComplete {
				if m.chains[index].CanLink(imcompletchain) {
					m.chains[index].AddChain(imcompletchain)
					deletIndex = append(deletIndex, i)
				}
			}
		}
	}
	//delete  imcomplete chain which have been added, or chain which need to giving up
	for _, index := range deletIndex {
		m.chains = append(m.chains[:index], m.chains[index+1:]...)
	}
	return isUpdated
}

/**
updateChainIndex
aim:update mainChainIndex from mainChain
TODO need to test
*/
func (m *ChainManage) updateChainIndex() bool {
	m.chainMtx.Lock()
	defer m.chainMtx.Unlock()

	forkNode := m.mainChain.GetLastElement()
	forkPosition := len(m.mainChainIndex) - 1
	endNode := forkNode.Value.(poameta.ChainNode)
	if forkPosition < 0 {
		//init mainchain index
		for e := m.mainChain.GetFristElement(); e != nil; e = e.Next() {
			node := e.Value.(poameta.ChainNode)

			//add indexs(block status)
			b, err := GetManager().BlockManager.GetBlockByID(node.GetNodeHash())
			if err != nil {
				log.Error("ChainManage", "init new chain account failed. block hash", b.GetBlockID().GetString())
				return false
			}
			errorStatus := m.updateStatus(b, true)
			if errorStatus != nil {
				log.Error("ChainManage", "init new chain account failed", errorStatus)
				m.removeErrorNode(endNode)
				return false
			}

			m.mainChainIndex = append(m.mainChainIndex, node)
		}
		return true
	}

	for ; forkNode != nil && forkPosition >= 0; forkNode = forkNode.Prev() {
		node := forkNode.Value.(poameta.ChainNode)
		nodeHash := node.GetNodeHash()
		if node.GetNodeHeight() > uint32(forkPosition) {
			continue
		} else if int(node.GetNodeHeight()) < forkPosition {
			forkPosition--
			continue
		}
		checkIndexHash := m.mainChainIndex[forkPosition].GetNodeHash()
		if checkIndexHash.IsEqual(nodeHash) {
			break
		}
		forkPosition--
	}

	//delete indexs after forkpoint
	//delete indexs(block status)
	for i := len(m.mainChainIndex) - 1; i >= forkPosition+1; i-- {
		b, err := GetManager().BlockManager.GetBlockByID(m.mainChainIndex[i].GetNodeHash())
		if err != nil {
			log.Error("ChainManage", "remove old chain account failed. block hash", b.GetBlockID().GetString())
			return false
		}
		errorStatus := m.updateStatus(b, false)
		if errorStatus != nil {
			log.Error("ChainManage", "remove old chain account failed", errorStatus)
			m.removeErrorNode(endNode)
			return false
		}
	}
	m.mainChainIndex = m.mainChainIndex[:forkPosition+1]

	//push index from the behind of forkNode which from mainChain
	for forkNode = forkNode.Next(); forkNode != nil; forkNode = forkNode.Next() {
		node := forkNode.Value.(poameta.ChainNode)

		//add indexs(block status)
		b, err := GetManager().BlockManager.GetBlockByID(node.GetNodeHash())
		if err != nil {
			log.Error("ChainManage", "add new chain account failed. block hash", b.GetBlockID().GetString())
			return false
		}
		errorStatus := m.updateStatus(b, true)
		if errorStatus != nil {
			log.Error("ChainManage", "add new chain account failed", errorStatus)
			log.Error("ChainManage", "removeErrorNode", endNode.GetNodeHash().GetString())
			m.removeErrorNode(endNode)
			return false
		}

		m.mainChainIndex = append(m.mainChainIndex, node)
	}
	return true
}

/**
updateChain
aim:update mainChain from chains
TODO need to test
*/
func (m *ChainManage) updateChain() bool {
	m.chainMtx.Lock()
	defer m.chainMtx.Unlock()

	longestChain, _ := m.GetLongestChain()
	bestBlock := longestChain.GetLastBlock()
	log.Info("ChainManage UpdateChain", "bestheight", bestBlock.GetHeight(), "bestblock", bestBlock.GetBlockID(), "prev", bestBlock.GetPrevBlockID())
	m.mainChain.AddNode(poameta.NewPOAChainNode(bestBlock))

	err := m.mainChain.FillChain(GetManager().BlockManager)
	if err != nil {
		log.Error("ChainManage", "updateChain failed", err)
		return false
	}
	return true
}

func (m *ChainManage) updateStatus(block block.IBlock, isAdd bool) error {
	//GetManager().AccountManager.GetAllAccounts()
	//update mine account status
	poablock := *block.(*poameta.Block)

	amount := poameta.NewAmout(50)
	signerIndex := block.GetHeight() % uint32(len(globalconfig.SignMiners))
	minerPK, err := hex.DecodeString(globalconfig.SignMiners[signerIndex])
	if err != nil {
		return err
	}
	pk, err := btcec.ParsePubKey(minerPK, btcec.S256())
	sAccount := poameta.NewAccountId(pk)
	tp := poameta.NewTransactionPeer(*sAccount, poablock.Header.Sign)
	mineTx := poameta.NewTransaction(0, poameta.TransactionPeer{}, *tp, *amount, poablock.Header.Time, poablock.Header.Nonce, nil, poameta.FromSign{})
	cachTxs := block.GetTxs()
	mineIndex := len(cachTxs)
	cachTxs = append(cachTxs, mineTx)
	if isAdd {
		err := GetManager().AccountManager.UpdateAccountsByTxs(cachTxs, mineIndex)
		if err != nil {
			return err
		}
	} else {
		err := GetManager().AccountManager.RevertAccountsByTxs(cachTxs, mineIndex)
		if err != nil {
			return err
		}
	}

	//update tx pool
	//update normal account status
	//cachTxs = append(cachTxs[:mineIndex], cachTxs[mineIndex+1:]...) //Delete mineTx
	for _, tx := range block.GetTxs() {
		if isAdd {
			GetManager().TransactionManager.RemoveTransaction(tx.GetTxID())
		} else {
			GetManager().TransactionManager.AddTransaction(tx)
		}
	}
	return nil
}

func (m *ChainManage) removeErrorNode(node poameta.ChainNode) {
	deleteChain := -1
	deleteNode := -1
	for chainId, chain := range m.chains {
		for index, checkNode := range chain.Blocks {
			if node.IsEuqal(poameta.NewPOAChainNode(&checkNode)) {
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
		m.chains[deleteChain].Blocks = append(m.chains[deleteChain].Blocks[:deleteNode], m.chains[deleteChain].Blocks[deleteNode+1:]...)
		return
	}
}
