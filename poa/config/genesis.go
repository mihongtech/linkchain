package config

import (
	_ "bytes"
	_ "encoding/hex"
	_ "encoding/json"
	_ "errors"
	_ "fmt"
	"math/big"
	_ "strings"

	"github.com/linkchain/common/math"
)

type ChainConfig struct {
	ChainId *big.Int `json:"chainId"` // Chain id identifies the current chain and is used for replay protection
	Period  uint64   `json:"period"`  // Number of seconds between blocks to enforce
}

type Genesis struct {
	Config     *ChainConfig `json:"config"`
	Version    uint32       `json:"version"`
	Time       uint64       `json:"time"`
	Data       []byte       `json:"data"`
	Difficulty *big.Int     `json:"difficulty" gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Height uint32    `json:"height"`
	Prev   math.Hash `json:"prev"`
}
