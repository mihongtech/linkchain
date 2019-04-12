package meta

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/mihongtech/linkchain/common/btcec"
	_ "github.com/mihongtech/linkchain/common/btcec"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/protobuf"
	"github.com/mihongtech/linkchain/unittest"
)

var testPri, _ = hex.DecodeString("7a9c6f2b865c98c9fe174869de5818f4c62bc845441c08269487cdba6688f6b1")

//Create transaction for test.
func getTestTransaction() *Transaction {
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	utxos := make([]UTXO, 0)
	c := NewClearTime(0, 0)
	toAccount := NewAccount(*id, 0, utxos, c, *id)

	toId := *toAccount.GetAccountID()

	fcs := make([]FromCoin, 0)
	fcs = append(fcs, *getTestFromCoin())
	tf := NewTransactionFrom(fcs)

	tc := NewToCoin(toId, NewAmount(10))
	tcs := make([]ToCoin, 0)
	tcs = append(tcs, *tc)
	tt := NewTransactionTo(tcs)

	signs := make([]Signature, 0)
	testTx := NewTransaction(config.DefaultTransactionVersion, config.NormalTx, *tf, *tt, signs, nil)

	sign, _ := ex.Sign(testTx.GetTxID().CloneBytes())

	signature := NewSignature(sign.Serialize())
	testTx.AddSignature(signature)

	return testTx
}

//Create transactionFrom for test.
func getTransactionFrom() *TransactionFrom {
	fcs := make([]FromCoin, 0)
	fcs = append(fcs, *getTestFromCoin())
	return NewTransactionFrom(fcs)
}

//Create fromCoin for test.
func getTestFromCoin() *FromCoin {
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	ticket := NewTicket(*txid, 0)
	tickets := make([]Ticket, 0)
	tickets = append(tickets, *ticket)
	return NewFromCoin(*id, tickets)
}

//Create Signature for test.
func getTestSignature() *Signature {
	tx := getTestTransaction()
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	sign, _ := ex.Sign(tx.GetTxID().CloneBytes())
	return NewSignature(sign.Serialize())
}

//Create TransactionTo for test.
func getTestTransactionTo() *TransactionTo {
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	tc := NewToCoin(*id, NewAmount(10))
	tcs := make([]ToCoin, 0)
	tcs = append(tcs, *tc)
	tt := NewTransactionTo(tcs)
	return tt
}

//Create Ticket for test.
func getTestTicket() *Ticket {
	txid, _ := math.NewHashFromStr("5e6e12fc6cddbcdac39a9b265402960473fd2640a65ef32e558f89b47be40f64")
	return NewTicket(*txid, 0)
}

//Create ToCoin for test.
func getTestToCoin() *ToCoin {
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	tc := NewToCoin(*id, NewAmount(10))
	return tc
}

//Testing the method 'Verify' of transaction.
func TestTransaction_Verify(t *testing.T) {
	tx := getTestTransaction()
	err := tx.Verify()
	unittest.NotError(t, err)
}

//Testing the method 'Verify' of transaction with more sign.
func TestTransaction_Verify_More_Sign(t *testing.T) {
	str := "080110011a4f0a4d0a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012260a220a20640fe47bb4898f552ef35ea64026fd7304960254269b9ac3dabcdd6cfc126e5e1000222a0a280a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012010a2a490a473045022100b3b46c98236f2760344e5c9aaec44e7463f2435858c396fd731eed0e03f28d2502205cee7691880636a5643b5cc24fc05196ec2184939fccf289753efaaefe2795162a490a473045022100b3b46c98236f2760344e5c9aaec44e7463f2435858c396fd731eed0e03f28d2502205cee7691880636a5643b5cc24fc05196ec2184939fccf289753efaaefe279516"
	buffer, _ := hex.DecodeString(str)

	tx := &protobuf.Transaction{}

	err := proto.Unmarshal(buffer, tx)
	unittest.NotError(t, err)

	newTx := Transaction{}
	err = newTx.Deserialize(tx)
	unittest.NotError(t, err)

	err = newTx.Verify()
	unittest.NotEqual(t, err, nil)
}

