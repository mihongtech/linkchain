package meta

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/mihongtech/linkchain/common/btcec"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/serialize"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
)

type Ticket struct {
	Txid  TxID   `json:"txid"`
	Index uint32 `json:"index"`
}

func NewTicket(txid TxID, index uint32) *Ticket {
	return &Ticket{Txid: txid, Index: index}
}

func (t *Ticket) SetTxid(id TxID) {
	t.Txid = id
}

func (t *Ticket) GetTxid() *TxID {
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

func (fc *FromCoin) AddTicket(ticket *Ticket) {
	fc.Ticket = append(fc.Ticket, *ticket)
}

func (fc *FromCoin) GetTickets() []Ticket {
	tks := make([]Ticket, 0)
	for _, t := range fc.Ticket {
		tks = append(tks, t)
	}
	return tks
}

func (fc *FromCoin) SetId(id AccountID) {
	fc.Id = id
}

func (fc *FromCoin) GetId() AccountID {
	return fc.Id
}

//Serialize/Deserialize
func (fc *FromCoin) Serialize() serialize.SerializeStream {
	id := fc.Id.Serialize().(*protobuf.AccountID)

	ticket := make([]*protobuf.Ticket, 0)

	for _, c := range fc.Ticket {
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
	Id    AccountID `json:"id"`
	Value Amount    `json:"value"`
}

func NewToCoin(id AccountID, value *Amount) *ToCoin {
	return &ToCoin{Id: id, Value: *value}
}

func (tc *ToCoin) SetId(id AccountID) {
	tc.Id = id
}
func (tc *ToCoin) GetId() AccountID {
	return tc.Id
}

func (tc *ToCoin) SetValue(value *Amount) {
	tc.Value = *value
}

func (tc *ToCoin) GetValue() *Amount {
	return &tc.Value
}

func (tc *ToCoin) CheckValue() bool {
	return tc.Value.GetInt64() > 0
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
	tc.Id = AccountID{}
	if err := tc.Id.Deserialize(data.Id); err != nil {
		return err
	}

	tc.Value = *NewAmount(0)
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
	for index, _ := range tt.Coins {
		coins = append(coins, tt.Coins[index].Serialize().(*protobuf.ToCoin))
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

func NewSignature(code []byte) *Signature {
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

func (sign *Signature) Verify(hash []byte, address []byte) error {
	signer, err := btcec.GetSigner(hash, sign.Code)
	if err != nil {
		return err
	}
	id := NewAccountId(signer)

	if id.IsEqual(BytesToAccountID(address)) {
		return nil
	}
	return errors.New("Verify sign failed")
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

	txid TxID
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
	fromCoins := make([]FromCoin, 0)
	tf := *NewTransactionFrom(fromCoins)

	toCoins := make([]ToCoin, 0)
	tt := *NewTransactionTo(toCoins)

	signs := make([]Signature, 0)
	return NewTransaction(version, txtype, tf, tt, signs, nil)
}

func (tx *Transaction) GetTxID() *TxID {
	if tx.txid.IsEmpty() {
		s := tx.Serialize()
		err := tx.Deserialize(s)
		if err != nil {
			log.Error("Transaction", "GetTxID() error", err)
			return nil
		}
	}
	return &tx.txid
}

func (tx *Transaction) RebuildTxID() {
	s := tx.Serialize()
	err := tx.Deserialize(s)
	if err != nil {
		log.Error("Transaction", "GetTxID() error", err)
		return
	}
}

func (tx *Transaction) AddFromCoin(fromCoin ...FromCoin) {
	for i := range fromCoin {
		tx.From.AddFromCoin(fromCoin[i])
	}

}

func (tx *Transaction) AddToCoin(toCoin ...ToCoin) {
	for i := range toCoin {
		tx.To.AddToCoin(toCoin[i])
	}
}

func (tx *Transaction) SetTo(id AccountID, amount Amount) {
	index := -1
	for i := range tx.To.Coins {
		if tx.To.Coins[i].Id.IsEqual(id) {
			index = i
			break
		}
	}

	if index >= 0 {
		tx.To.Coins[index].Value.Addition(amount)
	} else {
		tc := *NewToCoin(id, &amount)
		tx.AddToCoin(tc)
	}
}

func (tx *Transaction) AddSignature(signature math.ISignature) {
	tx.Sign = append(tx.Sign, *signature.(*Signature))
}

func (tx *Transaction) GetFromCoins() []FromCoin {
	return tx.From.Coins
}

func (tx *Transaction) GetToCoins() []ToCoin {
	return tx.To.Coins
}

func (tx *Transaction) GetToValue() *Amount {
	sum := NewAmount(0)
	for _, tc := range tx.To.Coins {
		sum.Addition(tc.Value)
	}
	return sum
}

func (tx *Transaction) Verify() error {
	if len(tx.From.Coins) != len(tx.Sign) {
		return errors.New("tx from count must be equal to sign count in tx verify")
	}

	for index, sign := range tx.Sign {
		err := sign.Verify(tx.txid.CloneBytes(), tx.From.Coins[index].Id.CloneBytes())
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
func (tx *Transaction) GetNewFromCoins() []FromCoin {
	nfcs := make([]FromCoin, 0)
	for index, c := range tx.To.Coins {
		ticket := Ticket{}
		ticket.SetTxid(*tx.GetTxID())
		ticket.SetIndex(uint32(index))

		nfc := FromCoin{}
		nfc.SetId(c.Id)
		nfc.AddTicket(&ticket)

		nfcs = append(nfcs, nfc)
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

func (tx *Transaction) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.Transaction)
	tx.Version = *data.Version
	tx.Type = *data.Type

	if err := tx.From.Deserialize(data.From); err != nil {
		return err
	}

	if err := tx.To.Deserialize(data.To); err != nil {
		return err
	}

	tx.Sign = tx.Sign[:0]

	for _, cointent := range data.Sign {
		nSignatrue := Signature{}

		if err := nSignatrue.Deserialize(cointent); err != nil {
			return err
		}
		tx.Sign = append(tx.Sign, nSignatrue)
	}

	tx.Data = data.Data

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

	tx.txid = *MakeTxID(buffer)
	return nil
}

func (tx *Transaction) String() string {
	data, err := json.Marshal(tx)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func TxDifference(a, b []Transaction) (keep []Transaction) {
	keep = make([]Transaction, 0, len(a))

	remove := make(map[TxID]struct{})
	for _, tx := range b {
		remove[*tx.GetTxID()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[*tx.GetTxID()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}
