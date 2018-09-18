package full

import (
	_ "fmt"
	_ "io"
	"math/big"

	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/sync/full/protobufmsg"
)

// Constants to match up protocol versions and messages
const (
	full01 = 1
)

// Official short name of the protocol used during capability negotiation.
var ProtocolName = "full"

// Supported versions of the eth protocol (first is primary).
var ProtocolVersions = []uint64{full01}

// Number of implemented message corresponding to different protocol versions.
var ProtocolLengths = []uint64{8}

const ProtocolMaxMsgSize = 10 * 1024 * 1024 // Maximum cap on the size of a protocol message

// eth protocol message codes
const (
	// Protocol messages belonging to eth/62
	StatusMsg         = 0x00
	NewBlockHashesMsg = 0x01
	TxMsg             = 0x02
	GetBlockMsg       = 0x03
	BlockMsg          = 0x04
	NewBlockMsg       = 0x05
)

type errCode int

const (
	ErrMsgTooLarge = iota
	ErrDecode
	ErrInvalidMsgCode
	ErrProtocolVersionMismatch
	ErrNetworkIdMismatch
	ErrGenesisBlockMismatch
	ErrNoStatusMsg
	ErrExtraStatusMsg
	ErrSuspendedPeer
)

func (e errCode) String() string {
	return errorToString[int(e)]
}

// XXX change once legacy code is out
var errorToString = map[int]string{
	ErrMsgTooLarge:             "Message too long",
	ErrDecode:                  "Invalid message",
	ErrInvalidMsgCode:          "Invalid message code",
	ErrProtocolVersionMismatch: "Protocol version mismatch",
	ErrNetworkIdMismatch:       "NetworkId mismatch",
	ErrGenesisBlockMismatch:    "Genesis block mismatch",
	ErrNoStatusMsg:             "No status message",
	ErrExtraStatusMsg:          "Extra status message",
	ErrSuspendedPeer:           "Suspended peer",
}

//type txPool interface {
//	// AddRemotes should add the given transactions to the pool.
//	AddRemotes([]*types.Transaction) []error
//
//	// Pending should return pending transactions.
//	// The slice should be modifiable by the caller.
//	Pending() (map[common.Address]types.Transactions, error)
//
//	// SubscribeTxPreEvent should return an event subscription of
//	// TxPreEvent and send events to the given channel.
//	SubscribeTxPreEvent(chan<- core.TxPreEvent) event.Subscription
//}
//
// statusData is the network packet for the status message.
type statusData struct {
	ProtocolVersion uint32
	NetworkId       uint64
	TD              *big.Int
	CurrentBlock    meta.DataID
	GenesisBlock    meta.DataID
}

func (s *statusData) Serialize() serialize.SerializeStream {
	td := s.TD.Int64()
	currentBlock := s.CurrentBlock.Serialize().(*protobufmsg.Hash)
	genesisBlock := s.GenesisBlock.Serialize().(*protobufmsg.Hash)
	status := &protobufmsg.StatusData{
		ProtocolVersion: &s.ProtocolVersion,
		NetworkId:       &s.NetworkId,
		Td:              &td,
		CurrentBlock:    currentBlock,
		GenesisBlock:    genesisBlock,
	}

	return status
}

func (s *statusData) Deserialize(data serialize.SerializeStream) {
	d := data.(*protobufmsg.StatusData)
	s.ProtocolVersion = *d.ProtocolVersion
	s.NetworkId = *d.NetworkId
	s.TD.SetInt64(*d.Td)
	s.GenesisBlock.Deserialize(d.GenesisBlock)
	s.CurrentBlock.Deserialize(d.CurrentBlock)
}

// newBlockHashesData is the network packet for the block announcements.
type newBlockHashesData []newBlockHashData

type newBlockHashData struct {
	Hash   meta.DataID // Hash of one particular block being announced
	Number uint64      // Number of one particular block being announced
}

func (n *newBlockHashData) Serialize() serialize.SerializeStream {
	data := &protobufmsg.NewBlockHashData{
		Hash:   n.Hash.Serialize().(*protobufmsg.Hash),
		Number: &(n.Number),
	}
	return data
}

func (n *newBlockHashData) Deserialize(data serialize.SerializeStream) {
	d := data.(*protobufmsg.NewBlockHashData)
	n.Hash.Deserialize(d.Hash)
	n.Number = *(d.Number)
}

//// newBlockData is the network packet for the block propagation message.
//type newBlockData struct {
//	Block *types.Block
//	TD    *big.Int
//}
//
//// blockBody represents the data content of a single block.
//type blockBody struct {
//	Transactions []*types.Transaction // Transactions contained within a block
//	Uncles       []*types.Header      // Uncles contained within a block
//}
//
//// blockBodiesData is the network packet for block content distribution.
//type blockBodiesData []*blockBody
