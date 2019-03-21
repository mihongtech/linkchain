package config

import (
	"github.com/linkchain/contract/vm/params"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

type ChainConfig struct {
	ChainId *big.Int `json:"chainId"` // chain id identifies the current chain and is used for replay protection
	Period  uint64   `json:"period"`  // Number of seconds between blocks to enforce
}

// GasTable returns the gas table corresponding to the current phase (homestead or homestead reprice).
//
// The returned GasTable's fields shouldn't, under any circumstances, be changed.
func (c *ChainConfig) GasTable(num *big.Int) params.GasTable {
	return params.GasTableEIP158
}

type LinkChainConfig struct {
	// DataDir is the file system folder the node should use for any data storage
	// requirements. The configured data directory will not be directly shared with
	// registered services, instead those can use utility methods to create/access
	// databases or flat files. This enables ephemeral nodes which can fully reside
	// in memory.
	DataDir     string
	GenesisPath string
	//NodeService 	  common.Service
	ListenAddress  string
	NoDiscovery    bool
	BootstrapNodes string
	InterpreterAPI string
}

// DefaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Linkchain")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "Linkchain")
		} else {
			return filepath.Join(home, ".linkchain")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
