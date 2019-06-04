package chain

import (
	"errors"
	"fmt"
	"github.com/mihongtech/linkchain/node/bcsi"

	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mihongtech/linkchain/common/lcdb"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/common/util/event"
	"github.com/mihongtech/linkchain/common/util/log"
	"github.com/mihongtech/linkchain/core/meta"
	"github.com/mihongtech/linkchain/node/chain/storage"
	"github.com/mihongtech/linkchain/node/config"
	"github.com/mihongtech/linkchain/node/consensus"

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
// that's resident in a chain.
type CacheConfig struct {
	Disabled      bool          // Whether to disable trie write caching (archive node)
	TrieNodeLimit int           // Memory limit (MB) at which to flush the current in-memory trie to disk
	TrieTimeLimit time.Duration // Time limit after which to flush the current in-memory trie to disk
}

// ChainImpl represents the canonical chain given a database with a genesis
// block. The Blockchain manages chain imports, reverts, chain reorganisations.
//
// Importing blocks in to the block chain happens according to the set of rules
// defined by the two stage Validator. Processing of blocks is done using the
// Processor which processes the included transaction. The validation of the
// is done in the second part of the Validator. Failing results in aborting of
// the import.
//
// The ChainImpl also helps in returning blocks from **any** chain included
// in the database as well as blocks that represents the canonical chain. It's
// important to note that GetBlock can return any block and does not need to be
// included in the canonical one where as GetBlockByHeight always represents the
// canonical chain.
type ChainImpl struct {
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
	chainmu sync.RWMutex // chain insertion lock
	procmu  sync.RWMutex // block processor lock

	checkpoint       int          // checkpoint counts towards the new checkpoint
	currentBlock     atomic.Value // Current head of the block chain
	currentFastBlock atomic.Value // Current head of the fast-sync chain (may be above the block chain!)
	currentBlockHash math.Hash    // Hash of the current head of the header chain (prevent recomputing all the time)

	blockCache    *lru.Cache // Cache for the most recent entire blocks
	receiptsCache *lru.Cache // Cache for the most recent receipts per block
	futureBlocks  *lru.Cache // future blocks are blocks added for later processing
	numberCache   *lru.Cache // Cache for the most recent block numbers

	quit    chan struct{} // chain quit channel
	running int32         // running must be called atomically
	// procInterrupt must be atomically called
	procInterrupt int32          // interrupt signaler for block processing
	wg            sync.WaitGroup // chain processing wait group for shutting down

	bcsiAPI bcsi.BCSI //bcsi API

	badBlocks *lru.Cache // Bad block cache

	engine consensus.Engine //Block chain engine
}

// NewBlockChain returns a fully initialised block chain using information
// available in the database. It initialises the default Ethereum Validator and
// Processor.
func NewBlockChain(db lcdb.Database, genesisHash math.Hash, cacheConfig *CacheConfig, chainConfig *config.ChainConfig, bcsiAPI bcsi.BCSI, engine consensus.Engine) (*ChainImpl, error) {
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
	bc := &ChainImpl{
		chainConfig:   chainConfig,
		cacheConfig:   cacheConfig,
		db:            db,
		triegc:        prque.New(),
		quit:          make(chan struct{}),
		blockCache:    blockCache,
		futureBlocks:  futureBlocks,
		badBlocks:     badBlocks,
		numberCache:   numberCache,
		receiptsCache: receiptsCache,
		engine:        engine,
	}
	bc.bcsiAPI = bcsiAPI

	bc.genesisBlock, _ = bc.GetBlockByHeight(0)
	if bc.genesisBlock == nil || !bc.genesisBlock.GetBlockID().IsEqual(&genesisHash) {
		return nil, ErrNoGenesis
	}
	if err := bc.loadLastChain(); err != nil {
		return nil, err
	}
	bc.SetCurrentBlockHead(bc.CurrentBlock())

	go bc.update()
	return bc, nil
}

func (bc *ChainImpl) GetChainID() *big.Int {
	return bc.chainConfig.ChainId
}

func (bc *ChainImpl) getProcInterrupt() bool {
	return atomic.LoadInt32(&bc.procInterrupt) == 1
}

// loadLastChain loads the last known chain from the database. This method
// assumes that the chain manager mutex is held.
func (bc *ChainImpl) loadLastChain() error {
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

	//TODO the core best block should not be influence app best status.
	if err := bc.bcsiAPI.UpdateChain(currentBlock); err != nil {
		return err
	}
	return nil
}

// SetHead rewinds the local chain to a new head. In the case of headers, everything
// above the new head will be deleted and the new one set. In the case of blocks
// though, the head may be further rewound if block bodies are missing (non-archive
// nodes after a fast sync).
func (bc *ChainImpl) SetHead(head uint64) error {
	log.Warn("Rewinding chain", "target", head)

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

	// If either blocks reached nil, reset to the genesis
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
	return bc.loadLastChain()
}

// CurrentBlock retrieves the current head block of the canonical chain. The
// block is retrieved from the chain's internal cache.
func (bc *ChainImpl) CurrentBlock() *meta.Block {
	log.Debug("current is", "block", bc.currentBlock.Load().(*meta.Block))
	return bc.currentBlock.Load().(*meta.Block)
}

// CurrentFastBlock retrieves the current fast-sync head block of the canonical
// chain. The block is retrieved from the chain's internal cache.
func (bc *ChainImpl) CurrentFastBlock() *meta.Block {
	return bc.currentFastBlock.Load().(*meta.Block)
}

// Reset purges the entire chain, restoring it to its genesis .
func (bc *ChainImpl) Reset() error {
	return bc.ResetWithGenesisBlock(bc.genesisBlock)
}

// ResetWithGenesisBlock purges the entire chain, restoring it to the
// specified genesis .
func (bc *ChainImpl) ResetWithGenesisBlock(genesis *meta.Block) error {
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

// insert injects a new head block into the current block chain. This method
// assumes that the block is indeed a true head. It will also reset the head
// header and the head fast sync block to this very same block if they are older
// or if they are on a different side chain.
//
// Note, this function assumes that the `mu` mutex is held!
func (bc *ChainImpl) insert(block *meta.Block) {
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
func (bc *ChainImpl) Genesis() *meta.Block {
	return bc.genesisBlock
}

// HasBlock checks if a block is fully present in the database or not.
func (bc *ChainImpl) HasBlockAndNum(hash math.Hash, number uint64) bool {
	if bc.blockCache.Contains(hash) {
		return true
	}
	ok := storage.HasBlock(bc.db, hash, number)
	return ok
}

func (bc *ChainImpl) HasBlock(hash meta.BlockID) bool {
	if bc.blockCache.Contains(hash) {
		return true
	}
	ok := storage.HasBlock(bc.db, hash, bc.GetBlockNumber(hash))
	return ok
}

func (bc *ChainImpl) SetCurrentBlockHead(head *meta.Block) {
	bc.currentBlockHash = *head.GetBlockID()
}

// GetBlock retrieves a block from the database by hash and number,
// caching it if found.
func (bc *ChainImpl) GetBlock(hash math.Hash, number uint64) *meta.Block {
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
func (bc *ChainImpl) GetBlockByID(hash meta.BlockID) (*meta.Block, error) {
	block := bc.GetBlock(hash, bc.GetBlockNumber(hash))
	if block == nil {
		return block, errors.New("GetBlockByID:block not found")
	}
	return block, nil
}

func (bc *ChainImpl) GetBlockNumber(hash math.Hash) uint64 {
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
func (bc *ChainImpl) GetBlockByHeight(height uint32) (*meta.Block, error) {
	hash := storage.GetCanonicalHash(bc.db, uint64(height))
	if hash == (math.Hash{}) {
		return nil, errors.New("block not found")
	}
	return bc.GetBlock(hash, uint64(height)), nil
}

// Stop stops the chain service. If any imports are currently in progress
// it will abort them using the procInterrupt.
func (bc *ChainImpl) Stop() {
	if !atomic.CompareAndSwapInt32(&bc.running, 0, 1) {
		return
	}
	// Unsubscribe all subscriptions registered from chain
	bc.scope.Close()
	close(bc.quit)
	atomic.StoreInt32(&bc.procInterrupt, 1)

	bc.wg.Wait()

}

func (bc *ChainImpl) procFutureBlocks() {
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
func (bc *ChainImpl) Rollback(chain []math.Hash) {
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
// but does not write any. This is used to construct competing side forks
// up to the point where they exceed the canonical total difficulty.
func (bc *ChainImpl) WriteBlockWithoutState(block *meta.Block) (err error) {
	bc.wg.Add(1)
	defer bc.wg.Done()

	if err := storage.WriteBlock(bc.db, block); err != nil {
		return err
	}
	return nil
}

// WriteBlockWithState writes the block and all associated state to the database.
func (bc *ChainImpl) WriteBlockWithState(block *meta.Block) (status WriteStatus, err error) {
	bc.wg.Add(1)
	defer bc.wg.Done()

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

	if err := bc.bcsiAPI.Commit(*block.GetBlockID()); err != nil {
		return NonStatTy, err
	}

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
func (bc *ChainImpl) ProcessBlock(chain *meta.Block) error {
	if bc.HasBlock(*chain.GetBlockID()) {
		log.Trace("Block already exist, skip it", "hash", chain.GetBlockID())
		return nil
	}

	if err := bc.engine.ProcessBlock(chain); err != nil {
		return err
	}

	events, err := bc.insertChain(chain)
	bc.PostChainEvents(events)
	return err
}

// insertChain will execute the actual chain insertion and event aggregation. The
// only reason this method exists as a separate one is to make locking cleaner
// with deferred statements.
func (bc *ChainImpl) insertChain(chain *meta.Block) ([]interface{}, error) {
	// Pre-checks passed, start the full block imports
	bc.wg.Add(1)
	defer bc.wg.Done()

	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()

	// A queued approach to delivering events. This is generally
	// faster than direct delivery and requires much less mutex
	// acquiring.
	var (
		events    = make([]interface{}, 0, 1)
		lastCanon *meta.Block
	)

	err := bc.CheckBlock(chain)
	if err != nil {
		return events, err
	}

	// If the chain is terminating, stop processing blocks
	if atomic.LoadInt32(&bc.procInterrupt) == 1 {
		log.Debug("Premature abort during blocks processing")
		return events, nil
	}
	err = bc.CheckBlock(chain)
	switch {
	case err == consensus.ErrFutureBlock:
		// Allow up to MaxFuture second in the future blocks. If this limit is exceeded
		// the chain is discarded and processed at a later time if given.
		max := time.Now().Add(maxTimeFutureBlocks * time.Second)
		if chain.GetTime().After(max) {
			return events, fmt.Errorf("future block: %v > %v", chain.GetTime(), max)
		}
		bc.futureBlocks.Add(chain.GetBlockID(), chain)

	case err == consensus.ErrUnknownAncestor && bc.futureBlocks.Contains(chain.GetPrevBlockID()):
		bc.futureBlocks.Add(chain.GetBlockID(), chain)

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
		//TODO don't understand the code.

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

	// BCSI:Process block to app

	if err = bc.bcsiAPI.ProcessBlock(chain); err != nil {
		bc.reportBlock(chain, err)
		return events, err
	}
	// Write the block to the chain and get the status.
	status, err := bc.WriteBlockWithState(chain)
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

	case SideStatTy:
		log.Info("Inserted forked block", "number", chain.GetHeight(), "hash", chain.GetBlockID(),
			//"elapsed", common.PrettyDuration(time.Since(bstart)),
			"txs", len(chain.GetTxs()))

		events = append(events, meta.ChainSideEvent{chain})
	}

	// Append a single chain head event if we've progressed the chain
	if lastCanon != nil && bc.CurrentBlock().GetBlockID() == lastCanon.GetBlockID() {
		events = append(events, meta.ChainHeadEvent{lastCanon})
	}
	return events, nil
}

// reorgs takes two blocks, an old chain and a new chain and will reconstruct the blocks and inserts them
// to be part of the new canonical chain and accumulates potential missing transactions and post an
// event about them
func (bc *ChainImpl) reorg(oldBlock, newBlock *meta.Block) error {
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
func (bc *ChainImpl) PostChainEvents(events []interface{}) {
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

func (bc *ChainImpl) update() {
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

func (bc *ChainImpl) Engine() consensus.Engine {
	return bc.engine
}

func (bc *ChainImpl) GetHeader(hash math.Hash, height uint64) *meta.BlockHeader {
	block := bc.GetBlock(hash, height)
	return &block.Header
}

func (bc *ChainImpl) GetBestBlock() *meta.Block {
	return bc.CurrentBlock()
}

//CheckBlock Check block consensus,chain,app
func (bc *ChainImpl) CheckBlock(block *meta.Block) error {
	//log.Info("POA checkBlock ...")

	//Consensus
	if err := bc.engine.CheckBlock(block); err != nil {
		return err
	}

	//App
	if err := bc.bcsiAPI.CheckBlock(block); err != nil {
		return err
	}

	prevBlock, err := bc.GetBlockByID(*block.GetPrevBlockID())

	if err != nil {
		log.Error("BlockManage", "checkBlock", err)
		return err
	}

	if prevBlock.GetHeight()+1 != block.GetHeight() {
		log.Error("BlockManage", "checkBlock", "current block height is error")
		return errors.New("Check block height failed")
	}

	return nil
}

// BadBlockArgs represents the entries in the list returned when bad blocks are queried.
type BadBlockArgs struct {
	Hash  math.Hash   `json:"hash"`
	Block *meta.Block `json:"block"`
}

// BadBlocks returns a list of the last 'bad blocks' that the client has seen on the network
func (bc *ChainImpl) BadBlocks() ([]BadBlockArgs, error) {
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
func (bc *ChainImpl) addBadBlock(block *meta.Block) {
	bc.badBlocks.Add(block.GetBlockID(), block)
}

// reportBlock logs a bad block error.
func (bc *ChainImpl) reportBlock(block *meta.Block, err error) {
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

// Config retrieves the chain's chain configuration.
func (bc *ChainImpl) GetChainConfig() *config.ChainConfig { return bc.chainConfig }

// SubscribeChainEvent registers a subscription of ChainEvent.
func (bc *ChainImpl) SubscribeChainEvent(ch chan<- meta.ChainEvent) event.Subscription {
	return bc.scope.Track(bc.chainFeed.Subscribe(ch))
}

// SubscribeChainHeadEvent registers a subscription of ChainHeadEvent.
func (bc *ChainImpl) SubscribeChainHeadEvent(ch chan<- meta.ChainHeadEvent) event.Subscription {
	return bc.scope.Track(bc.chainHeadFeed.Subscribe(ch))
}

// SubscribeChainSideEvent registers a subscription of ChainSideEvent.
func (bc *ChainImpl) SubscribeChainSideEvent(ch chan<- meta.ChainSideEvent) event.Subscription {
	return bc.scope.Track(bc.chainSideFeed.Subscribe(ch))
}
