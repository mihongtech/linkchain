package downloader

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/linkchain/common/util/log"
	"github.com/linkchain/meta"
	"github.com/linkchain/meta/block"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"
)

var (
	blockCacheItems      = 8192             // Maximum number of blocks to cache before throttling the download
	blockCacheMemory     = 64 * 1024 * 1024 // Maximum amount of memory to use for block caching
	blockCacheSizeWeight = 0.1              // Multiplier to approximate the average block size based on past ones
)

var (
	errNoFetchesPending = errors.New("no fetches pending")
	errStaleDelivery    = errors.New("stale delivery")
)

// fetchRequest is a currently running data retrieval operation.
type fetchRequest struct {
	Peer   *peerConnection // Peer to which the request was sent
	From   uint64
	Blocks []block.IBlock
	Time   time.Time // Time when the request was made
}

// fetchResult is a struct collecting partial results from data fetchers until
// all outstanding pieces complete and the result as a whole can be processed.
type fetchResult struct {
	Pending int         // Number of data fetches still pending
	Hash    meta.DataID // Hash of the block to prevent recalculating
	Block   block.IBlock
}
type StorageSize float64

// queue represents hashes that are either need fetching or are being fetched
type queue struct {
	mode SyncMode // Synchronisation mode to decide on the block parts to schedule for fetching

	// Blocks are "special", they download in batches, supported by a skeleton chain
	blockHead      meta.DataID                    // [eth/62] Hash of the last queued block to verify order
	blockTaskPool  map[uint64]block.IBlock        // [eth/62] Pending block retrieval tasks, mapping starting indexes to skeleton headers
	blockTaskQueue *prque.Prque                   // [eth/62] Priority queue of the skeleton indexes to fetch the filling headers for
	blockPeerMiss  map[string]map[uint64]struct{} // [eth/62] Set of per-peer block batches known to be unavailable
	blockPendPool  map[string]*fetchRequest       // [eth/62] Currently pending block retrieval operations
	blockResults   []block.IBlock                 // [eth/62] Result cache accumulating the completed headers
	blockProced    int                            // [eth/62] Number of headers already processed from the results
	blockOffset    uint64                         // [eth/62] Number of the first block in the result cache
	blockContCh    chan bool                      // [eth/62] Channel to notify when block download finishes

	resultCache  []*fetchResult // Downloaded but not yet delivered fetch results
	resultOffset uint64         // Offset of the first cached fetch result in the block chain
	resultSize   StorageSize    // Approximate size of a block (exponential moving average)

	lock   *sync.Mutex
	active *sync.Cond
	closed bool
}

// newQueue creates a new download queue for scheduling block retrieval.
func newQueue() *queue {
	lock := new(sync.Mutex)
	return &queue{
		blockPendPool: make(map[string]*fetchRequest),
		blockContCh:   make(chan bool),
		resultCache:   make([]*fetchResult, blockCacheItems),
		active:        sync.NewCond(lock),
		lock:          lock,
	}
}

// Reset clears out the queue contents.
func (q *queue) Reset() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.closed = false
	q.mode = FullSync

	q.blockHead = nil

	q.resultCache = make([]*fetchResult, blockCacheItems)
	q.resultOffset = 0
}

// Close marks the end of the sync, unblocking WaitResults.
// It may be called even if the queue is already closed.
func (q *queue) Close() {
	q.lock.Lock()
	q.closed = true
	q.lock.Unlock()
	q.active.Broadcast()
}

// PendingBlocks retrieves the number of block requests pending for retrieval.
func (q *queue) PendingBlocks() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.blockTaskQueue.Size()
}

// InFlightBlocks retrieves whether there are block fetch requests currently
// in flight.
func (q *queue) InFlightBlocks() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.blockPendPool) > 0
}

// Idle returns if the queue is fully idle or has some data still inside.
func (q *queue) Idle() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	queued := q.blockTaskQueue.Size()

	return queued == 0
}

