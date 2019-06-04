package normal

import (
	"errors"

	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/interpreter"
	"github.com/mihongtech/linkchain/node/consensus"
)

func (n *Interpreter) ValidateBlockHeader(engine consensus.Engine, chain core.Chain, block *meta.Block) error {
	prevBlock, err := chain.GetBlockByID(*block.GetPrevBlockID())

	if err != nil {
		log.Error("BlockManage", "checkBlock", err)
		return err
	}

	if prevBlock.GetHeight()+1 != block.GetHeight() {
		log.Error("BlockManage", "checkBlock", "current block height is error")
		return errors.New("Check block height failed")
	}
	return err
}

func (n *Interpreter) ValidateBlockBody(txValidator interpreter.TransactionValidator, chain core.Chain, block *meta.Block) error {
	croot := block.CalculateTxTreeRoot()
	if !block.GetMerkleRoot().IsEqual(&croot) {
		log.Error("POA checkBlock", "check merkle root", false)
		return errors.New("Check block merkle root failed")
	}

	//check TXs only one coinBase
	txs := block.GetTxs()
	for i := range txs {
		if i != 0 && txs[i].Type == config.CoinBaseTx {
			return errors.New("the block must be only one coinBase tx")
		} else if i == 0 && txs[i].Type != config.CoinBaseTx {
			return errors.New("the first tx of block must be coinBase tx")
		}
	}

	//check txs have the same tx
	txCount := len(txs)
	for i := 0; i < txCount; i++ {
		for j := i + 1; j < txCount; j++ {
			if txs[i].GetTxID().IsEqual(txs[j].GetTxID()) {
				return errors.New("the block have two same tx")
			}
		}
	}

	//check tx body
	for i := range txs {
		if err := txValidator.CheckTx(&txs[i]); err != nil {
			return err
		}
	}
	return nil
}

func (n *Interpreter) VerifyBlockState(block *meta.Block, root math.Hash, actualReward *meta.Amount, fee *meta.Amount, headerData []byte) error {
	log.Debug("VerifyBlockState", "actualReward", actualReward.GetInt64(), "fee", fee.GetInt64())
	//Check block reward
	if actualReward.Subtraction(*meta.NewAmount(config.DefaultBlockReward)).GetInt64() != fee.GetInt64() && len(block.TXs) > 0 {
		return errors.New("coin base tx reward is error")
	}

	log.Debug("VerifyBlockState", "excute Status", root.String(), "header status", block.Header.Status.String())
	if !root.IsEqual(&block.Header.Status) {
		return errors.New("calculate status root must be equal to block status root")
	}

	return nil
}
