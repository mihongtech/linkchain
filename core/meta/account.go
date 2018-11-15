package meta

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
)

type AccountID struct {
	ID []byte `json:"id"`
}

func (a *AccountID) String() string {
	return hex.EncodeToString(a.ID)
}

func (a *AccountID) IsEqual(other AccountID) bool {
	return strings.Compare(a.String(), other.String()) == 0
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
	LocatedHeight uint32        `json:"locatedHeight"`
	EffectHeight  uint32        `json:"effectHeight"`
	Value         Amount 		`json:"value"`
}

func NewUTXO(tickets *Ticket, locatedHeight uint32, effectHeight uint32, value Amount) *UTXO {
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
	Id          AccountID `json:"accountId"`
	AccountType uint32    `json:"accountType"`
	UTXOs       []UTXO    `json:"AccountUXTO"`
	ClearTime   int64     `json:"clearTime"`
	SecurityId  AccountID `json:"securityId"`
}

func NewAccount(id AccountID, accountType uint32, utxos []UTXO, clearTime int64, securityId AccountID) *Account {
	return &Account{Id: id, AccountType: accountType, UTXOs: utxos, ClearTime: clearTime, SecurityId: securityId}
}

func (a *Account) GetAccountID() *AccountID {
	return &a.Id
}

func (a *Account) GetAmount() *Amount {
	sum := NewAmount(0)
	for _, u := range a.UTXOs {
		sum.Addition(u.Value)
	}
	return sum
}

func (a *Account) GetFromCoinValue(fromCoin *FromCoin) (*Amount, error) {
	sum := NewAmount(0)
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
		return nil, errors.New("fromCoin's AccountId is error")
	}
}

func (a *Account) CheckFromCoin(fromCoin *FromCoin) bool {
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

func (a *Account) RemoveUTXOByFromCoin(fromCoin *FromCoin) error {
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
				return errors.New("removeUTXO():The ticket is not exist in Account")

			}
		}
		return nil
	} else {
		return errors.New("fromCoin's AccountId is error")
	}
}

func (a *Account) getUTXOByTicket(ticket Ticket) (*UTXO, error) {
	for _, t := range a.UTXOs {
		if ticket.GetTxid().IsEqual(&t.Txid) && t.Index == ticket.GetIndex() {
			return &t, nil
		}
	}
	return nil, errors.New("the ticket is not exist in Account")
}

func (a *Account) Contains(ticket Ticket) bool {
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

func (a *Account) String() string {
	data, err := json.Marshal(a)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
