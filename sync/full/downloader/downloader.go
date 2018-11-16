package downloader

import (
	"errors"
	"fmt"
	_ "math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/linkchain/common/util/event"
	"github.com/linkchain/common/util/log"
	"github.com/linkchain/core/meta"
	"github.com/linkchain/node"
)

var (
	MaxBlockFetch   = 192 // Amount of blocks to be fetched per retrieval request
	MaxSkeletonSize = 128 // Number of header fetches to need for a skeleton assembly

	rttMinEstimate   = 2 * time.Second  // Minimum round-trip time to target for download requests
	rttMaxEstimate   = 20 * time.Second // Maximum rount-trip time to target for download requests
	rttMinConfidence = 0.1              // Worse confidence factor in our estimated RTT value
	ttlScaling       = 3                // Constant scaling factor for RTT -> TTL conversion
	ttlLimit         = time.Minute      // Maximum TTL allowance to prevent reaching crazy timeouts

	qosTuningPeers   = 5    // Number of peers to tune based on (best peers)
	qosConfidenceCap = 10   // Number of peers above which not to modify RTT confidence
	qosTuningImpact  = 0.25 // Impact that a new tuning target has on the previous value

	maxBlocksProcess  = 2048 // Number of header download results to import at once into the chain
	maxResultsProcess = 2048 // Number of content download results to import at once into the chain

	fsBlockContCheck = 3 * time.Second
)

var (
	errBusy                    = errors.New("busy")
	errUnknownPeer             = errors.New("peer is unknown or unhealthy")
	errBadPeer                 = errors.New("action from bad peer ignored")
	errStallingPeer            = errors.New("peer is stalling")
	errNoPeers                 = errors.New("no peers to keep download active")
	errTimeout                 = errors.New("timeout")
	errEmptyBlockSet           = errors.New("empty block set by peer")
	errPeersUnavailable        = errors.New("no peers available or all tried for download")
	errInvalidAncestor         = errors.New("retrieved ancestor is invalid")
	errInvalidChain            = errors.New("retrieved hash chain is invalid")
	errCancelBlockFetch        = errors.New("block download canceled (requested)")
	errCancelBlockProcessing   = errors.New("block processing canceled (requested)")
	errCancelContentProcessing = errors.New("content processing canceled (requested)")
	errNoSyncActive            = errors.New("no sync active")
	errTooOld                  = errors.New("peer doesn't speak recent enough protocol version")
)

type Downloader struct {
	mode SyncMode       // Synchronisation mode defining the strategy used (per sync cycle)
	mux  *event.TypeMux // Event multiplexer to announce sync operation events

	queue *queue   // Scheduler for selecting the hashes to download
	peers *peerSet // Set of active peers from which download can proceed

	rttEstimate   uint64 // Round trip time to target for download requests
	rttConfidence uint64 // Confidence in the estimated RTT (unit: millionths to allow atomic ops)

	nodeAPI *node.PublicNodeAPI

	// Callbacks
	dropPeer peerDropFn // Drops a peer for misbehaving

	// Status
	synchronising int32
	notified      int32
	committed     int32

	// Channels
	blockCh     chan dataPack      // [eth/62] Channel receiving inbound block headers
	blockProcCh chan []*meta.Block // [eth/62] Channel to feed the header processor new tasks

	// Cancellation and termination
	cancelPeer string        // Identifier of the peer currently being used as the master (cancel on drop)
	cancelCh   chan struct{} // Channel to cancel mid-flight syncs
	cancelLock sync.RWMutex  // Lock to protect the cancel channel and peer in delivers

	quitCh   chan struct{} // Quit channel to signal termination
	quitLock sync.RWMutex  // Lock to prevent double closes
}

// New creates a new downloader to fetch hashes and blocks from remote peers.
func New(mux *event.TypeMux, nodeSvc *node.PublicNodeAPI, dropPeer peerDropFn) *Downloader {

	dl := &Downloader{
		mode:          FullSync,
		mux:           mux,
		queue:         newQueue(),
		peers:         newPeerSet(),
		rttEstimate:   uint64(rttMaxEstimate),
		rttConfidence: uint64(1000000),
		nodeAPI:       nodeSvc,
		dropPeer:      dropPeer,
		blockCh:       make(chan dataPack, 1),
		blockProcCh:   make(chan []*meta.Block, 1),
		quitCh:        make(chan struct{}),
	}
	go dl.qosTuner()
	return dl
}

