package config

import (
	"github.com/linkchain/common"
)

type LinkChainConfig struct {
	ConsensusService common.IService
	StorageService   common.IService
	DataDir          string
	ListenPort       int
}
