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
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine/types"
	. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	cfg "github.com/dappledger/AnnChain/ann-module/lib/go-config"
	"github.com/dappledger/AnnChain/ann-module/lib/go-p2p"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
)

const (
	BlockchainChannel      = byte(0x40)
	defaultChannelCapacity = 100
	defaultSleepIntervalMS = 500
	trySyncIntervalMS      = 100

	// stop syncing when last block's time is within this much of the system time.
	stopSyncingDurationMinutes = 10

	// ask for best height every 10s
	statusUpdateIntervalSeconds = 10

	// check if we should switch to consensus reactor
	switchToConsensusIntervalSeconds = 1

	maxBlockchainResponseSize = types.MaxBlockSize + 2
)

// BlockchainReactor handles long-term catchup syncing.
type BlockchainReactor struct {
	p2p.BaseReactor

	config     cfg.Config
	store      *BlockStore
	pool       *BlockPool
	fastSync   bool
	requestsCh chan BlockRequest
	timeoutsCh chan string
	lastBlock  *types.Block

	blockVerifier func(types.BlockID, int, *types.Commit) error
	blockExecuter func(*types.Block, *types.PartSet, *types.Commit) error

	evsw types.EventSwitch

	logger *zap.Logger
}

func NewBlockchainReactor(logger *zap.Logger, config cfg.Config, lastBlockHeight int, store *BlockStore, fastSync bool) *BlockchainReactor {
	if lastBlockHeight == store.Height()-1 {
		store.height -= 1 // XXX HACK, make this better
	}
	if lastBlockHeight != store.Height() {
		PanicSanity(Fmt("state (%v) and store (%v) height mismatch", lastBlockHeight, store.Height()))
	}
	requestsCh := make(chan BlockRequest, defaultChannelCapacity)
	timeoutsCh := make(chan string, defaultChannelCapacity)
	pool := NewBlockPool(
		logger,
		store.Height()+1,
		requestsCh,
		timeoutsCh,
	)
	bcR := &BlockchainReactor{
		config:     config,
		store:      store,
		pool:       pool,
		fastSync:   fastSync,
		requestsCh: requestsCh,
		timeoutsCh: timeoutsCh,

		logger: logger,
	}
	bcR.BaseReactor = *p2p.NewBaseReactor(logger, "BlockchainReactor", bcR)
	return bcR
}

func (bcR *BlockchainReactor) SetBlockVerifier(v func(types.BlockID, int, *types.Commit) error) {
	bcR.blockVerifier = v
}

func (bcR *BlockchainReactor) SetBlockExecuter(x func(*types.Block, *types.PartSet, *types.Commit) error) {
	bcR.blockExecuter = x
}

func (bcR *BlockchainReactor) OnStart() error {
	bcR.BaseReactor.OnStart()
	if bcR.fastSync {
		_, err := bcR.pool.Start()
		if err != nil {
			return err
		}
		go bcR.poolRoutine()
	}
	return nil
}

func (bcR *BlockchainReactor) OnStop() {
	bcR.BaseReactor.OnStop()
	bcR.pool.Stop()
}

// Implements Reactor
func (bcR *BlockchainReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		&p2p.ChannelDescriptor{
			ID:                BlockchainChannel,
			Priority:          5,
			SendQueueCapacity: 100,
		},
	}
}

// Implements Reactor
func (bcR *BlockchainReactor) AddPeer(peer *p2p.Peer) {
	// Send peer our state.
	peer.Send(BlockchainChannel, struct{ BlockchainMessage }{&bcStatusResponseMessage{bcR.store.Height()}})
}

// Implements Reactor
func (bcR *BlockchainReactor) RemovePeer(peer *p2p.Peer, reason interface{}) {
	// Remove peer from the pool.
	bcR.pool.RemovePeer(peer.Key)
}