// Progress retrieves the synchronisation boundaries, specifically the origin
// block where synchronisation started at (may have failed/suspended); the block
// or header sync is currently at; and the latest known block which the sync targets.
//
// In addition, during the state download phase of fast synchronisation the number
// of processed and the total number of known states are also returned. Otherwise
// these are zero.
//func (d *Downloader) Progress() ethereum.SyncProgress {
//	// Lock the current stats and return the progress
//	d.syncStatsLock.RLock()
//	defer d.syncStatsLock.RUnlock()
//
//	current := uint64(0)
//	switch d.mode {
//	case FullSync:
//		current = uint64(d.blockchain.getBestBlock().GetHeight())
//		//	case FastSync:
//		//		current = d.blockchain.CurrentFastBlock().NumberU64()
//		//	case LightSync:
//		//		current = d.lightchain.CurrentHeader().Number.Uint64()
//	}
//	return ethereum.SyncProgress{
//		StartingBlock: d.syncStatsChainOrigin,
//		CurrentBlock:  current,
//		HighestBlock:  d.syncStatsChainHeight,
//	}
//}

// Synchronising returns whether the downloader is currently retrieving blocks.
func (d *Downloader) Synchronising() bool {
	return atomic.LoadInt32(&d.synchronising) > 0
}

// RegisterPeer injects a new download peer into the set of block source to be
// used for fetching hashes and blocks from.
func (d *Downloader) RegisterPeer(id string, version int, peer Peer) error {
	logger := log.New("peer", id)
	logger.Trace("Registering sync peer")
	if err := d.peers.Register(newPeerConnection(id, version, peer, logger)); err != nil {
		logger.Error("Failed to register sync peer", "err", err)
		return err
	}
	d.qosReduceConfidence()

	return nil
}

// RegisterLightPeer injects a light client peer, wrapping it so it appears as a regular peer.
func (d *Downloader) RegisterLightPeer(id string, version int, peer LightPeer) error {
	return d.RegisterPeer(id, version, &lightPeerWrapper{peer})
}

// UnregisterPeer remove a peer from the known list, preventing any action from
// the specified peer. An effort is also made to return any pending fetches into
// the queue.
func (d *Downloader) UnregisterPeer(id string) error {
	// Unregister the peer from the active peer set and revoke any fetch tasks
	logger := log.New("peer", id)
	logger.Trace("Unregistering sync peer")
	if err := d.peers.Unregister(id); err != nil {
		logger.Error("Failed to unregister sync peer", "err", err)
		return err
	}
	d.queue.Revoke(id)

	// If this peer was the master peer, abort sync immediately
	d.cancelLock.RLock()
	master := id == d.cancelPeer
	d.cancelLock.RUnlock()

	if master {
		d.Cancel()
	}
	return nil
}

// Synchronise tries to sync up our local block chain with a remote peer, both
// adding various sanity checks as well as wrapping it with various log entries.
func (d *Downloader) Synchronise(id string, head meta.BlockID) error {
	err := d.synchronise(id, head)
	switch err {
	case nil:
	case errBusy:

	case errTimeout, errBadPeer, errStallingPeer,
		errEmptyBlockSet, errPeersUnavailable, errTooOld,
		errInvalidAncestor, errInvalidChain:
		log.Warn("Synchronisation failed, dropping peer", "peer", id, "err", err)
		if d.dropPeer == nil {
			// The dropPeer method is nil when `--copydb` is used for a local copy.
			// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
			log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", id)
		} else {
			d.dropPeer(id)
		}
	default:
		log.Warn("Synchronisation failed, retrying", "err", err)
	}
	return err
}

// synchronise will select the peer and use it for synchronising. If an empty string is given
// it will use the best peer possible and synchronize if its TD is higher than our own. If any of the
// checks fail an error will be returned. This method is synchronous
func (d *Downloader) synchronise(id string, hash meta.BlockID) error {
	// Make sure only one goroutine is ever allowed past this point at once
	if !atomic.CompareAndSwapInt32(&d.synchronising, 0, 1) {
		return errBusy
	}
	defer atomic.StoreInt32(&d.synchronising, 0)

	// Post a user notification of the sync (only once per session)
	if atomic.CompareAndSwapInt32(&d.notified, 0, 1) {
		log.Info("Block synchronisation started")
	}
	// Reset the queue, peer set and wake channels to clean any internal leftover state
	d.queue.Reset()
	d.peers.Reset()

	for _, ch := range []chan dataPack{d.blockCh} {
		for empty := false; !empty; {
			select {
			case <-ch:
			default:
				empty = true
			}
		}
	}
	for empty := false; !empty; {
		select {
		case <-d.blockProcCh:
		default:
			empty = true
		}
	}
	// Create cancel channel for aborting mid-flight and mark the master peer
	d.cancelLock.Lock()
	d.cancelCh = make(chan struct{})
	d.cancelPeer = id
	d.cancelLock.Unlock()

	defer d.Cancel() // No matter what, we can't leave the cancel channel open

	// Retrieve the origin peer and initiate the downloading process
	p := d.peers.Peer(id)
	if p == nil {
		return errUnknownPeer
	}
	return d.syncWithPeer(p, hash)
}

