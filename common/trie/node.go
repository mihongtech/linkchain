package trie

import (
	_ "errors"
	"fmt"
	"io"
	"strings"

	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/serialize"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/protobuf"

	"github.com/golang/protobuf/proto"
)

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "[17]"}

type node interface {
	fstring(string) string
	cache() (hashNode, bool)
	canUnload(cachegen, cachelimit uint16) bool

	//serialize
	serialize.ISerialize
}

type (
	fullNode struct {
		Children [17]node // Actual trie node data to encode/decode (needs custom encoder)
		flags    nodeFlag
	}
	shortNode struct {
		Key   []byte
		Val   node
		flags nodeFlag
	}
	hashNode  []byte
	valueNode []byte
)

func (n *fullNode) copy() *fullNode   { copy := *n; return &copy }
func (n *shortNode) copy() *shortNode { copy := *n; return &copy }

// nodeFlag contains caching-related metadata about a node.
type nodeFlag struct {
	hash  hashNode // cached hash of the node (may be nil)
	gen   uint16   // cache generation counter
	dirty bool     // whether the node has changes that must be written to the database
}

// canUnload tells whether a node can be unloaded.
func (n *nodeFlag) canUnload(cachegen, cachelimit uint16) bool {
	return !n.dirty && cachegen-n.gen >= cachelimit
}

func (n *fullNode) canUnload(gen, limit uint16) bool  { return n.flags.canUnload(gen, limit) }
func (n *shortNode) canUnload(gen, limit uint16) bool { return n.flags.canUnload(gen, limit) }
func (n hashNode) canUnload(uint16, uint16) bool      { return false }
func (n valueNode) canUnload(uint16, uint16) bool     { return false }

func (n *fullNode) cache() (hashNode, bool)  { return n.flags.hash, n.flags.dirty }
func (n *shortNode) cache() (hashNode, bool) { return n.flags.hash, n.flags.dirty }
func (n hashNode) cache() (hashNode, bool)   { return nil, true }
func (n valueNode) cache() (hashNode, bool)  { return nil, true }

// Pretty printing.
func (n *fullNode) String() string  { return n.fstring("") }
func (n *shortNode) String() string { return n.fstring("") }
func (n hashNode) String() string   { return n.fstring("") }
func (n valueNode) String() string  { return n.fstring("") }

func (n *fullNode) fstring(ind string) string {
	resp := fmt.Sprintf("[\n%s  ", ind)
	for i, node := range n.Children {
		if node == nil {
			resp += fmt.Sprintf("%s: <nil> ", indices[i])
		} else {
			resp += fmt.Sprintf("%s: %v", indices[i], node.fstring(ind+"  "))
		}
	}
	return resp + fmt.Sprintf("\n%s] ", ind)
}
func (n *shortNode) fstring(ind string) string {
	return fmt.Sprintf("{%x: %v} ", n.Key, n.Val.fstring(ind+"  "))
}
func (n hashNode) fstring(ind string) string {
	return fmt.Sprintf("<%x> ", []byte(n))
}
func (n valueNode) fstring(ind string) string {
	return fmt.Sprintf("%x ", []byte(n))
}

//Serialize/Deserialize
func (n *fullNode) Serialize() serialize.SerializeStream {
	var children [][]byte

	for _, child := range n.Children {
		if nil == child {
			children = append(children, nil)
			continue
		}

		enc := child.Serialize()
		if nil == enc {
			children = append(children, nil)
			continue
		}

		buffer, err := proto.Marshal(enc)
		if err != nil {
			log.Error("node marshaling error: ", "full node", err, "child is", child, "enc", enc)
		}
		switch child.(type) {
		case valueNode:
			if len(child.(valueNode)) == 32 {
				data := append(child.(valueNode)[:], byte(0))
				enc = data.Serialize()
				buffer, _ = proto.Marshal(enc)
				children = append(children, buffer)
			} else {
				children = append(children, buffer)
			}
		case hashNode:
			children = append(children, buffer)
		default:
			children = append(children, buffer)
		}
	}

	node := protobuf.FullNode{
		Children: children,
	}
	return &node
}

func (n *shortNode) Serialize() serialize.SerializeStream {
	enc := n.Val.Serialize()
	buffer, err := proto.Marshal(enc)
	if err != nil {
		log.Error("node marshaling error: ", "short node", err, "n.Val", n.Val)
	}

	switch nt := n.Val.(type) {
	case valueNode:
		node := protobuf.ShortNode{
			Key: n.Key,
			Val: buffer,
		}
		return &node
	case hashNode:
		node := protobuf.ShortNode{
			Key: n.Key,
			Val: buffer,
		}
		return &node
	default:
		log.Info("%v: invalid node: %v", buffer, nt)
		hash := math.HashB(buffer)

		node := protobuf.ShortNode{
			Key: n.Key,
			Val: hash,
		}
		return &node
	}
}

