package fetcher

import (
	"errors"
	"math/rand"
	"time"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/consensus/manager"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

const (
	arriveTimeout = 500 * time.Millisecond // Time allowance before an announced block is explicitly requested
	gatherSlack   = 100 * time.Millisecond // Interval used to collate almost-expired announces with fetches
	fetchTimeout  = 5 * time.Second        // Maximum allotted time to return an explicitly requested block
	maxQueueDist  = 32                     // Maximum allowed distance from the chain head to queue
	hashLimit     = 256                    // Maximum number of unique blocks a peer may have announced
	blockLimit    = 64                     // Maximum number of unique blocks a peer may have delivered
)

var (
	errTerminated = errors.New("terminated")
)

// blockRetrievalFn is a callback type for retrieving a block from the local chain.
type blockRetrievalFn func(meta.DataID) (block.IBlock, error)

// blockRequesterFn is a callback type for sending a block retrieval request.
type blockRequesterFn func(meta.DataID) error

// blockVerifierFn is a callback type to verify a block for fast propagation.
type blockVerifierFn func(block block.IBlock) bool

// blockBroadcasterFn is a callback type for broadcasting a block to connected peers.
type blockBroadcasterFn func(block block.IBlock, propagate bool)

// chainHeightFn is a callback type to retrieve the current chain height.
type chainHeightFn func() uint64

// peerDropFn is a callback type for dropping a peer detected as malicious.
type peerDropFn func(id string)

// announce is the hash notification of the availability of a new block in the
// network.
type announce struct {
	hash   meta.DataID  // Hash of the block being announced
	number uint64       // Number of the block being announced (0 = unknown | old protocol)
	b      block.IBlock // Header of the block partially reassembled (new protocol)
	time   time.Time    // Timestamp of the announcement

	origin string // Identifier of the peer originating the notification

	fetchBlock blockRequesterFn // Fetcher function to retrieve the header of an announced block
}

// headerFilterTask represents a batch of headers needing fetcher filtering.
type blockFilterTask struct {
	peer   string         // The source peer of block headers
	blocks []block.IBlock // Collection of headers to filter
	time   time.Time      // Arrival time of the headers
}

// inject represents a schedules import operation.
type inject struct {
	origin string
	block  block.IBlock
}

// Fetcher is responsible for accumulating block announcements from various peers
// and scheduling them for retrieval.
type Fetcher struct {
	// Various event channels
	notify chan *announce
	inject chan *inject

	blockFilter chan chan *blockFilterTask

	done chan meta.DataID
	quit chan struct{}

	// Announce states
	announces  map[string]int              // Per peer announce counts to prevent memory exhaustion
	announced  map[meta.DataID][]*announce // Announced blocks, scheduled for fetching
	fetching   map[meta.DataID]*announce   // Announced blocks, currently fetching
	fetched    map[meta.DataID][]*announce // Blocks with headers fetched, scheduled for body retrieval
	completing map[meta.DataID]*announce   // Blocks with headers, currently body-completing

	// Block cache
	queue  *prque.Prque            // Queue containing the import operations (block number sorted)
	queues map[string]int          // Per peer block counts to prevent memory exhaustion
	queued map[meta.DataID]*inject // Set of already queued blocks (to dedup imports)

	// Callbacks
	getBlock       blockRetrievalFn   // Retrieves a block from the local chain
	verifyBlock    blockVerifierFn    // Checks if a block's headers have a valid proof of work
	broadcastBlock blockBroadcasterFn // Broadcasts a block to connected peers
	chainHeight    chainHeightFn      // Retrieves the current chain's height
	insertChain    manager.BlockManager
	dropPeer       peerDropFn // Drops a peer for misbehaving
}

// New creates a block fetcher to retrieve blocks based on hash announcements.
func New(getBlock blockRetrievalFn, verifyBlock blockVerifierFn, broadcastBlock blockBroadcasterFn, chainHeight chainHeightFn, insertChain manager.BlockManager, dropPeer peerDropFn) *Fetcher {
	return &Fetcher{
		notify:         make(chan *announce),
		inject:         make(chan *inject),
		blockFilter:    make(chan chan *blockFilterTask),
		done:           make(chan meta.DataID),
		quit:           make(chan struct{}),
		announces:      make(map[string]int),
		announced:      make(map[meta.DataID][]*announce),
		fetching:       make(map[meta.DataID]*announce),
		fetched:        make(map[meta.DataID][]*announce),
		completing:     make(map[meta.DataID]*announce),
		queue:          prque.New(),
		queues:         make(map[string]int),
		queued:         make(map[meta.DataID]*inject),
		getBlock:       getBlock,
		verifyBlock:    verifyBlock,
		broadcastBlock: broadcastBlock,
		chainHeight:    chainHeight,
		insertChain:    insertChain,
		dropPeer:       dropPeer,
	}
}

// Start boots up the announcement based synchroniser, accepting and processing
// hash notifications and block fetches until termination requested.
func (f *Fetcher) Start() {
	go f.loop()
}

// Stop terminates the announcement based synchroniser, canceling all pending
// operations.
func (f *Fetcher) Stop() {
	close(f.quit)
}

// Notify announces the fetcher of the potential availability of a new block in
// the network.
func (f *Fetcher) Notify(peer string, hash meta.DataID, number uint64, time time.Time,
	blockFetcher blockRequesterFn) error {
	block := &announce{
		hash:       hash,
		number:     number,
		time:       time,
		origin:     peer,
		fetchBlock: blockFetcher,
	}
	select {
	case f.notify <- block:
		return nil
	case <-f.quit:
		return errTerminated
	}
}

// Enqueue tries to fill gaps the the fetcher's future import queue.
func (f *Fetcher) Enqueue(peer string, block block.IBlock) error {
	op := &inject{
		origin: peer,
		block:  block,
	}
	select {
	case f.inject <- op:
		return nil
	case <-f.quit:
		return errTerminated
	}
}

// FilterHeaders extracts all the headers that were explicitly requested by the fetcher,
// returning those that should be handled differently.
func (f *Fetcher) FilterBlocks(peer string, blocks []block.IBlock, time time.Time) []block.IBlock {
	log.Trace("Filtering blocks", "peer", peer, "blocks", len(blocks))

	// Send the filter channel to the fetcher
	filter := make(chan *blockFilterTask)

	select {
	case f.blockFilter <- filter:
	case <-f.quit:
		return nil
	}
	// Request the filtering of the header list
	select {
	case filter <- &blockFilterTask{peer: peer, blocks: blocks, time: time}:
	case <-f.quit:
		return nil
	}
	// Retrieve the headers remaining after filtering
	select {
	case task := <-filter:
		return task.blocks
	case <-f.quit:
		return nil
	}
}

// Loop is the main fetcher loop, checking and processing various notification
// events.
func (f *Fetcher) loop() {
	// Iterate the block fetching until a quit is requested
	fetchTimer := time.NewTimer(0)
	completeTimer := time.NewTimer(0)

	for {
		// Clean up any expired block fetches
		for hash, announce := range f.fetching {
			if time.Since(announce.time) > fetchTimeout {
				f.forgetHash(hash)
			}
		}
		// Import any queued blocks that could potentially fit
		height := f.chainHeight()
		for !f.queue.Empty() {
			op := f.queue.PopItem().(*inject)
			// If too high up the chain or phase, continue later
			number := uint64(op.block.GetHeight())
			if number > height+1 {
				f.queue.Push(op, -float32(op.block.GetHeight()))
				break
			}
			// Otherwise if fresh and still unknown, try and import
			hash := op.block.GetBlockID()
			block, err := f.getBlock(hash)
			if err != nil {
				if number < height || block != nil {
					f.forgetBlock(hash)
					continue
				}
			}
			f.insert(op.origin, op.block)
		}
		// Wait for an outside event to occur
		select {
		case <-f.quit:
			// Fetcher terminating, abort all operations
			return

		case notification := <-f.notify:
			// A block was announced, make sure the peer isn't DOSing us

			count := f.announces[notification.origin] + 1
			if count > hashLimit {
				log.Trace("Peer exceeded outstanding announces", "peer", notification.origin, "limit", hashLimit)
				break
			}
			// If we have a valid block number, check that it's potentially useful
			if notification.number > 0 {
				if dist := int64(notification.number) - int64(f.chainHeight()); dist > maxQueueDist {
					log.Trace("Peer discarded announcement", "peer", notification.origin, "number", notification.number, "hash", notification.hash, "distance", dist)
					break
				}
			}
			// All is well, schedule the announce if block's not yet downloading
			if _, ok := f.fetching[notification.hash]; ok {
				break
			}
			if _, ok := f.completing[notification.hash]; ok {
				break
			}
			f.announces[notification.origin] = count
			f.announced[notification.hash] = append(f.announced[notification.hash], notification)
			if len(f.announced) == 1 {
				f.rescheduleFetch(fetchTimer)
			}

		case op := <-f.inject:
			// A direct block insertion was requested, try and fill any pending gaps
			f.enqueue(op.origin, op.block)

		case hash := <-f.done:
			// A pending import finished, remove all traces of the notification
			f.forgetHash(hash)
			f.forgetBlock(hash)

		case <-fetchTimer.C:
			// At least one block's timer ran out, check for needing retrieval
			request := make(map[string][]meta.DataID)

			for hash, announces := range f.announced {
				if time.Since(announces[0].time) > arriveTimeout-gatherSlack {
					// Pick a random peer to retrieve from, reset all others
					announce := announces[rand.Intn(len(announces))]
					f.forgetHash(hash)

					// If the block still didn't arrive, queue for fetching
					block, err := f.getBlock(hash)
					if block == nil || err != nil {
						request[announce.origin] = append(request[announce.origin], hash)
						f.fetching[hash] = announce
					}
				}
			}
			// Send out all block header requests
			for peer, hashes := range request {
				log.Trace("Fetching scheduled headers", "peer", peer, "list", hashes)

				// Create a closure of the fetch and schedule in on a new thread
				fetchBlock, hashes := f.fetching[hashes[0]].fetchBlock, hashes
				go func() {
					for _, hash := range hashes {
						fetchBlock(hash) // Suboptimal, but protocol doesn't allow batch header retrievals
					}
				}()
			}
			// Schedule the next fetch if blocks are still pending
			f.rescheduleFetch(fetchTimer)

		case <-completeTimer.C:
			// At least one header's timer ran out, retrieve everything
			request := make(map[string][]meta.DataID)

			for hash, announces := range f.fetched {
				// Pick a random peer to retrieve from, reset all others
				announce := announces[rand.Intn(len(announces))]
				f.forgetHash(hash)

				// If the block still didn't arrive, queue for completion
				block, err := f.getBlock(hash)
				if block == nil || err != nil {
					request[announce.origin] = append(request[announce.origin], hash)
					f.completing[hash] = announce
				}
			}
			// Send out all block body requests
			for peer, hashes := range request {
				log.Trace("Fetching scheduled bodies", "peer", peer, "list", hashes)
			}
			// Schedule the next fetch if blocks are still pending
			f.rescheduleComplete(completeTimer)

		case filter := <-f.blockFilter:
			log.Trace("blockFilter arrived", "filter", filter)
			var task *blockFilterTask
			select {
			case task = <-filter:
			case <-f.quit:
				return
			}

			// Split the batch of headers into unknown ones (to return to the caller),
			// known incomplete ones (requiring body retrievals) and completed blocks.
			unknown, incomplete, complete := []block.IBlock{}, []*announce{}, []block.IBlock{}
			for _, block := range task.blocks {
				hash := block.GetBlockID()

				// Filter fetcher-requested headers from other synchronisation algorithms
				if announce := f.fetching[hash]; announce != nil && announce.origin == task.peer && f.fetched[hash] == nil && f.completing[hash] == nil && f.queued[hash] == nil {
					// If the delivered header does not match the promised number, drop the announcer
					if uint64(block.GetHeight()) != announce.number {
						log.Trace("Invalid block number fetched", "peer", announce.origin, "hash", hash, "announced", announce.number, "provided", block.GetHeight())
						f.dropPeer(announce.origin)
						f.forgetHash(hash)
						continue
					}
					// Only keep if not imported by other means
					if bk, err := f.getBlock(hash); err != nil || bk == nil {
						announce.b = block
						announce.time = task.time

						complete = append(complete, block)
						f.completing[hash] = announce
						continue
						// Otherwise add to the list of blocks needing completion
						incomplete = append(incomplete, announce)
					} else {
						// log.Trace("Block already imported, discarding header", "peer", announce.origin, "number", header.Number, "hash", header.Hash())
						f.forgetHash(hash)
					}
				} else {
					// Fetcher doesn't know about it, add to the return list
					unknown = append(unknown, block)
				}
			}
			select {
			case filter <- &blockFilterTask{blocks: unknown, time: task.time}:
			case <-f.quit:
				return
			}
			// Schedule the retrieved headers for body completion
			for _, announce := range incomplete {
				hash := announce.b.GetBlockID()
				if _, ok := f.completing[hash]; ok {
					continue
				}
				f.fetched[hash] = append(f.fetched[hash], announce)
				if len(f.fetched) == 1 {
					f.rescheduleComplete(completeTimer)
				}
			}
			// Schedule the header-only blocks for import
			for _, block := range complete {
				if announce := f.completing[block.GetBlockID()]; announce != nil {
					f.enqueue(announce.origin, block)
				}
			}
		}
	}
}

// rescheduleFetch resets the specified fetch timer to the next announce timeout.
func (f *Fetcher) rescheduleFetch(fetch *time.Timer) {
	// Short circuit if no blocks are announced
	if len(f.announced) == 0 {
		return
	}
	// Otherwise find the earliest expiring announcement
	earliest := time.Now()
	for _, announces := range f.announced {
		if earliest.After(announces[0].time) {
			earliest = announces[0].time
		}
	}
	fetch.Reset(arriveTimeout - time.Since(earliest))
}

// rescheduleComplete resets the specified completion timer to the next fetch timeout.
func (f *Fetcher) rescheduleComplete(complete *time.Timer) {
	// Short circuit if no headers are fetched
	if len(f.fetched) == 0 {
		return
	}
	// Otherwise find the earliest expiring announcement
	earliest := time.Now()
	for _, announces := range f.fetched {
		if earliest.After(announces[0].time) {
			earliest = announces[0].time
		}
	}
	complete.Reset(gatherSlack - time.Since(earliest))
}

// enqueue schedules a new future import operation, if the block to be imported
// has not yet been seen.
func (f *Fetcher) enqueue(peer string, block block.IBlock) {
	hash := block.GetBlockID()

	// Ensure the peer isn't DOSing us
	count := f.queues[peer] + 1
	if count > blockLimit {
		log.Trace("Discarded propagated block, exceeded allowance", "peer", peer, "number", block.GetHeight(), "hash", hash, "limit", blockLimit)

		f.forgetHash(hash)
		return
	}
	// Discard any past or too distant blocks
	if dist := int64(block.GetHeight()) - int64(f.chainHeight()); dist > maxQueueDist {
		log.Trace("Discarded propagated block, too far away", "peer", peer, "number", block.GetHeight(), "hash", hash, "distance", dist)

		f.forgetHash(hash)
		return
	}
	// Schedule the block for future importing
	if _, ok := f.queued[hash]; !ok {
		op := &inject{
			origin: peer,
			block:  block,
		}
		f.queues[peer] = count
		f.queued[hash] = op
		f.queue.Push(op, -float32(block.GetHeight()))
		log.Trace("Queued propagated block", "peer", peer, "number", block.GetHeight(), "hash", hash, "queued", f.queue.Size())
	}
}

// insert spawns a new goroutine to run a block insertion into the chain. If the
// block's number is at the same height as the current import phase, it updates
// the phase states accordingly.
func (f *Fetcher) insert(peer string, block block.IBlock) {
	hash := block.GetBlockID()

	// Run the import on a new thread
	log.Trace("Importing propagated block", "peer", peer, "number", block.GetHeight(), "hash", hash)
	go func() {
		defer func() { f.done <- hash }()

		// If the parent's unknown, abort insertion
		parent, err := f.getBlock(block.GetPrevBlockID())
		if parent == nil || err != nil {
			log.Debug("Unknown parent of propagated block", "peer", peer, "number", block.GetHeight(), "hash", hash, "parent", block.GetPrevBlockID())
			return
		}
		// Quickly validate the block and propagate the block if it passes
		valid := f.verifyBlock(block)
		if valid {
			go f.broadcastBlock(block, true)
		} else {
			log.Debug("Propagated block verification failed", "peer", peer, "number", block.GetHeight(), "hash", hash)
			f.dropPeer(peer)
			return
		}
		log.Debug("insert block to chain", "block", block)
		if err := f.insertChain.ProcessBlock(block); err != nil {
			log.Debug("Propagated block import failed", "peer", peer, "number", block.GetHeight(), "hash", hash, "err", err)
			return
		}

		go f.broadcastBlock(block, false)
	}()
}

// forgetHash removes all traces of a block announcement from the fetcher's
// internal state.
func (f *Fetcher) forgetHash(hash meta.DataID) {
	// Remove all pending announces and decrement DOS counters
	for _, announce := range f.announced[hash] {
		f.announces[announce.origin]--
		if f.announces[announce.origin] == 0 {
			delete(f.announces, announce.origin)
		}
	}
	delete(f.announced, hash)
	// Remove any pending fetches and decrement the DOS counters
	if announce := f.fetching[hash]; announce != nil {
		f.announces[announce.origin]--
		if f.announces[announce.origin] == 0 {
			delete(f.announces, announce.origin)
		}
		delete(f.fetching, hash)
	}

	// Remove any pending completion requests and decrement the DOS counters
	for _, announce := range f.fetched[hash] {
		f.announces[announce.origin]--
		if f.announces[announce.origin] == 0 {
			delete(f.announces, announce.origin)
		}
	}
	delete(f.fetched, hash)

	// Remove any pending completions and decrement the DOS counters
	if announce := f.completing[hash]; announce != nil {
		f.announces[announce.origin]--
		if f.announces[announce.origin] == 0 {
			delete(f.announces, announce.origin)
		}
		delete(f.completing, hash)
	}
}

// forgetBlock removes all traces of a queued block from the fetcher's internal
// state.
func (f *Fetcher) forgetBlock(hash meta.DataID) {
	if insert := f.queued[hash]; insert != nil {
		f.queues[insert.origin]--
		if f.queues[insert.origin] == 0 {
			delete(f.queues, insert.origin)
		}
		delete(f.queued, hash)
	}
}
