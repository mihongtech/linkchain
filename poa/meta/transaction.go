package poameta

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/linkchain/common/btcec"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/amount"
	"github.com/linkchain/meta/coin"
	"github.com/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
	"github.com/linkchain/common/math"
)

type Ticket struct {
	Txid  meta.TxID `json:"txid"`
	Index uint32    `json:"index"`
}

func NewTicket(txid meta.TxID, index uint32) *Ticket {
	return &Ticket{Txid: txid, Index: index}
}

func (t *Ticket) SetTxid(id meta.TxID) {
	t.Txid = id
}

func (t *Ticket) GetTxid() *meta.TxID {
	return &t.Txid
}

func (t *Ticket) SetIndex(index uint32) {
	t.Index = index
}
func (t *Ticket) GetIndex() uint32 {
	return t.Index
}

//Serialize/Deserialize
func (t *Ticket) Serialize() serialize.SerializeStream {
	txid := t.Txid.Serialize().(*protobuf.Hash)
	ticket := protobuf.Ticket{
		Txid:  txid,
		Index: proto.Uint32(t.Index),
	}
	return &ticket
}

func (t *Ticket) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.Ticket)
	err := t.Txid.Deserialize(data.Txid)
	if err != nil {
		return err
	}
	t.Index = *data.Index
	return nil
}

func (t *Ticket) String() string {
	data, err := json.Marshal(t)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

type FromCoin struct {
	Id     AccountID `json:"accountId"`
	Ticket []Ticket  `json:"tickets"`
}

func NewFromCoin(id AccountID, ticket []Ticket) *FromCoin {
	return &FromCoin{Id: id, Ticket: ticket}
}

func (tc *FromCoin) AddTicket(ticket coin.ITicket) {
	tc.Ticket = append(tc.Ticket, *ticket.(*Ticket))
}

func (tc *FromCoin) GetTickets() []coin.ITicket {
	tks := make([]coin.ITicket, 0)
	for _, t := range tc.Ticket {
		tks = append(tks, &t)
	}
	return tks
}

func (tc *FromCoin) SetId(id meta.IAccountID) {
	tc.Id = *id.(*AccountID)
}

func (tc *FromCoin) GetId() meta.IAccountID {
	return &tc.Id
}

//Serialize/Deserialize
func (tc *FromCoin) Serialize() serialize.SerializeStream {
	id := tc.Id.Serialize().(*protobuf.AccountID)

	ticket := make([]*protobuf.Ticket, 0)

	for _, c := range tc.Ticket {
		ticket = append(ticket, c.Serialize().(*protobuf.Ticket))
	}

	peer := protobuf.FromCoin{
		Id:     id,
		Ticket: ticket,
	}
	return &peer
}

func (fc *FromCoin) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.FromCoin)
	err := fc.Id.Deserialize(data.Id)
	if err != nil {
		return err
	}

	fc.Ticket = fc.Ticket[:0]
	for _, ticket := range data.Ticket {
		nticket := Ticket{}
		err := nticket.Deserialize(ticket)
		if err != nil {
			return err
		}
		fc.Ticket = append(fc.Ticket, nticket)
	}
	return nil
}

func (fc *FromCoin) String() string {
	data, err := json.Marshal(fc)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

type TransactionFrom struct {
	Coins []FromCoin `json:"coins"`
}

func NewTransactionFrom(coin []FromCoin) *TransactionFrom {
	return &TransactionFrom{Coins: coin}
}

func (tf *TransactionFrom) AddFromCoin(coin FromCoin) {
	tf.Coins = append(tf.Coins, coin)
}

//Serialize/Deserialize
func (tf *TransactionFrom) Serialize() serialize.SerializeStream {

	coin := make([]*protobuf.FromCoin, 0)

	for _, c := range tf.Coins {
		coin = append(coin, c.Serialize().(*protobuf.FromCoin))
	}

	peer := protobuf.TransactionFrom{
		Coins: coin,
	}
	return &peer
}

func (tf *TransactionFrom) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.TransactionFrom)
	tf.Coins = tf.Coins[:0]
	for _, coin := range data.Coins {
		nCoin := FromCoin{}
		err := nCoin.Deserialize(coin)
		if err != nil {
			return err
		}
		tf.Coins = append(tf.Coins, nCoin)
	}
	return nil
}