// syncWithPeer starts a block synchronization based on the hash chain from the
// specified peer and head hash.
func (d *Downloader) syncWithPeer(p *peerConnection, hash meta.BlockID) (err error) {
	log.Debug("start to sync with peer")
	d.mux.Post(StartEvent{})
	defer func() {
		// reset on error
		if err != nil {
			d.mux.Post(FailedEvent{err})
		} else {
			d.mux.Post(DoneEvent{})
		}
	}()
	if p.version < 1 {
		return errTooOld
	}

	log.Trace("Synchronising with the network", "peer", p.id, "eth", p.version, "head", hash, "mode", d.mode)
	defer func(start time.Time) {
		log.Debug("Synchronisation terminated", "elapsed", time.Since(start))
	}(time.Now())

	// Look up the sync boundaries: the common ancestor and the target block
	latest, err := d.fetchHeight(p)
	if err != nil {
		return err
	}
	height := uint64(latest.GetHeight())

	origin, err := d.findAncestor(p, height)
	if err != nil {
		return err
	}

	d.committed = 1
	d.queue.Prepare(origin+1, d.mode)
	pivot := uint64(0)
	fetchers := []func() error{
		func() error { return d.fetchBlocks(p, origin+1, pivot) },
		func() error { return d.processBlocks(origin+1, pivot) },
	}
	if d.mode == FullSync {
		fetchers = append(fetchers, d.processFullSyncContent)
	}
	return d.spawnSync(fetchers)
}

// spawnSync runs d.process and all given fetcher functions to completion in
// separate goroutines, returning the first error that appears.
func (d *Downloader) spawnSync(fetchers []func() error) error {
	var wg sync.WaitGroup
	errc := make(chan error, len(fetchers))
	wg.Add(len(fetchers))
	for i, fn := range fetchers {
		fn := fn
		log.Debug("start sync fetchers", "index", i)
		go func() { defer wg.Done(); errc <- fn() }()
	}
	// Wait for the first error, then terminate the others.
	var err error
	for i := 0; i < len(fetchers); i++ {
		if i == len(fetchers)-1 {
			d.queue.Close()
		}
		if err = <-errc; err != nil {
			break
		}
	}
	d.queue.Close()
	d.Cancel()
	wg.Wait()
	return err
}

// Cancel cancels all of the operations and resets the queue. It returns true
// if the cancel operation was completed.
func (d *Downloader) Cancel() {
	// Close the current cancel channel
	d.cancelLock.Lock()
	if d.cancelCh != nil {
		select {
		case <-d.cancelCh:
			// Channel was already closed
		default:
			close(d.cancelCh)
		}
	}
	d.cancelLock.Unlock()
}

// Terminate interrupts the downloader, canceling all pending operations.
// The downloader cannot be reused after calling Terminate.
func (d *Downloader) Terminate() {
	// Close the termination channel (make sure double close is allowed)
	d.quitLock.Lock()
	select {
	case <-d.quitCh:
	default:
		close(d.quitCh)
	}
	d.quitLock.Unlock()

	// Cancel any pending download requests
	d.Cancel()
}

// fetchHeight retrieves the head header of the remote peer to aid in estimating
// the total time a pending synchronisation would take.
func (d *Downloader) fetchHeight(p *peerConnection) (*meta.Block, error) {
	p.log.Trace("Retrieving remote chain height")

	// Request the advertised remote head block and wait for the response
	head, _ := p.peer.Head()
	go p.peer.RequestBlocksByHash(head, 1, 0)

	ttl := d.requestTTL()
	timeout := time.After(ttl)
	for {
		select {
		case <-d.cancelCh:
			return nil, errCancelBlockFetch

		case packet := <-d.blockCh:
			// Discard anything not from the origin peer
			if packet.PeerId() != p.id {
				log.Trace("Received blocks from incorrect peer", "peer", packet.PeerId())
				break
			}
			// Make sure the peer actually gave something valid
			blocks := packet.(*blockPack).blocks
			if len(blocks) != 1 {
				p.log.Trace("Multiple blocks for single request", "blocks", len(blocks))
				return nil, errBadPeer
			}
			block := blocks[0]
			p.log.Trace("Remote blocks identified", "number", block.GetHeight(), "hash", block.GetBlockID())
			return block, nil

		case <-timeout:
			p.log.Trace("Waiting for head height header timed out", "elapsed", ttl)
			return nil, errTimeout
		}
	}
}

