package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/linkchain/client/httpclient"
	"github.com/linkchain/rpc/rpcjson"
)

var RootCmd = &cobra.Command{}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop",
	Long:  "This is stop command for stopping linkchain server",
	Run: func(cmd *cobra.Command, args []string) {
		method := "shutdown"
		//call
		out, err := rpc(method, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(out)
		os.Exit(0)
	},
}

func init() {
	RootCmd.AddCommand(stopCmd)
}

var httpConfig = &httpclient.Config{
	RPCUser:     "lc",
	RPCPassword: "lc",
	RPCServer:   "localhost:8082",
}

//rpc call
func rpc(method string, cmd interface{}) (string, error) {
	//param
	s, _ := rpcjson.MarshalCmd(1, method, cmd)
	//log.Info(method, "req", string(s))

	//response
	rawRet, err := httpclient.SendPostRequest(s, httpConfig)
	if err != nil {
		//log.Error(method, "error", err)
		return "", err
	}

	//log.Info(method, "rsp", string(rawRet))

	var out bytes.Buffer
	json.Indent(&out, rawRet, "", "  ")

	return out.String(), nil
}
