package helper

import (
	"encoding/hex"
	"sort"
	"time"

	"github.com/linkchain/common"
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/math"
	"github.com/linkchain/config"
	"github.com/linkchain/core/meta"
)

/*
	Account
*/

func CreateAccountIdByAddress(addr string) (*meta.AccountID, error) {
	buffer, err := hex.DecodeString(addr)
	if err != nil {
		return nil, err
	}

	id := meta.BytesToAccountID(buffer)
	return &id, nil
}

func CreateAccountIdByPubKey(pubKey string) (*meta.AccountID, error) {
	pkBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	pk, err := btcec.ParsePubKey(pkBytes, btcec.S256())
	if err != nil {
		return nil, err
	}
	return meta.NewAccountId(pk), nil
}

func CreateAccountIdByPrivKey(privKey string) (*meta.AccountID, error) {
	priv, err := hex.DecodeString(privKey)
	if err != nil {
		return nil, err
	}
	_, pk := btcec.PrivKeyFromBytes(btcec.S256(), priv)
	if err != nil {
		return nil, err
	}
	return meta.NewAccountId(pk), nil
}

func CreateTemplateAccount(id meta.AccountID) *meta.Account {
	u := make([]meta.UTXO, 0)
	c := meta.NewClearTime(0, 0)
	a := meta.NewAccount(id, config.NormalAccount, u, c, meta.AccountID{})
	return a
}

func CreateNormalAccount(key *btcec.PrivateKey) (*meta.Account, error) {
	privateStr := hex.EncodeToString(key.Serialize())
	id, err := CreateAccountIdByPrivKey(privateStr)
	if err != nil {
		return nil, err
	}

	a := CreateTemplateAccount(*id)
	return a, nil
}

/*

	Transaction
*/

func CreateToCoin(to meta.AccountID, amount *meta.Amount) *meta.ToCoin {
	return meta.NewToCoin(to, amount)
}

func CreateFromCoin(from meta.AccountID, ticket ...meta.Ticket) *meta.FromCoin {
	tickets := make([]meta.Ticket, 0)
	fc := meta.NewFromCoin(from, tickets)
	for _, c := range ticket {
		fc.AddTicket(&c)
	}
	return fc
}

func CreateTempleteTx(version uint32, txtype uint32) *meta.Transaction {
	return meta.NewEmptyTransaction(version, txtype)
}

func CreateTransaction(fromCoin meta.FromCoin, toCoin meta.ToCoin) *meta.Transaction {
	transaction := CreateTempleteTx(config.DefaultTransactionVersion, config.NormalTx)
	transaction.AddFromCoin(fromCoin)
	transaction.AddToCoin(toCoin)
	return transaction
}

func CreateCoinBaseTx(to meta.AccountID, amount *meta.Amount, height uint32) *meta.Transaction {
	toCoin := meta.NewToCoin(to, amount)
	transaction := meta.NewEmptyTransaction(config.DefaultDifficulty, config.CoinBaseTx)
	transaction.AddToCoin(*toCoin)
	transaction.Data = common.UInt32ToBytes(height)
	return transaction
}

func SortTransaction(tx *meta.Transaction) {
	//sort from
	sort.Slice(tx.From.Coins, func(i, j int) bool {
		if tx.From.Coins[i].Id.Big().Cmp(tx.From.Coins[j].Id.Big()) == 0 {
			return false
		} else if tx.From.Coins[i].Id.Big().Cmp(tx.From.Coins[j].Id.Big()) < 0 {
			return true
		} else {
			return false
		}
	})

	//sort from ticket
	for k := range tx.From.Coins {
		sort.Slice(tx.From.Coins[k].Ticket, func(i, j int) bool {
			if tx.From.Coins[k].Ticket[i].Txid.Big().Cmp(tx.From.Coins[k].Ticket[j].Txid.Big()) == 0 {
				return tx.From.Coins[k].Ticket[i].Index > tx.From.Coins[k].Ticket[j].Index
			} else if tx.From.Coins[k].Ticket[i].Txid.Big().Cmp(tx.From.Coins[k].Ticket[j].Txid.Big()) < 0 {
				return true
			} else {
				return false
			}
		})
	}

	//sort to
	sort.Slice(tx.To.Coins, func(i, j int) bool {
		if tx.To.Coins[i].Id.Big().Cmp(tx.To.Coins[j].Id.Big()) == 0 {
			return tx.To.Coins[i].Value.GetInt64() < tx.To.Coins[j].Value.GetInt64()
		} else if tx.To.Coins[i].Id.Big().Cmp(tx.To.Coins[j].Id.Big()) < 0 {
			return true
		} else {
			return false
		}
	})
}

/*

	Block
*/
func CreateBlock(prevHeight uint32, prevHash meta.BlockID) (*meta.Block, error) {
	var txs []meta.Transaction
	header := meta.NewBlockHeader(config.DefaultBlockVersion, prevHeight+1, time.Now(),
		config.DefaultNounce, config.DefaultDifficulty, prevHash,
		math.Hash{}, math.Hash{}, meta.Signature{}, nil)
	b := meta.NewBlock(*header, txs)
	return RebuildBlock(b)

}

func RebuildBlock(block *meta.Block) (*meta.Block, error) {
	pb := block
	root := pb.CalculateTxTreeRoot()
	pb.Header.SetMerkleRoot(root)
	return pb, nil
}