// findAncestor tries to locate the common ancestor link of the local chain and
// a remote peers blockchain. In the general case when our node was in sync and
// on the correct chain, checking the top N links should already get us a match.
// In the rare scenario when we ended up on a long reorganisation (i.e. none of
// the head links match), we do a binary search to find the common ancestor.
func (d *Downloader) findAncestor(p *peerConnection, height uint64) (uint64, error) {
	var ceil uint64
	floor := int64(-1)
	if d.mode == FullSync {
		ceil = uint64(d.nodeAPI.GetBestBlock().GetHeight())
	}

	p.log.Debug("Looking for common ancestor", "local", ceil, "remote", height)

	// Request the topmost blocks to short circuit binary ancestor lookup
	head := ceil
	if head > height {
		head = height
	}
	from := int64(head) - int64(MaxBlockFetch)
	if from < 0 {
		from = 0
	}
	// Span out with 15 block gaps into the future to catch bad head reports
	limit := 2 * MaxBlockFetch / 16
	count := 1 + int((int64(ceil)-from)/16)
	if count > limit {
		count = limit
	}
	log.Debug("findAncestor RequestBlocksByNumber", "from", from, "count", count)
	go p.peer.RequestBlocksByNumber(uint64(from), count, 15)

	// Wait for the remote response to the head fetch
	number := uint64(0)
	var hash meta.BlockID

	ttl := d.requestTTL()
	timeout := time.After(ttl)

	for finished := false; !finished; {
		select {
		case <-d.cancelCh:
			return 0, errCancelBlockFetch

		case packet := <-d.blockCh:
			// Discard anything not from the origin peer
			if packet.PeerId() != p.id {
				log.Debug("Received headers from incorrect peer", "peer", packet.PeerId())
				break
			}
			// Make sure the peer actually gave something valid
			blocks := packet.(*blockPack).blocks
			if len(blocks) == 0 {
				p.log.Warn("Empty head block set")
				return 0, errEmptyBlockSet
			}
			// Make sure the peer's reply conforms to the request
			for i := 0; i < len(blocks); i++ {
				if number := int64(blocks[i].GetHeight()); number != from+int64(i)*16 {
					p.log.Warn("Head blocks broke chain ordering", "index", i, "requested", from+int64(i)*16, "received", number)
					return 0, errInvalidChain
				}
			}
			// Check if a common ancestor was found
			finished = true
			for i := len(blocks) - 1; i >= 0; i-- {
				// Skip any headers that underflow/overflow our requested set
				if int64(blocks[i].GetHeight()) < from || uint64(blocks[i].GetHeight()) > ceil {
					continue
				}
				// Otherwise check if we already know the header or not
				if d.mode == FullSync && d.nodeAPI.HasBlock(*blocks[i].GetBlockID()) {
					number, hash = uint64(blocks[i].GetHeight()), *blocks[i].GetBlockID()

					// If every header is known, even future ones, the peer straight out lied about its head
					if number > height && i == limit-1 {
						p.log.Warn("Lied about chain head", "reported", height, "found", number)
						return 0, errStallingPeer
					}
					break
				}
			}

		case <-timeout:
			p.log.Debug("Waiting for Ancestor head header timed out", "elapsed", ttl)
			return 0, errTimeout
		}
	}
	// If the head fetch already found an ancestor, return
	if hash.IsEmpty() && hash.IsEmpty() {
		if int64(number) <= floor {
			p.log.Warn("Ancestor below allowance", "number", number, "hash", hash, "allowance", floor)
			return 0, errInvalidAncestor
		}
		p.log.Trace("Found common ancestor", "number", number, "hash", hash)
		return number, nil
	}
	// Ancestor not found, we need to binary search over our chain
	start, end := uint64(0), head

	for start+1 < end {
		// Split our chain interval in two, and request the hash to cross check
		check := (start + end) / 2

		ttl := d.requestTTL()
		timeout := time.After(ttl)

		go p.peer.RequestBlocksByNumber(check, 1, 0)

		// Wait until a reply arrives to this request
		for arrived := false; !arrived; {
			select {
			case <-d.cancelCh:
				return 0, errCancelBlockFetch

			case packer := <-d.blockCh:
				// Discard anything not from the origin peer
				if packer.PeerId() != p.id {
					log.Trace("Received headers from incorrect peer", "peer", packer.PeerId())
					break
				}
				// Make sure the peer actually gave something valid
				blocks := packer.(*blockPack).blocks
				if len(blocks) != 1 {
					p.log.Trace("Multiple blocks for single request", "blocks", len(blocks))
					return 0, errBadPeer
				}
				arrived = true

				// Modify the search interval based on the response
				if d.mode == FullSync && !d.nodeAPI.HasBlock(*blocks[0].GetBlockID()) {
					end = check
					break
				}
				start = check

			case <-timeout:
				p.log.Trace("Waiting for search header timed out", "elapsed", ttl)
				return 0, errTimeout
			}
		}
	}

	p.log.Trace("Found common ancestor", "number", start, "hash", hash)
	return start, nil
}

