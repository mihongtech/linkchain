package poamanager

import (
	"time"

	"github.com/linkchain/meta/tx"
	"github.com/linkchain/common/util/log"
	poameta "github.com/linkchain/poa/meta"
	"github.com/linkchain/meta/account"
	"github.com/linkchain/meta"
	"errors"
)

type POATxManager struct {
	txpool []poameta.POATransaction
}

/** interface: common.IService **/
func (m *POATxManager) Init(i interface{}) bool{
	log.Info("POABlockManager init...");
	m.txpool = make([]poameta.POATransaction,0)

	return true
}

func (m *POATxManager) Start() bool{
	log.Info("POABlockManager start...");
	return true
}

func (m *POATxManager) Stop(){
	log.Info("POABlockManager stop...");
}

func (m *POATxManager) AddTransaction(tx tx.ITx) error{
	newTx := *tx.(*poameta.POATransaction)
	m.txpool = append(m.txpool,newTx)
	return nil
}


func (m *POATxManager) GetAllTransaction() []tx.ITx{
	txs := make([]tx.ITx,0)
	for _,tx := range m.txpool {
		txs = append(txs,&tx)
	}
	return txs
}


func (m *POATxManager) RemoveTransaction(txid meta.DataID) error{
	deleteIndex := make([]int,0)
	for index,tx := range m.txpool{
		txHash := tx.GetTxID()
		if txHash.IsEqual(txid){
			deleteIndex = append(deleteIndex,index)
		}
	}
	for _,index := range deleteIndex{
		m.txpool = append(m.txpool[:index],m.txpool[index+1:]...)
	}
	return nil
}

func (m *POATxManager) NewTransaction(from account.IAccount,to account.IAccount,amount meta.IAmount) tx.ITx {
	newTx := poameta.POATransaction{Version:0,
		From:poameta.GetPOATransactionPeer(from,nil),
		To:poameta.GetPOATransactionPeer(to,nil),
		Amount:*amount.(*poameta.POAAmount),
		Time:time.Now(),
		Nounce:(from.GetNounce()+1),}
	return &newTx
}

func (m *POATxManager) CheckTx(tx tx.ITx) bool {
	log.Info("POA CheckTx ...")
	err := tx.Verify()
	if err != nil {
		log.Error("POA CheckTx","failed",err)
		return false
	}

	err = GetManager().AccountManager.CheckTxFromAccount(tx)

	if err != nil {
		log.Error("POA CheckTx","failed",err)
		return false
	}

	err = GetManager().AccountManager.CheckTxFromNounce(tx)

	if err != nil {
		log.Error("POA CheckTx","failed",err)
		return false
	}

	return true
}

func (m *POATxManager) ProcessTx(tx tx.ITx) error{
	log.Info("POA ProcessTx ...")
	//1.checkTx
	if !m.CheckTx(tx) {
		log.Error("POA checkTransaction failed")
		return errors.New("POA checkTransaction failed")
	}
	//2.push Tx into storage
	m.AddTransaction(tx)
	log.Info("POA Add Tranasaction Pool  ...")
	return nil
}

func (m *POATxManager) SignTransaction(tx tx.ITx) error  {
	return nil
}