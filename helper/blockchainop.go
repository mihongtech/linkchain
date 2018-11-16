package helper

import (
	"encoding/hex"
	"time"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/config"
	"github.com/linkchain/core/meta"
)

/*
	Account
*/
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

func CreateTempleteAccount(id meta.AccountID) *meta.Account {
	utxo := make([]meta.UTXO, 0)
	a := meta.NewAccount(id, config.NormalAccount, utxo, config.DafaultClearTime, meta.AccountID{})
	return a
}

func CreateNormalAccount(key *btcec.PrivateKey) (*meta.Account, error) {
	privStr := hex.EncodeToString(key.Serialize())
	id, err := CreateAccountIdByPrivKey(privStr)
	if err != nil {
		return nil, err
	}

	a := CreateTempleteAccount(*id)
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

func CreateCoinBaseTx(to meta.AccountID, amount *meta.Amount) *meta.Transaction {
	toCoin := meta.NewToCoin(to, amount)
	transaction := meta.NewEmptyTransaction(config.DefaultDifficulty, config.CoinBaseTx)
	transaction.AddToCoin(*toCoin)
	return transaction
}

/*
	Block
*/
var fristPrivMiner, _ = hex.DecodeString("55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b")

func GetGensisBlock() *meta.Block {
	txs := []meta.Transaction{}

	header := meta.NewBlockHeader(config.DefaultBlockVersion, 0, time.Unix(1487780010, 0), config.DefaultNounce, config.DefaultDifficulty, math.Hash{}, math.Hash{}, math.Hash{}, meta.Signature{Code: make([]byte, 0)}, nil)
	b := meta.NewBlock(*header, txs)
	id, _ := CreateAccountIdByPrivKey(hex.EncodeToString(fristPrivMiner))
	coinbase := CreateCoinBaseTx(*id, meta.NewAmount(50))
	b.SetTx(*coinbase)
	root := b.CalculateTxTreeRoot()
	b.Header.SetMerkleRoot(root)

	SignGensisBlock(b)
	return b
}

func SignGensisBlock(block *meta.Block) error {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), fristPrivMiner)
	log.Info("signGensisBlock", "block hash", block.GetBlockID().String())
	signature, err := priv.Sign(block.GetBlockID().CloneBytes())
	if err != nil {
		log.Error("signGensisBlock", "Sign", err)
		return nil
	}
	sign := meta.NewSignatrue(signature.Serialize())
	block.SetSign(sign)
	return nil
}

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
