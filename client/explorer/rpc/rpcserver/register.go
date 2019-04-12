package rpcserver

import (
	"reflect"

	"github.com/mihongtech/linkchain/client/explorer/rpc/rpcjson"
	"github.com/mihongtech/linkchain/client/explorer/rpc/rpcobject"
	"github.com/mihongtech/linkchain/client/httpclient"
)

type commandHandler func(*Server, interface{}, <-chan struct{}) (interface{}, error)

//handler pool
var handlerPool = map[string]commandHandler{
	"getBlockChainInfo": getBlockChainInfo,

	// peer
	"addPeer":    addPeer,
	"listPeer":   listPeer,
	"selfPeer":   selfPeer,
	"removePeer": removePeer,

	"getBestBlock":     getBestBlock,
	"getBlockByHeight": getBlockByHeight,
	"getBlockByHash":   getBlockByHash,

	//miner
	"getMineInfo": getMineInfo,
	"startMine":   startMine,
	"stopMine":    stopMine,
	"mine":        mine,

	//wallet
	"exportAccount": exportAccount,
	"importAccount": importAccount,

	"getWalletInfo":  getWalletInfo,
	"getAccountInfo": getAccountInfo,
	"newAcount":      newAcount,

	"sendMoneyTransaction": sendMoneyTransaction,

	//transaction
	"getTxByHash": getTxByHash,

	//shutdown
	//"shutdown": shutdown,

	//contract
	"publishContract":    publishContract,
	"callContract":       callContract,
	"call":               call,
	"transactionReceipt": getTransactionReceipt,
}

var cmdPool = map[string]reflect.Type{
	"version":    reflect.TypeOf((*rpcobject.VersionCmd)(nil)),
	"addPeer":    reflect.TypeOf((*rpcobject.PeerCmd)(nil)),
	"removePeer": reflect.TypeOf((*rpcobject.PeerCmd)(nil)),

	"getBlockByHeight": reflect.TypeOf((*rpcobject.GetBlockByHeightCmd)(nil)),
	"getBlockByHash":   reflect.TypeOf((*rpcobject.GetBlockByHashCmd)(nil)),

	"getAccountInfo": reflect.TypeOf((*rpcobject.SingleCmd)(nil)),

	"sendMoneyTransaction": reflect.TypeOf((*rpcobject.SendToTxCmd)(nil)),

	"getTxByHash": reflect.TypeOf((*rpcobject.GetTransactionByHashCmd)(nil)),

	"importAccount": reflect.TypeOf((*rpcobject.ImportAccountCmd)(nil)),
	"exportAccount": reflect.TypeOf((*rpcobject.ExportAccountCmd)(nil)),

	//contract
	"publishContract":    reflect.TypeOf((*rpcobject.PublishContractCmd)(nil)),
	"callContract":       reflect.TypeOf((*rpcobject.CallContractCmd)(nil)),
	"call":               reflect.TypeOf((*rpcobject.CallCmd)(nil)),
	"getCode":            reflect.TypeOf((*rpcobject.GetCodeCmd)(nil)),
	"transactionReceipt": reflect.TypeOf((*rpcobject.GetTransactionReceiptCmd)(nil)),
}

var httpConfig = &httpclient.Config{
	RPCUser:     "lc",
	RPCPassword: "lc",
	RPCServer:   "localhost:8082",
}

//rpc call
func rpc(method string, cmd interface{}) ([]byte, error) {
	//param
	s, _ := rpcjson.MarshalCmd(1, method, cmd)
	//log.Info(method, "req", string(s))

	//response
	rawRet, err := httpclient.SendPostRequest(s, httpConfig)
	if err != nil {
		//log.Error(method, "error", err)
		return nil, err
	}

	return rawRet, nil
}
