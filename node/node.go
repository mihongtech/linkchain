package node

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/linkchain/common"
	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/consensus"
	"github.com/linkchain/function/miner"
	"github.com/linkchain/function/wallet"
	"github.com/linkchain/p2p"
	"github.com/linkchain/storage"
)

var (
	//service collection
	svcList = []common.IService{
		&storage.Storage{},
		&consensus.Service{},
		&wallet.Wallet{},
		&p2p.Service{},
		&miner.Miner{},
	}
)

func Init() bool {
	log.Info("Node init...")

	//storage init
	if !svcList[0].Init(nil) {
		return false
	}

	//consensus init
	if !svcList[1].Init(GetStorage()) {
		return false
	}

	//wallet init
	if !svcList[2].Init(GetConsensusService().GetAccountManager()) {
		return false
	}

	//p2p init
	if !svcList[3].Init(GetConsensusService()) {
		return false
	}

	//storage init
	if !svcList[4].Init(nil) {
		return false
	}
	return true
}

func Run() {
	log.Info("Node is running...")

	//start all service
	for _, v := range svcList {
		if !v.Start() {
			return
		}
	}

	/*block :=svcList[1].(*consensus.Service).GetBlockManager().CreateBlock()
	svcList[1].(*consensus.Service).GetBlockManager().ProcessBlock(block)*/
}

// Config represents a small collection of configuration values to fine tune the
// P2P network layer of a protocol stack. These values can be further extended by
// all registered services.
type Config struct {
	// Name sets the instance name of the node. It must not contain the / character and is
	// used in the devp2p node identifier. The instance name of geth is "geth". If no
	// value is specified, the basename of the current executable is used.
	Name string `toml:"-"`

	// DataDir is the file system folder the node should use for any data storage
	// requirements. The configured data directory will not be directly shared with
	// registered services, instead those can use utility methods to create/access
	// databases or flat files. This enables ephemeral nodes which can fully reside
	// in memory.
	DataDir string
}

// Node is a container on which services can be registered.
type Node struct {
	config *Config
}

func (c *Config) instanceDir() string {
	if c.DataDir == "" {
		return ""
	}
	return filepath.Join(c.DataDir, c.name())
}

func (c *Config) name() string {
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
func (c *Config) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if c.DataDir == "" {
		return ""
	}

	return filepath.Join(c.instanceDir(), path)
}

// OpenDatabase opens an existing database with the given name (or creates one if no
// previous can be found) from within the node's instance directory. If the node is
// ephemeral, a memory database is returned.
func (n *Node) OpenDatabase(name string, cache, handles int) (lcdb.Database, error) {
	if n.config.DataDir == "" {
		return lcdb.NewMemDatabase()
	}
	return lcdb.NewLDBDatabase(n.config.resolvePath(name), cache, handles)
}

//get service

func GetStorage() *storage.Storage {
	return svcList[0].(*storage.Storage)
}

func GetConsensusService() *consensus.Service {
	return svcList[1].(*consensus.Service)
}

func GetWallet() *wallet.Wallet {
	return svcList[2].(*wallet.Wallet)
}

func GetP2pService() *p2p.Service {
	return svcList[3].(*p2p.Service)
}

func GetMiner() *miner.Miner {
	return svcList[4].(*miner.Miner)
}
