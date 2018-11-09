// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protobuf/transaction.proto

package protobuf

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Transactions struct {
	Txs                  []*Transaction `protobuf:"bytes,1,rep,name=txs" json:"txs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Transactions) Reset()         { *m = Transactions{} }
func (m *Transactions) String() string { return proto.CompactTextString(m) }
func (*Transactions) ProtoMessage()    {}
func (*Transactions) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_389ff9127b3b78a0, []int{0}
}
func (m *Transactions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transactions.Unmarshal(m, b)
}
func (m *Transactions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transactions.Marshal(b, m, deterministic)
}
func (dst *Transactions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transactions.Merge(dst, src)
}
func (m *Transactions) XXX_Size() int {
	return xxx_messageInfo_Transactions.Size(m)
}
func (m *Transactions) XXX_DiscardUnknown() {
	xxx_messageInfo_Transactions.DiscardUnknown(m)
}

var xxx_messageInfo_Transactions proto.InternalMessageInfo

func (m *Transactions) GetTxs() []*Transaction {
	if m != nil {
		return m.Txs
	}
	return nil
}

type Transaction struct {
	Version              *uint32          `protobuf:"varint,1,req,name=version" json:"version,omitempty"`
	Type                 *uint32          `protobuf:"varint,2,req,name=type" json:"type,omitempty"`
	From                 *TransactionFrom `protobuf:"bytes,3,req,name=from" json:"from,omitempty"`
	To                   *TransactionTo   `protobuf:"bytes,4,req,name=to" json:"to,omitempty"`
	Sign                 []*Signature     `protobuf:"bytes,5,rep,name=sign" json:"sign,omitempty"`
	Data                 []byte           `protobuf:"bytes,6,opt,name=data" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *Transaction) Reset()         { *m = Transaction{} }
func (m *Transaction) String() string { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()    {}
func (*Transaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_389ff9127b3b78a0, []int{1}
}
func (m *Transaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction.Unmarshal(m, b)
}
func (m *Transaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction.Marshal(b, m, deterministic)
}
func (dst *Transaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction.Merge(dst, src)
}
func (m *Transaction) XXX_Size() int {
	return xxx_messageInfo_Transaction.Size(m)
}
func (m *Transaction) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction proto.InternalMessageInfo

func (m *Transaction) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return 0
}

func (m *Transaction) GetType() uint32 {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return 0
}

func (m *Transaction) GetFrom() *TransactionFrom {
	if m != nil {
		return m.From
	}
	return nil
}

func (m *Transaction) GetTo() *TransactionTo {
	if m != nil {
		return m.To
	}
	return nil
}

func (m *Transaction) GetSign() []*Signature {
	if m != nil {
		return m.Sign
	}
	return nil
}

func (m *Transaction) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type FromCoin struct {
	Id                   *AccountID `protobuf:"bytes,1,req,name=id" json:"id,omitempty"`
	Ticket               []*Ticket  `protobuf:"bytes,2,rep,name=ticket" json:"ticket,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *FromCoin) Reset()         { *m = FromCoin{} }
func (m *FromCoin) String() string { return proto.CompactTextString(m) }
func (*FromCoin) ProtoMessage()    {}
func (*FromCoin) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_389ff9127b3b78a0, []int{2}
}
func (m *FromCoin) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FromCoin.Unmarshal(m, b)
}
func (m *FromCoin) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FromCoin.Marshal(b, m, deterministic)
}
func (dst *FromCoin) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FromCoin.Merge(dst, src)
}
func (m *FromCoin) XXX_Size() int {
	return xxx_messageInfo_FromCoin.Size(m)
}
func (m *FromCoin) XXX_DiscardUnknown() {
	xxx_messageInfo_FromCoin.DiscardUnknown(m)
}

var xxx_messageInfo_FromCoin proto.InternalMessageInfo

func (m *FromCoin) GetId() *AccountID {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *FromCoin) GetTicket() []*Ticket {
	if m != nil {
		return m.Ticket
	}
	return nil
}

type TransactionFrom struct {
	Coins                []*FromCoin `protobuf:"bytes,1,rep,name=coins" json:"coins,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *TransactionFrom) Reset()         { *m = TransactionFrom{} }
func (m *TransactionFrom) String() string { return proto.CompactTextString(m) }
func (*TransactionFrom) ProtoMessage()    {}
func (*TransactionFrom) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_389ff9127b3b78a0, []int{3}
}
func (m *TransactionFrom) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TransactionFrom.Unmarshal(m, b)
}
func (m *TransactionFrom) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TransactionFrom.Marshal(b, m, deterministic)
}
func (dst *TransactionFrom) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TransactionFrom.Merge(dst, src)
}
func (m *TransactionFrom) XXX_Size() int {
	return xxx_messageInfo_TransactionFrom.Size(m)
}
func (m *TransactionFrom) XXX_DiscardUnknown() {
	xxx_messageInfo_TransactionFrom.DiscardUnknown(m)
}

