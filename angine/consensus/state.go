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

package consensus

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"
	"unsafe"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	bc "github.com/dappledger/AnnChain/angine/blockchain"
	mempl "github.com/dappledger/AnnChain/angine/mempool"
	csspb "github.com/dappledger/AnnChain/angine/protos/consensus"
	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	sm "github.com/dappledger/AnnChain/angine/state"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

//-----------------------------------------------------------------------------
// Timeout Parameters

// TimeoutParams holds timeouts and deltas for each round step.
// All timeouts and deltas in milliseconds.
type TimeoutParams struct {
	Propose0          def.INT
	ProposeDelta      def.INT
	Prevote0          def.INT
	PrevoteDelta      def.INT
	Precommit0        def.INT
	PrecommitDelta    def.INT
	Commit0           def.INT
	SkipTimeoutCommit bool
}

// Wait this long for a proposal
func (tp *TimeoutParams) Propose(round def.INT) time.Duration {
	return time.Duration(tp.Propose0+tp.ProposeDelta*round) * time.Millisecond
}

// After receiving any +2/3 prevote, wait this long for stragglers
func (tp *TimeoutParams) Prevote(round def.INT) time.Duration {
	return time.Duration(tp.Prevote0+tp.PrevoteDelta*round) * time.Millisecond
}

// After receiving any +2/3 precommits, wait this long for stragglers
func (tp *TimeoutParams) Precommit(round def.INT) time.Duration {
	return time.Duration(tp.Precommit0+tp.PrecommitDelta*round) * time.Millisecond
}

// After receiving +2/3 precommits for a single block (a commit), wait this long for stragglers in the next height's RoundStepNewHeight
func (tp *TimeoutParams) Commit(t time.Time) time.Time {
	return t.Add(time.Duration(tp.Commit0) * time.Millisecond)
}

// InitTimeoutParamsFromConfig initializes parameters from config
func InitTimeoutParamsFromConfig(config *viper.Viper) *TimeoutParams {
	return &TimeoutParams{
		Propose0:          def.INT(config.GetInt("timeout_propose")),
		ProposeDelta:      def.INT(config.GetInt("timeout_propose_delta")),
		Prevote0:          def.INT(config.GetInt("timeout_prevote")),
		PrevoteDelta:      def.INT(config.GetInt("timeout_prevote_delta")),
		Precommit0:        def.INT(config.GetInt("timeout_precommit")),
		PrecommitDelta:    def.INT(config.GetInt("timeout_precommit_delta")),
		Commit0:           def.INT(config.GetInt("timeout_commit")),
		SkipTimeoutCommit: config.GetBool("skip_timeout_commit"),
	}
}

//-----------------------------------------------------------------------------
// Errors

var (
	ErrInvalidProposalSignature = errors.New("Error invalid proposal signature")
	ErrInvalidProposalPOLRound  = errors.New("Error invalid proposal POL round")
	ErrAddingVote               = errors.New("Error adding vote")
	ErrVoteHeightMismatch       = errors.New("Error vote height mismatch")
)

//-----------------------------------------------------------------------------

// Immutable when returned from ConsensusState.GetRoundState()
type RoundState struct {
	Height             def.INT // Height we are working on
	Round              def.INT
	Step               csspb.RoundStepType
	StartTime          time.Time
	CommitTime         time.Time // Subjective time when +2/3 precommits for Block at Round were found
	Validators         *agtypes.ValidatorSet
	Proposal           *pbtypes.Proposal
	ProposalBlock      *agtypes.BlockCache
	ProposalBlockParts *agtypes.PartSet
	LockedRound        def.INT
	LockedBlock        *agtypes.BlockCache
	LockedBlockParts   *agtypes.PartSet
	Votes              *HeightVoteSet
	CommitRound        def.INT          //
	LastCommit         *agtypes.VoteSet // Last precommits at Height-1
	LastValidators     *agtypes.ValidatorSet
}

func (rs *RoundState) RoundStateEvent() agtypes.EventDataRoundState {
	edrs := agtypes.EventDataRoundState{
		EventDataRoundStateJson: agtypes.EventDataRoundStateJson{
			Height:     rs.Height,
			Round:      rs.Round,
			Step:       rs.Step.CString(),
			RoundState: rs,
		},
	}
	return edrs
}

func (rs *RoundState) String() string {
	return rs.StringIndented("")
}

func (rs *RoundState) StringIndented(indent string) string {
	return fmt.Sprintf(`RoundState{
%s  H:%v R:%v S:%v
%s  StartTime:     %v
%s  CommitTime:    %v
%s  Validators:    %v
%s  Proposal:      %v
%s  ProposalBlock: %v %v
%s  LockedRound:   %v
%s  LockedBlock:   %v %v
%s  Votes:         %v
%s  LastCommit: %v
%s  LastValidators:    %v
%s}`,
		indent, rs.Height, rs.Round, rs.Step,
		indent, rs.StartTime,
		indent, rs.CommitTime,
		indent, rs.Validators.StringIndented(indent+"    "),
		indent, rs.Proposal,
		indent, rs.ProposalBlockParts.StringShort(), rs.ProposalBlock.StringShort(),
		indent, rs.LockedRound,
		indent, rs.LockedBlockParts.StringShort(), rs.LockedBlock.StringShort(),
		indent, rs.Votes.StringIndented(indent+"    "),
		indent, rs.LastCommit.StringShort(),
		indent, rs.LastValidators.StringIndented(indent+"    "),
		indent)
}

func (rs *RoundState) StringShort() string {
	return fmt.Sprintf(`RoundState{H:%v R:%v S:%v ST:%v}`,
		rs.Height, rs.Round, rs.Step, rs.StartTime)
}

//-----------------------------------------------------------------------------

var (
	msgQueueSize = 1000

	int0            def.INT
	sizeOfInt       = unsafe.Sizeof(int0)
	emptyRoundLimit = 1 << (*(*uint)(unsafe.Pointer(&sizeOfInt))*8 - 2)
)

const (
	CssMsgTypeProposal  = byte(0x01)
	CssMsgTypeBlockPart = byte(0x02)
	CssMsgTypeVote      = byte(0x03)
)

func cssMsgType(msg csspb.ConsensusMsgItfc) byte {
	switch csspb.GetMessageType(msg) {
	case csspb.MsgType_Proposal:
		return CssMsgTypeProposal
	case csspb.MsgType_BlockPart:
		return CssMsgTypeBlockPart
	case csspb.MsgType_Vote:
		return CssMsgTypeVote
	}
	return 0
}

type msgInfoJson struct {
	Msg     csspb.StConsensusMsg `json:"msg"`
	PeerKey string               `json:"peer_key"`
}

// msgs from the reactor which may update the state
type msgInfo struct {
	msgInfoJson
}

func (mi *msgInfo) GetMsg() csspb.ConsensusMsgItfc {
	return mi.Msg.ConsensusMsgItfc
}

func genMsgInfo(msg csspb.ConsensusMsgItfc, pk string) (retMsg msgInfo) {
	retMsg.Msg = csspb.StConsensusMsg{msg}
	retMsg.PeerKey = pk
	return
}

func (mi msgInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(mi.msgInfoJson)
}

func (mi *msgInfo) UnmarshalJSON(data []byte) error {
	var mij msgInfoJson
	err := json.Unmarshal(data, &mij)
	if err != nil {
		return err
	}
	mi.msgInfoJson = mij
	return nil
}

// internally generated messages which may update the state
type timeoutInfo struct {
	timeoutInfoJson
}

type timeoutInfoJson struct {
	Duration time.Duration       `json:"duration"`
	Height   def.INT             `json:"height"`
	Round    def.INT             `json:"round"`
	Step     csspb.RoundStepType `json:"step"`
}

func (ti timeoutInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(&ti.timeoutInfoJson)
}

func (ti *timeoutInfo) UnmarshalJSON(data []byte) error {
	var tij timeoutInfoJson
	err := json.Unmarshal(data, &tij)
	if err != nil {
		return err
	}
	ti.timeoutInfoJson = tij
	return nil
}

func (ti *timeoutInfo) String() string {
	return fmt.Sprintf("%v ; %d/%d %v", ti.Duration, ti.Height, ti.Round, ti.Step)
}

