package node

import (
	"errors"
	"fmt"
	"github.com/mihongtech/linkchain/interpreter"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mihongtech/linkchain/common/lcdb"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/trie"
	"github.com/mihongtech/linkchain/common/util/event"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/common/util/mclock"
	"github.com/mihongtech/linkchain/config"
	"github.com/mihongtech/linkchain/consensus"
	"github.com/mihongtech/linkchain/core"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/normal"
	"github.com/mihongtech/linkchain/storage"
	"github.com/mihongtech/linkchain/storage/state"

	"github.com/hashicorp/golang-lru"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

var (
	ErrNoGenesis = errors.New("Genesis not found in chain")
)

const (
	blockCacheLimit     = 256
	maxFutureBlocks     = 256
	maxTimeFutureBlocks = 30
	badBlockLimit       = 10
	numberCacheLimit    = 2048
	triesInMemory       = 128
	receiptsCacheLimit  = 32

	// BlockChainVersion ensures that an incompatible database forces a resync from scratch.
	BlockChainVersion = 3
)

// CacheConfig contains the configuration values for the trie caching/pruning
// that's resident in a blockchain.
type CacheConfig struct {
	Disabled      bool          // Whether to disable trie write caching (archive node)
	TrieNodeLimit int           // Memory limit (MB) at which to flush the current in-memory trie to disk
	TrieTimeLimit time.Duration // Time limit after which to flush the current in-memory trie to disk
}

// BlockChain represents the canonical chain given a database with a genesis
// block. The Blockchain manages chain imports, reverts, chain reorganisations.
//
// Importing blocks in to the block chain happens according to the set of rules
// defined by the two stage Validator. Processing of blocks is done using the
// Processor which processes the included transaction. The validation of the state
// is done in the second part of the Validator. Failing results in aborting of
// the import.
//
// The BlockChain also helps in returning blocks from **any** chain included
// in the database as well as blocks that represents the canonical chain. It's
// important to note that GetBlock can return any block and does not need to be
// included in the canonical one where as GetBlockByHeight always represents the
// canonical chain.
type BlockChain struct {
	chainConfig *config.ChainConfig // chain & network configuration
	cacheConfig *CacheConfig        // Cache configuration for pruning

	db     lcdb.Database // Low level persistent database to store final content in
	triegc *prque.Prque  // Priority queue mapping block numbers to tries to gc
	gcproc time.Duration // Accumulates canonical block processing for trie dumping

	chainFeed     event.Feed
	chainSideFeed event.Feed
	chainHeadFeed event.Feed
	scope         event.SubscriptionScope
	genesisBlock  *meta.Block

	mu      sync.RWMutex // global mutex for locking chain operations
	chainmu sync.RWMutex // blockchain insertion lock
	procmu  sync.RWMutex // block processor lock

	checkpoint       int          // checkpoint counts towards the new checkpoint
	currentBlock     atomic.Value // Current head of the block chain
	currentFastBlock atomic.Value // Current head of the fast-sync chain (may be above the block chain!)
	currentBlockHash math.Hash    // Hash of the current head of the header chain (prevent recomputing all the time)

	// stateCache   state.Database // State database to reuse between imports (contains state cache)
	blockCache    *lru.Cache // Cache for the most recent entire blocks
	receiptsCache *lru.Cache // Cache for the most recent receipts per block
	futureBlocks  *lru.Cache // future blocks are blocks added for later processing
	numberCache   *lru.Cache // Cache for the most recent block numbers

	stateCache state.Database
	quit       chan struct{} // blockchain quit channel
	running    int32         // running must be called atomically
	// procInterrupt must be atomically called
	procInterrupt int32          // interrupt signaler for block processing
	wg            sync.WaitGroup // chain processing wait group for shutting down

	processor interpreter.Processor
	validator interpreter.Validator // block and state validator interpreter

	badBlocks *lru.Cache // Bad block cache

	engine consensus.Engine //Block chain engine
}

// NewBlockChain returns a fully initialised block chain using information
// available in the database. It initialises the default Ethereum Validator and
// Processor.
func NewBlockChain(db lcdb.Database, genesisHash math.Hash, cacheConfig *CacheConfig, chainConfig *config.ChainConfig, intrepreterAPI interpreter.Interpreter, engine consensus.Engine) (*BlockChain, error) {
	if cacheConfig == nil {
		cacheConfig = &CacheConfig{
			TrieNodeLimit: 256 * 1024 * 1024,
			TrieTimeLimit: 5 * time.Minute,
		}
	}
	blockCache, _ := lru.New(blockCacheLimit)
	futureBlocks, _ := lru.New(maxFutureBlocks)
	badBlocks, _ := lru.New(badBlockLimit)
	numberCache, _ := lru.New(numberCacheLimit)
	receiptsCache, _ := lru.New(receiptsCacheLimit)
	bc := &BlockChain{
		chainConfig:   chainConfig,
		cacheConfig:   cacheConfig,
		db:            db,
		triegc:        prque.New(),
		stateCache:    state.NewDatabase(db),
		quit:          make(chan struct{}),
		blockCache:    blockCache,
		futureBlocks:  futureBlocks,
		badBlocks:     badBlocks,
		numberCache:   numberCache,
		receiptsCache: receiptsCache,
		engine:        engine,
	}
	bc.validator = intrepreterAPI
	bc.processor = intrepreterAPI

	bc.genesisBlock, _ = bc.GetBlockByHeight(0)
	if bc.genesisBlock == nil || !bc.genesisBlock.GetBlockID().IsEqual(&genesisHash) {
		return nil, ErrNoGenesis
	}
	if err := bc.loadLastState(); err != nil {
		return nil, err
	}
	bc.SetCurrentBlockHead(bc.CurrentBlock())
	// Take ownership of this particular state
	go bc.update()
	return bc, nil
}

func (bc *BlockChain) GetChainID() *big.Int {
	return bc.chainConfig.ChainId
}

func (bc *BlockChain) getProcInterrupt() bool {
	return atomic.LoadInt32(&bc.procInterrupt) == 1
}

// loadLastState loads the last known chain state from the database. This method
// assumes that the chain manager mutex is held.
func (bc *BlockChain) loadLastState() error {
	// Restore the last known head block
	head := storage.GetHeadBlockHash(bc.db)
	if head == (math.Hash{}) {
		// Corrupt or empty database, init from scratch
		log.Warn("Empty database, resetting chain")
		return bc.Reset()
	}
	// Make sure the entire head block is available
	currentBlock, _ := bc.GetBlockByID(head)
	if currentBlock == nil {
		// Corrupt or empty database, init from scratch
		log.Warn("Head block missing, resetting chain", "hash", head)
		return bc.Reset()
	}
	// Make sure the state associated with the block is available
	if _, err := state.New(*currentBlock.GetStatus(), bc.db); err != nil {
		// Dangling block without a state associated, init from scratch
		log.Warn("Head state missing, repairing chain", "number", currentBlock.GetHeight(), "hash", currentBlock.GetBlockID())
		if err := bc.repair(&currentBlock); err != nil {
			return err
		}
	}
	// Everything seems to be fine, set as the head block
	bc.currentBlock.Store(currentBlock)

	// Restore the last known head fast block
	bc.currentFastBlock.Store(currentBlock)
	if head := storage.GetHeadFastBlockHash(bc.db); head != (math.Hash{}) {
		if block, _ := bc.GetBlockByID(head); block != nil {
			bc.currentFastBlock.Store(block)
		}
	}

	// Issue a status log for the user
	// currentFastBlock := bc.CurrentFastBlock()
	log.Info("Loaded most recent local full block", "number", currentBlock.GetHeight(), "hash", currentBlock.GetBlockID())
	return nil
}

// SetHead rewinds the local chain to a new head. In the case of headers, everything
// above the new head will be deleted and the new one set. In the case of blocks
// though, the head may be further rewound if block bodies are missing (non-archive
// nodes after a fast sync).
func (bc *BlockChain) SetHead(head uint64) error {
	log.Warn("Rewinding blockchain", "target", head)

	bc.mu.Lock()
	defer bc.mu.Unlock()

	height := uint64(0)

	if hdr := bc.CurrentBlock(); hdr != nil {
		height = uint64(hdr.GetHeight())
	}

	for hdr := bc.CurrentBlock(); hdr != nil && uint64(hdr.GetHeight()) > head; hdr = bc.CurrentBlock() {
		hash := *hdr.GetBlockID()
		num := uint64(hdr.GetHeight())
		prev := *hdr.GetPrevBlockID()
		storage.DeleteBlock(bc.db, hash, num)
		bc.currentBlock.Store(bc.GetBlock(prev, num-1))
		bc.currentFastBlock.Store(bc.GetBlock(prev, num-1))
	}
	// Roll back the canonical chain numbering
	for i := height; i > head; i-- {
		storage.DeleteCanonicalHash(bc.db, i)
	}
	bc.SetCurrentBlockHead(bc.CurrentBlock())
	// Clear out any stale content from the caches
	bc.blockCache.Purge()
	bc.futureBlocks.Purge()
	bc.receiptsCache.Purge()
	bc.numberCache.Purge()

	// Rewind the block chain, ensuring we don't end up with a stateless head block
	if currentBlock := bc.CurrentBlock(); currentBlock != nil {
		if _, err := state.New(*currentBlock.GetStatus(), bc.db); err != nil {
			// Rewound state missing, rolled back to before pivot, reset to genesis
			bc.currentBlock.Store(bc.genesisBlock)
		}
	}

	// If either blocks reached nil, reset to the genesis state
	if currentBlock := bc.CurrentBlock(); currentBlock == nil {
		bc.currentBlock.Store(bc.genesisBlock)
	}
	if currentFastBlock := bc.CurrentFastBlock(); currentFastBlock == nil {
		bc.currentFastBlock.Store(bc.genesisBlock)
	}
	currentBlock := bc.CurrentBlock()
	currentFastBlock := bc.CurrentFastBlock()
	if err := storage.WriteHeadBlockHash(bc.db, *currentBlock.GetBlockID()); err != nil {
		log.Crit("Failed to reset head full block", "err", err)
	}
	if err := storage.WriteHeadFastBlockHash(bc.db, *currentFastBlock.GetBlockID()); err != nil {
		log.Crit("Failed to reset head fast block", "err", err)
	}
	return bc.loadLastState()
}

// CurrentBlock retrieves the current head block of the canonical chain. The
// block is retrieved from the blockchain's internal cache.
func (bc *BlockChain) CurrentBlock() *meta.Block {
	log.Debug("current is", "block", bc.currentBlock.Load().(*meta.Block))
	return bc.currentBlock.Load().(*meta.Block)
}

// CurrentFastBlock retrieves the current fast-sync head block of the canonical
// chain. The block is retrieved from the blockchain's internal cache.
func (bc *BlockChain) CurrentFastBlock() *meta.Block {
	return bc.currentFastBlock.Load().(*meta.Block)
}

// State returns a new mutable state based on the current HEAD block.
func (bc *BlockChain) State() (*state.StateDB, error) {
	return bc.StateAt(*bc.CurrentBlock().GetStatus())
}

// StateAt returns a new mutable state based on a particular point in time.
func (bc *BlockChain) StateAt(root math.Hash) (*state.StateDB, error) {
	return state.New(root, bc.db)
}

// Reset purges the entire blockchain, restoring it to its genesis state.
func (bc *BlockChain) Reset() error {
	return bc.ResetWithGenesisBlock(bc.genesisBlock)
}

// ResetWithGenesisBlock purges the entire blockchain, restoring it to the
// specified genesis state.
func (bc *BlockChain) ResetWithGenesisBlock(genesis *meta.Block) error {
	// Dump the entire block chain and purge the caches
	if err := bc.SetHead(0); err != nil {
		return err
	}
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Prepare the genesis block and reinitialise the chain
	if err := storage.WriteBlock(bc.db, genesis); err != nil {
		log.Crit("Failed to write genesis block", "err", err)
	}
	bc.genesisBlock = genesis
	bc.insert(bc.genesisBlock)
	bc.currentBlock.Store(bc.genesisBlock)
	bc.SetCurrentBlockHead(bc.genesisBlock)
	bc.currentFastBlock.Store(bc.genesisBlock)

	return nil
}

// repair tries to repair the current blockchain by rolling back the current block
// until one with associated state is found. This is needed to fix incomplete db
// writes caused either by crashes/power outages, or simply non-committed tries.
//
// This method only rolls back the current block. The current header and current
// fast block are left intact.
func (bc *BlockChain) repair(head **meta.Block) error {
	for {
		// Abort if we've rewound to a head block that does have associated state
		if _, err := state.New(*(*head).GetStatus(), bc.db); err == nil {
			log.Info("Rewound blockchain to past state", "number", (*head).GetHeight(), "hash", (*head).GetBlockID())
			return nil
		}
		// Otherwise rewind one block and recheck state availability there
		(*head) = bc.GetBlock(*(*head).GetPrevBlockID(), uint64((*head).GetHeight()-1))
	}
}

// insert injects a new head block into the current block chain. This method
// assumes that the block is indeed a true head. It will also reset the head
// header and the head fast sync block to this very same block if they are older
// or if they are on a different side chain.
//
// Note, this function assumes that the `mu` mutex is held!
func (bc *BlockChain) insert(block *meta.Block) {
	// If the block is on a side chain or an unknown one, force other heads onto it too
	updateHeads := (storage.GetCanonicalHash(bc.db, uint64(block.GetHeight()))) != *block.GetBlockID()

	// Add the block to the canonical chain number scheme and mark as the head
	if err := storage.WriteCanonicalHash(bc.db, *block.GetBlockID(), uint64(block.GetHeight())); err != nil {
		log.Crit("Failed to insert block number", "err", err)
	}
	if err := storage.WriteHeadBlockHash(bc.db, *block.GetBlockID()); err != nil {
		log.Crit("Failed to insert head block hash", "err", err)
	}
	bc.currentBlock.Store(block)

	// If the block is better than our head or is on a different chain, force update heads
	if updateHeads {
		bc.SetCurrentBlockHead(block)

		if err := storage.WriteHeadFastBlockHash(bc.db, *block.GetBlockID()); err != nil {
			log.Crit("Failed to insert head fast block hash", "err", err)
		}
		bc.currentFastBlock.Store(block)
	}
}

// Genesis retrieves the chain's genesis block.
func (bc *BlockChain) Genesis() *meta.Block {
	return bc.genesisBlock
}

// HasBlock checks if a block is fully present in the database or not.
func (bc *BlockChain) HasBlockAndNum(hash math.Hash, number uint64) bool {
	if bc.blockCache.Contains(hash) {
		return true
	}
	ok := storage.HasBlock(bc.db, hash, number)
	return ok
}

func (bc *BlockChain) HasBlock(hash math.Hash) bool {
	if bc.blockCache.Contains(hash) {
		return true
	}
	ok := storage.HasBlock(bc.db, hash, bc.GetBlockNumber(hash))
	return ok
}

// HasState checks if state trie is fully present in the database or not.
func (bc *BlockChain) HasState(hash math.Hash) bool {
	_, err := bc.stateCache.OpenTrie(hash)
	return err == nil
}

func (bc *BlockChain) SetCurrentBlockHead(head *meta.Block) {
	bc.currentBlockHash = *head.GetBlockID()
}

// HasBlockAndState checks if a block and associated state trie is fully present
// in the database or not, caching it if present.
func (bc *BlockChain) HasBlockAndState(hash math.Hash, number uint64) bool {
	// Check first that the block itself is known
	block := bc.GetBlock(hash, number)
	if block == nil {
		return false
	}
	return bc.HasState(*block.GetStatus())
}

// GetBlock retrieves a block from the database by hash and number,
// caching it if found.
func (bc *BlockChain) GetBlock(hash math.Hash, number uint64) *meta.Block {
	// Short circuit if the block's already in the cache, retrieve otherwise
	if block, ok := bc.blockCache.Get(hash); ok {
		return block.(*meta.Block)
	}
	block := storage.GetBlock(bc.db, hash, number)
	if block == nil {
		return nil
	}
	// Cache the found block for next time and return
	bc.blockCache.Add(*block.GetBlockID(), block)
	return block
}

// GetBlockByHash retrieves a block from the database by hash, caching it if found.
func (bc *BlockChain) GetBlockByID(hash math.Hash) (*meta.Block, error) {
	return bc.GetBlock(hash, bc.GetBlockNumber(hash)), nil
}

func (bc *BlockChain) GetBlockNumber(hash math.Hash) uint64 {
	if cached, ok := bc.numberCache.Get(hash); ok {
		return cached.(uint64)
	}
	number := storage.GetBlockNumber(bc.db, hash)
	if number != storage.MissingNumber {
		bc.numberCache.Add(hash, number)
	}
	return number
}

// GetBlockByHeight retrieves a block from the database by number, caching it
// (associated with its hash) if found.
func (bc *BlockChain) GetBlockByHeight(number uint32) (*meta.Block, error) {
	hash := storage.GetCanonicalHash(bc.db, uint64(number))
	if hash == (math.Hash{}) {
		return nil, errors.New("block not found")
	}
	return bc.GetBlock(hash, uint64(number)), nil
}

// TrieNode retrieves a blob of data associated with a trie node (or code hash)
// either from ephemeral in-memory cache, or from persistent storage.
func (bc *BlockChain) TrieNode(hash math.Hash) ([]byte, error) {
	return bc.stateCache.TrieDB().Node(hash)
}

// Stop stops the blockchain service. If any imports are currently in progress
// it will abort them using the procInterrupt.
func (bc *BlockChain) Stop() {
	if !atomic.CompareAndSwapInt32(&bc.running, 0, 1) {
		return
	}
	// Unsubscribe all subscriptions registered from blockchain
	bc.scope.Close()
	close(bc.quit)
	atomic.StoreInt32(&bc.procInterrupt, 1)

	bc.wg.Wait()

	// Ensure the state of a recent block is also stored to disk before exiting.
	// We're writing three different states to catch different restart scenarios:
	//  - HEAD:     So we don't need to reprocess any blocks in the general case
	//  - HEAD-1:   So we don't do large reorgs if our HEAD becomes an uncle
	//  - HEAD-127: So we have a hard limit on the number of blocks reexecuted
	if !bc.cacheConfig.Disabled {
		triedb := bc.stateCache.TrieDB()

		for _, offset := range []uint64{0, 1, triesInMemory - 1} {
			if number := uint64(bc.CurrentBlock().GetHeight()); number > offset {
				recent, _ := bc.GetBlockByHeight(uint32(number - offset))

				log.Info("Writing cached state to disk", "block", uint32(recent.GetHeight()), "hash", recent.GetBlockID(), "root", recent.GetStatus())
				if err := triedb.Commit(*recent.GetStatus(), true); err != nil {
					log.Error("Failed to commit recent state trie", "err", err)
				}
			}
		}
		for !bc.triegc.Empty() {
			triedb.Dereference(bc.triegc.PopItem().(math.Hash), math.Hash{})
		}
		if size := triedb.Size(); size != 0 {
			log.Error("Dangling trie nodes after full cleanup")
		}
	}
	log.Info("Blockchain manager stopped")
}

func (bc *BlockChain) procFutureBlocks() {
	blocks := make([]*meta.Block, 0, bc.futureBlocks.Len())
	for _, hash := range bc.futureBlocks.Keys() {
		if block, exist := bc.futureBlocks.Peek(hash); exist {
			blocks = append(blocks, block.(*meta.Block))
		}
	}
	if len(blocks) > 0 {
		meta.BlockBy(meta.Number).Sort(blocks)

		// Insert one by one as chain insertion needs contiguous ancestry between blocks
		for i := range blocks {
			bc.ProcessBlock(blocks[i])
		}
	}
}

// WriteStatus status of write
type WriteStatus byte

const (
	NonStatTy WriteStatus = iota
	CanonStatTy
	SideStatTy
)

// Rollback is designed to remove a chain of links from the database that aren't
// certain enough to be valid.
func (bc *BlockChain) Rollback(chain []math.Hash) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for i := len(chain) - 1; i >= 0; i-- {
		hash := chain[i]

		if currentFastBlock := bc.CurrentFastBlock(); currentFastBlock.GetBlockID().IsEqual(&hash) {
			newFastBlock := bc.GetBlock(*currentFastBlock.GetPrevBlockID(), uint64(currentFastBlock.GetHeight())-1)
			bc.currentFastBlock.Store(newFastBlock)
			storage.WriteHeadFastBlockHash(bc.db, *newFastBlock.GetBlockID())
		}
		if currentBlock := bc.CurrentBlock(); currentBlock.GetBlockID().IsEqual(&hash) {
			newBlock := bc.GetBlock(*currentBlock.GetPrevBlockID(), uint64(currentBlock.GetHeight())-1)
			bc.currentBlock.Store(newBlock)
			storage.WriteHeadBlockHash(bc.db, *newBlock.GetBlockID())
			bc.SetCurrentBlockHead(newBlock)
		}
	}
}

