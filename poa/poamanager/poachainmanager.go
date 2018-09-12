package poamanager

import (
	"errors"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/common/math"
	"github.com/linkchain/meta/block"
	poameta "github.com/linkchain/poa/meta"
)

type POAChainManager struct {
	chains []poameta.POAChain	//the chain tree for storing all chains
	mainChainIndex []poameta.POAChainNode //the mainChain is slice for search block
	mainChain poameta.BlockChain	//the mainChain is linked list for converting chain
}

func (m *POAChainManager) Init(i interface{}) bool{
	log.Info("POAChainManager init...");

	//create gensis chain
	gensisBlock := GetManager().BlockManager.GetGensisBlock()

	gensisChain := poameta.NewPOAChain(gensisBlock,nil)
	m.chains = make([]poameta.POAChain,0)
	m.chains = append(m.chains,gensisChain)

	gensisChainNode := poameta.NewPOAChainNode(gensisBlock)
	m.mainChainIndex = make([]poameta.POAChainNode,0)
	//m.mainChainIndex = append(m.mainChainIndex,gensisChainNode)

	m.mainChain = poameta.NewBlockChain(gensisChainNode)
	/*longestChain,_ := m.GetLongestChain()
	bestBlock := longestChain.GetLastBlock()
	m.mainChain.AddNode(poameta.NewPOAChainNode(bestBlock))*/
	//TODO need to load storage

	//TODO BlockManager need inited
	return m.UpdateChain()
}

func (m *POAChainManager) Start() bool{
	log.Info("POAChainManager start...");
	//TODO need to updateMainChain
	return true
}

func (m *POAChainManager) Stop(){
	log.Info("POAChainManager stop...");
}

func (m *POAChainManager) GetBestBlock() block.IBlock  {
	bestHeight,error := m.GetBestHeight()
	if error != nil {
		return nil
	}
	log.Info("GetBestBlock","bestHeight",bestHeight)
	hash := m.mainChainIndex[bestHeight].GetNodeHash()
	bestBlock,_ := GetManager().BlockManager.GetBlockByID(hash)
	return bestBlock
}

func (m *POAChainManager) GetBestNode() (poameta.POAChainNode,error)  {
	bestHeight,error := m.GetBestHeight()
	if error != nil {
		return poameta.POAChainNode{},errors.New("the chain is not init")
	}
	return m.mainChainIndex[bestHeight],nil
}

func (m *POAChainManager) GetBestBlockHash() block.IBlockID  {
	bestHeight,error := m.GetBestHeight()
	if error != nil {
		return nil
	}
	hash := m.mainChainIndex[bestHeight].GetNodeHash()
	return hash
}

func (m *POAChainManager) GetBestHeight() (uint32,error) {
	bestHeight := len(m.mainChainIndex) - 1
	if bestHeight < 0 {
		return uint32(0),errors.New("thechain is not Init")
	}
	return uint32(bestHeight),nil
}

func (m *POAChainManager) GetBlockByHash(hash math.Hash) block.IBlock  {
	block,_ := GetManager().BlockManager.GetBlockByID(hash)
	return block
}

func (m *POAChainManager) GetBlockByHeight(height uint32) block.IBlock  {
	block,_ := GetManager().BlockManager.GetBlockByID(m.mainChainIndex[height].GetNodeHash())
	return block
}

func (m *POAChainManager) GetBlockNodeByHeight(height uint32) (poameta.POAChainNode,error)  {
	if height > uint32(len(m.mainChainIndex)-1){
		return poameta.POAChainNode{},errors.New("the height is too large")
	}
	return m.mainChainIndex[height],nil
}

func (m *POAChainManager) GetBlockChainInfo() string  {

	log.Info("POAChainManager","chains",len(m.chains))
	for i,chain := range m.chains {
		log.Info("POAChainManager chain","chainId",i,"chainHeight",chain.GetHeight(),"bestHash",chain.GetLastBlock().GetBlockID().GetString())
	}

	for _,block := range m.mainChainIndex {
		log.Info("POAChainManager mainchain","chainHeight",block.GetNodeHeight(),"bestHash",block.GetNodeHash())
	}

	return "this is poa chain";
}

func (m *POAChainManager) AddBlock(block block.IBlock)  {
	newblock := *block.(*poameta.POABlock)

	GetManager().BlockManager.AddBlock(&newblock)

	_,error := m.GetBestNode()
	if error != nil {
		log.Error("POAChainManager","error",error)
		return
	}
	m.sortChains(newblock)
	longest,_ := m.GetLongestChain()
	log.Info("AddBlock","Longest Chain height",len(longest.Blocks),"Longest Chain bestHash",longest.GetLastBlock().GetBlockID().GetString())
}