// resultSlots calculates the number of results slots available for requests
// whilst adhering to both the item and the memory limit too of the results
// cache.
func (q *queue) resultSlots(pendPool map[string]*fetchRequest, donePool map[meta.DataID]struct{}) int {
	// Calculate the maximum length capped by the memory limit
	limit := len(q.resultCache)
	if StorageSize(len(q.resultCache))*q.resultSize > StorageSize(blockCacheMemory) {
		limit = int((StorageSize(blockCacheMemory) + q.resultSize - 1) / q.resultSize)
	}
	// Calculate the number of slots already finished
	finished := 0
	for _, result := range q.resultCache[:limit] {
		if result == nil {
			break
		}
		if _, ok := donePool[result.Hash]; ok {
			finished++
		}
	}
	// Calculate the number of slots currently downloading
	pending := 0
	for _, request := range pendPool {
		for _, block := range request.Blocks {
			if uint64(block.GetHeight()) < q.resultOffset+uint64(limit) {
				pending++
			}
		}
	}
	// Return the free slots to distribute
	return limit - finished - pending
}

// ScheduleSkeleton adds a batch of block retrieval tasks to the queue to fill
// up an already retrieved block skeleton.
func (q *queue) ScheduleSkeleton(from uint64, skeleton []block.IBlock) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// No skeleton retrieval can be in progress, fail hard if so (huge implementation bug)
	if q.blockResults != nil {
		panic("skeleton assembly already in progress")
	}
	// Shedule all the block retrieval tasks for the skeleton assembly
	q.blockTaskPool = make(map[uint64]block.IBlock)
	q.blockTaskQueue = prque.New()
	q.blockPeerMiss = make(map[string]map[uint64]struct{}) // Reset availability to correct invalid chains
	q.blockResults = make([]block.IBlock, len(skeleton)*MaxBlockFetch)
	q.blockProced = 0
	q.blockOffset = from
	q.blockContCh = make(chan bool, 1)

	for i, block := range skeleton {
		index := from + uint64(i*MaxBlockFetch)

		q.blockTaskPool[index] = block
		q.blockTaskQueue.Push(index, -float32(index))
	}
}

// RetrieveBlocks retrieves the block chain assemble based on the scheduled
// skeleton.
func (q *queue) RetrieveBlocks() ([]block.IBlock, int) {
	q.lock.Lock()
	defer q.lock.Unlock()

	blocks, proced := q.blockResults, q.blockProced
	q.blockResults, q.blockProced = nil, 0

	return blocks, proced
}

// Schedule adds a set of headers for the download queue for scheduling, returning
// the new headers encountered.
func (q *queue) Schedule(blocks []block.IBlock, from uint64) []block.IBlock {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Insert all the headers prioritised by the contained block number
	inserts := make([]block.IBlock, 0, len(blocks))
	for _, block := range blocks {
		// Make sure chain order is honoured and preserved throughout
		hash := block.GetBlockID()
		if uint64(block.GetHeight()) != from {
			log.Warn("Header broke chain ordering", "number", block.GetHeight(), "hash", hash, "expected", from)
			break
		}
		if q.blockHead != nil && !q.blockHead.IsEqual(block.GetPrevBlockID()) {
			log.Warn("Header broke chain ancestry", "number", block.GetHeight(), "hash", hash, "q.blockHead", q.blockHead, "block.GetPrevBlockID()", block.GetPrevBlockID())
			break
		}

		inserts = append(inserts, block)
		q.blockHead = hash
		from++
	}
	return inserts
}

// Results retrieves and permanently removes a batch of fetch results from
// the cache. the result slice will be empty if the queue has been closed.
func (q *queue) Results(block bool) []*fetchResult {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Count the number of items available for processing
	nproc := q.countProcessableItems()
	for nproc == 0 && !q.closed {
		if !block {
			return nil
		}
		q.active.Wait()
		nproc = q.countProcessableItems()
	}
	// Since we have a batch limit, don't pull more into "dangling" memory
	if nproc > maxResultsProcess {
		nproc = maxResultsProcess
	}
	results := make([]*fetchResult, nproc)
	copy(results, q.resultCache[:nproc])
	if len(results) > 0 {
		// Delete the results from the cache and clear the tail.
		copy(q.resultCache, q.resultCache[nproc:])
		for i := len(q.resultCache) - nproc; i < len(q.resultCache); i++ {
			q.resultCache[i] = nil
		}
		// Advance the expected block number of the first cache entry.
		q.resultOffset += uint64(nproc)

		// Recalculate the result item weights to prevent memory exhaustion
		//		for _, result := range results {
		//			size := result.Block.Size()
		//			q.resultSize = StorageSize(blockCacheSizeWeight)*size + (1-common.StorageSize(blockCacheSizeWeight))*q.resultSize
		//		}
	}
	return results
}