var lastWrite uint64

// WriteBlockWithoutState writes only the block and its metadata to the database,
// but does not write any state. This is used to construct competing side forks
// up to the point where they exceed the canonical total difficulty.
func (bc *BlockChain) WriteBlockWithoutState(block *meta.Block) (err error) {
	bc.wg.Add(1)
	defer bc.wg.Done()

	if err := storage.WriteBlock(bc.db, block); err != nil {
		return err
	}
	return nil
}

// GetReceiptsByHash retrieves the receipts for all transactions in a given block.
func (bc *BlockChain) GetReceiptsByHash(hash math.Hash) core.Receipts {
	if receipts, ok := bc.receiptsCache.Get(hash); ok {
		return receipts.(core.Receipts)
	}
	number := storage.GetBlockNumber(bc.db, hash)
	if number == storage.MissingNumber {
		return nil
	}
	receipts := storage.ReadReceipts(bc.db, hash, number)
	bc.receiptsCache.Add(hash, receipts)
	return receipts
}

// WriteBlockWithState writes the block and all associated state to the database.
func (bc *BlockChain) WriteBlockWithState(block *meta.Block, state *state.StateDB, results []interpreter.Result) (status WriteStatus, err error) {
	bc.wg.Add(1)
	defer bc.wg.Done()

	// Make sure no inconsistent state is leaked during insertion
	bc.mu.Lock()
	defer bc.mu.Unlock()

	currentBlock := bc.CurrentBlock()
	localHeight := uint64(currentBlock.GetHeight())
	externHeight := uint64(block.GetHeight())
	// Write other block data using a batch.
	batch := bc.db.NewBatch()
	if err := storage.WriteBlock(batch, block); err != nil {
		return NonStatTy, err
	}
	root, err := state.Commit()
	if err != nil {
		return NonStatTy, err
	}
	triedb := bc.stateCache.TrieDB()

	// If we're running an archive node, always flush
	if bc.cacheConfig.Disabled {
		if err := triedb.Commit(root, false); err != nil {
			return NonStatTy, err
		}
	} else {
		// Full but not archive node, do proper garbage collection
		triedb.Reference(root, math.Hash{}) // metadata reference to keep trie alive
		bc.triegc.Push(root, -float32(block.GetHeight()))

		if current := uint64(block.GetHeight()); current > triesInMemory {
			// Find the next state trie we need to commit
			newBlock, _ := bc.GetBlockByHeight(uint32(current - triesInMemory))
			chosen := uint64(newBlock.GetHeight())

			// Only write to disk if we exceeded our memory allowance *and* also have at
			// least a given number of tries gapped.
			var (
				size  = triedb.Size()
				limit = trie.StorageSize(bc.cacheConfig.TrieNodeLimit) * 1024 * 1024
			)
			if size > limit || bc.gcproc > bc.cacheConfig.TrieTimeLimit {
				// If we're exceeding limits but haven't reached a large enough memory gap,
				// warn the user that the system is becoming unstable.
				if chosen < lastWrite+triesInMemory {
					switch {
					case size >= 2*limit:
						log.Warn("State memory usage too high, committing", "size", size, "limit", limit, "optimum", float64(chosen-lastWrite)/triesInMemory)
					case bc.gcproc >= 2*bc.cacheConfig.TrieTimeLimit:
						log.Info("State in memory for too long, committing", "time", bc.gcproc, "allowance", bc.cacheConfig.TrieTimeLimit, "optimum", float64(chosen-lastWrite)/triesInMemory)
					}
				}
				// If optimum or critical limits reached, write to disk
				if chosen >= lastWrite+triesInMemory || size >= 2*limit || bc.gcproc >= 2*bc.cacheConfig.TrieTimeLimit {
					triedb.Commit(*newBlock.GetStatus(), true)
					lastWrite = chosen
					bc.gcproc = 0
				}
			}
			// Garbage collect anything below our required write retention
			for !bc.triegc.Empty() {
				root, number := bc.triegc.Pop()
				if uint64(-number) > chosen {
					bc.triegc.Push(root, number)
					break
				}
				triedb.Dereference(root.(math.Hash), math.Hash{})
			}
		}
	}

	// Write other block data using a batch.
	storage.WriteReceipts(batch, *block.GetBlockID(), uint64(block.GetHeight()), normal.GetReceiptsByResult(results))

	// If the externHeight is higher than our known, add it to the canonical chain
	reorg := externHeight > localHeight
	currentBlock = bc.CurrentBlock()
	if reorg {
		// Reorganise the chain if the parent is not the head block
		if !block.GetPrevBlockID().IsEqual(currentBlock.GetBlockID()) {
			if err := bc.reorg(currentBlock, block); err != nil {
				return NonStatTy, err
			}
		}
		// Write the positional metadata for transaction and receipt lookups
		if err := storage.WriteTxLookupEntries(batch, block); err != nil {
			return NonStatTy, err
		}

		status = CanonStatTy
	} else {
		status = SideStatTy
	}
	if err := batch.Write(); err != nil {
		return NonStatTy, err
	}

	// Set new head.
	if status == CanonStatTy {
		bc.insert(block)
	}
	bc.futureBlocks.Remove(*block.GetBlockID())
	return status, nil
}