func (tf *TransactionFrom) String() string {
	data, err := json.Marshal(tf)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

type ToCoin struct {
	Id    AccountID     `json:"id"`
	Value amount.Amount `json:"value"`
}

func NewToCoin(id AccountID, value *amount.Amount) *ToCoin {
	return &ToCoin{Id: id, Value: *value}
}

func (tc *ToCoin) SetId(id meta.IAccountID) {
	tc.Id = *id.(*AccountID)
}
func (tc *ToCoin) GetId() meta.IAccountID {
	return &tc.Id
}

func (tc *ToCoin) SetValue(value *amount.Amount) {
	tc.Value = *value
}
func (tc *ToCoin) GetValue() *amount.Amount {
	return &tc.Value
}

//Serialize/Deserialize
func (tc *ToCoin) Serialize() serialize.SerializeStream {
	peer := &protobuf.ToCoin{
		Id:    tc.Id.Serialize().(*protobuf.AccountID),
		Value: proto.NewBuffer(tc.Value.GetBytes()).Bytes(),
	}
	return peer
}

func (tc *ToCoin) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.ToCoin)
	if err := tc.Id.Deserialize(data.Id); err != nil {
		return err
	}
	tc.Value = *amount.NewAmount(0)
	tc.Value.SetBytes(data.Value)
	return nil
}

func (tc *ToCoin) String() string {
	data, err := json.Marshal(tc)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

type TransactionTo struct {
	Coins []ToCoin `json:"coins"`
}

func NewTransactionTo(coins []ToCoin) *TransactionTo {
	return &TransactionTo{Coins: coins}
}

func (tt *TransactionTo) AddToCoin(coin ToCoin) {
	tt.Coins = append(tt.Coins, coin)
}

//Serialize/Deserialize
func (tt *TransactionTo) Serialize() serialize.SerializeStream {
	coins := make([]*protobuf.ToCoin, 0)
	for _, c := range tt.Coins {
		coins = append(coins, c.Serialize().(*protobuf.ToCoin))
	}

	peer := protobuf.TransactionTo{
		Coins: coins,
	}
	return &peer
}

func (tt *TransactionTo) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.TransactionTo)

	tt.Coins = tt.Coins[:0]
	for _, c := range data.Coins {
		nCoin := ToCoin{}
		err := nCoin.Deserialize(c)
		if err != nil {
			return err
		}
		tt.Coins = append(tt.Coins, nCoin)
	}

	return nil
}

func (tt *TransactionTo) String() string {
	data, err := json.Marshal(tt)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

type Signature struct {
	Code []byte `json:"code"`
}

func NewSignatrue(code []byte) *Signature {
	return &Signature{Code: code}
}

//Serialize/Deserialize
func (sign *Signature) Serialize() serialize.SerializeStream {
	peer := protobuf.Signature{
		Code: proto.NewBuffer(sign.Code).Bytes(),
	}
	return &peer
}

func (sign *Signature) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.Signature)
	sign.Code = data.Code
	return nil
}

func (sign *Signature) String() string {
	return hex.EncodeToString(sign.Code)
}

func (sign *Signature) Verify(hash []byte, pubKey []byte) error {
	signature, err := btcec.ParseSignature(sign.Code, btcec.S256())
	if err != nil {
		return err
	}

	pk, err := btcec.ParsePubKey(pubKey, btcec.S256())
	if err != nil {
		return errors.New("Transaction VerifySign ParsePubKey is error")
	}

	verified := signature.Verify(hash, pk)
	if verified {
		return nil
	} else {
		return errors.New("Transaction VerifySign failed: Error Sign")
	}
}

type Transaction struct {
	// The version of the Transaction.  This is not the same as the Blocks version.
	Version uint32 `json:"version"`

	// The type of the Transaction.
	Type uint32 `json:"type"`

	//The accounts of the Transaction related to inputs.
	From TransactionFrom `json:"from"`

	//The accounts of the Transaction related to outputs.
	To TransactionTo `json:"to"`

	//The Sign of From, which is represent the Coins each Froms if not can put.
	Sign []Signature `json:"signs"`

	//The extra feild of Transaction.
	Data []byte `json:"data"`

	txid meta.TxID
}

func NewTransaction(version uint32, txtype uint32, from TransactionFrom, to TransactionTo, sign []Signature, data []byte) *Transaction {
	return &Transaction{
		Version: version,
		Type:    txtype,
		From:    from,
		To:      to,
		Sign:    sign,
		Data:    data,
	}
}