//Testing the method 'Verify' of transaction with error sign.
func TestTransaction_Verify_Error_Sign(t *testing.T) {
	ex, _ := btcec.NewPrivateKey(btcec.S256())
	id := NewAccountId(ex.PubKey())
	utxos := make([]UTXO, 0)
	c := NewClearTime(0, 0)
	toAccount := NewAccount(*id, 0, utxos, c, *id)

	toId := *toAccount.GetAccountID()

	fcs := make([]FromCoin, 0)
	fcs = append(fcs, *getTestFromCoin())
	tf := NewTransactionFrom(fcs)

	tc := NewToCoin(toId, NewAmount(10))
	tcs := make([]ToCoin, 0)
	tcs = append(tcs, *tc)
	tt := NewTransactionTo(tcs)

	signs := make([]Signature, 0)
	testTx := NewTransaction(config.DefaultTransactionVersion, config.NormalTx, *tf, *tt, signs, nil)

	sign, _ := ex.Sign(testTx.GetTxID().CloneBytes())

	signature := NewSignature(sign.Serialize())
	testTx.AddSignature(signature)

	err := testTx.Verify()
	unittest.NotEqual(t, err, nil)
}

//Testing the method 'Deserialize' of transaction.
func TestTransaction_Deserialize(t *testing.T) {
	txid, _ := math.NewHashFromStr("b97868f005dcbae2e0f110913962e0aeb85fd67797f750c21d7975614923e8d0")
	str := "080110011a4f0a4d0a230a21036f0954edd6804850409f8a7358a01a2a692856f314f0b91f03afc35e51eb863f12260a220a20640fe47bb4898f552ef35ea64026fd7304960254269b9ac3dabcdd6cfc126e5e1000222a0a280a230a21036f0954edd6804850409f8a7358a01a2a692856f314f0b91f03afc35e51eb863f12010a"
	buffer, _ := hex.DecodeString(str)
	tx := &protobuf.Transaction{}

	err := proto.Unmarshal(buffer, tx)
	unittest.NotError(t, err)

	newTx := Transaction{}
	err = newTx.Deserialize(tx)
	unittest.NotError(t, err)
	newTxHash := newTx.GetTxID()
	unittest.Equal(t, txid, newTxHash)
}

//Testing the method 'Serialize' of transaction.
func TestTransaction_Serialize(t *testing.T) {
	tx := getTestTransaction()
	s := tx.Serialize()

	_, err := proto.Marshal(s)
	unittest.NotError(t, err)
	//t.Log("tx Serialize", "txid", tx.GetTxID().String(), "buffer", hex.EncodeToString(buffer))
}

//Testing the method 'Deserialize' of transaction without sign.
func TestTransaction_Deserialize_Nil_Sign(t *testing.T) {
	txid, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")
	str := "080110011a4f0a4d0a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012260a220a20640fe47bb4898f552ef35ea64026fd7304960254269b9ac3dabcdd6cfc126e5e1000222a0a280a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012010a"
	buffer, _ := hex.DecodeString(str)
	tx := &protobuf.Transaction{}

	err := proto.Unmarshal(buffer, tx)
	unittest.NotError(t, err)

	newTx := Transaction{}
	err = newTx.Deserialize(tx)
	unittest.NotError(t, err)
	unittest.Equal(t, len(newTx.Sign), 0)
	newTxHash := newTx.GetTxID()
	unittest.Equal(t, txid, newTxHash)

}

//Testing the method 'Serialize' of transaction without sign.
func TestTransaction_Serialize_Nil_Sign(t *testing.T) {
	tx := getTestTransaction()
	tx.Sign = nil
	s := tx.Serialize()

	_, err := proto.Marshal(s)
	unittest.NotError(t, err)
}

//Testing the method 'Deserialize' of transaction with more sign.
func TestTransaction_Deserialize_more_Sign(t *testing.T) {
	txid, _ := math.NewHashFromStr("3361426edc0980b83404e2f5927d6579040fa26958d77cd5e35bc1fd1e084cf5")
	str := "080110011a4f0a4d0a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012260a220a20640fe47bb4898f552ef35ea64026fd7304960254269b9ac3dabcdd6cfc126e5e1000222a0a280a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012010a2a490a473045022100b3b46c98236f2760344e5c9aaec44e7463f2435858c396fd731eed0e03f28d2502205cee7691880636a5643b5cc24fc05196ec2184939fccf289753efaaefe2795162a490a473045022100b3b46c98236f2760344e5c9aaec44e7463f2435858c396fd731eed0e03f28d2502205cee7691880636a5643b5cc24fc05196ec2184939fccf289753efaaefe279516"
	buffer, _ := hex.DecodeString(str)

	tx := &protobuf.Transaction{}

	err := proto.Unmarshal(buffer, tx)
	unittest.NotError(t, err)

	newTx := Transaction{}
	err = newTx.Deserialize(tx)
	unittest.NotError(t, err)
	unittest.Equal(t, len(newTx.Sign), 2)
	newTxHash := newTx.GetTxID()
	unittest.Equal(t, txid, newTxHash)
}