// InsertChain attempts to insert the given batch of blocks in to the canonical
// chain or, otherwise, create a fork. If an error is returned it will return
// the index number of the failing block as well an error describing what went
// wrong.
//
// After insertion is done, all accumulated events will be fired.
func (bc *BlockChain) ProcessBlock(chain *meta.Block) error {
	if bc.HasBlock(*chain.GetBlockID()) {
		log.Trace("Block already exist, skip it", "hash", chain.GetBlockID())
		return nil
	}

	events, err := bc.insertChain(chain)
	bc.PostChainEvents(events)
	return err
}

// insertChain will execute the actual chain insertion and event aggregation. The
// only reason this method exists as a separate one is to make locking cleaner
// with deferred statements.
func (bc *BlockChain) insertChain(chain *meta.Block) ([]interface{}, error) {
	// Pre-checks passed, start the full block imports
	bc.wg.Add(1)
	defer bc.wg.Done()

	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()

	// A queued approach to delivering events. This is generally
	// faster than direct delivery and requires much less mutex
	// acquiring.
	var (
		stats     = insertStats{startTime: mclock.Now()}
		events    = make([]interface{}, 0, 1)
		lastCanon *meta.Block
	)

	err := bc.validator.ValidateBlockHeader(bc.engine, bc, chain)
	if err != nil {
		return events, err
	}

	// If the chain is terminating, stop processing blocks
	if atomic.LoadInt32(&bc.procInterrupt) == 1 {
		log.Debug("Premature abort during blocks processing")
		return events, nil
	}
	bstart := time.Now()
	err = bc.validator.ValidateBlockBody(bc.validator, bc, chain)
	switch {
	case err == consensus.ErrFutureBlock:
		// Allow up to MaxFuture second in the future blocks. If this limit is exceeded
		// the chain is discarded and processed at a later time if given.
		max := time.Now().Add(maxTimeFutureBlocks * time.Second)
		if chain.GetTime().After(max) {
			return events, fmt.Errorf("future block: %v > %v", chain.GetTime(), max)
		}
		bc.futureBlocks.Add(chain.GetBlockID(), chain)
		stats.queued++

	case err == consensus.ErrUnknownAncestor && bc.futureBlocks.Contains(chain.GetPrevBlockID()):
		bc.futureBlocks.Add(chain.GetBlockID(), chain)
		stats.queued++

	case err == consensus.ErrPrunedAncestor:
		// Block competing with the canonical chain, store in the db, but don't process
		// until the competitor TD goes above the canonical TD
		currentBlock := bc.CurrentBlock()
		localHeight := currentBlock.GetHeight()
		externHeight := chain.GetHeight()
		if localHeight > externHeight {
			if err = bc.WriteBlockWithoutState(chain); err != nil {
				return events, err
			}
			break
		}
		// Competitor chain beat canonical, gather all blocks from the common ancestor
		var winner []*meta.Block

		parent := bc.GetBlock(*chain.GetPrevBlockID(), uint64(chain.GetHeight()-1))
		for !bc.HasState(*parent.GetStatus()) {
			winner = append(winner, parent)
			parent = bc.GetBlock(*parent.GetPrevBlockID(), uint64(parent.GetHeight()-1))
		}
		for j := 0; j < len(winner)/2; j++ {
			winner[j], winner[len(winner)-1-j] = winner[len(winner)-1-j], winner[j]
		}
		// Import all the pruned blocks to make the state available
		bc.chainmu.Unlock()
		for j := 0; j < len(winner); j++ {
			evs, err := bc.insertChain(winner[j])
			events = evs
			if err != nil {
				return events, err
			}
		}
		bc.chainmu.Lock()

	case err != nil:
		bc.reportBlock(chain, err)
		return events, err
	}
	// Create a new statedb using the parent block and report an
	// error if it fails.
	parentHash := math.Hash{}
	if !chain.IsGensis() {
		prevBlock, err := bc.GetBlockByID(*chain.GetPrevBlockID())
		if err != nil {
			log.Error("Get block by hash failed", "prev block id", chain.GetPrevBlockID())
			return events, err
		}
		parentHash.SetBytes(prevBlock.Header.Status.CloneBytes())
	}

	state, err := state.New(parentHash, bc.db)
	if err != nil {
		return events, err
	}
	// Process block using the parent state as reference point.
	err, results := bc.processor.ProcessBlockState(chain, state, bc, bc.validator)
	if err != nil {
		bc.reportBlock(chain, err)
		return events, err
	}
	proctime := time.Since(bstart)
	// Write the block to the chain and get the status.
	status, err := bc.WriteBlockWithState(chain, state, results)
	if err != nil {
		return events, err
	}
	switch status {
	case CanonStatTy:
		log.Info("Inserted new block", "number", chain.GetHeight(), "hash", chain.GetBlockID(),
			"txs", len(chain.GetTxs())) //, "elapsed", common.PrettyDuration(time.Since(bstart)))

		events = append(events, meta.ChainEvent{chain, *chain.GetBlockID()})
		lastCanon = chain

		// Only count canonical blocks for GC processing time
		bc.gcproc += proctime

	case SideStatTy:
		log.Info("Inserted forked block", "number", chain.GetHeight(), "hash", chain.GetBlockID(),
			//"elapsed", common.PrettyDuration(time.Since(bstart)),
			"txs", len(chain.GetTxs()))

		events = append(events, meta.ChainSideEvent{chain})
	}
	stats.processed++
	stats.report(chain, bc.stateCache.TrieDB().Size())

	// Append a single chain head event if we've progressed the chain
	if lastCanon != nil && bc.CurrentBlock().GetBlockID() == lastCanon.GetBlockID() {
		events = append(events, meta.ChainHeadEvent{lastCanon})
	}
	return events, nil
}

