package consensus

import (
	"errors"

	"github.com/mihongtech/linkchain/core/meta"
)

var (
	// ErrUnknownAncestor is returned when validating a block requires an ancestor
	// that is unknown.
	ErrUnknownAncestor = errors.New("unknown ancestor")

	// ErrPrunedAncestor is returned when validating a block requires an ancestor
	// that is known, but the state of which is not available.
	ErrPrunedAncestor = errors.New("pruned ancestor")

	// ErrFutureBlock is returned when a block's timestamp is in the future according
	// to the current node.
	ErrFutureBlock = errors.New("block in the future")

	// ErrInvalidNumber is returned if a block's number doesn't equal it's parent's
	// plus one.
	ErrInvalidNumber = errors.New("invalid block number")
)

// Engine is an algorithm agnostic consensus engine.
type Engine interface {
	// Author retrieves the Rollchain address of the account that minted the given
	// block, which may be different from the header's coinbase if a consensus
	// engine is based on signatures.
	Author(header *meta.BlockHeader) ([]byte, error)

	// VerifyHeader checks whether a block conforms to the consensus rules of a
	// given engine.
	VerifyBlock(chain meta.ChainReader, block *meta.Block) error
}