func (n hashNode) Serialize() serialize.SerializeStream {
	if len(n) == 0 {
		return nil
	}

	node := protobuf.HashNode{
		Data: n,
	}

	return &node
}

func (n valueNode) Serialize() serialize.SerializeStream {
	if len(n) == 0 {
		return nil
	}

	node := protobuf.ValueNode{
		Data: n,
	}

	return &node
}

func (n *fullNode) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.FullNode)

	if len(data.Children) != 17 {
		return fmt.Errorf("parse childern error, children: %v,  len(data.Children) :%v", data.Children, len(data.Children))
	}
	for i := 0; i < 17; i++ {
		if len(data.Children[i]) == 0 || data.Children[i] == nil {
			n.Children[i] = nil
			continue
		}

		cld, err := decodeRef(data.Children[i], 0)
		if err != nil {
			// n.Children[i] = nil
			return err
		}
		n.Children[i] = cld
	}

	return nil
}

func (n *shortNode) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.ShortNode)
	n.Key = append(n.Key, data.Key...)

	key := compactToHex(data.Key)
	if hasTerm(key) {
		n.Val = append(valueNode{}, data.Val...)
	} else {
		val, err := decodeRef(data.Val, 0)
		if err != nil {
			return err
		}
		n.Val = val
	}
	// do not deserialize flags
	return nil
}

func (n hashNode) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.HashNode)
	copy(n, data.Data)
	return nil
}

func (n valueNode) Deserialize(s serialize.SerializeStream) error {
	data := *s.(*protobuf.ValueNode)
	copy(n, data.Data)
	return nil
}

func mustDecodeNode(hash, buf []byte, cachegen uint16) node {
	n, err := decodeNode(hash, buf, cachegen)
	if err != nil {
		panic(fmt.Sprintf("node %x: %v", hash, err))
	}
	return n
}

// decodeNode parses the RLP encoding of a trie node.
func decodeNode(hash, buf []byte, cachegen uint16) (node, error) {
	if len(buf) == 0 {
		return nil, io.ErrUnexpectedEOF
	}

	if n, err := decodeShort(hash, buf, cachegen); err != nil {
		if n, err := decodeFull(hash, buf, cachegen); err != nil {
			return nil, fmt.Errorf("invalid node data %v, %v", buf, err)
		} else {
			return n, wrapError(err, "full")

		}
	} else {
		return n, wrapError(err, "short")
	}
}

func decodeShort(hash, buf []byte, cachegen uint16) (node, error) {
	var node protobuf.ShortNode
	if err := proto.Unmarshal(buf, &node); err != nil {
		return nil, err
	}

	flag := nodeFlag{hash: hash, gen: cachegen}
	key := compactToHex(node.Key)
	if hasTerm(key) {
		var n protobuf.ValueNode
		if err := proto.Unmarshal(node.Val, &n); err != nil {
			return nil, err
		}
		return &shortNode{key, append(valueNode{}, n.Data...), flag}, nil
	}

	r, err := decodeRef(node.Val, cachegen)
	if err != nil {
		return nil, err
	}

	return &shortNode{key, r, flag}, nil
}

func decodeFull(hash, buf []byte, cachegen uint16) (*fullNode, error) {
	n := &fullNode{flags: nodeFlag{hash: hash, gen: cachegen}}

	var node protobuf.FullNode
	if err := proto.Unmarshal(buf, &node); err != nil {
		return nil, err
	}

	err := n.Deserialize(&node)
	if err != nil {
		return nil, err
	}

	n.flags = nodeFlag{hash: hash, gen: cachegen}

	return n, nil
}

const hashLen = len(math.Hash{})

func decodeRef(buf []byte, cachegen uint16) (node, error) {
	var short protobuf.ShortNode
	if err := proto.Unmarshal(buf, &short); err == nil {
		if sn, err := decodeShort(nil, buf, cachegen); err == nil {
			return sn, nil
		}
	}

	var n protobuf.HashNode
	if err := proto.Unmarshal(buf, &n); err != nil {
		return nil, fmt.Errorf("decodeRef data, buf:%v, err: %v", buf, err)
	}

	switch {
	case len(n.Data) == 0:
		// empty node
		return nil, nil
	case len(n.Data) == 32:
		return append(hashNode{}, n.Data...), nil
	default:
		if len(n.Data) == 33 {
			if n.Data[32] == byte(0) {
				return append(valueNode{}, n.Data[:32]...), nil
			}
		}

		return append(valueNode{}, n.Data...), nil
	}
}

// wraps a decoding error with information about the path to the
// invalid child node (for debugging encoding issues).
type decodeError struct {
	what  error
	stack []string
}

func wrapError(err error, ctx string) error {
	if err == nil {
		return nil
	}
	if decErr, ok := err.(*decodeError); ok {
		decErr.stack = append(decErr.stack, ctx)
		return decErr
	}
	return &decodeError{err, []string{ctx}}
}

func (err *decodeError) Error() string {
	return fmt.Sprintf("%v (decode path: %s)", err.what, strings.Join(err.stack, "<-"))
}
