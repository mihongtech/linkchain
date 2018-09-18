// Code generated by protoc-gen-go. DO NOT EDIT.
// source: poabuf.proto

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

type POABlockHeader struct {
	Version              *uint32  `protobuf:"varint,1,req,name=version" json:"version,omitempty"`
	PrevHash             *Hash    `protobuf:"bytes,2,req,name=prevHash" json:"prevHash,omitempty"`
	MerkleRoot           *Hash    `protobuf:"bytes,3,req,name=merkleRoot" json:"merkleRoot,omitempty"`
	Time                 *int64   `protobuf:"varint,4,req,name=time" json:"time,omitempty"`
	Difficulty           *uint32  `protobuf:"varint,5,req,name=difficulty" json:"difficulty,omitempty"`
	Nounce               *uint32  `protobuf:"varint,6,req,name=nounce" json:"nounce,omitempty"`
	Height               *uint32  `protobuf:"varint,7,req,name=height" json:"height,omitempty"`
	Extra                []byte   `protobuf:"bytes,8,opt,name=extra" json:"extra,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *POABlockHeader) Reset()         { *m = POABlockHeader{} }
func (m *POABlockHeader) String() string { return proto.CompactTextString(m) }
func (*POABlockHeader) ProtoMessage()    {}
func (*POABlockHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_poabuf_faee9b36d5426c40, []int{0}
}
func (m *POABlockHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_POABlockHeader.Unmarshal(m, b)
}
func (m *POABlockHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_POABlockHeader.Marshal(b, m, deterministic)
}
func (dst *POABlockHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_POABlockHeader.Merge(dst, src)
}
func (m *POABlockHeader) XXX_Size() int {
	return xxx_messageInfo_POABlockHeader.Size(m)
}
func (m *POABlockHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_POABlockHeader.DiscardUnknown(m)
}

var xxx_messageInfo_POABlockHeader proto.InternalMessageInfo

func (m *POABlockHeader) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return 0
}

func (m *POABlockHeader) GetPrevHash() *Hash {
	if m != nil {
		return m.PrevHash
	}
	return nil
}

func (m *POABlockHeader) GetMerkleRoot() *Hash {
	if m != nil {
		return m.MerkleRoot
	}
	return nil
}

func (m *POABlockHeader) GetTime() int64 {
	if m != nil && m.Time != nil {
		return *m.Time
	}
	return 0
}

func (m *POABlockHeader) GetDifficulty() uint32 {
	if m != nil && m.Difficulty != nil {
		return *m.Difficulty
	}
	return 0
}

func (m *POABlockHeader) GetNounce() uint32 {
	if m != nil && m.Nounce != nil {
		return *m.Nounce
	}
	return 0
}

func (m *POABlockHeader) GetHeight() uint32 {
	if m != nil && m.Height != nil {
		return *m.Height
	}
	return 0
}

func (m *POABlockHeader) GetExtra() []byte {
	if m != nil {
		return m.Extra
	}
	return nil
}

type POATransactions struct {
	Txs                  []*POATransaction `protobuf:"bytes,1,rep,name=txs" json:"txs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *POATransactions) Reset()         { *m = POATransactions{} }
func (m *POATransactions) String() string { return proto.CompactTextString(m) }
func (*POATransactions) ProtoMessage()    {}
func (*POATransactions) Descriptor() ([]byte, []int) {
	return fileDescriptor_poabuf_faee9b36d5426c40, []int{1}
}
func (m *POATransactions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_POATransactions.Unmarshal(m, b)
}
func (m *POATransactions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_POATransactions.Marshal(b, m, deterministic)
}
func (dst *POATransactions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_POATransactions.Merge(dst, src)
}
func (m *POATransactions) XXX_Size() int {
	return xxx_messageInfo_POATransactions.Size(m)
}
func (m *POATransactions) XXX_DiscardUnknown() {
	xxx_messageInfo_POATransactions.DiscardUnknown(m)
}

var xxx_messageInfo_POATransactions proto.InternalMessageInfo

func (m *POATransactions) GetTxs() []*POATransaction {
	if m != nil {
		return m.Txs
	}
	return nil
}

type POABlock struct {
	Header               *POABlockHeader  `protobuf:"bytes,1,req,name=header" json:"header,omitempty"`
	TxList               *POATransactions `protobuf:"bytes,2,opt,name=txList" json:"txList,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *POABlock) Reset()         { *m = POABlock{} }
func (m *POABlock) String() string { return proto.CompactTextString(m) }
func (*POABlock) ProtoMessage()    {}
func (*POABlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_poabuf_faee9b36d5426c40, []int{2}
}
func (m *POABlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_POABlock.Unmarshal(m, b)
}
func (m *POABlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_POABlock.Marshal(b, m, deterministic)
}
func (dst *POABlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_POABlock.Merge(dst, src)
}
func (m *POABlock) XXX_Size() int {
	return xxx_messageInfo_POABlock.Size(m)
}
func (m *POABlock) XXX_DiscardUnknown() {
	xxx_messageInfo_POABlock.DiscardUnknown(m)
}

var xxx_messageInfo_POABlock proto.InternalMessageInfo

func (m *POABlock) GetHeader() *POABlockHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *POABlock) GetTxList() *POATransactions {
	if m != nil {
		return m.TxList
	}
	return nil
}

type POATransaction struct {
	Version              *uint32             `protobuf:"varint,1,req,name=version" json:"version,omitempty"`
	From                 *POATransactionPeer `protobuf:"bytes,2,req,name=from" json:"from,omitempty"`
	To                   *POATransactionPeer `protobuf:"bytes,3,req,name=to" json:"to,omitempty"`
	Amount               *POAAmount          `protobuf:"bytes,4,req,name=amount" json:"amount,omitempty"`
	Time                 *int64              `protobuf:"varint,5,req,name=time" json:"time,omitempty"`
	Nounce               *uint32             `protobuf:"varint,6,req,name=nounce" json:"nounce,omitempty"`
	Extra                []byte              `protobuf:"bytes,7,opt,name=extra" json:"extra,omitempty"`
	Sign                 []byte              `protobuf:"bytes,8,opt,name=sign" json:"sign,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *POATransaction) Reset()         { *m = POATransaction{} }
func (m *POATransaction) String() string { return proto.CompactTextString(m) }
func (*POATransaction) ProtoMessage()    {}
func (*POATransaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_poabuf_faee9b36d5426c40, []int{3}
}
func (m *POATransaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_POATransaction.Unmarshal(m, b)
}
func (m *POATransaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_POATransaction.Marshal(b, m, deterministic)
}
func (dst *POATransaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_POATransaction.Merge(dst, src)
}
func (m *POATransaction) XXX_Size() int {
	return xxx_messageInfo_POATransaction.Size(m)
}
func (m *POATransaction) XXX_DiscardUnknown() {
	xxx_messageInfo_POATransaction.DiscardUnknown(m)
}

var xxx_messageInfo_POATransaction proto.InternalMessageInfo

func (m *POATransaction) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return 0
}

func (m *POATransaction) GetFrom() *POATransactionPeer {
	if m != nil {
		return m.From
	}
	return nil
}

func (m *POATransaction) GetTo() *POATransactionPeer {
	if m != nil {
		return m.To
	}
	return nil
}

func (m *POATransaction) GetAmount() *POAAmount {
	if m != nil {
		return m.Amount
	}
	return nil
}

func (m *POATransaction) GetTime() int64 {
	if m != nil && m.Time != nil {
		return *m.Time
	}
	return 0
}

func (m *POATransaction) GetNounce() uint32 {
	if m != nil && m.Nounce != nil {
		return *m.Nounce
	}
	return 0
}

func (m *POATransaction) GetExtra() []byte {
	if m != nil {
		return m.Extra
	}
	return nil
}

func (m *POATransaction) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

type POATransactionPeer struct {
	AccountID            *POAAccountID `protobuf:"bytes,1,req,name=accountID" json:"accountID,omitempty"`
	Extra                []byte        `protobuf:"bytes,2,opt,name=extra" json:"extra,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *POATransactionPeer) Reset()         { *m = POATransactionPeer{} }
func (m *POATransactionPeer) String() string { return proto.CompactTextString(m) }
func (*POATransactionPeer) ProtoMessage()    {}
func (*POATransactionPeer) Descriptor() ([]byte, []int) {
	return fileDescriptor_poabuf_faee9b36d5426c40, []int{4}
}
func (m *POATransactionPeer) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_POATransactionPeer.Unmarshal(m, b)
}
func (m *POATransactionPeer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_POATransactionPeer.Marshal(b, m, deterministic)
}
func (dst *POATransactionPeer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_POATransactionPeer.Merge(dst, src)
}
func (m *POATransactionPeer) XXX_Size() int {
	return xxx_messageInfo_POATransactionPeer.Size(m)
}
func (m *POATransactionPeer) XXX_DiscardUnknown() {
	xxx_messageInfo_POATransactionPeer.DiscardUnknown(m)
}

var xxx_messageInfo_POATransactionPeer proto.InternalMessageInfo

func (m *POATransactionPeer) GetAccountID() *POAAccountID {
	if m != nil {
		return m.AccountID
	}
	return nil
}

func (m *POATransactionPeer) GetExtra() []byte {
	if m != nil {
		return m.Extra
	}
	return nil
}

type POAAccountID struct {
	Id                   []byte   `protobuf:"bytes,1,req,name=id" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *POAAccountID) Reset()         { *m = POAAccountID{} }
func (m *POAAccountID) String() string { return proto.CompactTextString(m) }
func (*POAAccountID) ProtoMessage()    {}
func (*POAAccountID) Descriptor() ([]byte, []int) {
	return fileDescriptor_poabuf_faee9b36d5426c40, []int{5}
}
func (m *POAAccountID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_POAAccountID.Unmarshal(m, b)
}
func (m *POAAccountID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_POAAccountID.Marshal(b, m, deterministic)
}
func (dst *POAAccountID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_POAAccountID.Merge(dst, src)
}
func (m *POAAccountID) XXX_Size() int {
	return xxx_messageInfo_POAAccountID.Size(m)
}
func (m *POAAccountID) XXX_DiscardUnknown() {
	xxx_messageInfo_POAAccountID.DiscardUnknown(m)
}

var xxx_messageInfo_POAAccountID proto.InternalMessageInfo

func (m *POAAccountID) GetId() []byte {
	if m != nil {
		return m.Id
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
	return fileDescriptor_poabuf_faee9b36d5426c40, []int{6}
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

type POAAmount struct {
	Value                *int32   `protobuf:"varint,1,req,name=value" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *POAAmount) Reset()         { *m = POAAmount{} }
func (m *POAAmount) String() string { return proto.CompactTextString(m) }
func (*POAAmount) ProtoMessage()    {}
func (*POAAmount) Descriptor() ([]byte, []int) {
	return fileDescriptor_poabuf_faee9b36d5426c40, []int{7}
}
func (m *POAAmount) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_POAAmount.Unmarshal(m, b)
}
func (m *POAAmount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_POAAmount.Marshal(b, m, deterministic)
}
func (dst *POAAmount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_POAAmount.Merge(dst, src)
}
func (m *POAAmount) XXX_Size() int {
	return xxx_messageInfo_POAAmount.Size(m)
}
func (m *POAAmount) XXX_DiscardUnknown() {
	xxx_messageInfo_POAAmount.DiscardUnknown(m)
}

var xxx_messageInfo_POAAmount proto.InternalMessageInfo

func (m *POAAmount) GetValue() int32 {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return 0
}

func init() {
	proto.RegisterType((*POABlockHeader)(nil), "protobuf.POABlockHeader")
	proto.RegisterType((*POATransactions)(nil), "protobuf.POATransactions")
	proto.RegisterType((*POABlock)(nil), "protobuf.POABlock")
	proto.RegisterType((*POATransaction)(nil), "protobuf.POATransaction")
	proto.RegisterType((*POATransactionPeer)(nil), "protobuf.POATransactionPeer")
	proto.RegisterType((*POAAccountID)(nil), "protobuf.POAAccountID")
	proto.RegisterType((*Hash)(nil), "protobuf.Hash")
	proto.RegisterType((*POAAmount)(nil), "protobuf.POAAmount")
}

func init() { proto.RegisterFile("poabuf.proto", fileDescriptor_poabuf_faee9b36d5426c40) }

var fileDescriptor_poabuf_faee9b36d5426c40 = []byte{
	// 444 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0xdd, 0x8a, 0x13, 0x31,
	0x14, 0xc7, 0x69, 0x3a, 0xfd, 0xd8, 0xd3, 0x5a, 0xe1, 0x28, 0x4b, 0x14, 0x59, 0xc6, 0xb9, 0x1a,
	0x56, 0x29, 0x6b, 0xf1, 0xd6, 0x8b, 0x8a, 0x17, 0x2b, 0x08, 0x5b, 0x82, 0x0f, 0x60, 0x76, 0x26,
	0xb3, 0x0d, 0xdb, 0x4e, 0x4a, 0x92, 0x29, 0xf5, 0xce, 0x17, 0xf0, 0x9d, 0x25, 0x99, 0xb4, 0x4d,
	0xd5, 0xee, 0xd5, 0xe4, 0x9c, 0xff, 0xff, 0x7c, 0x70, 0x7e, 0x03, 0xe3, 0x8d, 0xe2, 0xf7, 0x4d,
	0x35, 0xdd, 0x68, 0x65, 0x15, 0x0e, 0xfd, 0xe7, 0xbe, 0xa9, 0xb2, 0x5f, 0x04, 0x26, 0x8b, 0xbb,
	0xf9, 0xe7, 0x95, 0x2a, 0x1e, 0x6f, 0x05, 0x2f, 0x85, 0x46, 0x0a, 0x83, 0xad, 0xd0, 0x46, 0xaa,
	0x9a, 0x76, 0x52, 0x92, 0x3f, 0x63, 0xfb, 0x10, 0xaf, 0x61, 0xb8, 0xd1, 0x62, 0x7b, 0xcb, 0xcd,
	0x92, 0x92, 0x94, 0xe4, 0xa3, 0xd9, 0x64, 0xba, 0xef, 0x34, 0x75, 0x59, 0x76, 0xd0, 0x71, 0x0a,
	0xb0, 0x16, 0xfa, 0x71, 0x25, 0x98, 0x52, 0x96, 0x76, 0xff, 0xeb, 0x8e, 0x1c, 0x88, 0x90, 0x58,
	0xb9, 0x16, 0x34, 0x49, 0x49, 0xde, 0x65, 0xfe, 0x8d, 0x57, 0x00, 0xa5, 0xac, 0x2a, 0x59, 0x34,
	0x2b, 0xfb, 0x93, 0xf6, 0xfc, 0x32, 0x51, 0x06, 0x2f, 0xa1, 0x5f, 0xab, 0xa6, 0x2e, 0x04, 0xed,
	0x7b, 0x2d, 0x44, 0x2e, 0xbf, 0x14, 0xf2, 0x61, 0x69, 0xe9, 0xa0, 0xcd, 0xb7, 0x11, 0xbe, 0x84,
	0x9e, 0xd8, 0x59, 0xcd, 0xe9, 0x30, 0xed, 0xe4, 0x63, 0xd6, 0x06, 0xd9, 0x27, 0x78, 0xbe, 0xb8,
	0x9b, 0x7f, 0xd7, 0xbc, 0x36, 0xbc, 0xb0, 0x52, 0xd5, 0x06, 0xaf, 0xa1, 0x6b, 0x77, 0x86, 0x76,
	0xd2, 0x6e, 0x3e, 0x9a, 0xd1, 0xe3, 0xd6, 0xa7, 0x3e, 0xe6, 0x4c, 0x99, 0x82, 0xe1, 0xfe, 0x80,
	0x78, 0xe3, 0x06, 0xbb, 0x23, 0xfa, 0xcb, 0xfd, 0x5d, 0x1a, 0x1d, 0x99, 0x05, 0x1f, 0x7e, 0x80,
	0xbe, 0xdd, 0x7d, 0x93, 0xc6, 0x52, 0x92, 0x76, 0xf2, 0xd1, 0xec, 0xd5, 0xb9, 0x61, 0x86, 0x05,
	0x63, 0xf6, 0xbb, 0x45, 0x16, 0x69, 0x4f, 0x20, 0xbb, 0x81, 0xa4, 0xd2, 0x6a, 0x1d, 0x70, 0xbd,
	0x39, 0xd7, 0x7d, 0x21, 0x84, 0x66, 0xde, 0x89, 0xef, 0x81, 0x58, 0x15, 0x80, 0x3d, 0xed, 0x27,
	0x56, 0xe1, 0x3b, 0xe8, 0xf3, 0xb5, 0x6a, 0x6a, 0xeb, 0xc1, 0x8d, 0x66, 0x2f, 0x4e, 0x2a, 0xe6,
	0x5e, 0x62, 0xc1, 0x72, 0x60, 0xdc, 0x8b, 0x18, 0x9f, 0x63, 0x78, 0x60, 0x35, 0x88, 0x58, 0xb9,
	0x0e, 0x46, 0x3e, 0xd4, 0x01, 0xa0, 0x7f, 0x67, 0x3f, 0x00, 0xff, 0x5d, 0x0e, 0x3f, 0xc2, 0x05,
	0x2f, 0x0a, 0x37, 0xf6, 0xeb, 0x97, 0x40, 0xe3, 0xf2, 0x74, 0xb7, 0xbd, 0xca, 0x8e, 0xc6, 0xe3,
	0x54, 0x12, 0xff, 0x21, 0x57, 0x30, 0x8e, 0x0b, 0x70, 0x02, 0x44, 0x96, 0xbe, 0xe9, 0x98, 0x11,
	0x59, 0x66, 0xaf, 0x21, 0xf1, 0xff, 0x3c, 0x42, 0x52, 0x72, 0xcb, 0x83, 0xe2, 0xdf, 0xd9, 0x5b,
	0xb8, 0x38, 0x1c, 0xc2, 0xb5, 0xdf, 0xf2, 0x55, 0x23, 0xbc, 0xa3, 0xc7, 0xda, 0xe0, 0x4f, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x03, 0xcd, 0x28, 0xe6, 0x9c, 0x03, 0x00, 0x00,
}