type PrivValidator interface {
	GetAddress() []byte
	SignVote(chainID string, vote *pbtypes.Vote) error
	SignProposal(chainID string, proposal *pbtypes.Proposal) error
}

// Tracks consensus state across block heights and rounds.
type ConsensusState struct {
	BaseService

	config     *viper.Viper
	blockStore *bc.BlockStore
	mempool    *mempl.Mempool

	conR *ConsensusReactor

	privValidator PrivValidator // for signing votes

	mtx sync.Mutex
	RoundState
	state *sm.State // State until height-1.

	peerMsgQueue     chan msgInfo   // serializes msgs affecting state (proposals, block parts, votes)
	internalMsgQueue chan msgInfo   // like peerMsgQueue but for our own proposals, parts, votes
	timeoutTicker    TimeoutTicker  // ticker for timeouts
	timeoutParams    *TimeoutParams // parameters and functions for timeout intervals

	evsw agtypes.EventSwitch

	wal        *WAL
	replayMode bool // so we don't log signing errors during replay

	nSteps int // used for testing to limit the number of transitions the state makes

	// allow certain function to be overwritten for testing
	decideProposal func(height, round def.INT)
	doPrevote      func(height, round def.INT)
	setProposal    func(proposal *pbtypes.Proposal) error

	done chan struct{}

	logger  *zap.Logger
	slogger *zap.SugaredLogger

	badvoteCollector agtypes.IBadVoteCollector
}

func NewConsensusState(logger *zap.Logger, config *viper.Viper, state *sm.State, blockStore *bc.BlockStore, mempool *mempl.Mempool) *ConsensusState {
	cs := &ConsensusState{
		config:           config,
		blockStore:       blockStore,
		mempool:          mempool,
		peerMsgQueue:     make(chan msgInfo, msgQueueSize),
		internalMsgQueue: make(chan msgInfo, msgQueueSize),
		timeoutTicker:    NewTimeoutTicker(logger),
		timeoutParams:    InitTimeoutParamsFromConfig(config),
		done:             make(chan struct{}),
		logger:           logger,
		slogger:          logger.Sugar(),
	}
	// set function defaults (may be overwritten before calling Start)
	cs.decideProposal = cs.defaultDecideProposal
	cs.doPrevote = cs.defaultDoPrevote
	cs.setProposal = cs.defaultSetProposal

	cs.updateToState(state)
	// Don't call scheduleRound0 yet.
	// We do that upon Start().
	cs.reconstructLastCommit(state)
	cs.BaseService = *NewBaseService(logger, "ConsensusState", cs)

	walDir := cs.config.GetString("cs_wal_dir")
	err := EnsureDir(walDir, 0700)
	if err != nil {
		logger.Error("Error ensuring ConsensusState wal dir", zap.String("error", err.Error()))
		return nil
	}
	err = cs.OpenWAL(walDir)
	if err != nil {
		logger.Error("Error loading ConsensusState wal", zap.String("error", err.Error()))
		return nil
	}

	return cs
}

//----------------------------------------
// Public interface

// implements events.Eventable
func (cs *ConsensusState) SetEventSwitch(evsw agtypes.EventSwitch) {
	cs.evsw = evsw
}

func (cs *ConsensusState) BindReactor(r *ConsensusReactor) {
	cs.conR = r
}

func (cs *ConsensusState) String() string {
	// better not to access shared variables
	return Fmt("ConsensusState") //(H:%v R:%v S:%v", cs.Height, cs.Round, cs.Step)
}

func (cs *ConsensusState) GetState() *sm.State {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	return cs.state.Copy()
}

func (cs *ConsensusState) GetRoundState() *RoundState {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	return cs.getRoundState()
}

func (cs *ConsensusState) getRoundState() *RoundState {
	rs := cs.RoundState // copy
	return &rs
}

func (cs *ConsensusState) GetTotalVotingPower() int64 {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	return cs.Validators.TotalVotingPower()
}

func (cs *ConsensusState) GetValidators() (def.INT, []*agtypes.Validator) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	return cs.state.LastBlockHeight, cs.state.Validators.Copy().Validators
}

// Sets our private validator account for signing votes.
func (cs *ConsensusState) SetPrivValidator(priv PrivValidator) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	cs.privValidator = priv
}

// Set the local timer
func (cs *ConsensusState) SetTimeoutTicker(timeoutTicker TimeoutTicker) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	cs.timeoutTicker = timeoutTicker
}

func (cs *ConsensusState) SetBadVoteCollector(c agtypes.IBadVoteCollector) {
	cs.badvoteCollector = c
}

func (cs *ConsensusState) LoadCommit(height def.INT) *agtypes.CommitCache {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	if height == cs.blockStore.Height() {
		return cs.blockStore.LoadSeenCommit(height)
	}
	return cs.blockStore.LoadBlockCommit(height)
}

func (cs *ConsensusState) OnStart() error {
	cs.BaseService.OnStart()

	// If the latest block was applied in the abci handshake,
	// we may not have written the current height to the wal,
	// so check here and write it if not found.
	// TODO: remove this and run the handhsake/replay
	// through the consensus state with a mock app
	gr, found, err := cs.wal.group.Search("#HEIGHT: ", makeHeightSearchFunc(cs.Height))
	if (err == io.EOF || !found) && cs.Step == csspb.RoundStepType_NewHeight {
		cs.logger.Warn("Height not found in wal. Writing new height", zap.Int64("height", cs.Height))
		rs := cs.RoundStateEvent()
		cs.wal.Save(&rs)
	} else if err != nil {
		return err
	}
	if gr != nil {
		gr.Close()
	}

	// we need the timeoutRoutine for replay so
	//  we don't block on the tick chan.
	// NOTE: we will get a build up of garbage go routines
	//  firing on the tockChan until the receiveRoutine is started
	//  to deal with them (by that point, at most one will be valid)
	cs.timeoutTicker.Start()

	// we may have lost some votes if the process crashed
	// reload from consensus log to catchup
	if err := cs.catchupReplay(cs.Height); err != nil {
		cs.logger.Error("Error on catchup replay", zap.String("error", err.Error()))
		// let's go for it anyways, maybe we're fine
	}

	// now start the receiveRoutine
	go cs.receiveRoutine(0)

	// schedule the first round!
	// use GetRoundState so we don't race the receiveRoutine for access
	cs.scheduleRound0(cs.GetRoundState())

	return nil
}

// timeoutRoutine: receive requests for timeouts on tickChan and fire timeouts on tockChan
// receiveRoutine: serializes processing of proposoals, block parts, votes; coordinates state transitions
func (cs *ConsensusState) startRoutines(maxSteps int) {
	cs.timeoutTicker.Start()
	go cs.receiveRoutine(maxSteps)
}

func (cs *ConsensusState) OnStop() {
	cs.BaseService.OnStop()

	cs.timeoutTicker.Stop()

	// Make BaseService.Wait() wait until cs.wal.Wait()
	if cs.wal != nil && cs.IsRunning() {
		cs.wal.Wait()
	}
}

// NOTE: be sure to Stop() the event switch and drain
// any event channels or this may deadlock
func (cs *ConsensusState) Wait() {
	<-cs.done
}

// Open file to log all consensus messages and timeouts for deterministic accountability
func (cs *ConsensusState) OpenWAL(walDir string) (err error) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	wal, err := NewWAL(cs.logger, walDir, cs.config.GetBool("cs_wal_light"))
	if err != nil {
		return err
	}
	cs.wal = wal
	return nil
}

//------------------------------------------------------------
// Public interface for passing messages into the consensus state,
// possibly causing a state transition
// TODO: should these return anything or let callers just use events?

// May block on send if queue is full.
func (cs *ConsensusState) AddVote(vote *pbtypes.Vote, peerKey string) (added bool, err error) {
	if peerKey == "" {
		cs.internalMsgQueue <- genMsgInfo(&csspb.VoteMessage{vote}, "")
	} else {
		cs.peerMsgQueue <- genMsgInfo(&csspb.VoteMessage{vote}, peerKey)
	}

	// TODO: wait for event?!
	return false, nil
}

