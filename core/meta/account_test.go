package meta

import (
	"encoding/hex"
	"testing"

	"github.com/mihongtech/linkchain/common/btcec"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/protobuf"
	"github.com/mihongtech/linkchain/unittest"

	"github.com/golang/protobuf/proto"
)

//var testSecurityPri, _ = hex.DecodeString("bea9c932a33a9bf947625a490b297f1fe83abdacd971ffa65a51011a75f888f9")

//Create Account for test.
func getTestAccount() *Account {
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	utxos := make([]UTXO, 0)
	c := NewClearTime(0, 0)
	testAccount := NewAccount(*id, config.NormalAccount, utxos, c, *id)
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)

	u := NewUTXO(ticket, 10, 10, *NewAmount(10))

	testAccount.UTXOs = append(testAccount.UTXOs, *u)
	return testAccount
}

//Testing the method 'Serialize' of  account.
func TestAccount_Serialize(t *testing.T) {
	account := getTestAccount()
	s := account.Serialize()

	_, err := proto.Marshal(s)
	unittest.NotError(t, err)
}

//Testing the method 'Deserializ' of  account.
func TestAccount_Deserialize(t *testing.T) {
	str := "0a230a21033bc87ab98c040f90b3fe0fb4ba0ec8c1907b2ec08926bb31621e8aa3b001d16910001a300a260a220a20640fe47bb4898f552ef35ea64026fd7304960254269b9ac3dabcdd6cfc126e5e1000107818960122010a220808001000180020002a230a21033bc87ab98c040f90b3fe0fb4ba0ec8c1907b2ec08926bb31621e8aa3b001d169"
	buffer, _ := hex.DecodeString(str)
	pa := &protobuf.Account{}

	err := proto.Unmarshal(buffer, pa)
	unittest.NotError(t, err)

	a := Account{}
	err = a.Deserialize(pa)
	unittest.NotError(t, err)
}

//Testing the method 'Serialize' of  account.
func TestAccount_Serialize_Empty_Security(t *testing.T) {
	account := getTestAccount()
	account.SecurityId = AccountID{}
	s := account.Serialize()

	_, err := proto.Marshal(s)
	unittest.NotError(t, err)
}

//Testing the method 'Deserializ' of  account.
func TestAccount_Deserialize_Empty_Security(t *testing.T) {
	str := "0a230a21025d17cfc6faed6565a193e44e3de54955976ed1f18eeee6ab55ce79adcfa3215110001a300a260a220a20640fe47bb4898f552ef35ea64026fd7304960254269b9ac3dabcdd6cfc126e5e1000107818960122010a"
	buffer, _ := hex.DecodeString(str)
	pa := &protobuf.Account{}

	err := proto.Unmarshal(buffer, pa)
	unittest.NotError(t, err)

	a := Account{}
	err = a.Deserialize(pa)
	unittest.NotError(t, err)
}

func TestAccount_IsCanSetClearTime(t *testing.T) {
	account := getTestAccount()
	unittest.Assert(t, account.IsCanSetClearTime(0), "IsCanSetClearTime")
	unittest.Assert(t, account.IsCanSetClearTime(1), "IsCanSetClearTime")

	//test one setClearTime
	account.Clear.NextEffectHeight = 10
	account.Clear.NextClearTime = 6
	unittest.Assert(t, !account.IsCanSetClearTime(0), "IsCanSetClearTime")
	unittest.Assert(t, !account.IsCanSetClearTime(5), "IsCanSetClearTime")
	unittest.Assert(t, !account.IsCanSetClearTime(6), "IsCanSetClearTime")
	unittest.Assert(t, !account.IsCanSetClearTime(7), "IsCanSetClearTime")
	unittest.Assert(t, !account.IsCanSetClearTime(9), "IsCanSetClearTime")
	unittest.Assert(t, account.IsCanSetClearTime(10), "IsCanSetClearTime")
	unittest.Assert(t, account.IsCanSetClearTime(11), "IsCanSetClearTime")

	//test init clearTime
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 0
	account.Clear.NextClearTime = 0
	unittest.Assert(t, account.IsCanSetClearTime(0), "IsCanSetClearTime")
	unittest.Assert(t, account.IsCanSetClearTime(5), "IsCanSetClearTime")
	unittest.Assert(t, account.IsCanSetClearTime(6), "IsCanSetClearTime")
	unittest.Assert(t, account.IsCanSetClearTime(7), "IsCanSetClearTime")
	unittest.Assert(t, account.IsCanSetClearTime(9), "IsCanSetClearTime")
	unittest.Assert(t, account.IsCanSetClearTime(10), "IsCanSetClearTime")
	unittest.Assert(t, account.IsCanSetClearTime(11), "IsCanSetClearTime")

	//test have set two time clearTime,but the recent set have not effecting
	//test init clearTime
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 18
	account.Clear.NextClearTime = 9
	unittest.Assert(t, !account.IsCanSetClearTime(0), "IsCanSetClearTime")
	unittest.Assert(t, !account.IsCanSetClearTime(11), "IsCanSetClearTime")
	unittest.Assert(t, !account.IsCanSetClearTime(17), "IsCanSetClearTime")

	//test have set two time clearTime,but the recent set had effecting
	//test init clearTime
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 18
	account.Clear.NextClearTime = 9
	unittest.Assert(t, account.IsCanSetClearTime(18), "IsCanSetClearTime")
	unittest.Assert(t, account.IsCanSetClearTime(19), "IsCanSetClearTime")
}