// Implements Reactor
func (bcR *BlockchainReactor) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
	_, msg, err := DecodeMessage(msgBytes)
	if err != nil {
		bcR.logger.Warn("Error decoding message", zap.String("error", err.Error()))
		return
	}

	bcR.logger.Sugar().Debugw("Receive", "src", src, "chID", chID, "msg", msg)

	switch msg := msg.(type) {
	case *bcBlockRequestMessage:
		// Got a request for a block. Respond with block if we have it.
		block := bcR.store.LoadBlock(msg.Height)
		if block != nil {
			msg := &bcBlockResponseMessage{Block: block}
			queued := src.TrySend(BlockchainChannel, struct{ BlockchainMessage }{msg})
			if !queued {
				// queue is full, just ignore.
			}
		} else {
			// TODO peer is asking for things we don't have.
		}
	case *bcBlockResponseMessage:
		// Got a block.
		bcR.pool.AddBlock(src.Key, msg.Block, len(msgBytes))
	case *bcStatusRequestMessage:
		// Send peer our state.
		queued := src.TrySend(BlockchainChannel, struct{ BlockchainMessage }{&bcStatusResponseMessage{bcR.store.Height()}})
		if !queued {
			// sorry
		}
	case *bcStatusResponseMessage:
		// Got a peer status. Unverified.
		bcR.pool.SetPeerHeight(src.Key, msg.Height)
	default:
		bcR.logger.Warn(Fmt("Unknown message type %v", reflect.TypeOf(msg)))
	}
}

// Handle messages from the poolReactor telling the reactor what to do.
// NOTE: Don't sleep in the FOR_LOOP or otherwise slow it down!
// (Except for the SYNC_LOOP, which is the primary purpose and must be synchronous.)
func (bcR *BlockchainReactor) poolRoutine() {
	trySyncTicker := time.NewTicker(trySyncIntervalMS * time.Millisecond)
	statusUpdateTicker := time.NewTicker(statusUpdateIntervalSeconds * time.Second)
	switchToConsensusTicker := time.NewTicker(switchToConsensusIntervalSeconds * time.Second)

FOR_LOOP:
	for {
		select {
		case request := <-bcR.requestsCh: // chan BlockRequest
			peer := bcR.Switch.Peers().Get(request.PeerID)
			if peer == nil {
				continue FOR_LOOP // Peer has since been disconnected.
			}
			msg := &bcBlockRequestMessage{request.Height}
			queued := peer.TrySend(BlockchainChannel, struct{ BlockchainMessage }{msg})
			if !queued {
				// We couldn't make the request, send-queue full.
				// The pool handles timeouts, just let it go.
				continue FOR_LOOP
			}
		case peerID := <-bcR.timeoutsCh:
			// Peer timed out.
			peer := bcR.Switch.Peers().Get(peerID)
			if peer != nil {
				bcR.Switch.StopPeerForError(peer, errors.New("BlockchainReactor Timeout"))
			}
		case _ = <-statusUpdateTicker.C:
			// ask for status updates
			go bcR.BroadcastStatusRequest()
		case _ = <-switchToConsensusTicker.C:
			height, numPending, _ := bcR.pool.GetStatus()
			outbound, inbound, _ := bcR.Switch.NumPeers()
			bcR.logger.Debug("Consensus ticker", zap.Int32("numPending", numPending), zap.Int("total", len(bcR.pool.requesters)),
				zap.Int("outbound", outbound), zap.Int("inbound", inbound))
			if bcR.pool.IsCaughtUp() {
				bcR.logger.Info("Time to switch to consensus reactor!", zap.Int("height", height))
				bcR.pool.Stop()
				types.FireEventSwitchToConsensus(bcR.evsw)
				break FOR_LOOP
			}
		case _ = <-trySyncTicker.C: // chan time
			// This loop can be slow as long as it's doing syncing work.
		SYNC_LOOP:
			for i := 0; i < 10; i++ {
				// See if there are any blocks to sync.
				first, second := bcR.pool.PeekTwoBlocks()
				if first == nil || second == nil {
					// We need both to sync the first block.
					break SYNC_LOOP
				}
				firstParts := first.MakePartSet(bcR.config.GetInt("block_part_size")) // TODO: put part size in parts header?
				firstPartsHeader := firstParts.Header()
				// Finally, verify the first block using the second's commit
				// NOTE: we can probably make this more efficient, but note that calling
				// first.Hash() doesn't verify the tx contents, so MakePartSet() is
				// currently necessary.

				if err := bcR.blockVerifier(types.BlockID{Hash: first.Hash(), PartsHeader: firstPartsHeader}, first.Height, second.LastCommit); err != nil {
					bcR.logger.Error("error in validation", zap.String("error", err.Error()))
					bcR.pool.RedoRequest(first.Height)
					break SYNC_LOOP
				} else {
					bcR.pool.PopRequest()
					if err := bcR.blockExecuter(first, firstParts, second.LastCommit); err != nil {
						// TODO This is bad, are we zombie?
						PanicQ(Fmt("Failed to process committed block (%d:%X): %v", first.Height, first.Hash(), err))
					}
				}
			}
			continue FOR_LOOP
		case <-bcR.Quit:
			break FOR_LOOP
		}
	}
}