var xxx_messageInfo_TransactionFrom proto.InternalMessageInfo

func (m *TransactionFrom) GetCoins() []*FromCoin {
	if m != nil {
		return m.Coins
	}
	return nil
}

type ToCoin struct {
	Id                   *AccountID `protobuf:"bytes,1,req,name=id" json:"id,omitempty"`
	Value                []byte     `protobuf:"bytes,2,req,name=value" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *ToCoin) Reset()         { *m = ToCoin{} }
func (m *ToCoin) String() string { return proto.CompactTextString(m) }
func (*ToCoin) ProtoMessage()    {}
func (*ToCoin) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_389ff9127b3b78a0, []int{4}
}
func (m *ToCoin) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ToCoin.Unmarshal(m, b)
}
func (m *ToCoin) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ToCoin.Marshal(b, m, deterministic)
}
func (dst *ToCoin) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ToCoin.Merge(dst, src)
}
func (m *ToCoin) XXX_Size() int {
	return xxx_messageInfo_ToCoin.Size(m)
}
func (m *ToCoin) XXX_DiscardUnknown() {
	xxx_messageInfo_ToCoin.DiscardUnknown(m)
}

var xxx_messageInfo_ToCoin proto.InternalMessageInfo

func (m *ToCoin) GetId() *AccountID {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *ToCoin) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type TransactionTo struct {
	Coins                []*ToCoin `protobuf:"bytes,1,rep,name=coins" json:"coins,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *TransactionTo) Reset()         { *m = TransactionTo{} }
func (m *TransactionTo) String() string { return proto.CompactTextString(m) }
func (*TransactionTo) ProtoMessage()    {}
func (*TransactionTo) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_389ff9127b3b78a0, []int{5}
}
func (m *TransactionTo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TransactionTo.Unmarshal(m, b)
}
func (m *TransactionTo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TransactionTo.Marshal(b, m, deterministic)
}
func (dst *TransactionTo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TransactionTo.Merge(dst, src)
}
func (m *TransactionTo) XXX_Size() int {
	return xxx_messageInfo_TransactionTo.Size(m)
}
func (m *TransactionTo) XXX_DiscardUnknown() {
	xxx_messageInfo_TransactionTo.DiscardUnknown(m)
}

var xxx_messageInfo_TransactionTo proto.InternalMessageInfo

func (m *TransactionTo) GetCoins() []*ToCoin {
	if m != nil {
		return m.Coins
	}
	return nil
}

type Ticket struct {
	Txid                 *Hash    `protobuf:"bytes,1,req,name=txid" json:"txid,omitempty"`
	Index                *uint32  `protobuf:"varint,2,req,name=index" json:"index,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Ticket) Reset()         { *m = Ticket{} }
func (m *Ticket) String() string { return proto.CompactTextString(m) }
func (*Ticket) ProtoMessage()    {}
func (*Ticket) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_389ff9127b3b78a0, []int{6}
}
func (m *Ticket) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Ticket.Unmarshal(m, b)
}
func (m *Ticket) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Ticket.Marshal(b, m, deterministic)
}
func (dst *Ticket) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Ticket.Merge(dst, src)
}
func (m *Ticket) XXX_Size() int {
	return xxx_messageInfo_Ticket.Size(m)
}
func (m *Ticket) XXX_DiscardUnknown() {
	xxx_messageInfo_Ticket.DiscardUnknown(m)
}

var xxx_messageInfo_Ticket proto.InternalMessageInfo

func (m *Ticket) GetTxid() *Hash {
	if m != nil {
		return m.Txid
	}
	return nil
}

func (m *Ticket) GetIndex() uint32 {
	if m != nil && m.Index != nil {
		return *m.Index
	}
	return 0
}

type AccountID struct {
	Id                   []byte   `protobuf:"bytes,1,req,name=id" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccountID) Reset()         { *m = AccountID{} }