// May block on send if queue is full.
func (cs *ConsensusState) SetProposal(proposal *pbtypes.Proposal, peerKey string) error {

	if peerKey == "" {
		cs.internalMsgQueue <- genMsgInfo(&csspb.ProposalMessage{proposal}, "")
	} else {
		cs.peerMsgQueue <- genMsgInfo(&csspb.ProposalMessage{proposal}, peerKey)
	}

	// TODO: wait for event?!
	return nil
}

// May block on send if queue is full.
func (cs *ConsensusState) AddProposalBlockPart(height, round def.INT, part *pbtypes.Part, peerKey string) error {

	if peerKey == "" {
		cs.internalMsgQueue <- genMsgInfo(&csspb.BlockPartMessage{height, round, part}, "")
	} else {
		cs.peerMsgQueue <- genMsgInfo(&csspb.BlockPartMessage{height, round, part}, peerKey)
	}

	// TODO: wait for event?!
	return nil
}

// May block on send if queue is full.
func (cs *ConsensusState) SetProposalAndBlock(proposal *pbtypes.Proposal, block *pbtypes.Block, parts *agtypes.PartSet, peerKey string) error {
	cs.SetProposal(proposal, peerKey)
	pdata := proposal.GetData()
	for i := 0; i < int(parts.Total()); i++ {
		part := parts.GetPart(i)
		cs.AddProposalBlockPart(pdata.Height, pdata.Round, part, peerKey)
	}
	return nil // TODO errors
}

//------------------------------------------------------------
// internal functions for managing the state

func (cs *ConsensusState) updateHeight(height def.INT) {
	cs.Height = height
}

func (cs *ConsensusState) updateRoundStep(round def.INT, step csspb.RoundStepType) {
	cs.Round = round
	cs.Step = step
}

// enterNewRound(height, 0) at cs.StartTime.
func (cs *ConsensusState) scheduleRound0(rs *RoundState) {
	sleepDuration := rs.StartTime.Sub(time.Now())
	cs.scheduleTimeout(sleepDuration, rs.Height, 0, csspb.RoundStepType_NewHeight)
}

// Attempt to schedule a timeout (by sending timeoutInfo on the tickChan)
func (cs *ConsensusState) scheduleTimeout(duration time.Duration, height, round def.INT, step csspb.RoundStepType) {
	cs.timeoutTicker.ScheduleTimeout(timeoutInfo{
		timeoutInfoJson: timeoutInfoJson{
			Duration: duration,
			Height:   height,
			Round:    round,
			Step:     step,
		},
	})
}

// send a msg into the receiveRoutine regarding our own proposal, block part, or vote
func (cs *ConsensusState) sendInternalMessage(mi msgInfo) {
	select {
	case cs.internalMsgQueue <- mi:
	default:
		// NOTE: using the go-routine means our votes can
		// be processed out of order.
		// TODO: use CList here for strict determinism and
		// attempt push to internalMsgQueue in receiveRoutine
		cs.logger.Warn("Internal msg queue is full. Using a go-routine")
		go func() { cs.internalMsgQueue <- mi }()
	}
}

// Reconstruct LastCommit from SeenCommit, which we saved along with the block,
// (which happens even before saving the state)
func (cs *ConsensusState) reconstructLastCommit(state *sm.State) {
	if state.LastBlockHeight == 0 {
		return
	}
	seenCommit := cs.blockStore.LoadSeenCommit(state.LastBlockHeight)
	lastPrecommits := agtypes.NewVoteSet(cs.config.GetString("chain_id"), state.LastBlockHeight, seenCommit.Round(), pbtypes.VoteType_Precommit, state.LastValidators)
	for _, precommit := range seenCommit.Precommits {
		if !precommit.Exist() {
			continue
		}
		added, err := lastPrecommits.AddVote(precommit)
		if !added || err != nil {
			PanicCrisis(Fmt("Failed to reconstruct LastCommit: %v", err))
		}
	}
	if !lastPrecommits.HasTwoThirdsMajority() {
		PanicSanity("Failed to reconstruct LastCommit: Does not have +2/3 maj")
	}
	cs.LastCommit = lastPrecommits
}

// Updates ConsensusState and increments height to match that of state.
// The round becomes 0 and cs.Step becomes RoundStepNewHeight.
func (cs *ConsensusState) updateToState(state *sm.State) {
	if cs.CommitRound > -1 && 0 < cs.Height && cs.Height != state.LastBlockHeight {
		PanicSanity(Fmt("updateToState() expected state height of %v but found %v",
			cs.Height, state.LastBlockHeight))
	}
	if cs.state != nil && cs.state.LastBlockHeight+1 != cs.Height {
		// This might happen when someone else is mutating cs.state.
		// Someone forgot to pass in state.Copy() somewhere?!
		PanicSanity(Fmt("Inconsistent cs.state.LastBlockHeight+1 %v vs cs.Height %v", cs.state.LastBlockHeight+1, cs.Height))
	}

	// If state isn't further out than cs.state, just ignore.
	// This happens when SwitchToConsensus() is called in the reactor.
	// We don't want to reset e.g. the Votes.
	if cs.state != nil && (state.LastBlockHeight <= cs.state.LastBlockHeight) {
		cs.logger.Debug("Ignoring updateToState()", zap.Int64("newHeight", state.LastBlockHeight+1), zap.Int64("oldHeight", cs.state.LastBlockHeight+1))
		return
	}

	// Reset fields based on state.
	validators := state.Validators
	height := state.LastBlockHeight + 1 // Next desired block height
	lastPrecommits := (*agtypes.VoteSet)(nil)
	if cs.CommitRound > -1 && cs.Votes != nil {
		if !cs.Votes.Precommits(cs.CommitRound).HasTwoThirdsMajority() {
			PanicSanity("updateToState(state) called but last Precommit round didn't have +2/3")
		}
		lastPrecommits = cs.Votes.Precommits(cs.CommitRound)
	}

	// RoundState fields
	cs.updateHeight(height)
	cs.updateRoundStep(0, csspb.RoundStepType_NewHeight)
	if cs.CommitTime.IsZero() {
		// "Now" makes it easier to sync up dev nodes.
		// We add timeoutCommit to allow transactions
		// to be gathered for the first block.
		// And alternative solution that relies on clocks:
		//  cs.StartTime = state.LastBlockTime.Add(timeoutCommit)
		cs.StartTime = cs.timeoutParams.Commit(time.Now())
	} else {
		cs.StartTime = cs.timeoutParams.Commit(cs.CommitTime)
	}
	cs.Validators = validators
	cs.Proposal = nil
	cs.ProposalBlock = nil
	cs.ProposalBlockParts = nil
	cs.LockedRound = 0
	cs.LockedBlock = nil
	cs.LockedBlockParts = nil
	cs.Votes = NewHeightVoteSet(cs.config.GetString("chain_id"), height, validators)
	cs.CommitRound = -1
	cs.LastCommit = lastPrecommits
	cs.LastValidators = state.LastValidators

	cs.state = state

	// Finally, broadcast RoundState
	cs.newStep()
}

func (cs *ConsensusState) newStep() {
	rs := cs.RoundStateEvent()
	cs.wal.Save(&rs)
	cs.nSteps++
	// newStep is called by updateToStep in NewConsensusState before the evsw is set!
	if cs.evsw != nil {
		agtypes.FireEventNewRoundStep(cs.evsw, rs)
	}
}

//-----------------------------------------
// the main go routines

// a nice idea but probably more trouble than its worth
func (cs *ConsensusState) stopTimer() {
	cs.timeoutTicker.Stop()
}