func (d *Downloader) fetchBlocks(p *peerConnection, from uint64, pivot uint64) error {
	p.log.Debug("Directing block downloads", "origin", from)
	defer p.log.Debug("Block download terminated")

	// Create a timeout timer, and the associated header fetcher
	skeleton := false           // Skeleton assembly phase or finishing up
	request := time.Now()       // time of the last skeleton fetch request
	timeout := time.NewTimer(0) // timer to dump a non-responsive active peer
	<-timeout.C                 // timeout channel should be initially empty
	defer timeout.Stop()

	var ttl time.Duration
	getBlocks := func(from uint64) {
		request = time.Now()

		ttl = d.requestTTL()
		timeout.Reset(ttl)

		if skeleton {
			p.log.Trace("Fetching skeleton blocks", "count", MaxBlockFetch, "from", from)
			go p.peer.RequestBlocksByNumber(from+uint64(MaxBlockFetch)-1, MaxSkeletonSize, MaxBlockFetch-1)
		} else {
			p.log.Trace("Fetching full headers", "count", MaxBlockFetch, "from", from)
			go p.peer.RequestBlocksByNumber(from, MaxBlockFetch, 0)
		}
	}
	// Start pulling the header chain skeleton until all is done
	getBlocks(from)

	for {
		select {
		case <-d.cancelCh:
			return errCancelBlockFetch

		case packet := <-d.blockCh:
			// Make sure the active peer is giving us the skeleton headers
			if packet.PeerId() != p.id {
				log.Debug("Received skeleton from incorrect peer", "peer", packet.PeerId())
				break
			}

			timeout.Stop()
			log.Trace("Received skeleton result", "packet.Items()", packet.Items())
			// If the skeleton's finished, pull any remaining head headers directly from the origin
			if packet.Items() == 0 && skeleton {
				skeleton = false
				getBlocks(from)
				continue
			}
			// If no more headers are inbound, notify the content fetchers and return
			if packet.Items() == 0 {
				// Don't abort header fetches while the pivot is downloading
				if atomic.LoadInt32(&d.committed) == 0 && pivot <= from {
					p.log.Trace("No headers, waiting for pivot commit")
					select {
					case <-time.After(fsBlockContCheck):
						getBlocks(from)
						continue
					case <-d.cancelCh:
						return errCancelBlockFetch
					}
				}
				// Pivot done (or not in fast sync) and no more headers, terminate the process
				p.log.Trace("No more headers available")
				select {
				case d.blockProcCh <- nil:
					return nil
				case <-d.cancelCh:
					return errCancelBlockFetch
				}
			}
			blocks := packet.(*blockPack).blocks

			// If we received a skeleton batch, resolve internals concurrently
			if skeleton {
				filled, proced, err := d.fillBlockSkeleton(from, blocks)
				if err != nil {
					p.log.Trace("Skeleton chain invalid", "err", err)
					return errInvalidChain
				}
				blocks = filled[proced:]
				from += uint64(proced)
			}
			// Insert all the new headers and fetch the next batch
			if len(blocks) > 0 {
				p.log.Trace("Scheduling new blocks", "count", len(blocks), "from", from)
				select {
				case d.blockProcCh <- blocks:
				case <-d.cancelCh:
					return errCancelBlockFetch
				}
				from += uint64(len(blocks))
			}
			getBlocks(from)

		case <-timeout.C:
			if d.dropPeer == nil {
				// The dropPeer method is nil when `--copydb` is used for a local copy.
				// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
				p.log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", p.id)
				break
			}
			// Header retrieval timed out, consider the peer bad and drop
			p.log.Trace("Header request timed out", "elapsed", ttl)

			d.dropPeer(p.id)

			select {
			case d.blockProcCh <- nil:
			case <-d.cancelCh:
			}
			return errBadPeer
		}
	}
}

func (d *Downloader) fillBlockSkeleton(from uint64, skeleton []*meta.Block) ([]*meta.Block, int, error) {
	log.Trace("Filling up skeleton", "from", from)
	d.queue.ScheduleSkeleton(from, skeleton)

	var (
		deliver = func(packet dataPack) (int, error) {
			pack := packet.(*blockPack)
			return d.queue.DeliverBlocks(pack.peerId, pack.blocks, d.blockProcCh)
		}
		expire   = func() map[string]int { return d.queue.ExpireBlocks(d.requestTTL()) }
		throttle = func() bool { return false }
		reserve  = func(p *peerConnection, count int) (*fetchRequest, bool, error) {
			return d.queue.ReserveBlocks(p, count), false, nil
		}
		fetch    = func(p *peerConnection, req *fetchRequest) error { return p.FetchBlocks(req.From, MaxBlockFetch) }
		capacity = func(p *peerConnection) int { return p.BlockCapacity(d.requestRTT()) }
		setIdle  = func(p *peerConnection, accepted int) { p.SetBlocksIdle(accepted) }
	)
	err := d.fetchParts(errCancelBlockFetch, d.blockCh, deliver, d.queue.blockContCh, expire,
		d.queue.PendingBlocks, d.queue.InFlightBlocks, throttle, reserve,
		nil, fetch, d.queue.CancelBlocks, capacity, d.peers.BlockIdlePeers, setIdle, "blocks")

	log.Trace("Skeleton fill terminated", "err", err)

	filled, proced := d.queue.RetrieveBlocks()
	return filled, proced, err
}