//Testing the method 'Deserialize' of transaction with more sign.
func TestTransaction_Deserialize_Nil_From(t *testing.T) {
	txid, _ := math.NewHashFromStr("677bcee5837675ed03d92eb328fd30b23383b46acf7e90465063adf21cec2b16")
	str := "080110011a00222a0a280a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012010a"
	buffer, _ := hex.DecodeString(str)
	tx := &protobuf.Transaction{}

	err := proto.Unmarshal(buffer, tx)
	unittest.NotError(t, err)

	newTx := Transaction{}
	err = newTx.Deserialize(tx)
	unittest.NotError(t, err)
	unittest.Equal(t, len(newTx.From.Coins), 0)
	newTxHash := newTx.GetTxID()
	unittest.Equal(t, txid, newTxHash)
}

//Testing the method 'Serialize' of transaction.
func TestTransaction_Serialize_Nil_From(t *testing.T) {
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	id := NewAccountId(ex.PubKey())
	utxos := make([]UTXO, 0)
	c := NewClearTime(0, 0)
	toAccount := NewAccount(*id, 0, utxos, c, *id)

	toId := *toAccount.GetAccountID()

	tc := NewToCoin(toId, NewAmount(10))
	tcs := make([]ToCoin, 0)
	tcs = append(tcs, *tc)
	tt := NewTransactionTo(tcs)

	signs := make([]Signature, 0)
	tx := NewTransaction(config.DefaultTransactionVersion, config.NormalTx, TransactionFrom{}, *tt, signs, nil)

	//tx.From.Coins = tx.From.Coins[0:0]
	s := tx.Serialize()

	_, err := proto.Marshal(s)
	unittest.NotError(t, err)
}

//Testing the method 'Deserialize' of transactionFrom.
func TestTransactionFrom_Deserialize(t *testing.T) {
	hash, _ := math.NewHashFromStr("9ad0088ce685bee6407219040bdfa516e8134d3bdc96ddcc0c14d58a45f3e353")
	buffer, _ := hex.DecodeString("0a4d0a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012260a220a20640fe47bb4898f552ef35ea64026fd7304960254269b9ac3dabcdd6cfc126e5e1000")

	tf := &protobuf.TransactionFrom{}

	err := proto.Unmarshal(buffer, tf)
	unittest.NotError(t, err)

	newTf := TransactionFrom{}
	err = newTf.Deserialize(tf)
	unittest.NotError(t, err)

	newBuffer, err := proto.Marshal(newTf.Serialize())
	unittest.NotError(t, err)
	newHash := math.DoubleHashH(newBuffer)
	unittest.Equal(t, *hash, newHash)
}

func TestTransaction_GetToValue(t *testing.T) {
	tx := getTestTransaction()
	unittest.Equal(t, tx.GetToValue().GetInt64(), int64(10))
}

func TestTransaction_GetNewFromCoins(t *testing.T) {
	tx := getTestTransaction()
	fc := tx.GetNewFromCoins()
	unittest.Equal(t, len(fc), 1)
	unittest.Assert(t, fc[0].Id.IsEqual(tx.From.Coins[0].Id), "GetNewFromCoins")
	unittest.Equal(t, len(fc[0].Ticket), 1)
	unittest.Assert(t, fc[0].Ticket[0].Index == 0, "GetNewFromCoins")
	unittest.Assert(t, fc[0].Ticket[0].Txid.IsEqual(tx.GetTxID()), "GetNewFromCoins")
}