// receiveRoutine handles messages which may cause state transitions.
// it's argument (n) is the number of messages to process before exiting - use 0 to run forever
// It keeps the RoundState and is the only thing that updates it.
// Updates (state transitions) happen on timeouts, complete proposals, and 2/3 majorities
func (cs *ConsensusState) receiveRoutine(maxSteps int) {
	for {
		if maxSteps > 0 {
			if cs.nSteps >= maxSteps {
				cs.logger.Warn("reached max steps. exiting receive routine")
				cs.nSteps = 0
				return
			}
		}
		rs := cs.RoundState
		var mi msgInfo

		select {
		case mi = <-cs.peerMsgQueue:
			cs.wal.Save(&mi)
			// handles proposals, block parts, votes
			// may generate internal events (votes, complete proposals, 2/3 majorities)
			cs.handleMsg(mi, rs)
		case mi = <-cs.internalMsgQueue:
			cs.wal.Save(&mi)
			// handles proposals, block parts, votes
			cs.handleMsg(mi, rs)
		case ti := <-cs.timeoutTicker.Chan(): // tockChan:
			cs.wal.Save(&ti)
			// if the timeout is relevant to the rs
			// go to the next step
			cs.handleTimeout(ti, rs)
		case <-cs.Quit:

			// NOTE: the internalMsgQueue may have signed messages from our
			// priv_val that haven't hit the WAL, but its ok because
			// priv_val tracks LastSig

			// close wal now that we're done writing to it
			if cs.wal != nil {
				cs.wal.Stop()
			}

			close(cs.done)
			return
		}
	}
}

// state transitions on complete-proposal, 2/3-any, 2/3-one
func (cs *ConsensusState) handleMsg(mi msgInfo, rs RoundState) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	var err error
	msg, peerKey := mi.GetMsg(), mi.PeerKey
	switch msg := msg.(type) {
	case *csspb.ProposalMessage:
		// will not cause transition.
		// once proposal is set, we can receive block parts
		err = cs.setProposal(msg.Proposal)
	case *csspb.BlockPartMessage:
		// if the proposal is complete, we'll enterPrevote or tryFinalizeCommit
		_, err = cs.addProposalBlockPart(msg.Height, msg.Part, peerKey != "")
		if err != nil && msg.Round != cs.Round {
			err = nil
		}
	case *csspb.VoteMessage:
		// attempt to add the vote and dupeout the validator if its a duplicate signature
		// if the vote gives us a 2/3-any or 2/3-one, we transition
		err := cs.tryAddVote(msg.Vote, peerKey)
		var evidence *agtypes.HypoBadVoteEvidence
		if err == ErrAddingVote {
			evidence = &agtypes.HypoBadVoteEvidence{
				PubKey:   peerKey,
				VoteType: msg.Vote.Data.Type,
				Height:   cs.Height,
				Round:    cs.Round,
				Got:      msg.Vote,
			}
		} else if err, ok := err.(*agtypes.ErrVoteConflictingVotes); ok {
			evidence = &agtypes.HypoBadVoteEvidence{
				PubKey:   peerKey,
				VoteType: msg.Vote.Data.Type,
				Got:      msg.Vote,
				Height:   cs.Height,
				Round:    cs.Round,
				Expected: err.VoteA,
			}
		}
		if evidence != nil {
			pk := crypto.PubKeyEd25519{}
			pkb, _ := hex.DecodeString(peerKey)
			copy(pk[:], pkb)
			cs.badvoteCollector.ReportBadVote(&pk, evidence)
		}
		// NOTE: the vote is broadcast to peers by the reactor listening
		// for vote events

		// TODO: If rs.Height == vote.Height && rs.Round < vote.Round,
		// the peer is sending us CatchupCommit precommits.
		// We could make note of this and help filter in broadcastHasVoteMessage().
	default:
		cs.logger.Warn("Unknown msg type", zap.Reflect("type", reflect.TypeOf(msg)))
	}
	if err != nil {
		cs.slogger.Errorw("Error with msg", "type", reflect.TypeOf(msg), "peer", peerKey, "error", err, "msg", msg.CString())
	}
}

func (cs *ConsensusState) handleTimeout(ti timeoutInfo, rs RoundState) {
	cs.slogger.Debugw("Received tock", "timeout", ti.Duration, "height", ti.Height, "round", ti.Round, "step", ti.Step)

	// timeouts must be for current height, round, step
	if ti.Height != rs.Height || ti.Round < rs.Round || (ti.Round == rs.Round && ti.Step < rs.Step) {
		cs.slogger.Debugw("Ignoring tock because we're ahead", "height", rs.Height, "round", rs.Round, "step", rs.Step)
		return
	}

	// the timeout will now cause a state transition
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	switch ti.Step {
	case csspb.RoundStepType_NewHeight:
		// NewRound event fired from enterNewRound.
		// XXX: should we fire timeout here (for timeout commit)?
		cs.enterNewRound(ti.Height, 0)
	case csspb.RoundStepType_Propose:
		agtypes.FireEventTimeoutPropose(cs.evsw, cs.RoundStateEvent())
		cs.enterPrevote(ti.Height, ti.Round)
	case csspb.RoundStepType_PrevoteWait:
		agtypes.FireEventTimeoutWait(cs.evsw, cs.RoundStateEvent())
		cs.enterPrecommit(ti.Height, ti.Round)
	case csspb.RoundStepType_PrecommitWait:
		agtypes.FireEventTimeoutWait(cs.evsw, cs.RoundStateEvent())
		cs.enterNewRound(ti.Height, ti.Round+1)
	default:
		panic(Fmt("Invalid timeout step: %v", ti.Step))
	}

}

//-----------------------------------------------------------------------------
// State functions
// Used internally by handleTimeout and handleMsg to make state transitions