//test checkFromCoin with correct fc
func TestAccount_CheckFromCoin(t *testing.T) {
	account := getTestAccount()
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)
	isHave := account.CheckFromCoin(fc)
	unittest.Assert(t, isHave, "CheckFromCoin")

}

//test checkFromCoin with error fc ticket
func TestAccount_CheckFromCoin2(t *testing.T) {
	account := getTestAccount()
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)
	isHave := account.CheckFromCoin(fc)
	unittest.Assert(t, !isHave, "CheckFromCoin")

}

//test checkFromCoin with error fc account
func TestAccount_CheckFromCoin3(t *testing.T) {
	account := getTestAccount()
	ex, _ := btcec.NewPrivateKey(btcec.S256())
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)
	isHave := account.CheckFromCoin(fc)
	unittest.Assert(t, !isHave, "CheckFromCoin")
}

//test checkFromCoin with correct ticket
func TestAccount_Contains(t *testing.T) {
	account := getTestAccount()
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	isHave := account.Contains(*ticket)
	unittest.Assert(t, isHave, "Contains")
}

//test checkFromCoin with error ticket
func TestAccount_Contains2(t *testing.T) {
	account := getTestAccount()
	txid, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")
	ticket := NewTicket(*txid, 0)
	isHave := account.Contains(*ticket)
	unittest.Assert(t, !isHave, "Contains")
}

//test checkFromCoin with normal utxo
func TestAccount_IsFromEffect(t *testing.T) {
	account := getTestAccount()
	//correct fc
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)

	unittest.Assert(t, !account.IsFromEffect(fc, 9), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc, 10), "IsFromEffect")
	unittest.Assert(t, account.IsFromEffect(fc, 11), "IsFromEffect")

	//error fc account
	ex1, _ := btcec.NewPrivateKey(btcec.S256())
	id1 := NewAccountId(ex1.PubKey())
	txid1, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket1 := NewTicket(*txid1, 0)
	tickets1 := make([]Ticket, 0)
	tickets1 = append(tickets1, *ticket1)
	fc1 := NewFromCoin(*id1, tickets1)

	unittest.Assert(t, !account.IsFromEffect(fc1, 9), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc1, 10), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc1, 11), "IsFromEffect")

	//err fc ticket but correct fc account
	//correct fc
	ex2, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id2 := NewAccountId(ex2.PubKey())
	txid2, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")
	ticket2 := NewTicket(*txid2, 0)
	tickets2 := make([]Ticket, 0)
	tickets2 = append(tickets2, *ticket2)
	fc2 := NewFromCoin(*id2, tickets2)

	unittest.Assert(t, !account.IsFromEffect(fc2, 9), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc2, 10), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc2, 11), "IsFromEffect")
}

//test checkFromCoin with delay utxo
func TestAccount_IsFromEffect2(t *testing.T) {
	account := getTestAccount()
	account.UTXOs[0].EffectHeight += config.DefaultDelayTime
	//correct fc
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)

	unittest.Assert(t, !account.IsFromEffect(fc, 9), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc, 10), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc, 11), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc, 13), "IsFromEffect")
	unittest.Assert(t, account.IsFromEffect(fc, 14), "IsFromEffect")
	unittest.Assert(t, account.IsFromEffect(fc, 15), "IsFromEffect")

	//error fc account
	ex1, _ := btcec.NewPrivateKey(btcec.S256())
	id1 := NewAccountId(ex1.PubKey())
	txid1, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket1 := NewTicket(*txid1, 0)
	tickets1 := make([]Ticket, 0)
	tickets1 = append(tickets1, *ticket1)
	fc1 := NewFromCoin(*id1, tickets1)

	unittest.Assert(t, !account.IsFromEffect(fc1, 9), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc1, 10), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc1, 11), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc1, 13), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc1, 14), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc1, 15), "IsFromEffect")

	//err fc ticket but correct fc account
	//correct fc
	ex2, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id2 := NewAccountId(ex2.PubKey())
	txid2, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")
	ticket2 := NewTicket(*txid2, 0)
	tickets2 := make([]Ticket, 0)
	tickets2 = append(tickets2, *ticket2)
	fc2 := NewFromCoin(*id2, tickets2)

	unittest.Assert(t, !account.IsFromEffect(fc2, 9), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc2, 10), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc2, 11), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc2, 13), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc2, 14), "IsFromEffect")
	unittest.Assert(t, !account.IsFromEffect(fc2, 15), "IsFromEffect")
}