// countProcessableItems counts the processable items.
func (q *queue) countProcessableItems() int {
	for i, result := range q.resultCache {
		if result == nil || result.Pending > 0 {
			return i
		}
	}
	return len(q.resultCache)
}

// ReserveBlocks reserves a set of headers for the given peer, skipping any
// previously failed batches.
func (q *queue) ReserveBlocks(p *peerConnection, count int) *fetchRequest {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Short circuit if the peer's already downloading something (sanity check to
	// not corrupt state)
	if _, ok := q.blockPendPool[p.id]; ok {
		return nil
	}
	// Retrieve a batch of hashes, skipping previously failed ones
	send, skip := uint64(0), []uint64{}
	for send == 0 && !q.blockTaskQueue.Empty() {
		from, _ := q.blockTaskQueue.Pop()
		if q.blockPeerMiss[p.id] != nil {
			if _, ok := q.blockPeerMiss[p.id][from.(uint64)]; ok {
				skip = append(skip, from.(uint64))
				continue
			}
		}
		send = from.(uint64)
	}
	// Merge all the skipped batches back
	for _, from := range skip {
		q.blockTaskQueue.Push(from, -float32(from))
	}
	// Assemble and return the block download request
	if send == 0 {
		return nil
	}
	request := &fetchRequest{
		Peer: p,
		From: send,
		Time: time.Now(),
	}
	log.Info("start to ReserveBlocks", "id", p.id, "request", request)
	q.blockPendPool[p.id] = request
	return request
}

// reserveBlocks reserves a set of data download operations for a given peer,
// skipping any previously failed ones. This method is a generic version used
// by the individual special reservation functions.
//
// Note, this method expects the queue lock to be already held for writing. The
// reason the lock is not obtained in here is because the parameters already need
// to access the queue, so they already need a lock anyway.
func (q *queue) reserveBlocks(p *peerConnection, count int, taskPool map[meta.DataID]block.IBlock, taskQueue *prque.Prque,
	pendPool map[string]*fetchRequest, donePool map[meta.DataID]struct{}, isNoop func(block.IBlock) bool) (*fetchRequest, bool, error) {
	// Short circuit if the pool has been depleted, or if the peer's already
	// downloading something (sanity check not to corrupt state)
	if taskQueue.Empty() {
		return nil, false, nil
	}
	if _, ok := pendPool[p.id]; ok {
		return nil, false, nil
	}
	// Calculate an upper limit on the items we might fetch (i.e. throttling)
	space := q.resultSlots(pendPool, donePool)

	// Retrieve a batch of tasks, skipping previously failed ones
	send := make([]block.IBlock, 0, count)
	skip := make([]block.IBlock, 0)

	progress := false
	for proc := 0; proc < space && len(send) < count && !taskQueue.Empty(); proc++ {
		block := taskQueue.PopItem().(block.IBlock)
		hash := block.GetBlockID()

		// If we're the first to request this task, initialise the result container
		index := int(int64(block.GetHeight()) - int64(q.resultOffset))
		if index >= len(q.resultCache) || index < 0 {
			log.Error("index allocation went beyond available resultCache space")
			return nil, false, errInvalidChain
		}
		if q.resultCache[index] == nil {
			components := 1
			if q.mode == FastSync {
				components = 2
			}
			q.resultCache[index] = &fetchResult{
				Pending: components,
				Hash:    hash,
				Block:   block,
			}
		}
		// If this fetch task is a noop, skip this fetch operation
		if isNoop(block) {
			donePool[hash] = struct{}{}
			delete(taskPool, hash)

			space, proc = space-1, proc-1
			q.resultCache[index].Pending--
			progress = true
			continue
		}
		// Otherwise unless the peer is known not to have the data, add to the retrieve list
		if p.Lacks(hash) {
			skip = append(skip, block)
		} else {
			send = append(send, block)
		}
	}
	// Merge all the skipped headers back
	for _, block := range skip {
		taskQueue.Push(block, -float32(block.GetHeight()))
	}
	if progress {
		// Wake WaitResults, resultCache was modified
		q.active.Signal()
	}
	// Assemble and return the block download request
	if len(send) == 0 {
		return nil, progress, nil
	}
	request := &fetchRequest{
		Peer:   p,
		Blocks: send,
		Time:   time.Now(),
	}
	pendPool[p.id] = request

	return request, progress, nil
}