func (bcR *BlockchainReactor) BroadcastStatusResponse() error {
	bcR.Switch.Broadcast(BlockchainChannel, struct{ BlockchainMessage }{&bcStatusResponseMessage{bcR.store.Height()}})
	return nil
}

func (bcR *BlockchainReactor) BroadcastStatusRequest() error {
	bcR.Switch.Broadcast(BlockchainChannel, struct{ BlockchainMessage }{&bcStatusRequestMessage{bcR.store.Height()}})
	return nil
}

// implements events.Eventable
func (bcR *BlockchainReactor) SetEventSwitch(evsw types.EventSwitch) {
	bcR.evsw = evsw
}

//-----------------------------------------------------------------------------
// Messages

const (
	msgTypeBlockRequest   = byte(0x10)
	msgTypeBlockResponse  = byte(0x11)
	msgTypeStatusResponse = byte(0x20)
	msgTypeStatusRequest  = byte(0x21)
)

type BlockchainMessage interface{}

var _ = wire.RegisterInterface(
	struct{ BlockchainMessage }{},
	wire.ConcreteType{&bcBlockRequestMessage{}, msgTypeBlockRequest},
	wire.ConcreteType{&bcBlockResponseMessage{}, msgTypeBlockResponse},
	wire.ConcreteType{&bcStatusResponseMessage{}, msgTypeStatusResponse},
	wire.ConcreteType{&bcStatusRequestMessage{}, msgTypeStatusRequest},
)

// TODO: ensure that bz is completely read.
func DecodeMessage(bz []byte) (msgType byte, msg BlockchainMessage, err error) {
	msgType = bz[0]
	n := int(0)
	r := bytes.NewReader(bz)
	msg = wire.ReadBinary(struct{ BlockchainMessage }{}, r, maxBlockchainResponseSize, &n, &err).(struct{ BlockchainMessage }).BlockchainMessage
	if err != nil && n != len(bz) {
		err = errors.New("DecodeMessage() had bytes left over.")
	}
	return
}

//-------------------------------------

type bcBlockRequestMessage struct {
	Height int
}

func (m *bcBlockRequestMessage) String() string {
	return fmt.Sprintf("[bcBlockRequestMessage %v]", m.Height)
}

//-------------------------------------

// NOTE: keep up-to-date with maxBlockchainResponseSize
type bcBlockResponseMessage struct {
	Block *types.Block
}

func (m *bcBlockResponseMessage) String() string {
	return fmt.Sprintf("[bcBlockResponseMessage %v]", m.Block.Height)
}

//-------------------------------------

type bcStatusRequestMessage struct {
	Height int
}

func (m *bcStatusRequestMessage) String() string {
	return fmt.Sprintf("[bcStatusRequestMessage %v]", m.Height)
}

//-------------------------------------

type bcStatusResponseMessage struct {
	Height int
}

func (m *bcStatusResponseMessage) String() string {
	return fmt.Sprintf("[bcStatusResponseMessage %v]", m.Height)
}