// Enter: +2/3 precommits for nil at (height,round-1)
// Enter: `timeoutPrecommits` after any +2/3 precommits from (height,round-1)
// Enter: `startTime = commitTime+timeoutCommit` from NewHeight(height)
// NOTE: cs.StartTime was already set for height.
func (cs *ConsensusState) enterNewRound(height, round def.INT) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && cs.Step != csspb.RoundStepType_NewHeight) {
		cs.slogger.Debugf("enterNewRound(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)
		return
	}

	if now := time.Now(); cs.StartTime.After(now) {
		cs.logger.Warn("Need to set a buffer and cs.logger.Warn() here for sanity.", zap.Time("startTime", cs.StartTime), zap.Time("now", now))
	}
	// cs.stopTimer()

	cs.slogger.Infof("enterNewRound(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)

	// Increment validators if necessary
	validators := cs.Validators
	if cs.Round < round {
		validators = validators.Copy()
		validators.IncrementAccum(round - cs.Round)
	}

	// Setup new round
	// we don't fire newStep for this step,
	// but we fire an event, so update the round step first
	cs.updateRoundStep(round, csspb.RoundStepType_NewRound)
	cs.Validators = validators

	if round == 0 {
		// We've already reset these upon new height,
		// and meanwhile we might have received a proposal
		// for round 0.
	} else {
		cs.Proposal = nil
		cs.ProposalBlock = nil
		cs.ProposalBlockParts = nil
	}
	cs.Votes.SetRound(round + 1) // also track next round (round+1) to allow round-skipping

	rse := cs.RoundStateEvent()
	agtypes.FireEventNewRound(cs.evsw, rse)

	ed := agtypes.NewEventDataHookNewRound(rse.Height, rse.Round)
	agtypes.FireEventHookNewRound(cs.evsw, ed)
	<-ed.ResCh

	// Immediately go to enterPropose.
	cs.enterPropose(height, round)
}

// Enter: from NewRound(height,round).
func (cs *ConsensusState) enterPropose(height, round def.INT) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && csspb.RoundStepType_Propose <= cs.Step) {
		cs.slogger.Debugf("enterPropose(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)
		return
	}
	cs.slogger.Infof("enterPropose(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)

	defer func() {
		// Done enterPropose:
		cs.updateRoundStep(round, csspb.RoundStepType_Propose)
		cs.newStep()

		// If we have the whole proposal + POL, then goto Prevote now.
		// else, we'll enterPrevote when the rest of the proposal is received (in AddProposalBlockPart),
		// or else after timeoutPropose
		if cs.isProposalComplete() {
			cs.enterPrevote(height, cs.Round)
		}
	}()

	// If we don't get the proposal and all block parts quick enough, enterPrevote
	cs.scheduleTimeout(cs.timeoutParams.Propose(round), height, round, csspb.RoundStepType_Propose)

	// Nothing more to do if we're not a validator
	if cs.privValidator == nil {
		return
	}

	if !bytes.Equal(cs.Validators.Proposer().Address, cs.privValidator.GetAddress()) {
		cs.slogger.Infow("enterPropose: Not our turn to propose", "proposer", cs.Validators.Proposer().Address, "privValidator", cs.privValidator)
	} else {
		cs.slogger.Infow("enterPropose: Our turn to propose", "proposer", cs.Validators.Proposer().Address, "privValidator", cs.privValidator)
		cs.decideProposal(height, round)
	}
}

func (cs *ConsensusState) defaultDecideProposal(height, round def.INT) {
	var block *agtypes.BlockCache
	var blockParts *agtypes.PartSet

	// Decide on block
	if cs.LockedBlock != nil {
		// If we're locked onto a block, just choose that.
		block, blockParts = cs.LockedBlock, cs.LockedBlockParts
	} else {
		// Create a new proposal block from state/txs from the mempool.
		block, blockParts = cs.createProposalBlock()
		if block == nil { // on error
			return
		}
	}

	// Make proposal
	polRound, polBlockID := cs.Votes.POLInfo()
	proposal := agtypes.NewProposal(height, round, *(blockParts.Header()), polRound, polBlockID)
	err := cs.privValidator.SignProposal(cs.state.ChainID, proposal)
	if err == nil {
		// Set fields
		/*  fields set by setProposal and addBlockPart
		cs.Proposal = proposal
		cs.ProposalBlock = block
		cs.ProposalBlockParts = blockParts
		*/

		// send proposal and block parts on internal msg queue
		cs.sendInternalMessage(genMsgInfo(&csspb.ProposalMessage{proposal}, ""))
		for i := 0; i < int(blockParts.Total()); i++ {
			part := blockParts.GetPart(i)
			cs.sendInternalMessage(genMsgInfo(&csspb.BlockPartMessage{cs.Height, cs.Round, part}, ""))
		}
		cs.slogger.Infof("Signed proposal height %d round %d proposal %v", height, round, proposal.CString())
		//cs.slogger.Debugf("Signed proposal block: %v", block)
	} else {
		if !cs.replayMode {
			cs.logger.Warn("enterPropose: Error signing proposal", zap.Int64("height", height), zap.Int64("round", round), zap.String("error", err.Error()))
		}
	}
}

// Returns true if the proposal block is complete &&
// (if POLRound was proposed, we have +2/3 prevotes from there).
func (cs *ConsensusState) isProposalComplete() bool {
	if cs.Proposal == nil || cs.ProposalBlock == nil {
		return false
	}
	// we have the proposal. if there's a POLRound,
	// make sure we have the prevotes from it too
	pdata := cs.Proposal.GetData()
	if pdata.POLRound < 0 {
		return true
	} else {
		// if this is false the proposer is lying or we haven't received the POL yet
		return cs.Votes.Prevotes(pdata.POLRound).HasTwoThirdsMajority()
	}
}

// Create the next block to propose and return it.
// Returns nil block upon error.
// NOTE: keep it side-effect free for clarity.
func (cs *ConsensusState) createProposalBlock() (block *agtypes.BlockCache, blockParts *agtypes.PartSet) {
	var commit *agtypes.CommitCache
	if cs.Height == 1 {
		// We're creating a proposal for the first block.
		// The commit is empty, but not nil.
		commit = agtypes.NewCommitCache(&pbtypes.Commit{})
	} else if cs.LastCommit.HasTwoThirdsMajority() {
		// Make the commit from LastCommit
		commit = cs.LastCommit.MakeCommit()
	} else {
		// This shouldn't happen.
		cs.logger.Error("enterPropose: Cannot propose anything: No commit for the previous block.")
		return
	}

	// Mempool validated transactions
	txs := cs.mempool.Reap(cs.config.GetInt("block_size"))

	proposerPubkey, _ := cs.Validators.Proposer().GetPubKey().(*crypto.PubKeyEd25519)

	return agtypes.MakeBlock(proposerPubkey[:], cs.Height, cs.state.ChainID, txs, commit,
		cs.state.LastBlockID, cs.state.Validators.Hash(), cs.state.AppHash, cs.state.ReceiptsHash, cs.config.GetInt64("block_part_size"), cs.state.LastNonEmptyHeight)
}

// Enter: `timeoutPropose` after entering Propose.
// Enter: proposal block and POL is ready.
// Enter: any +2/3 prevotes for future round.
// Prevote for LockedBlock if we're locked, or ProposalBlock if valid.
// Otherwise vote nil.
func (cs *ConsensusState) enterPrevote(height, round def.INT) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && csspb.RoundStepType_Prevote <= cs.Step) {
		cs.slogger.Debugf("enterPrevote(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)
		return
	}

	defer func() {
		// Done enterPrevote:
		cs.updateRoundStep(round, csspb.RoundStepType_Prevote)
		cs.newStep()
	}()

	// fire event for how we got here
	if cs.isProposalComplete() {
		agtypes.FireEventCompleteProposal(cs.evsw, cs.RoundStateEvent())
	} else {
		// we received +2/3 prevotes for a future round
		// TODO: catchup event?
	}

	// cs.stopTimer()

	cs.slogger.Infof("enterPrevote(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)

	// Sign and broadcast vote as necessary
	cs.doPrevote(height, round)

	// Once `addVote` hits any +2/3 prevotes, we will go to PrevoteWait
	// (so we have more time to try and collect +2/3 prevotes for a single block)
}

func (cs *ConsensusState) defaultDoPrevote(height, round def.INT) {
	// If a block is locked, prevote that.
	var prevotedBlock *agtypes.BlockCache
	defer func() {
		agtypes.FireEventHookPrevote(cs.evsw, agtypes.EventDataHookPrevote{Height: height, Round: round, Block: prevotedBlock})
	}()

	if cs.LockedBlock != nil {
		cs.logger.Debug("enterPrevote: Block was locked")
		cs.signAddVote(pbtypes.VoteType_Prevote, cs.LockedBlock.Hash(), cs.LockedBlockParts.Header())
		prevotedBlock = cs.LockedBlock
		return
	}

	// If ProposalBlock is nil, prevote nil.
	if cs.ProposalBlock == nil {
		cs.logger.Debug("enterPrevote: ProposalBlock is nil")
		cs.signAddVote(pbtypes.VoteType_Prevote, nil, &(pbtypes.PartSetHeader{}))
		return
	}

	// Valdiate proposal block
	err := cs.state.ValidateBlock(cs.ProposalBlock)
	if err != nil {
		// ProposalBlock is invalid, prevote nil.
		cs.logger.Warn("enterPrevote: ProposalBlock is invalid", zap.String("error", err.Error()))
		cs.signAddVote(pbtypes.VoteType_Prevote, nil, &(pbtypes.PartSetHeader{}))
		return
	}

	// Prevote cs.ProposalBlock
	// NOTE: the proposal signature is validated when it is received,
	// and the proposal block parts are validated as they are received (against the merkle hash in the proposal)
	cs.signAddVote(pbtypes.VoteType_Prevote, cs.ProposalBlock.Hash(), cs.ProposalBlockParts.Header())
	prevotedBlock = cs.ProposalBlock
	return
}

// Enter: any +2/3 prevotes at next round.
func (cs *ConsensusState) enterPrevoteWait(height, round def.INT) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && csspb.RoundStepType_PrevoteWait <= cs.Step) {
		cs.slogger.Debugf("enterPrevoteWait(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)
		return
	}
	if !cs.Votes.Prevotes(round).HasTwoThirdsAny() {
		PanicSanity(Fmt("enterPrevoteWait(%v/%v), but Prevotes does not have any +2/3 votes", height, round))
	}
	cs.slogger.Infof("enterPrevoteWait(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)

	defer func() {
		// Done enterPrevoteWait:
		cs.updateRoundStep(round, csspb.RoundStepType_PrevoteWait)
		cs.newStep()
	}()

	// Wait for some more prevotes; enterPrecommit
	cs.scheduleTimeout(cs.timeoutParams.Prevote(round), height, round, csspb.RoundStepType_PrevoteWait)
}

// Enter: +2/3 precomits for block or nil.
// Enter: `timeoutPrevote` after any +2/3 prevotes.
// Enter: any +2/3 precommits for next round.
// Lock & precommit the ProposalBlock if we have enough prevotes for it (a POL in this round)
// else, unlock an existing lock and precommit nil if +2/3 of prevotes were nil,
// else, precommit nil otherwise.
func (cs *ConsensusState) enterPrecommit(height, round def.INT) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && csspb.RoundStepType_Precommit <= cs.Step) {
		cs.slogger.Debugf("enterPrecommit(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)
		return
	}
	var precommitedBlock *agtypes.BlockCache
	// cs.stopTimer()

	cs.slogger.Infof("enterPrecommit(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)

	defer func() {
		// Done enterPrecommit:
		cs.updateRoundStep(round, csspb.RoundStepType_Precommit)
		cs.newStep()
		agtypes.FireEventHookPrecommit(cs.evsw, agtypes.EventDataHookPrecommit{Height: height, Round: round, Block: precommitedBlock})
	}()

	blockID, ok := cs.Votes.Prevotes(round).TwoThirdsMajority()

	// If we don't have a polka, we must precommit nil
	if !ok {
		if cs.LockedBlock != nil {
			cs.logger.Warn("enterPrecommit: No +2/3 prevotes during enterPrecommit while we're locked. Precommitting nil")
		} else {
			cs.logger.Warn("enterPrecommit: No +2/3 prevotes during enterPrecommit. Precommitting nil.")
		}
		cs.signAddVote(pbtypes.VoteType_Precommit, nil, &(pbtypes.PartSetHeader{}))
		return
	}

	// At this point +2/3 prevoted for a particular block or nil
	agtypes.FireEventPolka(cs.evsw, cs.RoundStateEvent())

	// the latest POLRound should be this round
	polRound, _ := cs.Votes.POLInfo()
	if polRound < round {
		PanicSanity(Fmt("This POLRound should be %v but got %", round, polRound))
	}

	// +2/3 prevoted nil. Unlock and precommit nil.
	if len(blockID.Hash) == 0 {
		if cs.LockedBlock == nil {
			cs.logger.Debug("enterPrecommit: +2/3 prevoted for nil.")
		} else {
			cs.logger.Debug("enterPrecommit: +2/3 prevoted for nil. Unlocking")
			cs.LockedRound = 0
			cs.LockedBlock = nil
			cs.LockedBlockParts = nil
			agtypes.FireEventUnlock(cs.evsw, cs.RoundStateEvent())
		}
		cs.signAddVote(pbtypes.VoteType_Precommit, nil, &(pbtypes.PartSetHeader{}))
		return
	}

	// if cs.ProposalBlock.HashesTo(blockID.Hash) && cs.ProposalBlock.NumTxs == 0 && round < emptyRoundLimit {
	//	cs.logger.Debug("enterPrecommit: proposal block is empty, vote nil, no locking")
	//	cs.signAddVote(types.VoteTypePrecommit, nil, types.PartSetHeader{})
	//	return
	// }

	// if cs.LockedBlock.HashesTo(blockID.Hash) && cs.LockedBlock.NumTxs == 0 && round < emptyRoundLimit {
	//	cs.logger.Debug("enterPrecommit: locked block is empty, vote nil, unlocking")
	//	cs.signAddVote(types.VoteTypePrecommit, nil, types.PartSetHeader{})
	//	cs.LockedBlock = nil
	//	cs.LockedRound = 0
	//	cs.LockedBlockParts = nil
	//	return
	// }

	// At this point, +2/3 prevoted for a particular block.

	// If we're already locked on that block, precommit it, and update the LockedRound
	if cs.LockedBlock.HashesTo(blockID.Hash) {
		cs.logger.Debug("enterPrecommit: +2/3 prevoted locked block. Relocking")
		cs.LockedRound = round
		agtypes.FireEventRelock(cs.evsw, cs.RoundStateEvent())
		cs.signAddVote(pbtypes.VoteType_Precommit, blockID.Hash, blockID.PartsHeader)
		precommitedBlock = cs.LockedBlock
		return
	}

	// If +2/3 prevoted for proposal block, stage and precommit it
	if cs.ProposalBlock.HashesTo(blockID.Hash) {
		cs.slogger.Debugf("enterPrecommit: +2/3 prevoted proposal block. Locking hash %X", blockID.Hash)
		// Validate the block.
		if err := cs.state.ValidateBlock(cs.ProposalBlock); err != nil {
			PanicConsensus(Fmt("enterPrecommit: +2/3 prevoted for an invalid block: %v", err))
		}
		cs.LockedRound = round
		cs.LockedBlock = cs.ProposalBlock
		cs.LockedBlockParts = cs.ProposalBlockParts
		agtypes.FireEventLock(cs.evsw, cs.RoundStateEvent())
		cs.signAddVote(pbtypes.VoteType_Precommit, blockID.Hash, blockID.PartsHeader)
		precommitedBlock = cs.ProposalBlock
		return
	}

	// There was a polka in this round for a block we don't have.
	// Fetch that block, unlock, and precommit nil.
	// The +2/3 prevotes for this round is the POL for our unlock.
	// TODO: In the future save the POL prevotes for justification.
	cs.LockedRound = 0
	cs.LockedBlock = nil
	cs.LockedBlockParts = nil
	if !cs.ProposalBlockParts.HasHeader(blockID.PartsHeader) {
		cs.ProposalBlock = nil
		cs.ProposalBlockParts = agtypes.NewPartSetFromHeader(blockID.PartsHeader)
	}
	agtypes.FireEventUnlock(cs.evsw, cs.RoundStateEvent())
	cs.signAddVote(pbtypes.VoteType_Precommit, nil, &(pbtypes.PartSetHeader{}))
	return
}

// Enter: any +2/3 precommits for next round.
func (cs *ConsensusState) enterPrecommitWait(height, round def.INT) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && csspb.RoundStepType_PrecommitWait <= cs.Step) {
		cs.slogger.Debugf("enterPrecommitWait(%v/%v): Invalid args. Current step: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)
		return
	}
	if !cs.Votes.Precommits(round).HasTwoThirdsAny() {
		PanicSanity(Fmt("enterPrecommitWait(%v/%v), but Precommits does not have any +2/3 votes", height, round))
	}
	cs.slogger.Infof("enterPrecommitWait(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step)

	defer func() {
		// Done enterPrecommitWait:
		cs.updateRoundStep(round, csspb.RoundStepType_PrecommitWait)
		cs.newStep()
	}()

	// Wait for some more precommits; enterNewRound
	cs.scheduleTimeout(cs.timeoutParams.Precommit(round), height, round, csspb.RoundStepType_PrecommitWait)

}

// Enter: +2/3 precommits for block
func (cs *ConsensusState) enterCommit(height, commitRound def.INT) {
	if cs.Height != height || csspb.RoundStepType_Commit <= cs.Step {
		cs.slogger.Debugf("enterCommit(%v/%v): Invalid args. Current step: %v/%v/%v", height, commitRound, cs.Height, cs.Round, cs.Step)
		return
	}
	cs.slogger.Infof("enterCommit(%v/%v). Current: %v/%v/%v", height, commitRound, cs.Height, cs.Round, cs.Step)

	defer func() {
		// Done enterCommit:
		// keep cs.Round the same, commitRound points to the right Precommits set.
		cs.updateRoundStep(cs.Round, csspb.RoundStepType_Commit)
		cs.CommitRound = commitRound
		cs.CommitTime = time.Now()
		cs.newStep()

		// Maybe finalize immediately.
		cs.tryFinalizeCommit(height)
	}()

	blockID, ok := cs.Votes.Precommits(commitRound).TwoThirdsMajority()
	if !ok {
		PanicSanity("RunActionCommit() expects +2/3 precommits")
	}

	// The Locked* fields no longer matter.
	// Move them over to ProposalBlock if they match the commit hash,
	// otherwise they'll be cleared in updateToState.
	if cs.LockedBlock.HashesTo(blockID.Hash) {
		cs.ProposalBlock = cs.LockedBlock
		cs.ProposalBlockParts = cs.LockedBlockParts
	}

	// If we don't have the block being committed, set up to get it.
	if !cs.ProposalBlock.HashesTo(blockID.Hash) {
		if !cs.ProposalBlockParts.HasHeader(blockID.PartsHeader) {
			// We're getting the wrong block.
			// Set up ProposalBlockParts and keep waiting.
			cs.ProposalBlock = nil
			cs.ProposalBlockParts = agtypes.NewPartSetFromHeader(blockID.PartsHeader)
		} else {
			// We just need to keep waiting.
		}
	}
}

// If we have the block AND +2/3 commits for it, finalize.
func (cs *ConsensusState) tryFinalizeCommit(height def.INT) {
	if cs.Height != height {
		PanicSanity(Fmt("tryFinalizeCommit() cs.Height: %v vs height: %v", cs.Height, height))
	}

	blockID, ok := cs.Votes.Precommits(cs.CommitRound).TwoThirdsMajority()
	if !ok || len(blockID.Hash) == 0 {
		cs.logger.Warn("Attempt to finalize failed. There was no +2/3 majority, or +2/3 was for <nil>.")
		return
	}
	if !cs.ProposalBlock.HashesTo(blockID.Hash) {
		// TODO: this happens every time if we're not a validator (ugly logs)
		// TODO: ^^ wait, why does it matter that we're a validator?
		cs.logger.Warn("Attempt to finalize failed. We don't have the commit block.", zap.String("blockID", blockID.CString()))
		return
	}
	//	go
	cs.finalizeCommit(height)
}

// Increment height and goto RoundStepNewHeight
func (cs *ConsensusState) finalizeCommit(height def.INT) {
	if cs.Height != height || cs.Step != csspb.RoundStepType_Commit {
		cs.slogger.Debugf("finalizeCommit(%v): Invalid args. Current step: %v/%v/%v", height, cs.Height, cs.Round, cs.Step)
		return
	}
	// logger.Info("ann-stopwatch consensusTime elapsed ", cs.RoundState.CommitTime.Sub(cs.RoundState.StartTime).String())

	blockID, ok := cs.Votes.Precommits(cs.CommitRound).TwoThirdsMajority()
	block, blockParts := cs.ProposalBlock, cs.ProposalBlockParts

	if !ok {
		PanicSanity(Fmt("Cannot finalizeCommit, commit does not have two thirds majority"))
	}
	if !blockParts.HasHeader(blockID.PartsHeader) {
		PanicSanity(Fmt("Expected ProposalBlockParts header to be commit header"))
	}
	if !block.HashesTo(blockID.Hash) {
		PanicSanity(Fmt("Cannot finalizeCommit, ProposalBlock does not hash to commit hash"))
	}
	if err := cs.state.ValidateBlock(block); err != nil {
		PanicConsensus(Fmt("+2/3 committed an invalid block: %v", err))
	}

	bheader := block.GetHeader()
	cs.slogger.Infof("Finalizing commit of block with %d txs, height %d, hash %X, root %x", bheader.GetNumTxs(), bheader.GetHeight(), block.Hash(), bheader.GetAppHash())
	//cs.slogger.Debugf("%v", block)

	// Save to blockStore.
	if cs.blockStore.Height() < bheader.Height {
		// NOTE: the seenCommit is local justification to commit this block,
		// but may differ from the LastCommit included in the next block
		precommits := cs.Votes.Precommits(cs.CommitRound)
		seenCommit := precommits.MakeCommit()
		cs.blockStore.SaveBlock(block, blockParts, seenCommit)
	} else {
		cs.logger.Warn("Why are we finalizeCommitting a block height we already have?", zap.Int64("height", bheader.Height))
	}

	// Create a copy of the state for staging
	// and an event cache for txs
	stateCopy := cs.state.Copy()
	// eventCache := types.NewEventCache(cs.evsw)
	// Execute and commit the block, and update the mempool.
	// NOTE: the block.AppHash wont reflect these txs until the next block
	err := stateCopy.ApplyBlock(cs.evsw, block, blockParts.Header(), cs.mempool, cs.Round)
	if err != nil {
		// TODO!
	}

	// Fire off event for new block.
	// TODO: Handle app failure.  See #177
	agtypes.FireEventNewBlock(cs.evsw, agtypes.EventDataNewBlock{block})
	agtypes.FireEventNewBlockHeader(cs.evsw, agtypes.EventDataNewBlockHeader{block.Header})
	// eventCache.Flush()
	// Save the state.
	stateCopy.Save()
	// NewHeightStep!
	cs.updateToState(stateCopy)
	// cs.StartTime is already set.
	// Schedule Round0 to start soon.
	cs.scheduleRound0(&cs.RoundState)

	// By here,
	// * cs.Height has been increment to height+1
	// * cs.Step is now RoundStepNewHeight
	// * cs.StartTime is set to when we will start round0.
	return
}

//-----------------------------------------------------------------------------

func (cs *ConsensusState) defaultSetProposal(proposal *pbtypes.Proposal) error {
	// Already have one
	// TODO: possibly catch double proposals
	if cs.Proposal != nil {
		return nil
	}

	pdata := proposal.GetData()
	// Does not apply
	if pdata.Height != cs.Height || pdata.Round != cs.Round {
		return nil
	}

	// We don't care about the proposal if we're already in RoundStepCommit.
	if csspb.RoundStepType_Commit <= cs.Step {
		return nil
	}

	// Verify POLRound, which must be -1 or between 0 and proposal.Round exclusive.
	if pdata.POLRound != -1 &&
		(pdata.POLRound < 0 || pdata.Round <= pdata.POLRound) {
		return ErrInvalidProposalPOLRound
	}

	// Verify signature
	if !cs.Validators.Proposer().PubKey.VerifyBytes(agtypes.SignBytes(cs.state.ChainID, pdata), agtypes.NewDefaultSignature(proposal.Signature)) {
		return ErrInvalidProposalSignature
	}

	cs.Proposal = proposal
	cs.ProposalBlockParts = agtypes.NewPartSetFromHeader(pdata.BlockPartsHeader)
	return nil
}

// NOTE: block is not necessarily valid.
// Asynchronously triggers either enterPrevote (before we timeout of propose) or tryFinalizeCommit, once we have the full block.
func (cs *ConsensusState) addProposalBlockPart(height def.INT, part *pbtypes.Part, verify bool) (added bool, err error) {
	// Blocks might be reused, so round mismatch is OK
	if cs.Height != height {
		return false, nil
	}

	// We're not expecting a block part.
	if cs.ProposalBlockParts == nil {
		return false, nil // TODO: bad peer? Return error?
	}

	added, err = cs.ProposalBlockParts.AddPart(part, verify)
	if err != nil {
		return added, err
	}
	if added && cs.ProposalBlockParts.IsComplete() {
		// Added and completed!
		var err error
		cs.ProposalBlock = cs.ProposalBlockParts.AssembleToBlock(cs.config.GetInt64("block_part_size"))
		// NOTE: it's possible to receive complete proposal blocks for future rounds without having the proposal
		cs.logger.Debug("Received complete proposal block", zap.Int64("height", cs.ProposalBlock.Header.Height), zap.String("hash", Fmt("%X", cs.ProposalBlock.Hash())))
		if cs.Step == csspb.RoundStepType_Propose && cs.isProposalComplete() {
			// Move onto the next step
			cs.enterPrevote(height, cs.Round)
		} else if cs.Step == csspb.RoundStepType_Commit {
			// If we're waiting on the proposal block...
			cs.tryFinalizeCommit(height)
		}
		return true, err
	}
	return added, nil
}

// Attempt to add the vote. if its a duplicate signature, dupeout the validator
func (cs *ConsensusState) tryAddVote(vote *pbtypes.Vote, peerKey string) error {
	_, err := cs.addVote(vote, peerKey)
	if err != nil {
		// If the vote height is off, we'll just ignore it,
		// But if it's a conflicting sig, broadcast evidence tx for slashing.
		// If it's otherwise invalid, punish peer.
		if err == ErrVoteHeightMismatch {
			return err
		} else if _, ok := err.(*agtypes.ErrVoteConflictingVotes); ok {
			vdata := vote.GetData()
			if peerKey == "" {
				cs.logger.Warn("Found conflicting vote from ourselves. Did you unsafe_reset a validator?", zap.Int64("height", vdata.Height), zap.Int64("round", vdata.Round), zap.String("type", vdata.Type.String()))
				return err
			}
			cs.logger.Warn("Found conflicting vote. Publish evidence (TODO)")
			/* TODO
			evidenceTx := &types.DupeoutTx{
				Address: address,
				VoteA:   *errDupe.VoteA,
				VoteB:   *errDupe.VoteB,
			}
			cs.mempool.BroadcastTx(struct{???}{evidenceTx}) // shouldn't need to check returned err
			*/
			return err
		} else {
			// Probably an invalid signature. Bad peer.
			cs.logger.Warn("Error attempting to add vote", zap.String("error", err.Error()))
			return ErrAddingVote
		}
	}
	return nil
}

//-----------------------------------------------------------------------------

func (cs *ConsensusState) addVote(vote *pbtypes.Vote, peerKey string) (added bool, err error) {
	//cs.logger.Debug("addVote", zap.Int("voteHeight", vote.Height), zap.Binary("voteType", []byte{vote.Type}), zap.Int("csHeight", cs.Height))

	// A precommit for the previous height?
	// These come in while we wait timeoutCommit
	vdata := vote.Data
	if vdata.Height+1 == cs.Height {
		if !(cs.Step == csspb.RoundStepType_NewHeight && vdata.Type == pbtypes.VoteType_Precommit) {
			// TODO: give the reason ..
			// fmt.Errorf("tryAddVote: Wrong height, not a LastCommit straggler commit.")
			return added, ErrVoteHeightMismatch
		}
		added, err = cs.LastCommit.AddVote(vote)
		if added {
			cs.logger.Debug("Added to lastPrecommits: " + cs.LastCommit.StringShort())
			agtypes.FireEventVote(cs.evsw, agtypes.EventDataVote{vote})

			// if we can skip timeoutCommit and have all the votes now,
			if cs.timeoutParams.SkipTimeoutCommit && cs.LastCommit.HasAll() {
				// go straight to new round (skip timeout commit)
				// cs.scheduleTimeout(time.Duration(0), cs.Height, 0, RoundStepNewHeight)
				cs.enterNewRound(cs.Height, 0)
			}
		}

		return
	}

	// A prevote/precommit for this height?
	if vdata.Height == cs.Height {
		height := cs.Height
		added, err = cs.Votes.AddVote(vote, peerKey)
		if added {
			agtypes.FireEventVote(cs.evsw, agtypes.EventDataVote{vote})

			switch vdata.Type {
			case pbtypes.VoteType_Prevote:
				prevotes := cs.Votes.Prevotes(vdata.Round)
				cs.slogger.Debugw("Added to prevote", "vote", vote.CString(), "prevotes", prevotes.StringShort())
				// First, unlock if prevotes is a valid POL.
				// >> lockRound < POLRound <= unlockOrChangeLockRound (see spec)
				// NOTE: If (lockRound < POLRound) but !(POLRound <= unlockOrChangeLockRound),
				// we'll still enterNewRound(H,vote.R) and enterPrecommit(H,vote.R) to process it
				// there.
				if (cs.LockedBlock != nil) && (cs.LockedRound < vdata.Round) && (vdata.Round <= cs.Round) {
					blockID, ok := prevotes.TwoThirdsMajority()
					if ok && !cs.LockedBlock.HashesTo(blockID.Hash) {
						cs.logger.Info("Unlocking because of POL.", zap.Int64("lockedRound", cs.LockedRound), zap.Int64("POLRound", vdata.Round))
						cs.LockedRound = 0
						cs.LockedBlock = nil
						cs.LockedBlockParts = nil
						agtypes.FireEventUnlock(cs.evsw, cs.RoundStateEvent())
					}
				}
				pdata := cs.Proposal.GetData()
				if cs.Round <= vdata.Round && prevotes.HasTwoThirdsAny() {
					// Round-skip over to PrevoteWait or goto Precommit.
					cs.enterNewRound(height, vdata.Round) // if the vote is ahead of us
					if prevotes.HasTwoThirdsMajority() {
						cs.enterPrecommit(height, vdata.Round)
					} else {
						cs.enterPrevote(height, vdata.Round) // if the vote is ahead of us
						cs.enterPrevoteWait(height, vdata.Round)
					}
				} else if cs.Proposal != nil && 0 <= pdata.POLRound && pdata.POLRound == vdata.Round {
					// If the proposal is now complete, enter prevote of cs.Round.
					if cs.isProposalComplete() {
						cs.enterPrevote(height, cs.Round)
					}
				}
			case pbtypes.VoteType_Precommit:
				precommits := cs.Votes.Precommits(vdata.Round)
				//cs.slogger.Debugw("Added to precommit", "vote", vote, "precommits", precommits.StringShort())
				blockID, ok := precommits.TwoThirdsMajority()
				if ok {
					if len(blockID.Hash) == 0 {
						cs.enterNewRound(height, vdata.Round+1)
					} else {
						cs.enterNewRound(height, vdata.Round)
						cs.enterPrecommit(height, vdata.Round)
						cs.enterCommit(height, vdata.Round)

						if cs.timeoutParams.SkipTimeoutCommit && precommits.HasAll() {
							// if we have all the votes now,
							// go straight to new round (skip timeout commit)
							// cs.scheduleTimeout(time.Duration(0), cs.Height, 0, RoundStepNewHeight)
							cs.enterNewRound(cs.Height, 0)
						}

					}
				} else if cs.Round <= vdata.Round && precommits.HasTwoThirdsAny() {
					cs.enterNewRound(height, vdata.Round)
					cs.enterPrecommit(height, vdata.Round)
					cs.enterPrecommitWait(height, vdata.Round)
				}
			default:
				PanicSanity(Fmt("Unexpected vote type %X", vdata.Type)) // Should not happen.
			}
		}
		// Either duplicate, or error upon cs.Votes.AddByIndex()
		return
	} else {
		err = ErrVoteHeightMismatch
	}

	// Height mismatch, bad peer?
	cs.logger.Debug("Vote ignored and not added", zap.Int64("voteHeight", vdata.Height), zap.Int64("csHeight", cs.Height), zap.String("err", err.Error()))
	return
}

func (cs *ConsensusState) signVote(type_ pbtypes.VoteType, hash []byte, header *pbtypes.PartSetHeader) (*pbtypes.Vote, error) {
	addr := cs.privValidator.GetAddress()
	valIndex, _ := cs.Validators.GetByAddress(addr)
	vote := &pbtypes.Vote{
		Data: &pbtypes.VoteData{
			ValidatorAddress: addr,
			ValidatorIndex:   def.INT(valIndex),
			Height:           cs.Height,
			Round:            cs.Round,
			Type:             type_,
			BlockID:          &pbtypes.BlockID{hash, header},
		},
	}
	err := cs.privValidator.SignVote(cs.state.ChainID, vote)
	return vote, err
}

// sign the vote and publish on internalMsgQueue
func (cs *ConsensusState) signAddVote(type_ pbtypes.VoteType, hash []byte, header *pbtypes.PartSetHeader) *pbtypes.Vote {
	// if we don't have a key or we're not in the validator set, do nothing
	if cs.privValidator == nil || !cs.Validators.HasAddress(cs.privValidator.GetAddress()) {
		return nil
	}
	vote, err := cs.signVote(type_, hash, header)
	if err == nil {
		cs.sendInternalMessage(genMsgInfo(&csspb.VoteMessage{vote}, ""))
		cs.slogger.Debugf("Signed and pushed vote height %d, round %d, vote %X, error %v", cs.Height, cs.Round, vote.GetData().GetBlockID().CString(), err)
		return vote
	} else {
		if !cs.replayMode {
			cs.slogger.Warnf("Error signing vote height %d, round %d, vote %v, error %v", cs.Height, cs.Round, vote, err)
		}
		return nil
	}
}

//---------------------------------------------------------

func CompareHRS(h1, r1 def.INT, s1 csspb.RoundStepType, h2, r2 def.INT, s2 csspb.RoundStepType) int {
	if h1 < h2 {
		return -1
	} else if h1 > h2 {
		return 1
	}
	if r1 < r2 {
		return -1
	} else if r1 > r2 {
		return 1
	}
	if s1 < s2 {
		return -1
	} else if s1 > s2 {
		return 1
	}
	return 0
}
