package cmd

import (
	"fmt"
	"github.com/mihongtech/linkchain/rpc/rpcobject"
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
	Short: "peer command",
	Long:  "This is all network command for handling network",
}

var addPeerCmd = &cobra.Command{
	Use:     "add",
	Short:   "peer add <node>",
	Long:    "This is add peer node command",
	Example: "peer add enode://0b049941925387066ddf8c719a1b9126e5c4a6147112cae163ee2ca24d231f33141e7493b4bf63f3eebbec19f4bcc7d49f4ab94b9721641afb95047f045b902c@[::]:40000",
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
	Use:     "list",
	Short:   "peer list",
	Long:    "This is get connected nodes command",
	Example: "peer list",
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
	Use:     "remove",
	Short:   "peer remove <node>",
	Long:    "This is remove peer node command",
	Example: "peer remove enode://0b049941925387066ddf8c719a1b9126e5c4a6147112cae163ee2ca24d231f33141e7493b4bf63f3eebbec19f4bcc7d49f4ab94b9721641afb95047f045b902c@[::]:40000",
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
	Use:     "self",
	Short:   "peer self",
	Long:    "This is get self node command",
	Example: "peer self",
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