func (m *POAChainManager) GetBlockAncestor(block block.IBlock,height uint32) block.IBlock {
	if height > block.GetHeight(){
		log.Error("POAChainManager","GetBlockAncestor error", "height is plus block's height")
		return nil
	}else {
		ancestor := block
		var e error
		for height < block.GetHeight() {
			ancestor,e = GetManager().BlockManager.GetBlockByID(ancestor.GetPrevBlockID())
			if e != nil{
				log.Error("POAChainManager","GetBlockAncestor error", "can not find ancestor")
				return nil
			}
		}
		return ancestor
	}
}

func (m *POAChainManager) GetLongestChain() (poameta.POAChain,int)  {
	var mainChainIndex poameta.POAChain
	bestHeight := uint32(0);
	position := 0
	for i,chain := range m.chains {
		if bestHeight <= chain.GetHeight() {
			bestHeight = chain.GetHeight()
			mainChainIndex = chain
			position = i
		}
	}
	return mainChainIndex,position
}

func (m *POAChainManager) UpdateChain() bool  {
	return m.updateChain() && m.updateChainIndex()
}

func (m *POAChainManager) sortChains(block poameta.POABlock) bool  {
	isUpdated := false
	deletIndex := make([]int,0)
	blockNode := poameta.NewPOAChainNode(&block)

	prevBlock,error := GetManager().BlockManager.GetBlockByID(blockNode.GetPrevHash())
	//check the block's parent, if parent is not exist, the create is incomplete chain for create a chain when the parent is coming
	if error != nil {
		log.Error("POAChainManager sortChains","error",error)
		newChain := poameta.NewPOAChain(nil,prevBlock)
		newChain.AddNewBlock(&block)
		m.chains = append(m.chains,newChain)
		return isUpdated
	}

	prevNode := poameta.NewPOAChainNode(prevBlock)

	//1.find block Prev from mainChain
	if m.mainChain.IsOnChain(prevNode) {
		//If prevNodeInMain is bestNode: update chain; else : add new chain
		_,index := m.GetLongestChain()
		error = m.chains[index].UpdateChainTop(&block)
		if error != nil {
			newChain := m.chains[index].GetNewChain(prevBlock)
			newChain.AddNewBlock(&block)
			m.chains = append(m.chains,newChain)
			isUpdated = true
		}
	} else {
		//3.find block Prev from other sideChain,If cannot find then give up
		// a.update sidechain
		for index,_ := range m.chains {
			error = m.chains[index].UpdateChainTop(&block)
			if error == nil {
				// if update chain then check complete chain is the chain next
				isUpdated = true
			}
		}
		// b.add new sidechain
		for index,chain := range m.chains{

			if !chain.IsInComplete {
				continue
			}

			ancestorBlock := m.GetBlockAncestor(chain.GetLastBlock(),prevNode.GetNodeHeight())
			if ancestorBlock == nil {
				log.Info("sortChains :the chain is bad chain ,because the data of chain is imcomplete. the give up the chain")
				deletIndex = append(deletIndex,index)
				index--
			}
			ancestorNode := poameta.NewPOAChainNode(ancestorBlock)

			if ancestorNode.IsEuqal(prevNode){
				newChain := m.chains[index].GetNewChain(ancestorBlock)
				newChain.AddNewBlock(&block)
				m.chains = append(m.chains,newChain)
				isUpdated = true
			}
		}
	}

	//sort InCompleteChain
	for index,_ := range m.chains {
		// if update chain then check complete chain is the chain next
		for i,imcompletchain := range m.chains{
			if imcompletchain.IsInComplete {
				if m.chains[index].CanLink(imcompletchain){m.chains[index].AddChain(imcompletchain)
					deletIndex = append(deletIndex,i)
				}
			}
		}
	}
	//delete  imcomplete chain which have been added, or chain which need to giving up
	for _,index := range deletIndex {
		m.chains = append(m.chains[:index],m.chains[index+1:]...)
	}
	return true
}

/**
	updateChainIndex
	aim:update mainChainIndex from mainChain
	TODO need to test
 */