// CancelBlocks aborts a fetch request, returning all pending skeleton indexes to the queue.
func (q *queue) CancelBlocks(request *fetchRequest) {
	q.cancel(request, q.blockTaskQueue, q.blockPendPool)
}

// Cancel aborts a fetch request, returning all pending hashes to the task queue.
func (q *queue) cancel(request *fetchRequest, taskQueue *prque.Prque, pendPool map[string]*fetchRequest) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if request.From > 0 {
		taskQueue.Push(request.From, -float32(request.From))
	}
	for _, block := range request.Blocks {
		taskQueue.Push(block, -float32(block.GetHeight()))
	}
	delete(pendPool, request.Peer.id)
}

// Revoke cancels all pending requests belonging to a given peer. This method is
// meant to be called during a peer drop to quickly reassign owned data fetches
// to remaining nodes.
func (q *queue) Revoke(peerId string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if request, ok := q.blockPendPool[peerId]; ok {
		for _, block := range request.Blocks {
			q.blockTaskQueue.Push(block, -float32(block.GetHeight()))
		}
		delete(q.blockPendPool, peerId)
	}
}

// ExpireBlocks checks for in flight requests that exceeded a timeout allowance,
// canceling them and returning the responsible peers for penalisation.
func (q *queue) ExpireBlocks(timeout time.Duration) map[string]int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.expire(timeout, q.blockPendPool, q.blockTaskQueue)
}

// expire is the generic check that move expired tasks from a pending pool back
// into a task pool, returning all entities caught with expired tasks.
//
// Note, this method expects the queue lock to be already held. The
// reason the lock is not obtained in here is because the parameters already need
// to access the queue, so they already need a lock anyway.
func (q *queue) expire(timeout time.Duration, pendPool map[string]*fetchRequest, taskQueue *prque.Prque) map[string]int {
	// Iterate over the expired requests and return each to the queue
	expiries := make(map[string]int)
	for id, request := range pendPool {
		if time.Since(request.Time) > timeout {

			// Return any non satisfied requests to the pool
			if request.From > 0 {
				taskQueue.Push(request.From, -float32(request.From))
			}
			for _, block := range request.Blocks {
				taskQueue.Push(block, -float32(block.GetHeight()))
			}
			// Add the peer to the expiry report along the the number of failed requests
			expiries[id] = len(request.Blocks)
		}
	}
	// Remove the expired requests from the pending pool
	for id := range expiries {
		delete(pendPool, id)
	}
	return expiries
}

// DeliverBlocks injects a block retrieval response into the block results
// cache. This method either accepts all headers it received, or none of them
// if they do not map correctly to the skeleton.
//
// If the headers are accepted, the method makes an attempt to deliver the set
// of ready headers to the processor to keep the pipeline full. However it will
// not block to prevent stalling other pending deliveries.
func (q *queue) DeliverBlocks(id string, blocks []block.IBlock, blockProcCh chan []block.IBlock) (int, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Short circuit if the data was never requested
	request := q.blockPendPool[id]
	log.Info("start to DeliverBlocks", "id", id, "len(blocks)", len(blocks))
	if request == nil {
		return 0, errNoFetchesPending
	}
	// blockReqTimer.UpdateSince(request.Time)
	delete(q.blockPendPool, id)

	// Ensure headers can be mapped onto the skeleton chain
	target := q.blockTaskPool[request.From].GetBlockID()

	accepted := len(blocks) == MaxBlockFetch
	if accepted {
		if uint64(blocks[0].GetHeight()) != request.From {
			log.Trace("First block broke chain ordering", "peer", id, "number", blocks[0].GetHeight(), "hash", blocks[0].GetBlockID(), request.From)
			accepted = false
		} else if !blocks[len(blocks)-1].GetBlockID().IsEqual(target) {
			log.Trace("Last block broke skeleton structure ", "peer", id, "number", blocks[len(blocks)-1].GetHeight(), "hash", blocks[len(blocks)-1].GetBlockID(), "expected", target)
			accepted = false
		}
	}
	if accepted {
		for i, block := range blocks[1:] {
			hash := block.GetBlockID()
			if want := request.From + 1 + uint64(i); uint64(block.GetHeight()) != want {
				log.Warn("Header broke chain ordering", "peer", id, "number", block.GetHeight(), "hash", hash, "expected", want)
				accepted = false
				break
			}
			if !blocks[i].GetBlockID().IsEqual(block.GetPrevBlockID()) {
				log.Warn("Header broke chain ancestry", "peer", id, "number", block.GetHeight(), "hash", hash)
				accepted = false
				break
			}
		}
	}
	// If the batch of headers wasn't accepted, mark as unavailable
	if !accepted {
		log.Trace("Skeleton filling not accepted", "peer", id, "from", request.From)

		miss := q.blockPeerMiss[id]
		if miss == nil {
			q.blockPeerMiss[id] = make(map[uint64]struct{})
			miss = q.blockPeerMiss[id]
		}
		miss[request.From] = struct{}{}

		q.blockTaskQueue.Push(request.From, -float32(request.From))
		return 0, errors.New("delivery not accepted")
	}
	// Clean up a successful fetch and try to deliver any sub-results
	copy(q.blockResults[request.From-q.blockOffset:], blocks)
	delete(q.blockTaskPool, request.From)

	ready := 0
	for q.blockProced+ready < len(q.blockResults) && q.blockResults[q.blockProced+ready] != nil {
		ready += MaxBlockFetch
	}
	if ready > 0 {
		// Blocks are ready for delivery, gather them and push forward (non blocking)
		process := make([]block.IBlock, ready)
		copy(process, q.blockResults[q.blockProced:q.blockProced+ready])

		select {
		case blockProcCh <- process:
			log.Trace("Pre-scheduled new headers", "peer", id, "count", len(process), "from", process[0].GetHeight())
			q.blockProced += len(process)
		default:
		}
	}
	// Check for termination and return
	if len(q.blockTaskPool) == 0 {
		q.blockContCh <- false
	}
	return len(blocks), nil
}

