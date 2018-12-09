// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blockchain

import (
	"math"
	"sync"
	"time"

	"go.uber.org/zap"

	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	. "github.com/dappledger/AnnChain/module/lib/go-common"
	flow "github.com/dappledger/AnnChain/module/lib/go-flowrate/flowrate"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

const (
	requestIntervalMS         = 250
	maxTotalRequesters        = 300
	maxPendingRequests        = maxTotalRequesters
	maxPendingRequestsPerPeer = 75
	minRecvRate               = 10240 // 10Kb/s
)

var peerTimeoutSeconds = time.Duration(15) // not const so we can override with tests

/*
	Peers self report their heights when we join the block pool.
	Starting from our latest pool.height, we request blocks
	in sequence from peers that reported higher heights than ours.
	Every so often we ask peers what height they're on so we can keep going.

	Requests are continuously made for blocks of heigher heights until
	the limits. If most of the requests have no available peers, and we
	are not at peer limits, we can probably switch to consensus reactor
*/

type BlockPool struct {
	BaseService
	startTime time.Time

	mtx sync.Mutex
	// block requests
	requesters map[def.INT]*bpRequester
	height     def.INT // the lowest key in requesters.
	numPending int32   // number of requests pending assignment or block response
	// peers
	peers map[string]*bpPeer

	requestsCh chan<- BlockRequest
	timeoutsCh chan<- string

	logger *zap.Logger
}

func NewBlockPool(logger *zap.Logger, start def.INT, requestsCh chan<- BlockRequest, timeoutsCh chan<- string) *BlockPool {
	bp := &BlockPool{
		peers: make(map[string]*bpPeer),

		requesters: make(map[def.INT]*bpRequester),
		height:     start,
		numPending: 0,

		requestsCh: requestsCh,
		timeoutsCh: timeoutsCh,

		logger: logger,
	}
	bp.BaseService = *NewBaseService(logger, "BlockPool", bp)
	return bp
}

func (pool *BlockPool) OnStart() error {
	pool.BaseService.OnStart()
	go pool.makeRequestersRoutine()
	pool.startTime = time.Now()
	return nil
}

func (pool *BlockPool) OnStop() {
	pool.BaseService.OnStop()
}

// Run spawns requesters as needed.
func (pool *BlockPool) makeRequestersRoutine() {
	for {
		if !pool.IsRunning() {
			break
		}
		_, numPending, lenRequesters := pool.GetStatus()
		if numPending >= maxPendingRequests {
			// sleep for a bit.
			time.Sleep(requestIntervalMS * time.Millisecond)
			// check for timed out peers
			pool.removeTimedoutPeers()
		} else if lenRequesters >= maxTotalRequesters {
			// sleep for a bit.
			time.Sleep(requestIntervalMS * time.Millisecond)
			// check for timed out peers
			pool.removeTimedoutPeers()
		} else {
			// request for more blocks.
			pool.makeNextRequester()
		}
	}
}

func (pool *BlockPool) removeTimedoutPeers() {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	for _, peer := range pool.peers {
		if !peer.didTimeout && peer.numPending > 0 {
			curRate := peer.recvMonitor.Status().CurRate
			// XXX remove curRate != 0
			if curRate != 0 && curRate < minRecvRate {
				pool.sendTimeout(peer.id)
				pool.logger.Warn("SendTimeout", zap.String("peer", peer.id), zap.String("reason", "curRate too low"))
				peer.didTimeout = true
			}
		}
		if peer.didTimeout {
			pool.removePeer(peer.id)
		}
	}
}

func (pool *BlockPool) GetStatus() (height def.INT, numPending int32, lenRequesters int) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	return pool.height, pool.numPending, len(pool.requesters)
}

// TODO: relax conditions, prevent abuse.
func (pool *BlockPool) IsCaughtUp() bool {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	height := pool.height

	// Need at least 1 peer to be considered caught up.
	if len(pool.peers) == 0 {
		pool.logger.Debug("Blockpool has no peers")
		return false
	}

	var maxPeerHeight def.INT = 0
	for _, peer := range pool.peers {
		maxPeerHeight = MaxInt64(maxPeerHeight, peer.height)
	}

	isCaughtUp := (height > 0 || time.Now().Sub(pool.startTime) > 5*time.Second) && (maxPeerHeight == 0 || height >= maxPeerHeight)
	pool.logger.Info(Fmt("IsCaughtUp: %v", isCaughtUp), zap.Int64("height", height), zap.Int64("maxPeerHeight", maxPeerHeight))
	return isCaughtUp
}