func TestAccount_RemoveUTXOByFromCoin(t *testing.T) {
	account := getTestAccount()
	//correct fc
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)

	err := account.RemoveUTXOByFromCoin(fc)
	unittest.NotError(t, err)
}

func TestAccount_RemoveUTXOByFromCoin_Error_Ticket(t *testing.T) {
	account := getTestAccount()
	//correct fc
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)

	err := account.RemoveUTXOByFromCoin(fc)
	unittest.Error(t, err)
}

func TestAccount_RemoveUTXOByFromCoin_Error_Account(t *testing.T) {
	account := getTestAccount()
	//correct fc
	ex, _ := btcec.NewPrivateKey(btcec.S256())
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)

	err := account.RemoveUTXOByFromCoin(fc)
	unittest.Error(t, err)
}

func resetClearTime(account *Account) {
	account.Clear.LastEffectHeight = 0
	account.Clear.LastClearTime = 0
	account.Clear.NextEffectHeight = 0
	account.Clear.NextClearTime = 0
}

func TestAccount_SetClearTime(t *testing.T) {
	account := getTestAccount()
	unittest.Assert(t, account.SetClearTime(6, 10, 0), "SetClearTime")

	resetClearTime(account)
	unittest.Assert(t, account.SetClearTime(6, 10, 7), "SetClearTime")

	//test one setClearTime
	resetClearTime(account)
	account.Clear.NextEffectHeight = 10
	account.Clear.NextClearTime = 6
	unittest.Assert(t, !account.SetClearTime(9, 18, 0), "SetClearTime")

	resetClearTime(account)
	account.Clear.NextEffectHeight = 10
	account.Clear.NextClearTime = 6
	unittest.Assert(t, !account.SetClearTime(9, 18, 9), "SetClearTime")

	resetClearTime(account)
	account.Clear.NextEffectHeight = 10
	account.Clear.NextClearTime = 6
	unittest.Assert(t, account.SetClearTime(9, 18, 10), "SetClearTime")

	resetClearTime(account)
	account.Clear.NextEffectHeight = 10
	account.Clear.NextClearTime = 6
	unittest.Assert(t, account.SetClearTime(9, 18, 11), "SetClearTime")

	//test init clearTime
	resetClearTime(account)
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 0
	account.Clear.NextClearTime = 0
	unittest.Assert(t, account.SetClearTime(9, 18, 0), "SetClearTime")

	resetClearTime(account)
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 0
	account.Clear.NextClearTime = 0
	unittest.Assert(t, account.SetClearTime(9, 18, 9), "SetClearTime")

	resetClearTime(account)
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 0
	account.Clear.NextClearTime = 0
	unittest.Assert(t, account.SetClearTime(9, 18, 10), "SetClearTime")

	resetClearTime(account)
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 0
	account.Clear.NextClearTime = 0
	unittest.Assert(t, account.SetClearTime(9, 18, 11), "SetClearTime")

	//test have set two time clearTime,but the recent set have not effecting
	//test init clearTime
	resetClearTime(account)
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 18
	account.Clear.NextClearTime = 9
	unittest.Assert(t, !account.SetClearTime(9, 18, 0), "SetClearTime")

	resetClearTime(account)
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 18
	account.Clear.NextClearTime = 9
	unittest.Assert(t, !account.SetClearTime(10, 22, 10), "SetClearTime")

	resetClearTime(account)
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 18
	account.Clear.NextClearTime = 9
	unittest.Assert(t, !account.SetClearTime(10, 22, 11), "SetClearTime")

	resetClearTime(account)
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 18
	account.Clear.NextClearTime = 9
	unittest.Assert(t, !account.SetClearTime(10, 22, 17), "SetClearTime")

	//test have set two time clearTime,but the recent set had effecting
	resetClearTime(account)
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 18
	account.Clear.NextClearTime = 9
	unittest.Assert(t, account.SetClearTime(10, 22, 18), "SetClearTime")

	resetClearTime(account)
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 18
	account.Clear.NextClearTime = 9
	unittest.Assert(t, account.SetClearTime(10, 22, 19), "SetClearTime")

}

