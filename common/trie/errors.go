package trie

import (
	"fmt"

	"github.com/mihongtech/linkchain/common/math"
)

// MissingNodeError is returned by the trie functions (TryGet, TryUpdate, TryDelete)
// in the case where a trie node is not present in the local database. It contains
// information necessary for retrieving the missing node.
type MissingNodeError struct {
	NodeHash math.Hash // hash of the missing node
	Path     []byte    // hex-encoded path to the missing node
}

func (err *MissingNodeError) Error() string {
	return fmt.Sprintf("missing trie node %v (path %x)", err.NodeHash, err.Path)
}