func NewEmptyTransaction(version uint32, txtype uint32) *Transaction {
	fromcoins := make([]FromCoin, 0)
	tf := *NewTransactionFrom(fromcoins)

	tocoins := make([]ToCoin, 0)
	tt := *NewTransactionTo(tocoins)
	return NewTransaction(version, txtype, tf, tt, nil, nil)
}

func (tx *Transaction) GetTxID() *meta.TxID {
	if tx.txid.IsEmpty() {
		err := tx.Deserialize(tx.Serialize())
		if err != nil {
			log.Error("Transaction", "GetTxID() error", err)
			return nil
		}
	}
	return &tx.txid
}

func (tx *Transaction) AddFromCoin(fromCoin coin.IFromCoin) {
	tx.From.AddFromCoin(*fromCoin.(*FromCoin))
}

func (tx *Transaction) AddToCoin(toCoin coin.IToCoin) {
	tx.To.AddToCoin(*toCoin.(*ToCoin))
}

func (tx *Transaction) AddSignature(signature math.ISignature) {
	tx.Sign = append(tx.Sign, *signature.(*Signature))
}

func (t *Transaction) GetFromCoins() []coin.IFromCoin {
	fcs := make([]coin.IFromCoin, 0)
	for index, _ := range t.From.Coins {
		fcs = append(fcs, &t.From.Coins[index])
	}
	return fcs
}

func (t *Transaction) GetToCoins() []coin.IToCoin {
	tcs := make([]coin.IToCoin, 0)
	for index, _ := range t.To.Coins {
		tcs = append(tcs, &t.To.Coins[index])
	}
	return tcs
}

func (t *Transaction) GetToValue() *amount.Amount {
	sum := amount.NewAmount(0)
	for _, tc := range t.To.Coins {
		sum.Addition(tc.Value)
	}
	return sum
}

func (tx *Transaction) Verify() error {
	for index, sign := range tx.Sign {
		err := sign.Verify(tx.txid.CloneBytes(), tx.From.Coins[index].Id.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tx *Transaction) GetVersion() uint32 {
	return tx.Version
}

func (tx *Transaction) GetType() uint32 {
	return tx.Type
}

// the method is prepared for create new fromcoin from tx
func (t *Transaction) GetNewFromCoins() []coin.IFromCoin {
	nfcs := make([]coin.IFromCoin, 0)
	for index, c := range t.To.Coins {
		ticket := Ticket{}
		ticket.SetTxid(*t.GetTxID())
		ticket.SetIndex(uint32(index))

		nfc := FromCoin{}
		nfc.SetId(&c.Id)
		nfc.AddTicket(&ticket)

		nfcs = append(nfcs, &nfc)
	}

	return nfcs
}

//Serialize/Deserialize
func (tx *Transaction) Serialize() serialize.SerializeStream {
	from := tx.From.Serialize().(*protobuf.TransactionFrom)
	to := tx.To.Serialize().(*protobuf.TransactionTo)

	signature := make([]*protobuf.Signature, 0)

	for _, content := range tx.Sign {
		signature = append(signature, content.Serialize().(*protobuf.Signature))
	}

	t := protobuf.Transaction{
		Version: proto.Uint32(tx.Version),
		Type:    proto.Uint32(tx.Type),
		From:    from,
		To:      to,
		Sign:    signature,
		Data:    proto.NewBuffer(tx.Data).Bytes(),
	}
	return &t
}

func (t *Transaction) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.Transaction)
	t.Version = *data.Version
	t.Type = *data.Type

	if err := t.From.Deserialize(data.From); err != nil {
		return err
	}

	if err := t.To.Deserialize(data.To); err != nil {
		return err
	}

	t.Sign = t.Sign[:0]

	for _, cointent := range data.Sign {
		nSignatrue := Signature{}

		if err := nSignatrue.Deserialize(cointent); err != nil {
			return err
		}
		t.Sign = append(t.Sign, nSignatrue)
	}

	t.Data = data.Data

	pt := protobuf.Transaction{
		Version: data.Version,
		Type:    data.Type,
		From:    data.From,
		To:      data.To,
		Data:    data.Data,
	}
	buffer, err := proto.Marshal(&pt)
	if err != nil {
		return err
	}

	t.txid = *meta.MakeTxID(buffer)
	return nil
}

func (tx *Transaction) String() string {
	data, err := json.Marshal(tx)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
