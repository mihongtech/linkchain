package rpcserver

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/helper"
	"github.com/linkchain/node"
	"github.com/linkchain/rpc/rpcobject"
)

func getWalletInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	accounts := GetWalletAPI(s).GetAllWAccount()
	accountsArray := make([]*rpcobject.WalletAccountRSP, 0, len(accounts))

	//inflate all the accounts
	for _, a := range accounts {
		//log.Info("wallet info", "accountId", wa.GetAccountID().String(), "type", wa.GetAccountType(), "value", wa.GetAmount())
		accountsArray = append(accountsArray, &rpcobject.WalletAccountRSP{a.GetAccountID().String(), a.AccountType, a.GetAmount().GetInt64()})
	}

	return rpcobject.WalletInfoRSP{accountsArray}, nil
}

func getAccountInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.SingleCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}

	//get account id
	accountId, err := helper.CreateAccountIdByAddress(c.Key)
	if err != nil {
		return nil, err
	}

	//get account detail
	account, err := GetWalletAPI(s).GetAccount(accountId.String())
	if err != nil {
		return nil, err
	}

	//populate account info
	best := GetNodeAPI(s).GetBestBlock()
	utxos := account.UTXOs
	txArray := make([]*rpcobject.TxRSP, 0, len(utxos))

	//get all utxo
	for _, u := range utxos {
		log.Info("utxo", "TxId", u.Txid.String(), "index", u.Index, "value", u.Value.GetInt64(), "locatedHeight", u.LocatedHeight, "effectHeight", u.EffectHeight)
		txArray = append(txArray, &rpcobject.TxRSP{
			u.Txid.String(),
			u.Index,
			u.Value.GetInt64(),
			u.LocatedHeight,
			u.EffectHeight})
	}
	code, err := GetNodeAPI(s).GetCode(*account.GetAccountID())
	if err != nil {
		code = []byte(err.Error())
	}

	log.Info("rpc wallet", "code", hex.EncodeToString(code), "code hash", account.CodeHash.String())
	return &rpcobject.AccountRSP{
		account.GetAccountID().String(),
		account.AccountType,
		account.GetAmount().GetInt64(),
		account.SecurityId.String(),
		account.GetClearTime(best.GetHeight()),
		account.Clear,
		txArray,
		account.StorageRoot.String(),
		account.CodeHash.String(),
		hex.EncodeToString(code),
	}, nil
}

func newAcount(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	a, err := GetWalletAPI(s).NewAccount()
	if err != nil {
		return nil, nil
	}
	return a.String(), nil
}

func sendMoneyTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.SendToTxCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}
	bestHeight := GetNodeAPI(s).GetBestBlock().GetHeight()
	amount := meta.NewAmount(int64(c.Amount))
	toID, err := helper.CreateAccountIdByAddress(c.ToAccountId)
	toCoin := helper.CreateToCoin(*toID, amount)

	from, err := GetWalletAPI(s).GetAccount(c.FromAccountId)
	if err != nil {
		return nil, err
	}
	fromCoin, fromAmount, err := from.MakeFromCoin(amount, bestHeight)
	if err != nil {
		return nil, err
	}

	transaction := helper.CreateTransaction(*fromCoin, *toCoin)
	backChange := helper.CreateToCoin(from.Id, fromAmount.Subtraction(*amount))
	if backChange.Value.GetInt64() > 0 {
		transaction.AddToCoin(*backChange)
	}

	transaction, err = GetWalletAPI(s).SignTransaction(*transaction)
	if err != nil {
		return nil, err
	}

	if err = GetNodeAPI(s).ProcessTx(transaction); err == nil {
		GetNodeAPI(s).GetTxEvent().Send(node.TxEvent{transaction})
	}

	return &rpcobject.TransactionWithIDRSP{transaction.GetTxID().GetString(), transaction}, err
}

func importAccount(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.ImportAccountCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}
	accountId, err := GetWalletAPI(s).ImportAccount(c.Signer)

	if err != nil {
		log.Error("importSigner ", "error", err)
		return nil, err
	}
	return accountId.String(), err
}

func exportAccount(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c, ok := cmd.(*rpcobject.ExportAccountCmd)
	if !ok {
		fmt.Println("Type error:", reflect.TypeOf(cmd))
		return nil, nil
	}
	accountId, err := meta.NewAccountIdFromStr(c.AccountId)
	if err != nil {
		return "", err
	}
	privateKey, err := GetWalletAPI(s).ExportAccount(*accountId)

	return privateKey, err
}
