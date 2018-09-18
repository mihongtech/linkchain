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

//// getBlockHeadersData represents a block header query.
//type getBlockHeadersData struct {
//	Origin  hashOrNumber // Block from which to retrieve headers
//	Amount  uint64       // Maximum number of headers to retrieve
//	Skip    uint64       // Blocks to skip between consecutive headers
//	Reverse bool         // Query direction (false = rising towards latest, true = falling towards genesis)
//}
//
//// hashOrNumber is a combined field for specifying an origin block.
//type hashOrNumber struct {
//	Hash   common.Hash // Block hash from which to retrieve headers (excludes Number)
//	Number uint64      // Block hash from which to retrieve headers (excludes Hash)
//}
//
//// EncodeRLP is a specialized encoder for hashOrNumber to encode only one of the
//// two contained union fields.
//func (hn *hashOrNumber) EncodeRLP(w io.Writer) error {
//	if hn.Hash == (common.Hash{}) {
//		return rlp.Encode(w, hn.Number)
//	}
//	if hn.Number != 0 {
//		return fmt.Errorf("both origin hash (%x) and number (%d) provided", hn.Hash, hn.Number)
//	}
//	return rlp.Encode(w, hn.Hash)
//}
//
//// DecodeRLP is a specialized decoder for hashOrNumber to decode the contents
//// into either a block hash or a block number.
//func (hn *hashOrNumber) DecodeRLP(s *rlp.Stream) error {
//	_, size, _ := s.Kind()
//	origin, err := s.Raw()
//	if err == nil {
//		switch {
//		case size == 32:
//			err = rlp.DecodeBytes(origin, &hn.Hash)
//		case size <= 8:
//			err = rlp.DecodeBytes(origin, &hn.Number)
//		default:
//			err = fmt.Errorf("invalid input size %d for origin", size)
//		}
//	}
//	return err
//}
//
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
