package cmd

import (
	"github.com/linkchain/app"
	"github.com/linkchain/common/math"
	"github.com/linkchain/common/util/log"
	"github.com/spf13/cobra"
	"strconv"
)

func init() {
	RootCmd.AddCommand(chainInfoCmd)
	RootCmd.AddCommand(blockCmd)
	blockCmd.AddCommand(heightCmd, hashCmd)
}

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "block command",
}

var heightCmd = &cobra.Command{
	Use:   "height",
	Short: "get a block by height in chainmanage",
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

		if uint32(height) > app.GetNodeAPI().GetBestBlock().GetHeight() || height < 0 {
			log.Error("getblockbyheight ", "error", "height is out of range", example[0], example[1])
			return
		}
		block, err := app.GetNodeAPI().GetBlockByHeight(uint32(height))
		if err != nil {
			log.Error("getblockbyheight ", "error", err)
		} else {
			log.Info("block", "data", block.String())
		}
	},
}

var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "get a block by hash in chainmanage",
	Run: func(cmd *cobra.Command, args []string) {
		example := []string{"example", "block hash 98acd27a58c79eaab05ea4abd0daa8e63021df3bf2e65fcb38e2474fb706c3fe"}
		if len(args) != 1 {
			log.Error("getblockbyhash", "error", "please input hash", example[0], example[1])
			return
		}

		hash, err := math.NewHashFromStr(args[0])
		if err != nil {
			log.Error("getblockbyhash", "error", err, example[0], example[1])
		}
		block, err := app.GetNodeAPI().GetBlockByID(*hash)
		if err != nil {
			log.Error("getblockbyhash ", "error", err)
		} else {
			log.Info("block", "data", block)
		}
	},
}