//Testing the method 'Serialize' of transactionFrom.
func TestTransactionFrom_Serialize(t *testing.T) {
	tf := getTransactionFrom()
	s := tf.Serialize()
	_, err := proto.Marshal(s)
	unittest.NotError(t, err)
	//t.Log("transactionFrom Serialize", "transactionFrom hash", math.DoubleHashH(buffer), "buffer", hex.EncodeToString(buffer))
}

//Testing the method 'Deserialize' of fromCoin.
func TestFromCoin_Deserialize(t *testing.T) {
	hash, _ := math.NewHashFromStr("4b655a37f61225098195259f14f1c35941ca60d0cc44866053253b91dc8f33c8")
	buffer, _ := hex.DecodeString("0a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012260a220a20640fe47bb4898f552ef35ea64026fd7304960254269b9ac3dabcdd6cfc126e5e1000")

	fc := &protobuf.FromCoin{}

	err := proto.Unmarshal(buffer, fc)
	unittest.NotError(t, err)

	newFc := FromCoin{}
	err = newFc.Deserialize(fc)
	unittest.NotError(t, err)

	newBuffer, err := proto.Marshal(newFc.Serialize())
	unittest.NotError(t, err)
	newHash := math.DoubleHashH(newBuffer)
	unittest.Equal(t, *hash, newHash)
}

//Testing the method 'Serialize' of fromCoin.
func TestFromCoin_Serialize(t *testing.T) {
	fc := getTestFromCoin()
	s := fc.Serialize()
	_, err := proto.Marshal(s)
	unittest.Equal(t, err, nil)
}

//Testing the method 'Verify' of Signature.
func TestSignature_Verify(t *testing.T) {
	tx := getTestTransaction()
	ex, _ := btcec.PrivKeyFromBytes(btcec.S256(), testPri)
	signature := getTestSignature()
	err := signature.Verify(tx.GetTxID().CloneBytes(), ex.PubKey().SerializeCompressed())
	unittest.NotError(t, err)
}

//Testing the method 'Deserialize' of Signature.
func TestSignature_Deserialize(t *testing.T) {
	hash, _ := math.NewHashFromStr("6045e1be843b2d7292a7ecd512df315d81e77b7817dbd1c6cb379926f4d235e9")
	buffer, _ := hex.DecodeString("0a473045022100b3b46c98236f2760344e5c9aaec44e7463f2435858c396fd731eed0e03f28d2502205cee7691880636a5643b5cc24fc05196ec2184939fccf289753efaaefe279516")

	signature := &protobuf.Signature{}

	err := proto.Unmarshal(buffer, signature)
	unittest.NotError(t, err)

	newSignature := Signature{}
	err = newSignature.Deserialize(signature)
	unittest.NotError(t, err)

	newBuffer, err := proto.Marshal(newSignature.Serialize())
	unittest.NotError(t, err)
	newHash := math.DoubleHashH(newBuffer)
	unittest.Equal(t, *hash, newHash)
}

//Testing the method 'Serialize' of Signature.
func TestSignature_Serialize(t *testing.T) {
	signature := getTestSignature()
	s := signature.Serialize()
	_, err := proto.Marshal(s)
	unittest.Equal(t, err, nil)
}

//Testing the method 'Deserialize' of Signature.
func TestTransactionTo_Deserialize(t *testing.T) {
	hash, _ := math.NewHashFromStr("a6ba19875aca291d175942f2dfbc860ba85c49d71a3565f48f928cb2e4a9043e")
	buffer, _ := hex.DecodeString("0a280a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012010a")

	tt := &protobuf.TransactionTo{}

	err := proto.Unmarshal(buffer, tt)
	unittest.NotError(t, err)

	newTt := TransactionTo{}
	err = newTt.Deserialize(tt)
	unittest.NotError(t, err)

	newBuffer, err := proto.Marshal(newTt.Serialize())
	unittest.NotError(t, err)
	newHash := math.DoubleHashH(newBuffer)
	unittest.Equal(t, *hash, newHash)
}

//Testing the method 'Serialize' of Signature.
func TestTransactionTo_Serialize(t *testing.T) {
	tt := getTestTransactionTo()
	s := tt.Serialize()
	_, err := proto.Marshal(s)
	unittest.Equal(t, err, nil)
}