func TestAccount_GetClearTime(t *testing.T) {
	account := getTestAccount()
	unittest.Equal(t, account.GetClearTime(0), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(1), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(2), account.Clear.LastClearTime)

	//test one setClearTime
	account.Clear.NextEffectHeight = 10
	account.Clear.NextClearTime = 6
	unittest.Equal(t, account.GetClearTime(0), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(9), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(10), account.Clear.NextClearTime)
	unittest.Equal(t, account.GetClearTime(11), account.Clear.NextClearTime)

	//test init clearTime
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 0
	account.Clear.NextClearTime = 0
	unittest.Equal(t, account.GetClearTime(0), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(9), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(10), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(11), account.Clear.LastClearTime)

	//test have set two time clearTime,but the recent set have not effecting
	//test init clearTime
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 18
	account.Clear.NextClearTime = 9
	unittest.Equal(t, account.GetClearTime(9), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(10), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(11), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(13), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(17), account.Clear.LastClearTime)
	unittest.Equal(t, account.GetClearTime(18), account.Clear.NextClearTime)
	unittest.Equal(t, account.GetClearTime(19), account.Clear.NextClearTime)
}

func TestAccount_GetAmount(t *testing.T) {
	account := getTestAccount()
	unittest.Equal(t, account.GetAmount().GetInt64(), int64(10))
}

func TestAccount_GetFromCoinValue_Null(t *testing.T) {
	account := getTestAccount()
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)
	_, err := account.GetFromCoinValue(fc)
	unittest.Error(t, err)
}

func TestAccount_GetFromCoinValue(t *testing.T) {
	account := getTestAccount()
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)
	amount, err := account.GetFromCoinValue(fc)
	unittest.NotError(t, err)
	unittest.Equal(t, amount.GetInt64(), int64(10))
}

func TestAccount_GetFromCoinValue_Error_Account(t *testing.T) {
	account := getTestAccount()
	ex, _ := btcec.NewPrivateKey(btcec.S256())
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	fc := NewFromCoin(*id, tickets)
	_, err := account.GetFromCoinValue(fc)
	unittest.Error(t, err)
}

func TestAccount_GetUTXO(t *testing.T) {
	account := getTestAccount()
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	u := account.GetUTXO(*ticket)
	unittest.Equal(t, *u, account.UTXOs[0])
}

func TestAccount_GetUTXO_Error_Account(t *testing.T) {
	account := getTestAccount()
	txid, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")
	ticket := NewTicket(*txid, 0)
	u := account.GetUTXO(*ticket)
	unittest.NotEqual(t, u, account.UTXOs[0])
}

func TestClearTime_IsNextEffect(t *testing.T) {
	account := getTestAccount()
	unittest.Assert(t, !account.Clear.IsNextEffect(0), "IsNextEffect")
	unittest.Assert(t, !account.Clear.IsNextEffect(0), "IsNextEffect")

	//test init clearTime
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	unittest.Assert(t, !account.Clear.IsNextEffect(0), "IsNextEffect")
	unittest.Assert(t, !account.Clear.IsNextEffect(9), "IsNextEffect")
	unittest.Assert(t, !account.Clear.IsNextEffect(10), "IsNextEffect")
	unittest.Assert(t, !account.Clear.IsNextEffect(11), "IsNextEffect")

	//test have set two time clearTime
	//test init clearTime
	account.Clear.LastEffectHeight = 10
	account.Clear.LastClearTime = 6
	account.Clear.NextEffectHeight = 18
	account.Clear.NextClearTime = 9
	unittest.Assert(t, !account.Clear.IsNextEffect(0), "IsNextEffect")
	unittest.Assert(t, !account.Clear.IsNextEffect(9), "IsNextEffect")
	unittest.Assert(t, !account.Clear.IsNextEffect(10), "IsNextEffect")
	unittest.Assert(t, !account.Clear.IsNextEffect(11), "IsNextEffect")
	unittest.Assert(t, !account.Clear.IsNextEffect(12), "IsNextEffect")
	unittest.Assert(t, !account.Clear.IsNextEffect(17), "IsNextEffect")
	unittest.Assert(t, account.Clear.IsNextEffect(18), "IsNextEffect")
	unittest.Assert(t, account.Clear.IsNextEffect(19), "IsNextEffect")

}
