package cmd

import (
	"github.com/linkchain/app"
	"github.com/linkchain/common/util/log"
	p2p_node "github.com/linkchain/p2p/node"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(peerCmd)
	peerCmd.AddCommand(addPeerCmd, listPeerCmd, removePeerCmd)
}

var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: "all p2p peer related command",
}

var addPeerCmd = &cobra.Command{
	Use:   "add",
	Short: "add a new peer",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			println("Please input peer id")
			return
		}
		println("Add new peer")
		println("args is %s", args[0])
		server := app.GetP2PAPI()
		node, err := p2p_node.ParseNode(args[0])
		if err != nil {
			log.Error("parse node failes", "url", args[0])
			return
		}
		server.AddPeer(node)
	},
}

var listPeerCmd = &cobra.Command{
	Use:   "list",
	Short: "list all peers",
	Run: func(cmd *cobra.Command, args []string) {
		println("List all peers")
		server := app.GetP2PAPI()
		peers := server.Peers()

		if len(peers) == 0 {
			println("Peer count is 0")
			return
		}
		println("Peer count is %d", len(peers))
		for _, peer := range peers {
			println("peer: %s", peer.String())
		}
	},
}

var removePeerCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a new peer",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			println("Please input peer id")
			return
		}
		println("Remove  peer")
		println("args is %s", args[0])
		server := app.GetP2PAPI()
		node, err := p2p_node.ParseNode(args[0])
		if err != nil {
			log.Error("par"+
				""+
				"se node failes", "url", args[0])
			return
		}
		server.RemovePeer(node)
	},
}
