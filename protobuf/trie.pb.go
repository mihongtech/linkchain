// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protobuf/trie.proto

package protobuf

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type HashNode struct {
	Data                 []byte   `protobuf:"bytes,1,req,name=data" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HashNode) Reset()         { *m = HashNode{} }
func (m *HashNode) String() string { return proto.CompactTextString(m) }
func (*HashNode) ProtoMessage()    {}
func (*HashNode) Descriptor() ([]byte, []int) {
	return fileDescriptor_77173f6eb0822a18, []int{0}
}

func (m *HashNode) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HashNode.Unmarshal(m, b)
}
func (m *HashNode) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HashNode.Marshal(b, m, deterministic)
}
func (m *HashNode) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HashNode.Merge(m, src)
}
func (m *HashNode) XXX_Size() int {
	return xxx_messageInfo_HashNode.Size(m)
}
func (m *HashNode) XXX_DiscardUnknown() {
	xxx_messageInfo_HashNode.DiscardUnknown(m)
}

var xxx_messageInfo_HashNode proto.InternalMessageInfo

func (m *HashNode) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type ValueNode struct {
	Data                 []byte   `protobuf:"bytes,1,req,name=data" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValueNode) Reset()         { *m = ValueNode{} }
func (m *ValueNode) String() string { return proto.CompactTextString(m) }
func (*ValueNode) ProtoMessage()    {}
func (*ValueNode) Descriptor() ([]byte, []int) {
	return fileDescriptor_77173f6eb0822a18, []int{1}
}

func (m *ValueNode) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValueNode.Unmarshal(m, b)
}
func (m *ValueNode) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValueNode.Marshal(b, m, deterministic)
}
func (m *ValueNode) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValueNode.Merge(m, src)
}
func (m *ValueNode) XXX_Size() int {
	return xxx_messageInfo_ValueNode.Size(m)
}
func (m *ValueNode) XXX_DiscardUnknown() {
	xxx_messageInfo_ValueNode.DiscardUnknown(m)
}

var xxx_messageInfo_ValueNode proto.InternalMessageInfo

func (m *ValueNode) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type NodeFlag struct {
	Hash                 *HashNode `protobuf:"bytes,1,req,name=hash" json:"hash,omitempty"`
	Gen                  *uint32   `protobuf:"varint,2,req,name=gen" json:"gen,omitempty"`
	Dirty                *bool     `protobuf:"varint,3,req,name=dirty" json:"dirty,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *NodeFlag) Reset()         { *m = NodeFlag{} }
func (m *NodeFlag) String() string { return proto.CompactTextString(m) }
func (*NodeFlag) ProtoMessage()    {}
func (*NodeFlag) Descriptor() ([]byte, []int) {
	return fileDescriptor_77173f6eb0822a18, []int{2}
}

func (m *NodeFlag) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeFlag.Unmarshal(m, b)
}
func (m *NodeFlag) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeFlag.Marshal(b, m, deterministic)
}
func (m *NodeFlag) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeFlag.Merge(m, src)
}
func (m *NodeFlag) XXX_Size() int {
	return xxx_messageInfo_NodeFlag.Size(m)
}
func (m *NodeFlag) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeFlag.DiscardUnknown(m)
}

var xxx_messageInfo_NodeFlag proto.InternalMessageInfo

func (m *NodeFlag) GetHash() *HashNode {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *NodeFlag) GetGen() uint32 {
	if m != nil && m.Gen != nil {
		return *m.Gen
	}
	return 0
}

func (m *NodeFlag) GetDirty() bool {
	if m != nil && m.Dirty != nil {
		return *m.Dirty
	}
	return false
}

type FullNode struct {
	Children             []*HashNode `protobuf:"bytes,1,rep,name=children" json:"children,omitempty"`
	Flags                *NodeFlag   `protobuf:"bytes,2,req,name=flags" json:"flags,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *FullNode) Reset()         { *m = FullNode{} }
func (m *FullNode) String() string { return proto.CompactTextString(m) }
func (*FullNode) ProtoMessage()    {}
func (*FullNode) Descriptor() ([]byte, []int) {
	return fileDescriptor_77173f6eb0822a18, []int{3}
}

func (m *FullNode) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FullNode.Unmarshal(m, b)
}
func (m *FullNode) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FullNode.Marshal(b, m, deterministic)
}
func (m *FullNode) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FullNode.Merge(m, src)
}
func (m *FullNode) XXX_Size() int {
	return xxx_messageInfo_FullNode.Size(m)
}
func (m *FullNode) XXX_DiscardUnknown() {
	xxx_messageInfo_FullNode.DiscardUnknown(m)
}

var xxx_messageInfo_FullNode proto.InternalMessageInfo

func (m *FullNode) GetChildren() []*HashNode {
	if m != nil {
		return m.Children
	}
	return nil
}

func (m *FullNode) GetFlags() *NodeFlag {
	if m != nil {
		return m.Flags
	}
	return nil
}

