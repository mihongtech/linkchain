package poa

import (
	"errors"
	"sync"
	"time"

	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/helper"
	"github.com/mihongtech/linkchain/node/bcsi"
	"github.com/mihongtech/linkchain/node/chain"
	"github.com/mihongtech/linkchain/node/pool"
)

type Miner struct {
	poa      *Poa
	chain    chain.Chain
	txPool   pool.TxPool
	bcsiAPI  bcsi.BCSI
	isMining bool
	minerMtx sync.Mutex
}

func NewMiner(poa *Poa) *Miner {
	return &Miner{isMining: false, poa: poa}
}

func (m *Miner) Setup(i interface{}) bool {
	return true
}

func (m *Miner) Start() bool {
	log.Info("Miner start...")
	go m.StartMine()
	return true
}

func (m *Miner) Stop() {
	log.Info("Miner stop...")
	go m.StopMine()
}

func (m *Miner) MineBlock() (*meta.Block, error) {
	best := m.chain.GetBestBlock()
	block, err := helper.CreateBlock(best.GetHeight(), *best.GetBlockID())
	if err != nil {
		log.Error("Miner", "New Block error", err)
		return nil, err
	}
	signer := m.poa.getBlockSigner(block)
	coinbase := helper.CreateCoinBaseTx(signer, meta.NewAmount(config.DefaultBlockReward), block.GetHeight())
	block.SetTx(*coinbase)

	txs := m.txPool.GetAllTransaction()
	txs = m.bcsiAPI.FilterTx(txs)
	block.SetTx(txs...)

	if !IsBestBlockOffspring(m.chain, block) {
		m.removeBlockTxs(block)
		return nil, errors.New("current block is not block prev")
	}

	block.Header.Status = m.bcsiAPI.GetBlockState(*best.GetBlockID()) //The block status is prev block status

	block, err = helper.RebuildBlock(block)
	if err != nil {
		log.Error("Miner", "Rebuild Block error", err)
		m.removeBlockTxs(block)
		return nil, err
	}

	err = m.signBlock(signer, block)
	log.Debug("Miner", "signer", signer.String())
	if err != nil {
		log.Error("Miner", "sign Block status error", err)
		m.removeBlockTxs(block)
		return nil, err
	}

	err = m.chain.ProcessBlock(block)
	if err != nil {
		m.removeBlockTxs(block)
		return nil, err
	}

	return block, nil
}

func (m *Miner) signBlock(signer meta.AccountID, block *meta.Block) error {
	//TODO need to add poa sign
	//sign, err := m.walletAPI.SignMessage(signer, block.GetBlockID().CloneBytes())
	//if err != nil {
	//	return err
	//}
	//block.SetSign(sign)
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

func (m *Miner) removeBlockTxs(block *meta.Block) {
	for index := range block.TXs {
		m.txPool.RemoveTransaction(*block.TXs[index].GetTxID())
	}
}

func IsBestBlockOffspring(chain chain.ChainReader, block *meta.Block) bool {
	return block.GetPrevBlockID().IsEqual(chain.GetBestBlock().GetBlockID())
}
