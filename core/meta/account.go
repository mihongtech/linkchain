package meta

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"strings"

	"github.com/mihongtech/linkchain/common/btcec"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/serialize"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
)

const AccountLength = 20

type AccountID [AccountLength]byte

func CreateAccountId(b []byte) AccountID {
	hash := math.HashB(b)
	return BytesToAccountID(hash[12:])
}

func BytesToAccountID(b []byte) AccountID {
	var a AccountID
	if len(b) > len(a) {
		b = b[len(b)-AccountLength:]
	}
	copy(a[AccountLength-len(b):], b)
	return a
}

func HexToAccountID(str string) (AccountID, error) {
	if len(str) > 1 {
		if str[0:2] == "0x" || str[0:2] == "0X" {
			str = str[2:]
		}
	}
	if len(str)%2 == 1 {
		str = "0" + str
	}

	data, err := hex.DecodeString(str)
	if err != nil {
		return AccountID{}, err
	}
	return BytesToAccountID(data), nil
}

func (a AccountID) String() string {
	return hex.EncodeToString(a[:])
}

func (a AccountID) IsEqual(other AccountID) bool {
	return strings.Compare(a.String(), other.String()) == 0
}

func (a AccountID) IsEmpty() bool {
	isEmpty := true
	l := len(a)
	for i := 0; i < l; i++ {
		if a[i] != 0 {
			isEmpty = false
			break
		}
	}
	return isEmpty
}

//Serialize/Deserialize
func (a *AccountID) Serialize() serialize.SerializeStream {
	accountId := protobuf.AccountID{
		Id: proto.NewBuffer(a[:]).Bytes(),
	}
	return &accountId
}

func (a *AccountID) Deserialize(s serialize.SerializeStream) error {
	data := s.(*protobuf.AccountID)

	return a.SetBytes(data.Id)
}

func (a *AccountID) SetBytes(b []byte) error {
	if len(b) > AccountLength {
		return errors.New("byte's len more than max account length")
	}
	copy(a[:], b)
	return nil
}

func (a AccountID) CloneBytes() []byte {
	return a[:]
}

// Big converts an address to a big integer.
func (a AccountID) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

// BigToAccountId returns Address with byte values of b.
// If b is larger than len(h), b will be cropped from the left.
func BigToAccountId(b *big.Int) AccountID { return BytesToAccountID(b.Bytes()) }

//Json Hash convert to Hex
func (a AccountID) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *AccountID) UnmarshalJSON(data []byte) error {
	var str string = ""
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	account, err := HexToAccountID(str)
	if err != nil {
		return err
	}
	return a.SetBytes(account.CloneBytes())
}

func NewAccountId(pubkey *btcec.PublicKey) *AccountID {
	// TODO: maybe use bitcion account generate function
	id := BytesToAccountID(math.HashB(pubkey.SerializeCompressed())[12:])
	return &id
}

func NewAccountIdFromStr(str string) (*AccountID, error) {
	buff, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	id := BytesToAccountID(buff)
	return &id, nil
}

