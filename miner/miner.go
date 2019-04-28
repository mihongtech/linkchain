package miner

import (
	"errors"
	"sync"
	"time"

	"github.com/mihongtech/linkchain/app/context"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/helper"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/node"
	"github.com/mihongtech/linkchain/txpool"
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

	coinbase := helper.CreateCoinBaseTx(*signerId, meta.NewAmount(config.DefaultBlockReward), block.GetHeight())
	block.SetTx(*coinbase)

	txs := m.txPoolAPI.GetAllTransaction()
	txs = m.executor.ChooseTransaction(txs, best, m.nodeAPI.GetOffChain(), m.walletAPI, signerId)
	block.SetTx(txs...)

	if !IsBestBlockOffspring(m.nodeAPI, block) {
		m.removeBlockTxs(block)
		return nil, errors.New("current block is not block prev")
	}

	//excute block status
	err, results, rootStatus, txFee := m.nodeAPI.ExecuteBlock(block)
	if err != nil {
		log.Error("Miner", "update Block status error", err, "block", block.String())
		m.removeBlockTxs(block)
		return nil, err
	}

	if err := m.executor.ExecuteResult(results, txFee, block); err != nil {
		m.removeBlockTxs(block)
		return nil, err
	}

	if !IsBestBlockOffspring(m.nodeAPI, block) {
		m.removeBlockTxs(block)
		return nil, errors.New("current block is not block prev")
	}

	block.Header.Status = rootStatus

	block, err = helper.RebuildBlock(block)
	if err != nil {
		log.Error("Miner", "Rebuild Block error", err)
		m.removeBlockTxs(block)
		return nil, err
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
		time.Sleep(time.Duration(config.DefaultPeriod) * time.Second)
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

func IsBestBlockOffspring(nodeAPI *node.PublicNodeAPI, block *meta.Block) bool {
	return block.GetPrevBlockID().IsEqual(nodeAPI.GetBestBlock().GetBlockID())
}