func (bc *BlockChain) executeBlock(block *meta.Block) (error, []interpreter.Result, math.Hash, *meta.Amount) {
	prevBlock, err := bc.GetBlockByID(*block.GetPrevBlockID())
	if err != nil {
		return err, nil, math.Hash{}, nil
	}
	stateDb, err := bc.StateAt(prevBlock.Header.Status)
	if err != nil {
		return err, nil, math.Hash{}, nil
	}

	// Process block using the parent state as reference point.
	return bc.processor.ExecuteBlockState(block, stateDb, bc, bc.validator)
}

// insertStats tracks and reports on block insertion.
type insertStats struct {
	queued, processed, ignored int
	startTime                  mclock.AbsTime
}

// statsReportLimit is the time limit during import after which we always print
// out progress. This avoids the user wondering what's going on.
const statsReportLimit = 8 * time.Second

// report prints statistics if some number of blocks have been processed
// or more than a few seconds have passed since the last message.
func (st *insertStats) report(chain *meta.Block, cache trie.StorageSize) {
	// Fetch the timings for the batch
	var (
		now     = mclock.Now()
		elapsed = time.Duration(now) - time.Duration(st.startTime)
	)
	// If we're at the last block of the batch or report period reached, log
	if elapsed >= statsReportLimit {
		var (
			end = chain
			txs = countTransactions(chain)
		)
		context := []interface{}{
			"blocks", st.processed, "txs", txs,
			"elapsed", elapsed,
			"number", end.GetHeight(), "hash", end.GetBlockID(), "cache", cache,
		}
		if st.queued > 0 {
			context = append(context, []interface{}{"queued", st.queued}...)
		}
		if st.ignored > 0 {
			context = append(context, []interface{}{"ignored", st.ignored}...)
		}
		log.Info("Imported new chain segment", context...)

		*st = insertStats{startTime: now}
	}
}

