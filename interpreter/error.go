package interpreter

import "errors"

var (
	// ErrKnownBlock is returned when a block to import is already known locally.
	ErrKnownTxType = errors.New("the tx type is unknown")
)