func (m *AccountID) String() string { return proto.CompactTextString(m) }
func (*AccountID) ProtoMessage()    {}
func (*AccountID) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_389ff9127b3b78a0, []int{7}
}
func (m *AccountID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountID.Unmarshal(m, b)
}
func (m *AccountID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountID.Marshal(b, m, deterministic)
}
func (dst *AccountID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountID.Merge(dst, src)
}
func (m *AccountID) XXX_Size() int {
	return xxx_messageInfo_AccountID.Size(m)
}
func (m *AccountID) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountID.DiscardUnknown(m)
}

var xxx_messageInfo_AccountID proto.InternalMessageInfo

func (m *AccountID) GetId() []byte {
	if m != nil {
		return m.Id
	}
	return nil
}

type Signature struct {
	Code                 []byte   `protobuf:"bytes,1,opt,name=code" json:"code,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Signature) Reset()         { *m = Signature{} }
func (m *Signature) String() string { return proto.CompactTextString(m) }
func (*Signature) ProtoMessage()    {}
func (*Signature) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_389ff9127b3b78a0, []int{8}
}
func (m *Signature) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Signature.Unmarshal(m, b)
}
func (m *Signature) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Signature.Marshal(b, m, deterministic)
}
func (dst *Signature) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Signature.Merge(dst, src)
}
func (m *Signature) XXX_Size() int {
	return xxx_messageInfo_Signature.Size(m)
}
func (m *Signature) XXX_DiscardUnknown() {
	xxx_messageInfo_Signature.DiscardUnknown(m)
}

var xxx_messageInfo_Signature proto.InternalMessageInfo

func (m *Signature) GetCode() []byte {
	if m != nil {
		return m.Code
	}
	return nil
}

type Hash struct {
	Data                 []byte   `protobuf:"bytes,1,req,name=data" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Hash) Reset()         { *m = Hash{} }
func (m *Hash) String() string { return proto.CompactTextString(m) }
func (*Hash) ProtoMessage()    {}
func (*Hash) Descriptor() ([]byte, []int) {
	return fileDescriptor_transaction_389ff9127b3b78a0, []int{9}
}
func (m *Hash) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Hash.Unmarshal(m, b)
}
func (m *Hash) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Hash.Marshal(b, m, deterministic)
}
func (dst *Hash) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Hash.Merge(dst, src)
}
func (m *Hash) XXX_Size() int {
	return xxx_messageInfo_Hash.Size(m)
}
func (m *Hash) XXX_DiscardUnknown() {
	xxx_messageInfo_Hash.DiscardUnknown(m)
}

var xxx_messageInfo_Hash proto.InternalMessageInfo

func (m *Hash) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*Transactions)(nil), "protobuf.Transactions")
	proto.RegisterType((*Transaction)(nil), "protobuf.Transaction")
	proto.RegisterType((*FromCoin)(nil), "protobuf.FromCoin")
	proto.RegisterType((*TransactionFrom)(nil), "protobuf.TransactionFrom")
	proto.RegisterType((*ToCoin)(nil), "protobuf.ToCoin")
	proto.RegisterType((*TransactionTo)(nil), "protobuf.TransactionTo")
	proto.RegisterType((*Ticket)(nil), "protobuf.Ticket")
	proto.RegisterType((*AccountID)(nil), "protobuf.AccountID")
	proto.RegisterType((*Signature)(nil), "protobuf.Signature")
	proto.RegisterType((*Hash)(nil), "protobuf.Hash")
}

func init() {
	proto.RegisterFile("protobuf/transaction.proto", fileDescriptor_transaction_389ff9127b3b78a0)
}