func (d *Downloader) fetchParts(errCancel error, deliveryCh chan dataPack, deliver func(dataPack) (int, error), wakeCh chan bool,
	expire func() map[string]int, pending func() int, inFlight func() bool, throttle func() bool, reserve func(*peerConnection, int) (*fetchRequest, bool, error),
	fetchHook func([]*meta.Block), fetch func(*peerConnection, *fetchRequest) error, cancel func(*fetchRequest), capacity func(*peerConnection) int,
	idle func() ([]*peerConnection, int), setIdle func(*peerConnection, int), kind string) error {

	// Create a ticker to detect expired retrieval tasks
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	update := make(chan struct{}, 1)

	// Prepare the queue and fetch block parts until the block header fetcher's done
	finished := false
	for {
		select {
		case <-d.cancelCh:
			return errCancel

		case packet := <-deliveryCh:
			// If the peer was previously banned and failed to deliver its pack
			// in a reasonable time frame, ignore its message.
			if peer := d.peers.Peer(packet.PeerId()); peer != nil {
				// Deliver the received chunk of data and check chain validity
				accepted, err := deliver(packet)
				if err == errInvalidChain {
					return err
				}
				// Unless a peer delivered something completely else than requested (usually
				// caused by a timed out request which came through in the end), set it to
				// idle. If the delivery's stale, the peer should have already been idled.
				if err != errStaleDelivery {
					setIdle(peer, accepted)
				}
				// Issue a log to the user to see what's going on
				switch {
				case err == nil && packet.Items() == 0:
					peer.log.Trace("Requested data not delivered", "type", kind)
				case err == nil:
					peer.log.Trace("Delivered new batch of data", "type", kind, "count", packet.Stats())
				default:
					peer.log.Trace("Failed to deliver retrieved data", "type", kind, "err", err)
				}
			}
			// Blocks assembled, try to update the progress
			select {
			case update <- struct{}{}:
			default:
			}

		case cont := <-wakeCh:
			// The header fetcher sent a continuation flag, check if it's done
			if !cont {
				finished = true
			}
			// Headers arrive, try to update the progress
			select {
			case update <- struct{}{}:
			default:
			}

		case <-ticker.C:
			// Sanity check update the progress
			select {
			case update <- struct{}{}:
			default:
			}

		case <-update:
			// Short circuit if we lost all our peers
			if d.peers.Len() == 0 {
				return errNoPeers
			}
			// Check for fetch request timeouts and demote the responsible peers
			for pid, fails := range expire() {
				if peer := d.peers.Peer(pid); peer != nil {
					// If a lot of retrieval elements expired, we might have overestimated the remote peer or perhaps
					// ourselves. Only reset to minimal throughput but don't drop just yet. If even the minimal times
					// out that sync wise we need to get rid of the peer.
					//
					// The reason the minimum threshold is 2 is because the downloader tries to estimate the bandwidth
					// and latency of a peer separately, which requires pushing the measures capacity a bit and seeing
					// how response times reacts, to it always requests one more than the minimum (i.e. min 2).
					if fails > 2 {
						peer.log.Trace("Data delivery timed out", "type", kind)
						setIdle(peer, 0)
					} else {
						peer.log.Trace("Stalling delivery, dropping", "type", kind)
						if d.dropPeer == nil {
							// The dropPeer method is nil when `--copydb` is used for a local copy.
							// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
							peer.log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", pid)
						} else {
							d.dropPeer(pid)
						}
					}
				}
			}
			// If there's nothing more to fetch, wait or terminate
			if pending() == 0 {
				if !inFlight() && finished {
					log.Trace("Data fetching completed", "type", kind)
					return nil
				}
				break
			}
			// Send a download request to all idle peers, until throttled
			progressed, throttled, running := false, false, inFlight()
			idles, total := idle()

			for _, peer := range idles {
				// Short circuit if throttling activated
				if throttle() {
					throttled = true
					break
				}
				// Short circuit if there is no more available task.
				if pending() == 0 {
					break
				}
				// Reserve a chunk of fetches for a peer. A nil can mean either that
				// no more headers are available, or that the peer is known not to
				// have them.
				request, progress, err := reserve(peer, capacity(peer))
				if err != nil {
					return err
				}
				if progress {
					progressed = true
				}
				if request == nil {
					continue
				}
				if request.From > 0 {
					peer.log.Trace("Requesting new batch of data", "type", kind, "from", request.From)
				} else {
					peer.log.Trace("Requesting new batch of data", "type", kind, "count", len(request.Blocks), "from", request.Blocks[0].GetHeight())
				}
				// Fetch the chunk and make sure any errors return the hashes to the queue
				if fetchHook != nil {
					fetchHook(request.Blocks)
				}
				if err := fetch(peer, request); err != nil {
					// Although we could try and make an attempt to fix this, this error really
					// means that we've double allocated a fetch task to a peer. If that is the
					// case, the internal state of the downloader and the queue is very wrong so
					// better hard crash and note the error instead of silently accumulating into
					// a much bigger issue.
					panic(fmt.Sprintf("%v: %s fetch assignment failed", peer, kind))
				}
				running = true
			}
			// Make sure that we have peers available for fetching. If all peers have been tried
			// and all failed throw an error
			if !progressed && !throttled && !running && len(idles) == total && pending() > 0 {
				return errPeersUnavailable
			}
		}
	}
}

