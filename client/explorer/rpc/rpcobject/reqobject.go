package rpcobject

//concrete command
type InnerCmd struct {
	Ia int    `json:"ia"`
	Ib string `json:"ib"`
}

type VersionCmd struct {
	A int
	B string `json:""`
	C InnerCmd
}

//peer
type PeerCmd struct {
	Peer string `json:"peer"`
}

//block
type GetBlockByHeightCmd struct {
	Height int `json:"height"`
}

type GetBlockByHashCmd struct {
	Hash string `json:"hash"`
}

type SingleCmd struct {
	Key string `json:"key"`
}

//Transactions
type SendToTxCmd struct {
	FromAccountId string `json:"fromAccountId"`
	ToAccountId   string `json:"toAccountId"`
	Amount        int    `json:"amount"`
}

type GetTransactionByHashCmd struct {
	Hash string `json:"hash"`
}

type PublishContractCmd struct {
	FromAccountId string `json:"fromAccountId"`
	Contract      string `json:"contract"`
	Amount        int64  `json:"amount"`
	GasPrice      int64  `json:"gasPrice"`
	GasLimit      uint64 `json:"gasLimit"`
}

type CallContractCmd struct {
	FromAccountId string `json:"fromAccountId"`
	Contract      string `json:"contract"`
	CallMethod    string `json:"callMethod"`
	Amount        int64  `json:"amount"`
	GasPrice      int64  `json:"gasPrice"`
	GasLimit      uint64 `json:"gasLimit"`
}

type GetCodeCmd struct {
	FromAccountId string `json:"fromAccountId"`
	Height        int64  `json:"height"`
}

type CallCmd struct {
	FromAccountId string `json:"fromAccountId"`
	Contract      string `json:"contract"`
	Data          string `json:"data"`
	Height        int64  `json:"height"`
}

type GetTransactionReceiptCmd struct {
	Hash string `json:"hash"`
}

//Wallet
type ImportAccountCmd struct {
	Signer string `json:"accountPrivateKey"`
}

type ExportAccountCmd struct {
	AccountId string `json:"insuranceID"`
}