// deliver injects a data retrieval response into the results queue.
//
// Note, this method expects the queue lock to be already held for writing. The
// reason the lock is not obtained in here is because the parameters already need
// to access the queue, so they already need a lock anyway.
func (q *queue) deliver(id string, taskPool map[meta.DataID]block.IBlock, taskQueue *prque.Prque,
	pendPool map[string]*fetchRequest, donePool map[meta.DataID]struct{},
	results int, reconstruct func(block block.IBlock, index int, result *fetchResult) error) (int, error) {

	// Short circuit if the data was never requested
	log.Info("start to deliver", "id", id)
	request := pendPool[id]
	if request == nil {
		return 0, errNoFetchesPending
	}

	delete(pendPool, id)

	// If no data items were retrieved, mark them as unavailable for the origin peer
	if results == 0 {
		for _, block := range request.Blocks {
			request.Peer.MarkLacking(block.GetBlockID())
		}
	}
	// Assemble each of the results with their headers and retrieved data parts
	var (
		accepted int
		failure  error
		useful   bool
	)
	for i, block := range request.Blocks {
		// Short circuit assembly if no more fetch results are found
		if i >= results {
			break
		}
		// Reconstruct the next result if contents match up
		index := int(int64(block.GetHeight()) - int64(q.resultOffset))
		if index >= len(q.resultCache) || index < 0 || q.resultCache[index] == nil {
			failure = errInvalidChain
			break
		}
		if err := reconstruct(block, i, q.resultCache[index]); err != nil {
			failure = err
			break
		}
		hash := block.GetBlockID()

		donePool[hash] = struct{}{}
		q.resultCache[index].Pending--
		useful = true
		accepted++

		// Clean up a successful fetch
		request.Blocks[i] = nil
		delete(taskPool, hash)
	}
	// Return all failed or missing fetches to the queue
	for _, block := range request.Blocks {
		if block != nil {
			taskQueue.Push(block, -float32(block.GetHeight()))
		}
	}
	// Wake up WaitResults
	if accepted > 0 {
		q.active.Signal()
	}
	// If none of the data was good, it's a stale delivery
	switch {
	case failure == nil || failure == errInvalidChain:
		return accepted, failure
	case useful:
		return accepted, fmt.Errorf("partial failure: %v", failure)
	default:
		return accepted, errStaleDelivery
	}
}

// Prepare configures the result cache to allow accepting and caching inbound
// fetch results.
func (q *queue) Prepare(offset uint64, mode SyncMode) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Prepare the queue for sync results
	if q.resultOffset < offset {
		q.resultOffset = offset
	}
	q.mode = mode
}
