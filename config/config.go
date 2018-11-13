package config

import (
	"github.com/linkchain/common"
)

type LinkChainConfig struct {
	ConsensusService common.IService
	ListenPort       int
}
