package miner

import (
	"container/heap"
	"errors"
	"github.com/mihongtech/linkchain/app/context"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/helper"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/node"
	"github.com/mihongtech/linkchain/txpool"
	"sync"
	"time"
)

type Miner struct {
	nodeAPI   *node.PublicNodeAPI
	executor  interpreter.Executor
	walletAPI interpreter.Wallet
	txPoolAPI *txpool.TxPool
	isMining  bool
	minerMtx  sync.Mutex
}

func NewMiner() *Miner {
	return &Miner{isMining: false}
}
func (m *Miner) Setup(i interface{}) bool {
	m.nodeAPI = i.(*context.Context).NodeAPI.(*node.PublicNodeAPI)
	m.walletAPI = i.(*context.Context).WalletAPI
	m.txPoolAPI = i.(*context.Context).TxpoolAPI.(*txpool.TxPool)
	m.executor = i.(*context.Context).InterpreterAPI
	return true
}

func (m *Miner) Start() bool {
	log.Info("Miner start...")
	return true
}

func (m *Miner) Stop() {
	log.Info("Miner stop...")
}

func (m *Miner) MineBlock() (*meta.Block, error) {
	signerId, err := m.getMineBlock()
	if err != nil {
		return nil, errors.New("the node can not mine block" + err.Error())
	}
	best := m.nodeAPI.GetBestBlock()
	block, err := helper.CreateBlock(best.GetHeight(), *best.GetBlockID())
	if err != nil {
		log.Error("Miner", "New Block error", err)
		return nil, err
	}

	difficulty, err := m.nodeAPI.CalcNextRequiredDifficulty()
	if err != nil {
		m.removeBlockTxs(block)
		return nil, err
	}

	coinbase := helper.CreateCoinBaseTx(*signerId, meta.NewAmount(config.DefaultBlockReward), block.GetHeight())
	block.SetTx(*coinbase)

	txs := m.txPoolAPI.GetAllTransaction()
	txs = m.executor.ChooseTransaction(txs, best, m.nodeAPI.GetOffChain(), m.walletAPI, signerId)

	tq := make(TxDescQueue, 0)
	for _, tx := range txs {
		tq.Push(&TxDesc{
			tx:  &tx,
			fee: m.calcGasFee(&tx),
		})
	}
	heap.Init(&tq)

	err = m.generateValidBlock(tq, block, signerId, false)
	if err != nil {
		m.removeBlockTxs(block)
		return nil, err
	}

	for extraNonce := uint32(0); extraNonce < ^uint32(0); extraNonce++ {
		block.Header.Nonce = extraNonce
		err := block.Deserialize(block.Serialize())
		if err != nil {
			return nil, err
		}
		if node.HashToBig(block.GetBlockID()).Cmp(node.CompactToBig(difficulty)) < 0 {
			block.Header.Difficulty = difficulty
			block.Header.Time = time.Now()
			break
		}
	}

	err = m.signBlock(*signerId, block)
	log.Debug("Miner", "signer", signerId.String())
	if err != nil {
		log.Error("Miner", "sign Block status error", err)
		m.removeBlockTxs(block)
		return nil, err
	}

	err = m.nodeAPI.ProcessBlock(block)
	if err != nil {
		m.removeBlockTxs(block)
		return nil, err
	}
	m.nodeAPI.GetBlockEvent().Post(node.NewMinedBlockEvent{Block: block})

	return block, nil
}

func (m *Miner) signBlock(signer meta.AccountID, block *meta.Block) error {
	sign, err := m.walletAPI.SignMessage(signer, block.GetBlockID().CloneBytes())
	if err != nil {
		return err
	}
	block.SetSign(sign)
	return nil
}

func (m *Miner) StartMine() error {
	m.minerMtx.Lock()
	if m.isMining {
		m.minerMtx.Unlock()
		return errors.New("the node is mining")
	}
	m.isMining = true
	m.minerMtx.Unlock()
	for true {
		m.minerMtx.Lock()
		tempMing := m.isMining
		m.minerMtx.Unlock()
		if !tempMing {
			break
		}
		m.MineBlock()
		//time.Sleep(time.Duration(config.DefaultPeriod) * time.Second)
	}
	return nil
}

func (m *Miner) StopMine() {
	m.minerMtx.Lock()
	defer m.minerMtx.Unlock()
	m.isMining = false
}

func (m *Miner) GetInfo() bool {
	m.minerMtx.Lock()
	defer m.minerMtx.Unlock()
	return m.isMining
}