func countTransactions(chain *meta.Block) (c int) {
	return len(chain.GetTxs())
}

// reorgs takes two blocks, an old chain and a new chain and will reconstruct the blocks and inserts them
// to be part of the new canonical chain and accumulates potential missing transactions and post an
// event about them
func (bc *BlockChain) reorg(oldBlock, newBlock *meta.Block) error {
	var (
		newChain    meta.Blocks
		oldChain    meta.Blocks
		commonBlock *meta.Block
		deletedTxs  []meta.Transaction
		// collectLogs collects the logs that were generated during the
		// processing of the block that corresponds with the given hash.
		// These logs are later announced as deleted.
	)

	// first reduce whoever is higher bound
	if oldBlock.GetHeight() > newBlock.GetHeight() {
		// reduce old chain
		for ; oldBlock != nil && oldBlock.GetHeight() != newBlock.GetHeight(); oldBlock = bc.GetBlock(*oldBlock.GetPrevBlockID(), uint64(oldBlock.GetHeight()-1)) {
			oldChain = append(oldChain, oldBlock)
			deletedTxs = append(deletedTxs, oldBlock.GetTxs()...)
		}
	} else {
		// reduce new chain and append new chain blocks for inserting later on
		for ; newBlock != nil && newBlock.GetHeight() != oldBlock.GetHeight(); newBlock = bc.GetBlock(*newBlock.GetPrevBlockID(), uint64(newBlock.GetHeight()-1)) {
			newChain = append(newChain, newBlock)
		}
	}
	if oldBlock == nil {
		return fmt.Errorf("Invalid old chain")
	}
	if newBlock == nil {
		return fmt.Errorf("Invalid new chain")
	}

	for {
		if oldBlock.GetBlockID() == newBlock.GetBlockID() {
			commonBlock = oldBlock
			break
		}

		oldChain = append(oldChain, oldBlock)
		newChain = append(newChain, newBlock)
		deletedTxs = append(deletedTxs, oldBlock.GetTxs()...)

		oldBlock, newBlock = bc.GetBlock(*oldBlock.GetPrevBlockID(), uint64(oldBlock.GetHeight()-1)), bc.GetBlock(*newBlock.GetPrevBlockID(), uint64(newBlock.GetHeight()-1))
		if oldBlock == nil {
			return fmt.Errorf("Invalid old chain")
		}
		if newBlock == nil {
			return fmt.Errorf("Invalid new chain")
		}
	}
	// Ensure the user sees large reorgs
	if len(oldChain) > 0 && len(newChain) > 0 {
		logFn := log.Debug
		if len(oldChain) > 63 {
			logFn = log.Warn
		}
		logFn("chain split detected", "number", commonBlock.GetHeight(), "hash", commonBlock.GetBlockID(),
			"drop", len(oldChain), "dropfrom", oldChain[0].GetBlockID(), "add", len(newChain), "addfrom", newChain[0].GetBlockID())
	} else {
		log.Error("Impossible reorg, please file an issue", "oldnum", oldBlock.GetHeight(), "oldhash", oldBlock.GetBlockID(), "newnum", newBlock.GetHeight(), "newhash", newBlock.GetBlockID())
	}
	// Insert the new chain, taking care of the proper incremental order
	var addedTxs []meta.Transaction
	for i := len(newChain) - 1; i >= 0; i-- {
		// insert the block in the canonical way, re-writing history
		bc.insert(newChain[i])
		// write lookup entries for hash based transaction/receipt searches
		if err := storage.WriteTxLookupEntries(bc.db, newChain[i]); err != nil {
			return err
		}
		addedTxs = append(addedTxs, newChain[i].GetTxs()...)
	}
	// calculate the difference between deleted and added transactions
	diff := meta.TxDifference(deletedTxs, addedTxs)
	// When transactions get deleted from the database that means the
	// receipts that were created in the fork must also be deleted
	for _, tx := range diff {
		// transaction := &tx
		storage.DeleteTxLookupEntry(bc.db, *tx.GetTxID())
	}

	if len(oldChain) > 0 {
		go func() {
			for _, block := range oldChain {
				bc.chainSideFeed.Send(meta.ChainSideEvent{Block: block})
			}
		}()
	}

	return nil
}