//Testing the method 'Deserialize' of toCoin.
func TestToCoin_Deserialize(t *testing.T) {
	hash, _ := math.NewHashFromStr("4207cca8763724cf7a61dbdf6999da6902ff772833358dffa65d698f55b07ca2")
	buffer, _ := hex.DecodeString("0a230a2102ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf5012010a")

	tc := &protobuf.ToCoin{}

	err := proto.Unmarshal(buffer, tc)
	unittest.NotError(t, err)

	newTc := ToCoin{}
	err = newTc.Deserialize(tc)
	unittest.NotError(t, err)

	newBuffer, err := proto.Marshal(newTc.Serialize())
	unittest.NotError(t, err)
	newHash := math.DoubleHashH(newBuffer)
	unittest.Equal(t, *hash, newHash)
}

//Testing the method 'Serialize' of toCoin.
func TestToCoin_Serialize(t *testing.T) {
	tc := getTestToCoin()
	s := tc.Serialize()
	_, err := proto.Marshal(s)
	unittest.Equal(t, err, nil)
}

//Testing the method 'Deserialize' of Ticket.
func TestTicket_Deserialize(t *testing.T) {
	hash, _ := math.NewHashFromStr("cbd2621a9eba9b52fc8626a2620e3ef502d73bbf29da52d3924234a570e29180")
	buffer, _ := hex.DecodeString("0a220a20640fe47bb4898f552ef35ea64026fd7304960254269b9ac3dabcdd6cfc126e5e1000")

	ticket := &protobuf.Ticket{}

	err := proto.Unmarshal(buffer, ticket)
	unittest.NotError(t, err)

	newTicket := Ticket{}
	err = newTicket.Deserialize(ticket)
	unittest.NotError(t, err)

	newBuffer, err := proto.Marshal(newTicket.Serialize())
	unittest.NotError(t, err)
	newHash := math.DoubleHashH(newBuffer)
	unittest.Equal(t, *hash, newHash)
}

//Testing the method 'Serialize' of Ticket.
func TestTicket_Serialize(t *testing.T) {
	ticket := getTestTicket()
	s := ticket.Serialize()
	_, err := proto.Marshal(s)
	unittest.Equal(t, err, nil)
}

func TestUmarsh(t *testing.T) {
	txstr := "{\"version\":1,\"type\":7,\"from\":{\"coins\":[{\"accountId\":\"56c5636befbe7cc23f5157c9278fca4e09109ffc\",\"tickets\":[{\"txid\":\"480110260b5af6c98fe6745b492aee94e335456ff8cadbbca3185af5d702b103\",\"index\":0}]}]},\"to\":{\"coins\":[{\"id\":\"0000000000000000000000000000000000000000\",\"value\":3000000},{\"id\":\"56c5636befbe7cc23f5157c9278fca4e09109ffc\",\"value\":264597000000}]},\"signs\":[{\"code\":\"IE7PZzk4FeVciaKg1VmUmrTG1J6m4+0uvBEn98VGERmSfdzGRAQZKDfkaY3UiZNLU0FX5kDk5viUuofu1DM/Ddg=\"}],\"data\":\"CAEQgMLXLxqoAwgBEIDC1y8angNggGBAUmBAUWAggGEBfoM5gQFgQJCBUpBRYAFgoGACCgMzFmAAkIFSYCCBkFKRkJEgVWEBPoBhAEBgADlgAPMAYIBgQFJgBDYQYQBLV2P/////fAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAYAA1BBZjcKCCMYEUYQBQV4BjqQWcuxRhAI5XW2AAgP1bNBVhAFtXYACA/VthAHxz//////////////////////////9gBDUWYQC/VltgQIBRkYJSUZCBkANgIAGQ81s0FWEAmVdgAID9W2EAvXP//////////////////////////2AENRZgJDVhANFWWwBbYABgIIGQUpCBUmBAkCBUgVZbc///////////////////////////M4EWYACQgVJgIIGQUmBAgIIggFSFkAOQVZOQkRaBUpGQkSCAVJCRAZBVVgChZWJ6enIwWCBCtIvUVjfd6ZAPiWDHeg/xs5yIFMX0opOqnEwVMKiFogApAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAASw=\"}"
	tx := Transaction{}
	if err := json.Unmarshal([]byte(txstr), &tx); err != nil {
		unittest.Error(t, err)
	}
	t.Log(tx)
}