// We need to see the second block's Commit to validate the first block.
// So we peek two blocks at a time.
// The caller will verify the commit.
func (pool *BlockPool) PeekTwoBlocks() (first *agtypes.BlockCache, second *agtypes.BlockCache) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	if r := pool.requesters[pool.height]; r != nil {
		first = r.getBlock()
	}
	if r := pool.requesters[pool.height+1]; r != nil {
		second = r.getBlock()
	}
	return
}

// Pop the first block at pool.height
// It must have been validated by 'second'.Commit from PeekTwoBlocks().
func (pool *BlockPool) PopRequest() {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	if r := pool.requesters[pool.height]; r != nil {
		/*  The block can disappear at any time, due to removePeer().
		if r := pool.requesters[pool.height]; r == nil || r.block == nil {
			PanicSanity("PopRequest() requires a valid block")
		}
		*/
		r.Stop()
		delete(pool.requesters, pool.height)
		pool.height++
	} else {
		PanicSanity(Fmt("Expected requester to pop, got nothing at height %v", pool.height))
	}
}

// Invalidates the block at pool.height,
// Remove the peer and redo request from others.
func (pool *BlockPool) RedoRequest(height def.INT) {
	pool.mtx.Lock()
	request := pool.requesters[height]
	pool.mtx.Unlock()

	if request.block == nil {
		PanicSanity("Expected block to be non-nil")
	}
	// RemovePeer will redo all requesters associated with this peer.
	// TODO: record this malfeasance
	pool.RemovePeer(request.peerID)
}

// TODO: ensure that blocks come in order for each peer.
func (pool *BlockPool) AddBlock(peerID string, block *pbtypes.Block, blockSize int) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	requester := pool.requesters[block.Header.Height]
	if requester == nil {
		return
	}

	if requester.setBlock(block, peerID) {
		pool.numPending--
		peer := pool.peers[peerID]
		peer.decrPending(blockSize)
	} else {
		// Bad peer?
	}
}

// Sets the peer's alleged blockchain height.
func (pool *BlockPool) SetPeerHeight(peerID string, height def.INT) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	peer := pool.peers[peerID]
	if peer != nil {
		peer.height = height
	} else {
		peer = newBPPeer(pool, peerID, height)
		pool.peers[peerID] = peer
	}
}

func (pool *BlockPool) RemovePeer(peerID string) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	pool.removePeer(peerID)
}

func (pool *BlockPool) removePeer(peerID string) {
	for _, requester := range pool.requesters {
		if requester.getPeerID() == peerID {
			pool.numPending++
			go requester.redo() // pick another peer and ...
		}
	}
	delete(pool.peers, peerID)
}

// Pick an available peer with at least the given minHeight.
// If no peers are available, returns nil.
func (pool *BlockPool) pickIncrAvailablePeer(minHeight def.INT) *bpPeer {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	for _, peer := range pool.peers {
		if peer.didTimeout {
			pool.removePeer(peer.id)
			continue
		} else {
		}
		if peer.numPending >= maxPendingRequestsPerPeer {
			continue
		}
		if peer.height < minHeight {
			continue
		}
		peer.incrPending()
		return peer
	}
	return nil
}

func (pool *BlockPool) makeNextRequester() {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	nextHeight := pool.height + def.INT(len(pool.requesters))
	request := newBPRequester(pool, nextHeight)

	pool.requesters[nextHeight] = request
	pool.numPending++

	request.Start()
}

func (pool *BlockPool) sendRequest(height def.INT, peerID string) {
	if !pool.IsRunning() {
		return
	}
	pool.requestsCh <- BlockRequest{height, peerID}
}

func (pool *BlockPool) sendTimeout(peerID string) {
	if !pool.IsRunning() {
		return
	}
	pool.timeoutsCh <- peerID
}

func (pool *BlockPool) debug() string {
	pool.mtx.Lock() // Lock
	defer pool.mtx.Unlock()

	str := ""
	for h := pool.height; h < pool.height+def.INT(len(pool.requesters)); h++ {
		if pool.requesters[h] == nil {
			str += Fmt("H(%v):X ", h)
		} else {
			str += Fmt("H(%v):", h)
			str += Fmt("B?(%v) ", pool.requesters[h].block != nil)
		}
	}
	return str
}

//-------------------------------------

type bpPeer struct {
	pool        *BlockPool
	id          string
	recvMonitor *flow.Monitor

	mtx        sync.Mutex
	height     def.INT
	numPending int32
	timeout    *time.Timer
	didTimeout bool
}

func newBPPeer(pool *BlockPool, peerID string, height def.INT) *bpPeer {
	peer := &bpPeer{
		pool:       pool,
		id:         peerID,
		height:     height,
		numPending: 0,
	}
	return peer
}

