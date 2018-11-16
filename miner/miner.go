package miner

import (
	"encoding/hex"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/wallet"
	"sync"
)

var fristPrivMiner, _ = hex.DecodeString("55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b")
var secondPrivMiner, _ = hex.DecodeString("7a9c6f2b865c98c9fe174869de5818f4c62bc845441c08269487cdba6688f6b1")
var thirdPrivMiner, _ = hex.DecodeString("6647e717248720f1b50f3f1f765b731783205f2de2fedc9e447438966af7df85")

type Miner struct {
	signers  []wallet.WAccount
	isMining bool
	minerMtx sync.Mutex
}

func (w *Miner) Setup(i interface{}) bool {
	log.Info("Miner init...")
	w.signers = make([]wallet.WAccount, 0)
	w.isMining = false
	return true
}

//func (w *Miner) Start() bool {
//	log.Info("Miner start...")
//	fristWA := wallet.CreateWAccountFromBytes(fristPrivMiner, meta.NewAmount(0))
//	secondWA := wallet.CreateWAccountFromBytes(secondPrivMiner, meta.NewAmount(0))
//	thirdWA := wallet.CreateWAccountFromBytes(thirdPrivMiner, meta.NewAmount(0))
//	w.signers = append(w.signers, fristWA)
//	w.signers = append(w.signers, secondWA)
//	w.signers = append(w.signers, thirdWA)
//	return true
//}
//
//func (w *Miner) Stop() {
//	log.Info("Miner stop...")
//
//}
//
//func (w *Miner) MineBlock() {
//	best := node.getBestBlock()
//	block, err := node.CreateBlock(best.GetHeight(), *best.GetBlockID())
//	if err != nil {
//		log.Error("Miner", "New Block error", err)
//		return
//	}
//	mineIndex := block.GetHeight() % 3
//
//	id := w.signers[mineIndex].GetAccountID()
//	coinbase := node.CreateCoinBaseTx(id, meta.NewAmount(50))
//	block.SetTx(*coinbase)
//
//	txs := node.getAllTransaction()
//	block.SetTx(txs...)
//
//	w.SignBlock(w.signers[mineIndex], block)
//
//	block, err = node.RebuildBlock(block)
//	if err != nil {
//		log.Error("Miner", "Rebuild Block error", err)
//		return
//	}
//	node.processBlock(block)
//	//node.GetManager().NewBlockEvent.Post(meta_block.NewMinedBlockEvent{Block: block})
//}
//
//func (w *Miner) SignBlock(signer wallet.WAccount, block *meta.Block) {
//	sign := signer.Sign(block.GetBlockID().CloneBytes())
//	block.SetSign(sign)
//}
//
//func (w *Miner) StartMine() {
//	w.minerMtx.Lock()
//	w.isMining = true
//	w.minerMtx.Unlock()
//	for true {
//		w.minerMtx.Lock()
//		tempMing := w.isMining
//		w.minerMtx.Unlock()
//		if !tempMing {
//			break
//		}
//		w.MineBlock()
//		time.Sleep(time.Duration(config.DefaultPeriod) * time.Second)
//	}
//}
//
//func (w *Miner) StopMine() {
//	w.minerMtx.Lock()
//	defer w.minerMtx.Unlock()
//	w.isMining = false
//}
//
//func (w *Miner) GetInfo() {
//	w.minerMtx.Lock()
//	defer w.minerMtx.Unlock()
//	log.Info("Miner", "isMing", w.isMining)
//}
