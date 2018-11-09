package util

import (
	"encoding/hex"
	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/amount"
	"github.com/linkchain/meta/block"
	"github.com/linkchain/poa/config"
	poameta "github.com/linkchain/poa/meta"
	"time"
)

var fristPrivMiner, _ = hex.DecodeString("55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b")

func GetGensisBlock() block.IBlock {
	txs := []poameta.Transaction{}

	header := poameta.NewBlockHeader(config.DefaultBlockVersion, 0, time.Unix(1487780010, 0), config.DefaultNounce, config.DefaultDifficulty, math.Hash{}, math.Hash{}, math.Hash{}, poameta.Signature{Code: make([]byte, 0)}, nil)
	b := poameta.NewBlock(*header, txs)
	id, _ := CreateAccountIdByPrivKey(hex.EncodeToString(fristPrivMiner))
	coinbase := CreateCoinBaseTx(id, amount.NewAmount(50))
	b.SetTx(coinbase)
	root := b.CalculateTxTreeRoot()
	b.Header.SetMerkleRoot(root)

	SignGensisBlock(b)
	return b
}

func SignGensisBlock(block block.IBlock) error {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), fristPrivMiner)
	log.Info("SignGensisBlock", "block hash", block.GetBlockID().String())
	signature, err := priv.Sign(block.GetBlockID().CloneBytes())
	if err != nil {
		log.Error("SignGensisBlock", "Sign", err)
		return nil
	}
	sign := poameta.NewSignatrue(signature.Serialize())
	block.SetSign(sign)
	return nil
}

func CreateBlock(prevHeight uint32, prevHash meta.BlockID) (block.IBlock, error) {
	txs := []poameta.Transaction{}
	header := poameta.NewBlockHeader(config.DefaultBlockVersion, prevHeight+1, time.Now(), config.DefaultNounce, config.DefaultDifficulty, prevHash, math.Hash{}, math.Hash{}, poameta.Signature{}, nil)
	b := poameta.NewBlock(*header, txs)
	return RebuildBlock(b)

}

func RebuildBlock(block block.IBlock) (block.IBlock, error) {
	pb := *block.(*poameta.Block)
	root := pb.CalculateTxTreeRoot()
	pb.Header.SetMerkleRoot(root)
	return &pb, nil
}