type ShortNode struct {
	Key                  []byte    `protobuf:"bytes,1,req,name=key" json:"key,omitempty"`
	Val                  *HashNode `protobuf:"bytes,2,req,name=val" json:"val,omitempty"`
	Flags                *NodeFlag `protobuf:"bytes,3,req,name=flags" json:"flags,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *ShortNode) Reset()         { *m = ShortNode{} }
func (m *ShortNode) String() string { return proto.CompactTextString(m) }
func (*ShortNode) ProtoMessage()    {}
func (*ShortNode) Descriptor() ([]byte, []int) {
	return fileDescriptor_77173f6eb0822a18, []int{4}
}

func (m *ShortNode) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ShortNode.Unmarshal(m, b)
}
func (m *ShortNode) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ShortNode.Marshal(b, m, deterministic)
}
func (m *ShortNode) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ShortNode.Merge(m, src)
}
func (m *ShortNode) XXX_Size() int {
	return xxx_messageInfo_ShortNode.Size(m)
}
func (m *ShortNode) XXX_DiscardUnknown() {
	xxx_messageInfo_ShortNode.DiscardUnknown(m)
}

var xxx_messageInfo_ShortNode proto.InternalMessageInfo

func (m *ShortNode) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *ShortNode) GetVal() *HashNode {
	if m != nil {
		return m.Val
	}
	return nil
}

func (m *ShortNode) GetFlags() *NodeFlag {
	if m != nil {
		return m.Flags
	}
	return nil
}

func init() {
	proto.RegisterType((*HashNode)(nil), "protobuf.HashNode")
	proto.RegisterType((*ValueNode)(nil), "protobuf.ValueNode")
	proto.RegisterType((*NodeFlag)(nil), "protobuf.NodeFlag")
	proto.RegisterType((*FullNode)(nil), "protobuf.FullNode")
	proto.RegisterType((*ShortNode)(nil), "protobuf.ShortNode")
}

func init() { proto.RegisterFile("protobuf/trie.proto", fileDescriptor_77173f6eb0822a18) }

var fileDescriptor_77173f6eb0822a18 = []byte{
	// 231 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x8f, 0xc1, 0x4a, 0x03, 0x31,
	0x10, 0x86, 0xe9, 0xa6, 0x85, 0x74, 0x54, 0x28, 0xa3, 0x87, 0x3d, 0xe9, 0x12, 0x44, 0x72, 0x8a,
	0xe0, 0x43, 0x14, 0x4f, 0x1e, 0x22, 0x78, 0xf0, 0x36, 0x9a, 0x74, 0xb3, 0x18, 0x1b, 0xc9, 0x66,
	0x85, 0xbe, 0xbd, 0x24, 0x35, 0x15, 0x91, 0xe2, 0xed, 0x9f, 0xfc, 0x1f, 0xf3, 0x65, 0xe0, 0xfc,
	0x23, 0x86, 0x14, 0x5e, 0xa6, 0xcd, 0x6d, 0x8a, 0x83, 0x55, 0x65, 0x42, 0x5e, 0x1f, 0xc5, 0x25,
	0xf0, 0x7b, 0x1a, 0xdd, 0x43, 0x30, 0x16, 0x11, 0xe6, 0x86, 0x12, 0xb5, 0xb3, 0xae, 0x91, 0xa7,
	0xba, 0x64, 0x71, 0x05, 0xcb, 0x27, 0xf2, 0x93, 0x3d, 0x0a, 0x3c, 0x03, 0xcf, 0xdd, 0xda, 0x53,
	0x8f, 0x37, 0x30, 0x77, 0x34, 0xba, 0xd2, 0x9f, 0xdc, 0xa1, 0xaa, 0x16, 0x55, 0x15, 0xba, 0xf4,
	0xb8, 0x02, 0xd6, 0xdb, 0x6d, 0xdb, 0x74, 0x8d, 0x3c, 0xd3, 0x39, 0xe2, 0x05, 0x2c, 0xcc, 0x10,
	0xd3, 0xae, 0x65, 0x5d, 0x23, 0xb9, 0xde, 0x0f, 0xc2, 0x00, 0x5f, 0x4f, 0xde, 0x17, 0xb7, 0x02,
	0xfe, 0xea, 0x06, 0x6f, 0xa2, 0xdd, 0xb6, 0xb3, 0x8e, 0x1d, 0xd9, 0x7f, 0x60, 0x50, 0xc2, 0x62,
	0xe3, 0xa9, 0x1f, 0x8b, 0xe5, 0x17, 0x5c, 0xbf, 0xab, 0xf7, 0x80, 0x78, 0x87, 0xe5, 0xa3, 0x0b,
	0x31, 0x15, 0xcd, 0x0a, 0xd8, 0x9b, 0xdd, 0x7d, 0x5f, 0x98, 0x23, 0x5e, 0x03, 0xfb, 0x24, 0xff,
	0x77, 0xcd, 0xc1, 0x99, 0xeb, 0x1f, 0x1d, 0xfb, 0x47, 0xf7, 0x15, 0x00, 0x00, 0xff, 0xff, 0x82,
	0x50, 0xe6, 0x30, 0x91, 0x01, 0x00, 0x00,
}
