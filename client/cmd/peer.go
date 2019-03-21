package cmd

import (
	"fmt"
	"github.com/linkchain/rpc/rpcobject"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(peerCmd)
	peerCmd.AddCommand(
		addPeerCmd,
		listPeerCmd,
		removePeerCmd,
		selfPeerCmd)
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

		peer := args[0]
		method := "addPeer"

		//call
		out, err := rpc(method, &rpcobject.PeerCmd{peer})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

var listPeerCmd = &cobra.Command{
	Use:   "list",
	Short: "list all peers",
	Run: func(cmd *cobra.Command, args []string) {
		method := "listPeer"

		//call
		out, err := rpc(method, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
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
		peer := args[0]
		method := "removePeer"

		//call
		out, err := rpc(method, &rpcobject.PeerCmd{peer})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}

var selfPeerCmd = &cobra.Command{
	Use:   "self",
	Short: "print self peer node info",
	Run: func(cmd *cobra.Command, args []string) {
		method := "selfPeer"

		//call
		out, err := rpc(method, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
	},
}