func (m *POAChainManager) updateChainIndex() bool  {
	forkNode := m.mainChain.GetLastElement()
	forkPosition := len(m.mainChainIndex) - 1
	endNode := forkNode.Value.(poameta.POAChainNode)
	defer GetManager().AccountManager.GetAllAccounts()
	if forkPosition < 0 {
		//init mainchain index
		for e := m.mainChain.GetFristElement(); e != nil; e = e.Next(){
			node := e.Value.(poameta.POAChainNode)

			//add indexs(block status)
			block,error := GetManager().BlockManager.GetBlockByID(node.GetNodeHash())
			if error != nil {
				log.Error("POAChainManager","add new chain account failed. block hash",block.GetBlockID().GetString())
				return false
			}
			errorStatus := m.updateStatus(block,true)
			if errorStatus != nil {
				log.Error("POAChainManager","add new chain account failed",errorStatus)
				m.removeErrorNode(endNode)
				return false
			}

			m.mainChainIndex = append(m.mainChainIndex,node)
		}
		return true
	}

	for ; forkNode != nil && forkPosition >= 0 ; forkNode = forkNode.Prev() {
		node := forkNode.Value.(poameta.POAChainNode)
		nodeHash := node.GetNodeHash().(math.Hash)
		if node.GetNodeHeight() > uint32(forkPosition) {
			continue
		} else if int(node.GetNodeHeight()) < forkPosition{
			forkPosition--
			continue
		}
		checkIndexHash := m.mainChainIndex[forkPosition].GetNodeHash().(math.Hash)
		if checkIndexHash.IsEqual(&nodeHash) {
			break
		}
		forkPosition--
	}

	//delete indexs after forkpoint
	//delete indexs(block status)
	for i := len(m.mainChainIndex) - 1 ; i >= forkPosition+1; i-- {
		block,error := GetManager().BlockManager.GetBlockByID(m.mainChainIndex[i].GetNodeHash())
		if error != nil {
			log.Error("POAChainManager","remove old chain account failed. block hash",block.GetBlockID().GetString())
			return false
		}
		errorStatus := m.updateStatus(block,false)
		if errorStatus != nil {
			log.Error("POAChainManager","remove old chain account failed",errorStatus)
			m.removeErrorNode(endNode)
			return false
		}
	}
	m.mainChainIndex = m.mainChainIndex[:forkPosition+1]


	//push index from the behind of forkNode which from mainChain
	for forkNode = forkNode.Next(); forkNode != nil; forkNode = forkNode.Next() {
		node := forkNode.Value.(poameta.POAChainNode)

		//add indexs(block status)
		block,error := GetManager().BlockManager.GetBlockByID(node.GetNodeHash())
		if error != nil {
			log.Error("POAChainManager","add new chain account failed. block hash",block.GetBlockID().GetString())
			return false
		}
		errorStatus := m.updateStatus(block,true)
		if errorStatus != nil {
			log.Error("POAChainManager","add new chain account failed",errorStatus)
			m.removeErrorNode(endNode)
			return false
		}

		m.mainChainIndex = append(m.mainChainIndex,node)
	}
	return true
}


/**
	updateChain
	aim:update mainChain from chains
	TODO need to test
*/
func (m *POAChainManager) updateChain() bool  {
	longestChain,_ := m.GetLongestChain()
	bestBlock := longestChain.GetLastBlock()
	m.mainChain.AddNode(poameta.NewPOAChainNode(bestBlock))

	error := m.mainChain.FillChain(GetManager().BlockManager)
	if error != nil {
		log.Error("POAChainManager","updateChain failed",error)
		return false
	}
	return true
}

func (m *POAChainManager) updateStatus(block block.IBlock,isAdd bool) error {
	//check all from account
	for _,tx := range block.GetTxs() {
		error := GetManager().AccountManager.UpdateAccountByTX(tx)
		if error != nil {
			return error
		}
	}
	//update mine account status
	poablock := *block.(*poameta.POABlock)
	mineAccountId := *poablock.Header.GetMineAccount().(*poameta.POAAccountID)
	var mineAccount poameta.POAAccount

	if isAdd {
		mineAccount = poameta.POAAccount{AccountID:mineAccountId, Value:poameta.POAAmount{Value:50}}
	} else {
		mineAccount = poameta.POAAccount{AccountID:mineAccountId, Value:poameta.POAAmount{Value:-50}}
	}

	error := GetManager().AccountManager.UpdateAccount(&mineAccount)
	if error != nil {
		return error
	}

	//update normal account status
	for _,tx := range block.GetTxs() {
		var actualTx poameta.POATransaction
		if isAdd {
			actualTx = *tx.(*poameta.POATransaction)
		} else {
			actualTx = *tx.ChangeFromTo().(*poameta.POATransaction)
		}
		error := GetManager().AccountManager.UpdateAccountByTX(&actualTx)
		if error != nil {
			return error
		}
	}
	return nil
}

func (m *POAChainManager) removeErrorNode(node poameta.POAChainNode)  {
	deleteChain := -1
	deleteNode := -1
	for chainId,chain := range m.chains {
		for index,checkNode := range chain.Blocks {
			if node.IsEuqal(poameta.NewPOAChainNode(&checkNode)){
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
		m.chains[deleteChain].Blocks = append(m.chains[deleteChain].Blocks[:deleteNode],m.chains[deleteChain].Blocks[deleteNode+1:]...)
		return
	}
}


