package cmd

import (
	"github.com/linkchain/common/util/log"
	meta_block "github.com/linkchain/meta/block"
	"github.com/linkchain/poa/manage"
	"github.com/spf13/cobra"
	"strconv"
)

func init() {
	RootCmd.AddCommand(mineCmd)
	RootCmd.AddCommand(chainInfoCmd)
	RootCmd.AddCommand(blockCmd)
	blockCmd.AddCommand(heightCmd)
	RootCmd.AddCommand(minetestCmd)
}

var mineCmd = &cobra.Command{
	Use:   "mine",
	Short: "generate a new block",
	Run: func(cmd *cobra.Command, args []string) {
		block, err := manage.GetManager().BlockManager.NewBlock()
		if err != nil {
			log.Error("mine", "New Block error", err)
			return
		}
		txs := manage.GetManager().TransactionManager.GetAllTransaction()
		block.SetTx(txs)

		block, err = manage.GetManager().BlockManager.RebuildBlock(block)
		if err != nil {
			log.Error("mine", "Rebuild Block error", err)
			return
		}
		manage.GetManager().BlockManager.ProcessBlock(block)
		manage.GetManager().NewBlockEvent.Post(meta_block.NewMinedBlockEvent{Block: block})
	},
}

var minetestCmd = &cobra.Command{
	Use:   "test",
	Short: "generate a new block",
	Run: func(cmd *cobra.Command, args []string) {
		block, err := manage.GetManager().BlockManager.NewBlock()
		if err != nil {
			log.Error("mine", "New Block error", err)
			return
		}
		txs := manage.GetManager().TransactionManager.GetAllTransaction()
		block.SetTx(txs)

		block, err = manage.GetManager().BlockManager.RebuildTestBlock(block)
		if err != nil {
			log.Error("mine", "Rebuild Block error", err)
			return
		}
		manage.GetManager().BlockManager.ProcessBlock(block)
		manage.GetManager().NewBlockEvent.Post(meta_block.NewMinedBlockEvent{Block: block})
	},
}

var chainInfoCmd = &cobra.Command{
	Use:   "chaininfo",
	Short: "getBlockChainInfo",
	Run: func(cmd *cobra.Command, args []string) {
		manage.GetManager().ChainManager.GetBlockChainInfo()
	},
}

var loadChainCmd = &cobra.Command{
	Use:   "loadchain",
	Short: "loadchain",
	Run: func(cmd *cobra.Command, args []string) {
		manage.GetManager().ChainManager.UpdateChain()
	},
}

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "block command",
}

var heightCmd = &cobra.Command{
	Use:   "height",
	Short: "get a block by height",
	Run: func(cmd *cobra.Command, args []string) {
		example := []string{"example", "block height 0"}
		if len(args) != 1 {
			log.Error("getblockbyheight", "error", "please input height", example[0], example[1])
			return
		}

		height, error := strconv.Atoi(args[0])
		if error != nil {
			log.Error("getblockbyheight ", "error", error, example[0], example[1])
			return
		}

		if uint32(height) > manage.GetManager().ChainManager.GetBestBlock().GetHeight() || height < 0 {
			log.Error("getblockbyheight ", "error", "height is out of range", example[0], example[1])
			return
		}
		block, err := manage.GetManager().ChainManager.GetBlockByHeight(uint32(height))
		if err != nil {
			log.Error("getblockbyheight ", "error", err)
		} else {
			log.Info("block", "data", block)
		}

	},
}