// PostChainEvents iterates over the events generated by a chain insertion and
// posts them into the event feed.
// TODO: Should not expose PostChainEvents. The chain events should be posted in WriteBlock.
func (bc *BlockChain) PostChainEvents(events []interface{}) {
	// post event logs for further processing

	for _, event := range events {
		switch ev := event.(type) {
		case meta.ChainEvent:
			bc.chainFeed.Send(ev)

		case meta.ChainHeadEvent:
			bc.chainHeadFeed.Send(ev)

		case meta.ChainSideEvent:
			bc.chainSideFeed.Send(ev)
		}
	}
}

func (bc *BlockChain) update() {
	futureTimer := time.NewTicker(5 * time.Second)
	defer futureTimer.Stop()
	for {
		select {
		case <-futureTimer.C:
			bc.procFutureBlocks()
		case <-bc.quit:
			return
		}
	}
}

func (bc *BlockChain) Engine() consensus.Engine {
	return bc.engine
}

func (bc *BlockChain) GetHeader(hash math.Hash, height uint64) *meta.BlockHeader {
	block := bc.GetBlock(hash, height)
	return &block.Header
}

// BadBlockArgs represents the entries in the list returned when bad blocks are queried.
type BadBlockArgs struct {
	Hash  math.Hash   `json:"hash"`
	Block *meta.Block `json:"block"`
}