type UTXO struct {
	Ticket
	LocatedHeight uint32 `json:"locatedHeight"`
	EffectHeight  uint32 `json:"effectHeight"`
	Value         Amount `json:"value"`
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

type ClearTime struct {
	LastClearTime    int64  `json:"currentClearTime"`
	LastEffectHeight uint32 `json:"lastEffectHeight"`
	NextClearTime    int64  `json:"nextClearTime"`    // the next clear time.If it is puls 0,user can not change clear time
	NextEffectHeight uint32 `json:"nextEffectHeight"` // the effective height of next clear time
}

// get ClearTime.
// when ClearTime init,the nextClearTime and NextEffectHeight must be 0.
func NewClearTime(lastClearTime int64, lastEffectHeight uint32) *ClearTime {
	return &ClearTime{LastClearTime: lastClearTime, LastEffectHeight: lastEffectHeight, NextClearTime: 0, NextEffectHeight: 0}
}

// get clearTime.
// If nextClearTime had effect,then return nextClearTime.
func (c *ClearTime) GetClearTime(blockHeight uint32) int64 {
	if c.IsNextEffect(blockHeight) {
		//last setClearTime have effect
		return c.NextClearTime
	} else {
		return c.LastClearTime
	}
}

// set clearTime.
// If nextClearTime had effect,then can set clearTime.
// If NextEffectHeight == 0 (ClearTime had not be set clearTime),then then can set clearTime.
func (c *ClearTime) SetClearTime(clearTime int64, effectHeight uint32, blockHeight uint32) bool {
	if c.IsCanSet(blockHeight) {
		c.LastClearTime = c.NextClearTime
		c.LastEffectHeight = c.NextEffectHeight
		c.NextClearTime = clearTime
		c.NextEffectHeight = effectHeight
		return true
	} else {
		//the account have ineffective clear time.
		return false
	}
}

// Check nextClear have effected with current block height.
// If NextEffectHeight > 0 (ClearTime had be set clearTime),then check block have reached to nextEffectHeight.
// If NextEffectHeight == 0 (ClearTime had not be set clearTime),then nextClear.
func (c *ClearTime) IsNextEffect(blockHeight uint32) bool {
	return c.NextEffectHeight <= blockHeight && c.NextEffectHeight > 0
}

func (c *ClearTime) IsCanSet(blockHeight uint32) bool {
	return c.IsNextEffect(blockHeight) || c.NextEffectHeight == 0
}

func (c *ClearTime) Serialize() serialize.SerializeStream {
	s := &protobuf.ClearTime{
		LastClearTime:    proto.Int64(c.LastClearTime),
		LastEffectHeight: proto.Uint32(c.LastEffectHeight),
		NextClearTime:    proto.Int64(c.NextClearTime),
		NextEffectHeight: proto.Uint32(c.NextEffectHeight),
	}
	return s
}

func (c *ClearTime) Deserialize(s serialize.SerializeStream) error {
	data := s.(*protobuf.ClearTime)
	c.LastClearTime = *data.LastClearTime
	c.LastEffectHeight = *data.LastEffectHeight
	c.NextClearTime = *data.NextClearTime
	c.NextEffectHeight = *data.NextEffectHeight
	return nil
}

func (c *ClearTime) String() string {
	data, err := json.Marshal(c)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

type Account struct {
	Id          AccountID `json:"accountId"`
	AccountType uint32    `json:"accountType"`
	UTXOs       []UTXO    `json:"AccountUXTO"`
	Clear       ClearTime `json:"clearTime"`
	SecurityId  AccountID `json:"securityId"`
	StorageRoot TreeID    `json:"storageRoot"`
	CodeHash    math.Hash `json:"codeHash"`
}

func NewAccount(id AccountID, accountType uint32, utxos []UTXO, clearTime *ClearTime, securityId AccountID) *Account {
	return &Account{Id: id, AccountType: accountType, UTXOs: utxos, Clear: *clearTime, SecurityId: securityId, StorageRoot: math.Hash{}, CodeHash: math.Hash{}}
}

func (a Account) GetAccountID() *AccountID {
	return &a.Id
}

func (a Account) GetAmount() *Amount {
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

//check fromCoin ticket effect.
//the current block height must be > effectHeight
func (a *Account) IsFromEffect(fromCoin *FromCoin, height uint32) bool {
	if a.GetAccountID().IsEqual(fromCoin.GetId()) {
		tickets := fromCoin.GetTickets()
		for _, t := range tickets {
			u, err := a.getUTXOByTicket(t)
			if err != nil {
				return false
			}
			if height < u.EffectHeight {
				return false
			}
		}
		return true
	} else {
		return false
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

func (a *Account) GetUTXO(ticket Ticket) *UTXO {
	for _, t := range a.UTXOs {
		if ticket.GetTxid().IsEqual(&t.Txid) && t.Index == ticket.GetIndex() {
			return &t
		}
	}
	return nil
}

func (a *Account) GetClearTime(height uint32) int64 {
	return a.Clear.GetClearTime(height)
}

func (a *Account) SetClearTime(clearTime int64, effectHeight uint32, blockHeight uint32) bool {
	return a.Clear.SetClearTime(clearTime, effectHeight, blockHeight)
}

func (a *Account) IsCanSetClearTime(blockHeight uint32) bool {
	return a.Clear.IsCanSet(blockHeight)
}

//Serialize/Deserialize
func (a *Account) Serialize() serialize.SerializeStream {
	us := make([]*protobuf.UTXO, 0)
	for index := range a.UTXOs {
		t := NewTicket(a.UTXOs[index].Txid, a.UTXOs[index].Index)
		u := &protobuf.UTXO{
			Id:            t.Serialize().(*protobuf.Ticket),
			LocatedHeight: proto.Uint32(a.UTXOs[index].LocatedHeight),
			EffectHeight:  proto.Uint32(a.UTXOs[index].EffectHeight),
			Value:         proto.NewBuffer(a.UTXOs[index].Value.GetBytes()).Bytes(),
		}
		us = append(us, u)
	}
	s := &protobuf.Account{
		Id:    a.Id.Serialize().(*protobuf.AccountID),
		Type:  proto.Uint32(a.AccountType),
		Utxos: us,
	}
	if !a.SecurityId.IsEmpty() {
		clearTime := a.Clear.Serialize().(*protobuf.ClearTime)
		s.Clear = clearTime
		s.SecurityId = a.SecurityId.Serialize().(*protobuf.AccountID)
	}

	if !a.CodeHash.IsEmpty() {
		s.CodeHash = a.CodeHash.Serialize().(*protobuf.Hash)
	}

	if !a.StorageRoot.IsEmpty() {
		s.StorageRoot = a.StorageRoot.Serialize().(*protobuf.Hash)
	}

	return s

}

func (a *Account) Deserialize(s serialize.SerializeStream) error {
	data := s.(*protobuf.Account)
	if err := a.Id.Deserialize(data.Id); err != nil {
		return err
	}

	if data.SecurityId != nil {
		if err := a.SecurityId.Deserialize(data.SecurityId); err != nil {
			return err
		}

		if err := a.Clear.Deserialize(data.Clear); err != nil {
			return err
		}
	}

	if data.StorageRoot != nil {
		if err := a.StorageRoot.Deserialize(data.StorageRoot); err != nil {
			return err
		}
	}

	if data.CodeHash != nil {
		if err := a.CodeHash.Deserialize(data.CodeHash); err != nil {
			return err
		}
	}

	a.AccountType = *data.Type

	a.UTXOs = a.UTXOs[:0] // UTXOs clear

	for _, u := range data.Utxos {
		newUtxo := UTXO{}
		newTicket := &Ticket{}
		if err := newTicket.Deserialize(u.Id); err != nil {
			return err
		}
		newUtxo.Txid = newTicket.Txid
		newUtxo.Index = newTicket.Index
		newUtxo.Value = *NewAmount(0)
		newUtxo.Value.SetBytes(u.Value)
		newUtxo.LocatedHeight = *u.LocatedHeight
		newUtxo.EffectHeight = *u.EffectHeight

		a.UTXOs = append(a.UTXOs, newUtxo)
	}
	return nil
}

func (a *Account) String() string {
	data, err := json.Marshal(a)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (a *Account) MakeFromCoin(value *Amount, blockHeight uint32) (*FromCoin, *Amount, error) {
	if a.GetAmount().GetInt64() < value.GetInt64() {
		log.Error("MakeFromCoin failed", "a.GetAmount().GetInt64()", a.GetAmount().GetInt64(), "value.GetInt64()", value.GetInt64())
		return nil, nil, errors.New("Account MakeFromCoin() amount is too large")
	}
	tempUTXOs := make([]UTXO, 0)
	tempUTXOs = append(tempUTXOs, a.UTXOs...)
	sort.Slice(tempUTXOs, func(i, j int) bool {
		if tempUTXOs[i].Txid.Big().Cmp(tempUTXOs[j].Txid.Big()) == 0 {
			return tempUTXOs[i].Index > tempUTXOs[j].Index
		} else if tempUTXOs[i].Txid.Big().Cmp(tempUTXOs[j].Txid.Big()) < 0 {
			return true
		} else {
			return false
		}
	})
	tickets := make([]Ticket, 0)
	fc := NewFromCoin(a.Id, tickets)
	fromAmount := NewAmount(0)
	for _, v := range tempUTXOs {
		if blockHeight < v.EffectHeight {
			continue
		}
		// if not enough add a ticket
		if fromAmount.IsLessThan(*value) {
			fromAmount.Addition(v.Value)
			t := NewTicket(v.Txid, v.Index)
			fc.AddTicket(t)
		} else {
			break
		}
	}
	if len(fc.Ticket) == 0 || fromAmount.GetInt64() < value.GetInt64() {
		log.Error("MakeFromCoin failed", "len(fc.Ticket)", len(fc.Ticket), "a.GetAmount().GetInt64()", a.GetAmount().GetInt64(), "value.GetInt64()", value.GetInt64())
		return nil, nil, errors.New("Account MakeFromCoin() can not cover value.the value is too large")
	}
	return fc, fromAmount, nil
}

func GetAccountHash(id AccountID) math.Hash {
	return math.HashH(id.CloneBytes())
}