//check miner if not can mine next block
func (m *Miner) getMineBlock() (*meta.AccountID, error) {
	best := m.nodeAPI.GetBestBlock()
	newBlock, err := helper.CreateBlock(best.GetHeight(), *best.GetBlockID())
	if err != nil {
		return nil, err
	}
	signerStr := m.nodeAPI.GetEngine().GetBlockSigner(&newBlock.Header)

	if _, err = meta.NewAccountIdFromStr(signerStr); err != nil {
		log.Error("Get signer account id failed", "err", err)
		return nil, err
	}

	signer, err := m.walletAPI.GetAccount(signerStr)
	if err != nil {

		log.Error("GetAccount failed", "err", err, "signerStr", signerStr)
		return nil, err
	}
	return signer.GetAccountID(), nil
}

func (m *Miner) removeBlockTxs(block *meta.Block) {
	for index := range block.TXs {
		m.txPoolAPI.RemoveTransaction(*block.TXs[index].GetTxID())
	}
}

// choose transaction from txpool
// ensure not oversized block size limit
func (m *Miner) generateValidBlock(tq TxDescQueue, block *meta.Block, signerId *meta.AccountID, oversized bool) error {
	// if prev generate block ovsized, remove last transaction
	if oversized {
		err := block.RemoveLastTx()
		if err != nil {
			return err
		}
	} else {
		// if has transaction add a transaction
		if tq.Len() > 0 {
			txDesc := tq.Pop().(*TxDesc)
			err := block.SetTx(*txDesc.tx)
			if err != nil {
				return err
			}
		}
	}

	if !IsBestBlockOffspring(m.nodeAPI, block) {
		m.removeBlockTxs(block)
		return errors.New("current block is not block prev")
	}

	//excute block status
	err, results, rootStatus, txFee := m.nodeAPI.ExecuteBlock(block)
	if err != nil {
		log.Error("Miner", "update Block status error", err, "block", block.String())
		m.removeBlockTxs(block)
		return err
	}

	if err := m.executor.ExecuteResult(results, txFee, block); err != nil {
		m.removeBlockTxs(block)
		return err
	}

	if !IsBestBlockOffspring(m.nodeAPI, block) {
		m.removeBlockTxs(block)
		return errors.New("current block is not block prev")
	}

	block.Header.Status = rootStatus

	block, err = helper.RebuildBlock(block)
	if err != nil {
		log.Error("Miner", "Rebuild Block error", err)
		m.removeBlockTxs(block)
		return err
	}

	// verify block size
	err = block.VerifySize()
	if err != nil {
		// oversized error
		if err == meta.BlockOversizeErr {
			return m.generateValidBlock(tq, block, signerId, true)
		} else {
			return err
		}
	}

	if oversized {
		return nil
	} else {
		if tq.Len() > 0 {
			return m.generateValidBlock(tq, block, signerId, false)
		} else {
			return nil
		}
	}

}

func (m *Miner) calcGasFee(tx *meta.Transaction) int64 {
	if tx.Type == config.CoinBaseTx {
		return 0
	}
	var inTotal int64
	for _, coin := range tx.From.Coins {
		for _, ticket := range coin.Ticket {
			tx, _, _, _ := m.nodeAPI.GetTXByID(ticket.Txid)
			inTotal += tx.To.Coins[ticket.Index].GetValue().GetInt64()
		}
	}
	var toTotal int64
	for _, coin := range tx.To.Coins {
		toTotal += coin.GetValue().GetInt64()
	}
	return toTotal - inTotal
}

func IsBestBlockOffspring(nodeAPI *node.PublicNodeAPI, block *meta.Block) bool {
	return block.GetPrevBlockID().IsEqual(nodeAPI.GetBestBlock().GetBlockID())
}

type TxDesc struct {
	tx  *meta.Transaction
	fee int64
}

type TxDescQueue []*TxDesc

func (tq *TxDescQueue) Len() int {
	return len(*tq)
}

func (tq *TxDescQueue) Less(i, j int) bool {
	return (*tq)[i].fee < (*tq)[j].fee
}

func (tq *TxDescQueue) Swap(i, j int) {
	(*tq)[i], (*tq)[j] = (*tq)[j], (*tq)[i]
}

func (tq *TxDescQueue) Push(x interface{}) {
	*tq = append(*tq, x.(*TxDesc))
}

func (tq *TxDescQueue) Pop() interface{} {
	n := len(*tq)
	item := (*tq)[n-1]
	*tq = (*tq)[0 : n-1]
	return item
}