// BadBlocks returns a list of the last 'bad blocks' that the client has seen on the network
func (bc *BlockChain) BadBlocks() ([]BadBlockArgs, error) {
	datas := make([]BadBlockArgs, 0, bc.badBlocks.Len())
	for _, hash := range bc.badBlocks.Keys() {
		if hdr, exist := bc.badBlocks.Peek(hash); exist {
			block := hdr.(*meta.Block)
			datas = append(datas, BadBlockArgs{*block.GetBlockID(), block})
		}
	}
	return datas, nil
}

// addBadBlock adds a bad block to the bad-block LRU cache
func (bc *BlockChain) addBadBlock(block *meta.Block) {
	bc.badBlocks.Add(block.GetBlockID(), block)
}

// reportBlock logs a bad block error.
func (bc *BlockChain) reportBlock(block *meta.Block, err error) {
	bc.addBadBlock(block)

	log.Error(fmt.Sprintf(`
########## BAD BLOCK #########
chain config: %v

Number: %v
Hash: 0x%x

Error: %v
##############################
`, bc.chainConfig, block.GetHeight(), block.GetBlockID(), err))
}

// Config retrieves the blockchain's chain configuration.
func (bc *BlockChain) Config() *config.ChainConfig { return bc.chainConfig }

// SubscribeChainEvent registers a subscription of ChainEvent.
func (bc *BlockChain) SubscribeChainEvent(ch chan<- meta.ChainEvent) event.Subscription {
	return bc.scope.Track(bc.chainFeed.Subscribe(ch))
}

// SubscribeChainHeadEvent registers a subscription of ChainHeadEvent.
func (bc *BlockChain) SubscribeChainHeadEvent(ch chan<- meta.ChainHeadEvent) event.Subscription {
	return bc.scope.Track(bc.chainHeadFeed.Subscribe(ch))
}

// SubscribeChainSideEvent registers a subscription of ChainSideEvent.
func (bc *BlockChain) SubscribeChainSideEvent(ch chan<- meta.ChainSideEvent) event.Subscription {
	return bc.scope.Track(bc.chainSideFeed.Subscribe(ch))
}
