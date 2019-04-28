package rpcserver

import (
	"reflect"

	"github.com/mihongtech/linkchain/miner"
	"github.com/mihongtech/linkchain/node"
	"github.com/mihongtech/linkchain/p2p"
	"github.com/mihongtech/linkchain/rpc/rpcobject"
	"github.com/mihongtech/linkchain/txpool"
	"github.com/mihongtech/linkchain/wallet"
)

type commandHandler func(*Server, interface{}, <-chan struct{}) (interface{}, error)

//handler pool
var handlerPool = map[string]commandHandler{
	"getBlockChainInfo": getBlockChainInfo,

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
	"shutdown": shutdown,

	//contract
	"publishContract":    publishContract,
	"callContract":       callContract,
	"getCode":            getCode,
	"call":               call,
	"transactionReceipt": GetTransactionReceipt,
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

func GetNodeAPI(s *Server) *node.PublicNodeAPI {
	return s.appContext.NodeAPI.(*node.PublicNodeAPI)
}

func GetP2PAPI(s *Server) *p2p.Service {
	return s.appContext.P2PAPI.(*p2p.Service)
}

func GetMinerAPI(s *Server) *miner.Miner {
	return s.appContext.MinerAPI.(*miner.Miner)
}

func GetWalletAPI(s *Server) *wallet.Wallet {
	return s.appContext.WalletAPI.(*wallet.Wallet)
}

func GetTxpoolAPI(s *Server) *txpool.TxPool {
	return s.appContext.TxpoolAPI.(*txpool.TxPool)
}
