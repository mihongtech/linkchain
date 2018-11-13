package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/linkchain/common"
)

type LinkChainConfig struct {
	// Name sets the instance name of the node. It must not contain the / character and is
	// used in the devp2p node identifier. The instance name of geth is "geth". If no
	// value is specified, the basename of the current executable is used.
	Name string `toml:"-"`

	// DataDir is the file system folder the node should use for any data storage
	// requirements. The configured data directory will not be directly shared with
	// registered services, instead those can use utility methods to create/access
	// databases or flat files. This enables ephemeral nodes which can fully reside
	// in memory.
	DataDir          string
	ConsensusService common.IService
	StorageService   common.IService
	ListenAddress    string
}

func (c *LinkChainConfig) instanceDir() string {
	if c.DataDir == "" {
		return ""
	}
	return filepath.Join(c.DataDir, c.name())
}

func (c *LinkChainConfig) name() string {
	if c.Name == "" {
		progname := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if progname == "" {
			panic("empty executable name, set Config.Name")
		}
		return progname
	}
	return c.Name
}

// resolvePath resolves path in the instance directory.
func (c *LinkChainConfig) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if c.DataDir == "" {
		return ""
	}

	return filepath.Join(c.instanceDir(), path)
}