var fileDescriptor_transaction_389ff9127b3b78a0 = []byte{
	// 382 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x91, 0xcf, 0x6e, 0xda, 0x40,
	0x10, 0xc6, 0xe5, 0xc5, 0xb8, 0x30, 0x18, 0x5a, 0x6d, 0x5b, 0x75, 0x4b, 0x0f, 0xb5, 0xb6, 0x52,
	0xf1, 0x25, 0x44, 0xe2, 0xc2, 0x21, 0xa7, 0x84, 0x28, 0x4a, 0xae, 0x1b, 0x2e, 0x39, 0x3a, 0xb6,
	0x21, 0xab, 0x84, 0x1d, 0x64, 0xaf, 0x11, 0x79, 0xcb, 0x3c, 0x52, 0xb4, 0x6b, 0xfc, 0x87, 0x88,
	0x43, 0x6e, 0xb3, 0x33, 0xbf, 0xd9, 0xef, 0x9b, 0x19, 0x18, 0x6f, 0x33, 0xd4, 0xf8, 0x58, 0xac,
	0xce, 0x75, 0x16, 0xa9, 0x3c, 0x8a, 0xb5, 0x44, 0x35, 0xb5, 0x49, 0xda, 0xab, 0x6a, 0x7c, 0x0e,
	0xfe, 0xb2, 0x29, 0xe7, 0x74, 0x02, 0x1d, 0xbd, 0xcf, 0x99, 0x13, 0x74, 0xc2, 0xc1, 0xec, 0xe7,
	0xb4, 0xe2, 0xa6, 0x2d, 0x48, 0x18, 0x82, 0xbf, 0x39, 0x30, 0x68, 0x25, 0x29, 0x83, 0x2f, 0xbb,
	0x34, 0xcb, 0x25, 0x2a, 0xe6, 0x04, 0x24, 0x1c, 0x8a, 0xea, 0x49, 0x29, 0xb8, 0xfa, 0x75, 0x9b,
	0x32, 0x62, 0xd3, 0x36, 0xa6, 0x67, 0xe0, 0xae, 0x32, 0xdc, 0xb0, 0x4e, 0x40, 0xc2, 0xc1, 0xec,
	0xf7, 0x49, 0x9d, 0x9b, 0x0c, 0x37, 0xc2, 0x62, 0x74, 0x02, 0x44, 0x23, 0x73, 0x2d, 0xfc, 0xeb,
	0x24, 0xbc, 0x44, 0x41, 0x34, 0xd2, 0x09, 0xb8, 0xb9, 0x5c, 0x2b, 0xd6, 0xb5, 0xfe, 0xbf, 0x37,
	0xe8, 0xbd, 0x5c, 0xab, 0x48, 0x17, 0x59, 0x2a, 0x2c, 0x60, 0x4c, 0x25, 0x91, 0x8e, 0x98, 0x17,
	0x38, 0xa1, 0x2f, 0x6c, 0xcc, 0x1f, 0xa0, 0x67, 0x34, 0x17, 0x28, 0x15, 0xfd, 0x07, 0x44, 0x26,
	0x76, 0x92, 0xa3, 0x6f, 0x2e, 0xe3, 0x18, 0x0b, 0xa5, 0xef, 0xae, 0x05, 0x91, 0x09, 0x0d, 0xc1,
	0xd3, 0x32, 0x7e, 0x4e, 0x35, 0x23, 0x56, 0xef, 0x5b, 0xcb, 0x9a, 0xcd, 0x8b, 0x43, 0x9d, 0x5f,
	0xc0, 0xd7, 0x0f, 0x93, 0xd1, 0x10, 0xba, 0x31, 0x4a, 0x55, 0xed, 0x9a, 0x36, 0xbd, 0x95, 0x09,
	0x51, 0x02, 0x7c, 0x01, 0xde, 0x12, 0x3f, 0xef, 0xea, 0x07, 0x74, 0x77, 0xd1, 0x4b, 0x51, 0x2e,
	0xdc, 0x17, 0xe5, 0x83, 0xcf, 0x61, 0x78, 0xb4, 0x2e, 0xfa, 0xff, 0x58, 0xbf, 0xed, 0x1d, 0xdb,
	0xea, 0x57, 0xe0, 0x95, 0xc3, 0x50, 0x0e, 0xae, 0xde, 0xd7, 0xfa, 0xa3, 0xa6, 0xe1, 0x36, 0xca,
	0x9f, 0x84, 0xad, 0x19, 0x71, 0xa9, 0x92, 0x74, 0x7f, 0xb8, 0x76, 0xf9, 0xe0, 0x7f, 0xa0, 0x5f,
	0x7b, 0xa4, 0xa3, 0x7a, 0x08, 0xdf, 0xf8, 0xe5, 0x7f, 0xa1, 0x5f, 0x5f, 0xc7, 0xdc, 0x25, 0xc6,
	0x24, 0x65, 0x4e, 0x79, 0x17, 0x13, 0xf3, 0x31, 0xb8, 0x46, 0xa1, 0xbe, 0x59, 0xd9, 0x6a, 0xe3,
	0xf7, 0x00, 0x00, 0x00, 0xff, 0xff, 0x3a, 0x1c, 0xf3, 0x8b, 0xe6, 0x02, 0x00, 0x00,
}
