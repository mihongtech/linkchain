package storage

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/linkchain/common/lcdb"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
	"github.com/linkchain/meta/chain"
	"github.com/linkchain/meta/tx"
)

type IStroage interface {
	//block
	storeBlock(block block.IBlock)
	loadBlockById(id meta.BlockID) block.IBlock
	loadBlockByHeight(height int) block.IBlock

	//tx
	storeTx(iTx tx.ITx)
	loadTxById(id meta.BlockID) tx.ITx

	//chain info
	storeChain(chain chain.IChain)
	storeChainGraph(graph chain.IChainGraph)
	loadChainGraph() chain.IChainGraph
}

type Storage struct {
	Name    string
	db      lcdb.Database
	dataDir string
}

func (s *Storage) Init(i interface{}) bool {
	log.Info("Stroage init...")

	//load genesis from storage
	var err error
	name := "chaindata"

	s.db, err = s.OpenDatabase(name, 1024, 256)
	if err != nil {
		return false
	}

	return true
}

func (s *Storage) OpenDatabase(name string, cache, handles int) (lcdb.Database, error) {
	if s.dataDir == "" {
		return lcdb.NewMemDatabase()
	}
	return lcdb.NewLDBDatabase(s.resolvePath(name), cache, handles)
}

func (s *Storage) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if s.dataDir == "" {
		return ""
	}

	return filepath.Join(s.instanceDir(), path)
}

func (s *Storage) instanceDir() string {
	if s.dataDir == "" {
		return ""
	}
	return filepath.Join(s.dataDir, s.name())
}

func (s *Storage) name() string {
	if s.Name == "" {
		progname := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if progname == "" {
			panic("empty executable name, set Config.Name")
		}
		return progname
	}
	return s.Name
}

func (m *Storage) Start() bool {
	log.Info("Stroage start...")
	return true
}

func (m *Storage) Stop() {
	log.Info("Stroage stop...")
}

func (m *Storage) GetDB() lcdb.Database {
	return m.db
}