func (peer *bpPeer) resetMonitor() {
	peer.recvMonitor = flow.New(time.Second, time.Second*40)
	var initialValue = float64(minRecvRate) * math.E
	peer.recvMonitor.SetREMA(initialValue)
}

func (peer *bpPeer) resetTimeout() {
	if peer.timeout == nil {
		peer.timeout = time.AfterFunc(time.Second*peerTimeoutSeconds, peer.onTimeout)
	} else {
		peer.timeout.Reset(time.Second * peerTimeoutSeconds)
	}
}

func (peer *bpPeer) incrPending() {
	if peer.numPending == 0 {
		peer.resetMonitor()
		peer.resetTimeout()
	}
	peer.numPending++
}

func (peer *bpPeer) decrPending(recvSize int) {
	peer.numPending--
	if peer.numPending == 0 {
		peer.timeout.Stop()
	} else {
		peer.recvMonitor.Update(recvSize)
		peer.resetTimeout()
	}
}

func (peer *bpPeer) onTimeout() {
	peer.pool.mtx.Lock()
	defer peer.pool.mtx.Unlock()

	peer.pool.sendTimeout(peer.id)
	peer.pool.logger.Warn("SendTimeout", zap.String("peer", peer.id), zap.String("reason", "onTimeout"))
	peer.didTimeout = true
}

//-------------------------------------

type bpRequester struct {
	BaseService
	pool       *BlockPool
	height     def.INT
	gotBlockCh chan struct{}
	redoCh     chan struct{}

	mtx    sync.Mutex
	peerID string
	block  *agtypes.BlockCache
}

func newBPRequester(pool *BlockPool, height def.INT) *bpRequester {
	bpr := &bpRequester{
		pool:       pool,
		height:     height,
		gotBlockCh: make(chan struct{}),
		redoCh:     make(chan struct{}),

		peerID: "",
		block:  nil,
	}
	bpr.BaseService = *NewBaseService(nil, "bpRequester", bpr)
	return bpr
}

func (bpr *bpRequester) OnStart() error {
	bpr.BaseService.OnStart()
	go bpr.requestRoutine()
	return nil
}

// Returns true if the peer matches
func (bpr *bpRequester) setBlock(block *pbtypes.Block, peerID string) bool {
	bpr.mtx.Lock()
	if bpr.block != nil || bpr.peerID != peerID {
		bpr.mtx.Unlock()
		return false
	}
	bpr.block = agtypes.MakeBlockCache(block)
	bpr.mtx.Unlock()

	bpr.gotBlockCh <- struct{}{}
	return true
}

func (bpr *bpRequester) getBlock() *agtypes.BlockCache {
	bpr.mtx.Lock()
	defer bpr.mtx.Unlock()
	return bpr.block
}

func (bpr *bpRequester) getPeerID() string {
	bpr.mtx.Lock()
	defer bpr.mtx.Unlock()
	return bpr.peerID
}

func (bpr *bpRequester) reset() {
	bpr.mtx.Lock()
	bpr.peerID = ""
	bpr.block = nil
	bpr.mtx.Unlock()
}

// Tells bpRequester to pick another peer and try again.
// NOTE: blocking
func (bpr *bpRequester) redo() {
	bpr.redoCh <- struct{}{}
}

// Responsible for making more requests as necessary
// Returns only when a block is found (e.g. AddBlock() is called)
func (bpr *bpRequester) requestRoutine() {
OUTER_LOOP:
	for {
		// Pick a peer to send request to.
		var peer *bpPeer = nil
	PICK_PEER_LOOP:
		for {
			if !bpr.IsRunning() || !bpr.pool.IsRunning() {
				return
			}
			peer = bpr.pool.pickIncrAvailablePeer(bpr.height)
			if peer == nil {
				time.Sleep(requestIntervalMS * time.Millisecond)
				continue PICK_PEER_LOOP
			}
			break PICK_PEER_LOOP
		}
		bpr.mtx.Lock()
		bpr.peerID = peer.id
		bpr.mtx.Unlock()

		// Send request and wait.
		bpr.pool.sendRequest(bpr.height, peer.id)
		select {
		case <-bpr.pool.Quit:
			bpr.Stop()
			return
		case <-bpr.Quit:
			return
		case <-bpr.redoCh:
			bpr.reset()
			continue OUTER_LOOP // When peer is removed
		case <-bpr.gotBlockCh:
			// We got the block, now see if it's good.
			select {
			case <-bpr.pool.Quit:
				bpr.Stop()
				return
			case <-bpr.Quit:
				return
			case <-bpr.redoCh:
				bpr.reset()
				continue OUTER_LOOP
			}
		}
	}
}

//-------------------------------------

type BlockRequest struct {
	Height def.INT
	PeerID string
}