func (d *Downloader) processBlocks(origin uint64, pivot uint64) error {
	// Keep a count of uncertain headers to roll back
	rollback := []*meta.Block{}
	defer func() {
		if len(rollback) > 0 {
			// Flatten the headers and roll them back
			hashes := make([]meta.BlockID, len(rollback))
			for i, block := range rollback {
				hashes[i] = *block.GetBlockID()
			}

			lastBlock := d.nodeAPI.GetBestBlock().GetHeight()

			// TODO: add rollback code
			// d.lightchain.Rollback(hashes)

			curBlock := d.nodeAPI.GetBestBlock().GetHeight()
			log.Warn("Rolled back blocks", "count", len(hashes),
				"block", fmt.Sprintf("%d->%d", lastBlock, curBlock))
		}
	}()

	// Wait for batches of blocks to process
	// gotBlocks := false

	for {
		select {
		case <-d.cancelCh:
			return errCancelBlockProcessing

		case blocks := <-d.blockProcCh:
			// Terminate header processing if we synced up
			if len(blocks) == 0 {
				// Disable any rollback and return
				rollback = nil
				return nil
			}
			// Otherwise split the chunk of headers into batches and process them
			// gotBlocks = true

			for len(blocks) > 0 {
				// Terminate if something failed in between processing chunks
				select {
				case <-d.cancelCh:
					return errCancelBlockProcessing
				default:
				}
				// Select the next chunk of headers to import
				limit := maxBlocksProcess
				if limit > len(blocks) {
					limit = len(blocks)
				}
				chunk := blocks[:limit]

				// Unless we're doing light chains, schedule the headers for associated content retrieval
				if d.mode == FullSync {
					// Otherwise insert the headers for content retrieval
					inserts := d.queue.Schedule(chunk, origin)
					if len(inserts) != len(chunk) {
						log.Trace("Stale headers")
						return errBadPeer
					}
				}
				blocks = blocks[limit:]
				origin += uint64(limit)
			}
		}
	}
	return nil
}

// processFullSyncContent takes fetch results from the queue and imports them into the chain.
func (d *Downloader) processFullSyncContent() error {
	for {
		results := d.queue.Results(true)
		if len(results) == 0 {
			return nil
		}
		if err := d.importBlockResults(results); err != nil {
			log.Error("importBlockResults failed", "err", err)
			return err
		}
	}
}

func (d *Downloader) ImportBlocks(id string, blocks []*meta.Block) error {
	var results []*fetchResult
	for _, block := range blocks {
		results = append(results, &fetchResult{Hash: *block.GetBlockID(), Block: block})
	}
	return d.importBlockResults(results)
}

func (d *Downloader) importBlockResults(results []*fetchResult) error {
	// Check for any early termination requests
	if len(results) == 0 {
		return nil
	}
	select {
	case <-d.quitCh:
		return errCancelContentProcessing
	default:
	}
	// Retrieve the a batch of results to import
	first, last := results[0].Block, results[len(results)-1].Block
	log.Trace("Inserting downloaded chain", "items", len(results),
		"firstnum", first.GetHeight(), "firsthash", first.GetBlockID(),
		"lastnum", last.GetHeight(), "lasthash", last.GetBlockID(),
	)
	//blocks := make([]block.IBlock, len(results))

	for _, result := range results {
		log.Trace("Downloaded item processing block", "number", result.Block.GetHeight(), "hash", result.Block.GetBlockID(), "block", result.Block)
		if d.nodeAPI.HasBlock(*result.Block.GetBlockID()) {
			continue
		}

		if err := d.nodeAPI.ProcessBlock(result.Block); err != nil {
			log.Error("Downloaded item processing failed", "number", result.Block.GetHeight(), "hash", result.Block.GetBlockID(), "err", err)
			return errInvalidChain
		}
	}
	return nil
}

