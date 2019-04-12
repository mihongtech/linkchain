package bind

import (
	"context"
	"fmt"
	"time"

	"github.com/mihongtech/linkchain/common/util/log"
	_ "github.com/mihongtech/linkchain/contract"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
)

// WaitMined waits for tx to be mined on the blockchain.
// It stops waiting when the context is canceled.
func WaitMined(ctx context.Context, b DeployBackend, tx *meta.Transaction) (*core.Receipt, error) {
	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()

	logger := log.New("hash", *tx.GetTxID())
	for {
		receipt, err := b.TransactionReceipt(ctx, *tx.GetTxID())
		if receipt != nil {
			return receipt, nil
		}
		if err != nil {
			logger.Trace("Receipt retrieval failed", "err", err)
		} else {
			logger.Trace("Transaction not yet mined")
		}
		// Wait for the next round.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
		}
	}
}

// WaitDeployed waits for a contract deployment transaction and returns the on-chain
// contract address when it is mined. It stops waiting when ctx is canceled.
func WaitDeployed(ctx context.Context, b DeployBackend, tx *meta.Transaction) (meta.AccountID, error) {
	//	if tx.GetToCoins() != nil {
	//		return meta.AccountID{}, fmt.Errorf("tx is not contract creation")
	//	}
	if len(tx.GetToCoins()) < 1 {
		log.Error("tx error", "tx", tx.GetToCoins())
		return meta.AccountID{}, fmt.Errorf("tx is not contract creation")
	}
	if tx.GetToCoins()[0].GetId() != (meta.AccountID{}) {
		log.Error("tx error", "tx", tx)
		return meta.AccountID{}, fmt.Errorf("tx contract equal zero")
	}

	receipt, err := WaitMined(ctx, b, tx)
	if err != nil {
		return meta.AccountID{}, err
	}
	if receipt.ContractAddress == (meta.AccountID{}) {
		log.Error("receipt error", "receipt", receipt)
		return meta.AccountID{}, fmt.Errorf("zero address")
	}
	// Check that code has indeed been deployed at the address.
	// This matters on pre-Homestead chains: OOG in the constructor
	// could leave an empty account behind.
	code, err := b.CodeAt(ctx, receipt.ContractAddress, nil)
	if err == nil && len(code) == 0 {
		err = ErrNoCodeAfterDeploy
	}
	return receipt.ContractAddress, err
}
