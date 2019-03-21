package contract

import (
	"bytes"
	"errors"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/core/meta"
)

func (v *Interpreter) VerifyBlockState(block *meta.Block, root math.Hash, actualReward *meta.Amount, fee *meta.Amount, headerData []byte) error {
	log.Debug("VerifyBlockState", "actualReward", actualReward.GetInt64(), "fee", fee.GetInt64())
	//Check block reward
	if actualReward.Subtraction(*meta.NewAmount(config.DefaultBlockReward)).GetInt64() != fee.GetInt64() && len(block.TXs) > 0 {
		return errors.New("coin base tx reward is error")
	}

	log.Debug("VerifyBlockState", "excute Status", root.String(), "header status", block.Header.Status.String())
	if !root.IsEqual(&block.Header.Status) {
		return errors.New("calculate status root must be equal to block status root")
	}

	if bytes.Compare(block.Header.Data, headerData) != 0 {
		return errors.New("header Data is error")
	}

	return nil
}