func splitAroundPivot(pivot uint64, results []*fetchResult) (p *fetchResult, before, after []*fetchResult) {
	for _, result := range results {
		num := uint64(result.Block.GetHeight())
		switch {
		case num < pivot:
			before = append(before, result)
		case num == pivot:
			p = result
		default:
			after = append(after, result)
		}
	}
	return p, before, after
}

// DeliverHeaders injects a new batch of block headers received from a remote
// node into the download schedule.
func (d *Downloader) DeliverBlocks(id string, blocks []*meta.Block) (err error) {
	return d.deliver(id, d.blockCh, &blockPack{id, blocks})
}

// deliver injects a new batch of data received from a remote node.
func (d *Downloader) deliver(id string, destCh chan dataPack, packet dataPack) (err error) {
	// Deliver or abort if the sync is canceled while queuing
	d.cancelLock.RLock()
	cancel := d.cancelCh
	d.cancelLock.RUnlock()
	if cancel == nil {
		return errNoSyncActive
	}
	select {
	case destCh <- packet:
		return nil
	case <-cancel:
		return errNoSyncActive
	}
}

// qosTuner is the quality of service tuning loop that occasionally gathers the
// peer latency statistics and updates the estimated request round trip time.
func (d *Downloader) qosTuner() {
	for {
		// Retrieve the current median RTT and integrate into the previoust target RTT
		rtt := time.Duration((1-qosTuningImpact)*float64(atomic.LoadUint64(&d.rttEstimate)) + qosTuningImpact*float64(d.peers.medianRTT()))
		atomic.StoreUint64(&d.rttEstimate, uint64(rtt))

		// A new RTT cycle passed, increase our confidence in the estimated RTT
		conf := atomic.LoadUint64(&d.rttConfidence)
		conf = conf + (1000000-conf)/2
		atomic.StoreUint64(&d.rttConfidence, conf)

		// Log the new QoS values and sleep until the next RTT
		log.Trace("Recalculated downloader QoS values", "rtt", rtt, "confidence", float64(conf)/1000000.0, "ttl", d.requestTTL())
		select {
		case <-d.quitCh:
			return
		case <-time.After(rtt):
		}
	}
}

// qosReduceConfidence is meant to be called when a new peer joins the downloader's
// peer set, needing to reduce the confidence we have in out QoS estimates.
func (d *Downloader) qosReduceConfidence() {
	// If we have a single peer, confidence is always 1
	peers := uint64(d.peers.Len())
	if peers == 0 {
		// Ensure peer connectivity races don't catch us off guard
		return
	}
	if peers == 1 {
		atomic.StoreUint64(&d.rttConfidence, 1000000)
		return
	}
	// If we have a ton of peers, don't drop confidence)
	if peers >= uint64(qosConfidenceCap) {
		return
	}
	// Otherwise drop the confidence factor
	conf := atomic.LoadUint64(&d.rttConfidence) * (peers - 1) / peers
	if float64(conf)/1000000 < rttMinConfidence {
		conf = uint64(rttMinConfidence * 1000000)
	}
	atomic.StoreUint64(&d.rttConfidence, conf)

	rtt := time.Duration(atomic.LoadUint64(&d.rttEstimate))
	log.Trace("Relaxed downloader QoS values", "rtt", rtt, "confidence", float64(conf)/1000000.0, "ttl", d.requestTTL())
}

// requestRTT returns the current target round trip time for a download request
// to complete in.
//
// Note, the returned RTT is .9 of the actually estimated RTT. The reason is that
// the downloader tries to adapt queries to the RTT, so multiple RTT values can
// be adapted to, but smaller ones are preffered (stabler download stream).
func (d *Downloader) requestRTT() time.Duration {
	return time.Duration(atomic.LoadUint64(&d.rttEstimate)) * 9 / 10
}

// requestTTL returns the current timeout allowance for a single download request
// to finish under.
func (d *Downloader) requestTTL() time.Duration {
	var (
		rtt  = time.Duration(atomic.LoadUint64(&d.rttEstimate))
		conf = float64(atomic.LoadUint64(&d.rttConfidence)) / 1000000.0
	)
	ttl := time.Duration(ttlScaling) * time.Duration(float64(rtt)/conf)
	if ttl > ttlLimit {
		ttl = ttlLimit
	}
	return ttl
}
