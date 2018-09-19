package full

import (
	_ "fmt"
	_ "io"
	_ "math/big"

	"github.com/linkchain/common/math"
	"github.com/linkchain/common/serialize"
	"github.com/linkchain/meta"
	"github.com/linkchain/protobuf"
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

type statusData struct {
	ProtocolVersion uint32
	NetworkId       uint64
	CurrentBlock    meta.DataID
	GenesisBlock    meta.DataID
}

func (s *statusData) Serialize() serialize.SerializeStream {
	currentBlock := s.CurrentBlock.Serialize().(*protobuf.Hash)
	genesisBlock := s.GenesisBlock.Serialize().(*protobuf.Hash)
	status := &protobuf.StatusData{
		ProtocolVersion: &s.ProtocolVersion,
		NetworkId:       &s.NetworkId,
		CurrentBlock:    currentBlock,
		GenesisBlock:    genesisBlock,
	}

	return status
}

func (s *statusData) Deserialize(data serialize.SerializeStream) {
	d := data.(*protobuf.StatusData)
	s.ProtocolVersion = *d.ProtocolVersion
	s.NetworkId = *d.NetworkId
	genesis := &math.Hash{}
	genesis.Deserialize(d.GenesisBlock)
	s.GenesisBlock = genesis
	current := &math.Hash{}
	current.Deserialize(d.CurrentBlock)
	s.CurrentBlock = current
}

// newBlockHashesData is the network packet for the block announcements.
type newBlockHashesData []newBlockHashData

type newBlockHashData struct {
	Hash   meta.DataID // Hash of one particular block being announced
	Number uint64      // Number of one particular block being announced
}

func (n *newBlockHashData) Serialize() serialize.SerializeStream {
	data := &protobuf.NewBlockHashData{
		Hash:   n.Hash.Serialize().(*protobuf.Hash),
		Number: &(n.Number),
	}
	return data
}

func (n *newBlockHashData) Deserialize(data serialize.SerializeStream) {
	d := data.(*protobuf.NewBlockHashData)
	n.Hash = &math.Hash{}
	n.Hash.Deserialize(d.Hash)
	n.Number = *(d.Number)
}

type getBlockHeadersData struct {
	Hash    meta.DataID // Hash of one particular block being announced
	Number  uint64      // Number of one particular block being announced
	Amount  uint64      // Maximum number of headers to retrieve
	Skip    uint64      // Blocks to skip between consecutive headers
	Reverse bool        // Query direction (false = rising towards latest, true = falling towards genesis)

}

func (n *getBlockHeadersData) Serialize() serialize.SerializeStream {
	var hashdata *protobuf.Hash
	if n.Hash != nil {
		hashdata = n.Hash.Serialize().(*protobuf.Hash)
	} else {
		empty := &math.Hash{}
		hashdata = empty.Serialize().(*protobuf.Hash)
	}
	data := &protobuf.GetBlockHeadersData{
		Hash:    hashdata,
		Number:  &(n.Number),
		Amount:  &(n.Amount),
		Skip:    &(n.Skip),
		Reverse: &(n.Reverse),
	}
	return data
}

func (n *getBlockHeadersData) Deserialize(data serialize.SerializeStream) {
	d := data.(*protobuf.GetBlockHeadersData)
	n.Hash = &math.Hash{}
	n.Hash.Deserialize(d.Hash)
	n.Number = *(d.Number)
	n.Amount = *(d.Amount)
	n.Skip = *(d.Skip)
	n.Reverse = *(d.Reverse)
}
