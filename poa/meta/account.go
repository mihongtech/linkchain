package poameta

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/amount"
	"github.com/linkchain/meta/coin"
	"github.com/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
)

type AccountID struct {
	ID []byte
}

func (id *AccountID) String() string {
	return hex.EncodeToString(id.ID)
}

func (id *AccountID) IsEqual(other meta.IAccountID) bool {
	return strings.Compare(id.String(), other.String()) == 0
}

//Serialize/Deserialize
func (a *AccountID) Serialize() serialize.SerializeStream {
	accountId := protobuf.AccountID{
		Id: proto.NewBuffer(a.ID).Bytes(),
	}
	return &accountId
}

func (a *AccountID) Deserialize(s serialize.SerializeStream) error {
	data := s.(*protobuf.AccountID)
	pk, err := btcec.ParsePubKey(data.Id, btcec.S256())
	if err != nil {
		log.Error("Id", "Deserialize failed", err)
		return err
	}
	a.ID = pk.SerializeCompressed()
	return nil
}

//Json Hash convert to Hex
func (a AccountID) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *AccountID) UnmarshalJSON(data []byte) error {
	return nil
}

func NewAccountId(id *btcec.PublicKey) *AccountID {
	return &AccountID{ID: id.SerializeCompressed()}
}

type UTXO struct {
	Ticket
	LocatedHeight uint32
	EffectHeight  uint32
	Value         amount.Amount
}

func NewUTXO(tickets *Ticket, locatedHeight uint32, effectHeight uint32, value amount.Amount) *UTXO {
	return &UTXO{*tickets, locatedHeight, effectHeight, value}
}

func (u *UTXO) String() string {
	data, err := json.Marshal(u)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

type Account struct {
	Id          AccountID
	AccountType uint32
	UTXOs       []UTXO
	ClearTime   int64
	SecurityId  AccountID
}

func NewAccount(id AccountID, accountType uint32, utxos []UTXO, clearTime int64, securityId AccountID) *Account {
	return &Account{Id: id, AccountType: accountType, UTXOs: utxos, ClearTime: clearTime, SecurityId: securityId}
}

func (a *Account) GetAccountID() meta.IAccountID {
	return &a.Id
}

func (a *Account) GetAmount() *amount.Amount {
	sum := amount.NewAmount(0)
	for _, u := range a.UTXOs {
		sum.Addition(u.Value)
	}
	return sum
}

func (a *Account) GetFromCoinValue(fromCoin coin.IFromCoin) (*amount.Amount, error) {
	sum := amount.NewAmount(0)
	if a.GetAccountID().IsEqual(fromCoin.GetId()) {
		tickets := fromCoin.GetTickets()
		for _, t := range tickets {
			u, err := a.getUTXOByTicket(t)
			if err != nil {
				return nil, err
			}
			sum.Addition(u.Value)
		}
		return sum, nil
	} else {
		return nil, errors.New("FromCoin's AccountId is error")
	}
}

func (a *Account) CheckFromCoin(fromCoin coin.IFromCoin) bool {
	if a.GetAccountID().IsEqual(fromCoin.GetId()) {
		tickets := fromCoin.GetTickets()
		for _, t := range tickets {
			if a.Contains(t) {
				return true
			}
		}
	}
	return false
}

func (a *Account) RemoveUTXOByFromCoin(fromCoin coin.IFromCoin) error {
	if a.GetAccountID().IsEqual(fromCoin.GetId()) {
		tickets := fromCoin.GetTickets()
		for _, t := range tickets {
			removeOk := false
			for index, u := range a.UTXOs {
				if t.GetTxid().IsEqual(&u.Txid) && t.GetIndex() == u.GetIndex() {
					a.UTXOs = append(a.UTXOs[:index], a.UTXOs[index+1:]...)
					removeOk = true
					break
				}
			}
			if !removeOk {
				return errors.New("RemoveUTXO():The ticket is not exist in Account")

			}
		}
		return nil
	} else {
		return errors.New("FromCoin's AccountId is error")
	}
}

func (a *Account) getUTXOByTicket(ticket coin.ITicket) (*UTXO, error) {
	for _, t := range a.UTXOs {
		if ticket.GetTxid().IsEqual(&t.Txid) && t.Index == ticket.GetIndex() {
			return &t, nil
		}
	}
	return nil, errors.New("The ticket is not exist in Account")
}

func (a *Account) Contains(ticket coin.ITicket) bool {
	for _, t := range a.UTXOs {
		if ticket.GetTxid().IsEqual(&t.Txid) && t.Index == ticket.GetIndex() {
			return true
		}
	}
	return false
}

//Serialize/Deserialize
func (a *Account) Serialize() serialize.SerializeStream {
	return nil
}

func (a *Account) Deserialize(s serialize.SerializeStream) error {
	return nil
}

func (id *Account) String() string {
	data, err := json.Marshal(id)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
