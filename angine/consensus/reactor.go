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
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"go.uber.org/zap"

	csspb "github.com/dappledger/AnnChain/angine/protos/consensus"
	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	sm "github.com/dappledger/AnnChain/angine/state"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-p2p"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

const (
	StateChannel       = byte(0x20)
	DataChannel        = byte(0x21)
	VoteChannel        = byte(0x22)
	VoteSetBitsChannel = byte(0x23)

	peerGossipSleepDuration     = 100 * time.Millisecond // Time to sleep if there's nothing to send.
	peerQueryMaj23SleepDuration = 2 * time.Second        // Time to sleep after each VoteSetMaj23Message sent
	maxConsensusMessageSize     = 1048576                // 1MB; NOTE: keep in sync with types.PartSet sizes.
)

//-----------------------------------------------------------------------------

type ConsensusReactor struct {
	p2p.BaseReactor // BaseService + p2p.Switch

	conS     *ConsensusState
	fastSync bool
	evsw     agtypes.EventSwitch

	logger  *zap.Logger
	slogger *zap.SugaredLogger
}

func NewConsensusReactor(logger *zap.Logger, consensusState *ConsensusState, fastSync bool) *ConsensusReactor {
	conR := &ConsensusReactor{
		conS:     consensusState,
		fastSync: fastSync,

		logger:  logger,
		slogger: logger.Sugar(),
	}
	conR.BaseReactor = *p2p.NewBaseReactor(logger, "ConsensusReactor", conR)
	return conR
}

func (conR *ConsensusReactor) OnStart() error {
	conR.logger.Info("ConsensusReactor ", zap.Bool("fastSync", conR.fastSync))
	conR.BaseReactor.OnStart()

	// callbacks for broadcasting new steps and votes to peers
	// upon their respective events (ie. uses evsw)
	conR.registerEventCallbacks()

	if !conR.fastSync {
		_, err := conR.conS.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (conR *ConsensusReactor) OnStop() {
	conR.BaseReactor.OnStop()
	conR.conS.Stop()
}

// Switch from the fast_sync to the consensus:
// reset the state, turn off fast_sync, start the consensus-state-machine
func (conR *ConsensusReactor) SwitchToConsensus(state *sm.State) {
	cs := conR.conS

	// Reset fields based on state.
	validators := state.Validators
	height := state.LastBlockHeight + 1 // Next desired block height

	// RoundState fields
	cs.mtx.Lock()
	cs.updateHeight(height)
	cs.updateRoundStep(0, csspb.RoundStepType_NewHeight)
	cs.StartTime = cs.timeoutParams.Commit(time.Now())
	cs.Validators = validators
	cs.Proposal = nil
	cs.ProposalBlock = nil
	cs.ProposalBlockParts = nil
	cs.LockedRound = 0
	cs.LockedBlock = nil
	cs.LockedBlockParts = nil
	cs.Votes = NewHeightVoteSet(cs.config.GetString("chain_id"), height, validators)
	cs.CommitRound = -1
	cs.reconstructLastCommit(state)
	cs.LastValidators = state.LastValidators
	cs.state = state
	cs.mtx.Unlock()
	// Finally, broadcast RoundState
	cs.newStep()

	conR.fastSync = false
	cs.Start()
}

// Implements Reactor
func (conR *ConsensusReactor) GetChannels() []*p2p.ChannelDescriptor {
	// TODO optimize
	return []*p2p.ChannelDescriptor{
		&p2p.ChannelDescriptor{
			ID:                StateChannel,
			Priority:          5,
			SendQueueCapacity: 100,
		},
		&p2p.ChannelDescriptor{
			ID:                 DataChannel, // maybe split between gossiping current block and catchup stuff
			Priority:           10,          // once we gossip the whole block there's nothing left to send until next height or round
			SendQueueCapacity:  100,
			RecvBufferCapacity: 50 * 4096,
		},
		&p2p.ChannelDescriptor{
			ID:                 VoteChannel,
			Priority:           5,
			SendQueueCapacity:  100,
			RecvBufferCapacity: 100 * 100,
		},
		&p2p.ChannelDescriptor{
			ID:                 VoteSetBitsChannel,
			Priority:           1,
			SendQueueCapacity:  2,
			RecvBufferCapacity: 1024,
		},
	}
}

// Implements Reactor
func (conR *ConsensusReactor) AddPeer(peer *p2p.Peer) {
	if !conR.IsRunning() {
		return
	}

	// Create peerState for peer
	peerState := NewPeerState(conR.slogger, peer)
	peer.Data.Set(agtypes.PeerStateKey, peerState)

	// Begin routines for this peer.
	go conR.gossipDataRoutine(peer, peerState)
	go conR.gossipVotesRoutine(peer, peerState)
	go conR.queryMaj23Routine(peer, peerState)

	// Send our state to peer.
	// If we're fast_syncing, broadcast a RoundStepMessage later upon SwitchToConsensus().
	if !conR.fastSync {
		conR.sendNewRoundStepMessage(peer)
	}
}

// Implements Reactor
func (conR *ConsensusReactor) RemovePeer(peer *p2p.Peer, reason interface{}) {
	if !conR.IsRunning() {
		return
	}
	// TODO
	//peer.Data.Get(PeerStateKey).(*PeerState).Disconnect()
}

// Implements Reactor
// NOTE: We process these messages even when we're fast_syncing.
// Messages affect either a peer state or the consensus state.
// Peer state updates can happen in parallel, but processing of
// proposals, block parts, and votes are ordered by the receiveRoutine
// NOTE: blocks on consensus state for proposals, block parts, and votes
func (conR *ConsensusReactor) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
	if !conR.IsRunning() {
		conR.slogger.Debugw("Receive", "src", src, "chId", chID, "bytes", msgBytes)
		return
	}

	msg, err := csspb.UnmarshalCssMsg(msgBytes)
	if err != nil {
		conR.slogger.Warnw("Error decoding message", "src", src, "chId", chID, "msg", msg, "error", err, "bytes", msgBytes)
		// TODO punish peer?
		return
	}

	// Get peer states
	ps := src.Data.Get(agtypes.PeerStateKey).(*PeerState)

	switch chID {
	case StateChannel:
		switch msg := msg.(type) {
		case *csspb.NewRoundStepMessage:
			ps.ApplyNewRoundStepMessage(msg)
		case *csspb.CommitStepMessage:
			ps.ApplyCommitStepMessage(msg)
		case *csspb.HasVoteMessage:
			ps.ApplyHasVoteMessage(msg)
		case *csspb.VoteSetMaj23Message:
			cs := conR.conS
			cs.mtx.Lock()
			height, votes := cs.Height, cs.Votes
			cs.mtx.Unlock()
			if height != msg.Height {
				return
			}
			// Peer claims to have a maj23 for some BlockID at H,R,S,
			votes.SetPeerMaj23(msg.Round, msg.Type, ps.Peer.Key, msg.BlockID)
			// Respond with a VoteSetBitsMessage showing which votes we have.
			// (and consequently shows which we don't have)
			var ourVotes *BitArray
			switch msg.Type {
			case pbtypes.VoteType_Prevote:
				ourVotes = votes.Prevotes(msg.Round).BitArrayByBlockID(msg.BlockID)
			case pbtypes.VoteType_Precommit:
				ourVotes = votes.Precommits(msg.Round).BitArrayByBlockID(msg.BlockID)
			default:
				conR.logger.Warn("Bad VoteSetBitsMessage field Type")
				return
			}
			src.TrySendBytes(VoteSetBitsChannel, csspb.MarshalDataToCssMsg(&csspb.VoteSetBitsMessage{
				Height:  msg.Height,
				Round:   msg.Round,
				Type:    msg.Type,
				BlockID: msg.BlockID,
				Votes:   csspb.TransferBitArray(ourVotes),
			}))

		default:
			conR.slogger.Warnf("Unknown message type %T", reflect.TypeOf(msg))
		}

	case DataChannel:
		if conR.fastSync {
			conR.slogger.Warnw("Ignoring message received during fastSync", "msg", msg)
			return
		}
		switch msg := msg.(type) {
		case *csspb.ProposalMessage:
			ps.SetHasProposal(msg.Proposal)
			conR.conS.peerMsgQueue <- genMsgInfo(msg, src.Key)
		case *csspb.ProposalPOLMessage:
			ps.ApplyProposalPOLMessage(msg)
		case *csspb.BlockPartMessage:
			ps.SetHasProposalBlockPart(msg.Height, msg.Round, int(msg.Part.Index))
			conR.conS.peerMsgQueue <- genMsgInfo(msg, src.Key)
		default:
			conR.slogger.Warnf("Unknown message type %T", msg)
		}

	case VoteChannel:
		if conR.fastSync {
			conR.slogger.Warnw("Ignoring message received during fastSync", "msg", msg)
			return
		}
		switch msg := msg.(type) {
		case *csspb.VoteMessage:
			cs := conR.conS
			cs.mtx.Lock()
			height, valSize, lastCommitSize := cs.Height, cs.Validators.Size(), cs.LastCommit.Size()
			cs.mtx.Unlock()
			ps.EnsureVoteBitArrays(height, valSize)
			ps.EnsureVoteBitArrays(height-1, lastCommitSize)
			ps.SetHasVote(msg.Vote)

			conR.conS.peerMsgQueue <- genMsgInfo(msg, src.Key)

		default:
			// don't punish (leave room for soft upgrades)
			conR.slogger.Warnf("Unknown message type %T", msg)
		}

	case VoteSetBitsChannel:
		if conR.fastSync {
			conR.slogger.Warnw("Ignoring message received during fastSync", "msg", msg)
			return
		}
		switch msg := msg.(type) {
		case *csspb.VoteSetBitsMessage:
			cs := conR.conS
			cs.mtx.Lock()
			height, votes := cs.Height, cs.Votes
			cs.mtx.Unlock()

			if height == msg.Height {
				var ourVotes *BitArray
				switch msg.Type {
				case pbtypes.VoteType_Prevote:
					ourVotes = votes.Prevotes(msg.Round).BitArrayByBlockID(msg.BlockID)
				case pbtypes.VoteType_Precommit:
					ourVotes = votes.Precommits(msg.Round).BitArrayByBlockID(msg.BlockID)
				default:
					conR.logger.Warn("Bad VoteSetBitsMessage field Type")
					return
				}
				ps.ApplyVoteSetBitsMessage(msg, ourVotes)
			} else {
				ps.ApplyVoteSetBitsMessage(msg, nil)
			}
		default:
			// don't punish (leave room for soft upgrades)
			conR.slogger.Warnf("Unknown message type %T", msg)
		}

	default:
		conR.slogger.Warnf("Unknown chId %X", chID)
	}

	if err != nil {
		conR.logger.Warn("Error in Receive()", zap.String("error", err.Error()))
	}
}

// implements events.Eventable
func (conR *ConsensusReactor) SetEventSwitch(evsw agtypes.EventSwitch) {
	conR.evsw = evsw
	conR.conS.SetEventSwitch(evsw)
}

//--------------------------------------

// Listens for new steps and votes,
// broadcasting the result to peers
func (conR *ConsensusReactor) registerEventCallbacks() {

	agtypes.AddListenerForEvent(conR.evsw, "conR", agtypes.EventStringNewRoundStep(), func(data agtypes.TMEventData) {
		rs := data.(agtypes.EventDataRoundState).RoundState.(*RoundState)
		conR.broadcastNewRoundStep(rs)
	})

	agtypes.AddListenerForEvent(conR.evsw, "conR", agtypes.EventStringVote(), func(data agtypes.TMEventData) {
		edv := data.(agtypes.EventDataVote)
		conR.broadcastHasVoteMessage(edv.Vote)
	})

	agtypes.AddListenerForEvent(conR.evsw, "conR", agtypes.EventStringSwitchToConsensus(), func(data agtypes.TMEventData) {
		conR.SwitchToConsensus(conR.conS.state)
	})
}

func (conR *ConsensusReactor) broadcastNewRoundStep(rs *RoundState) {

	nrsMsg, csMsg := makeRoundStepMessages(rs)
	if nrsMsg != nil {
		conR.Switch.BroadcastBytes(StateChannel, csspb.MarshalDataToCssMsg(nrsMsg))
	}
	if csMsg != nil {
		conR.Switch.BroadcastBytes(StateChannel, csspb.MarshalDataToCssMsg(csMsg))
	}
}

// Broadcasts HasVoteMessage to peers that care.
func (conR *ConsensusReactor) broadcastHasVoteMessage(vote *pbtypes.Vote) {
	vdata := vote.GetData()
	msg := &csspb.HasVoteMessage{
		Height: vdata.Height,
		Round:  vdata.Round,
		Type:   vdata.Type,
		Index:  vdata.ValidatorIndex,
	}
	conR.Switch.BroadcastBytes(StateChannel, csspb.MarshalDataToCssMsg(msg))
	/*
		// TODO: Make this broadcast more selective.
		for _, peer := range conR.Switch.Peers().List() {
			ps := peer.Data.Get(PeerStateKey).(*PeerState)
			prs := ps.GetRoundState()
			if prs.Height == vote.Height {
				// TODO: Also filter on round?
				peer.TrySend(StateChannel, struct{ ConsensusMessage }{msg})
			} else {
				// Height doesn't match
				// TODO: check a field, maybe CatchupCommitRound?
				// TODO: But that requires changing the struct field comment.
			}
		}
	*/
}

func makeRoundStepMessages(rs *RoundState) (nrsMsg *csspb.NewRoundStepMessage, csMsg *csspb.CommitStepMessage) {
	nrsMsg = &csspb.NewRoundStepMessage{
		Height: rs.Height,
		Round:  rs.Round,
		Step:   rs.Step,
		SecondsSinceStartTime: def.INT(time.Now().Sub(rs.StartTime).Seconds()),
		LastCommitRound:       rs.LastCommit.Round(),
	}
	if rs.Step == csspb.RoundStepType_Commit {
		header := rs.ProposalBlockParts.Header()
		csMsg = &csspb.CommitStepMessage{
			Height:           rs.Height,
			BlockPartsHeader: header,
			BlockParts:       csspb.TransferBitArray(rs.ProposalBlockParts.BitArray()),
		}
	}
	return
}

func (conR *ConsensusReactor) sendNewRoundStepMessage(peer *p2p.Peer) {
	rs := conR.conS.GetRoundState()
	nrsMsg, csMsg := makeRoundStepMessages(rs)
	if nrsMsg != nil {
		peer.SendBytes(StateChannel, csspb.MarshalDataToCssMsg(nrsMsg))
	}
	if csMsg != nil {
		peer.SendBytes(StateChannel, csspb.MarshalDataToCssMsg(csMsg))
	}
}

func (conR *ConsensusReactor) gossipDataRoutine(peer *p2p.Peer, ps *PeerState) {
OUTER_LOOP:
	for {
		// Manage disconnects from self or peer.
		if !peer.IsRunning() || !conR.IsRunning() {
			conR.slogger.Infof("Stopping gossipDataRoutine for %v.", peer)
			return
		}
		rs := conR.conS.GetRoundState()
		prs := ps.GetRoundState()

		// Send proposal Block parts?
		if rs.ProposalBlockParts.HasHeader(&prs.ProposalBlockPartsHeader) {
			if index, ok := rs.ProposalBlockParts.BitArray().Sub(prs.ProposalBlockParts.Copy()).PickRandom(); ok {
				part := rs.ProposalBlockParts.GetPart(index)
				msg := &csspb.BlockPartMessage{
					Height: rs.Height, // This tells peer that this part applies to us.
					Round:  rs.Round,  // This tells peer that this part applies to us.
					Part:   part,
				}
				peer.SendBytes(DataChannel, csspb.MarshalDataToCssMsg(msg))
				ps.SetHasProposalBlockPart(prs.Height, prs.Round, index)
				continue OUTER_LOOP
			}
		}

		// If the peer is on a previous height, help catch up.
		if (0 < prs.Height) && (prs.Height < rs.Height) {
			if index, ok := prs.ProposalBlockParts.Not().PickRandom(); ok {
				// Ensure that the peer's PartSetHeader is correct
				blockMeta := conR.conS.blockStore.LoadBlockMeta(prs.Height)
				if blockMeta == nil {
					conR.slogger.Warnw("Failed to load block meta", "peer height", prs.Height, "our height", rs.Height, "blockstore height", conR.conS.blockStore.Height(), "pv", conR.conS.privValidator)
					time.Sleep(peerGossipSleepDuration)
					continue OUTER_LOOP
				} else if !blockMeta.PartsHeader.Equals(&prs.ProposalBlockPartsHeader) {
					conR.slogger.Debugw("Peer ProposalBlockPartsHeader mismatch, sleeping",
						"peerHeight", prs.Height, "blockPartsHeader", blockMeta.PartsHeader, "peerBlockPartsHeader", prs.ProposalBlockPartsHeader)
					time.Sleep(peerGossipSleepDuration)
					continue OUTER_LOOP
				}
				// Load the part
				part := conR.conS.blockStore.LoadBlockPart(prs.Height, index)
				if part == nil {
					conR.slogger.Warnw("Could not load part", "index", index,
						"peerHeight", prs.Height, "blockPartsHeader", blockMeta.PartsHeader, "peerBlockPartsHeader", prs.ProposalBlockPartsHeader)
					time.Sleep(peerGossipSleepDuration)
					continue OUTER_LOOP
				}
				// Send the part
				msg := &csspb.BlockPartMessage{
					Height: prs.Height, // Not our height, so it doesn't matter.
					Round:  prs.Round,  // Not our height, so it doesn't matter.
					Part:   part,
				}
				peer.SendBytes(DataChannel, csspb.MarshalDataToCssMsg(msg))
				ps.SetHasProposalBlockPart(prs.Height, prs.Round, index)
				continue OUTER_LOOP
			} else {
				time.Sleep(peerGossipSleepDuration)
				continue OUTER_LOOP
			}
		}

		// If height and round don't match, sleep.
		if (rs.Height != prs.Height) || (rs.Round != prs.Round) {
			time.Sleep(peerGossipSleepDuration)
			continue OUTER_LOOP
		}

		// By here, height and round match.
		// Proposal block parts were already matched and sent if any were wanted.
		// (These can match on hash so the round doesn't matter)
		// Now consider sending other things, like the Proposal itself.

		// Send Proposal && ProposalPOL BitArray?
		if rs.Proposal != nil && !prs.Proposal {
			// Proposal: share the proposal metadata with peer.
			{
				msg := &csspb.ProposalMessage{Proposal: rs.Proposal}
				peer.SendBytes(DataChannel, csspb.MarshalDataToCssMsg(msg))

				ps.SetHasProposal(rs.Proposal)
			}
			// ProposalPOL: lets peer know which POL votes we have so far.
			// Peer must receive ProposalMessage first.
			// rs.Proposal was validated, so rs.Proposal.POLRound <= rs.Round,
			// so we definitely have rs.Votes.Prevotes(rs.Proposal.POLRound).
			pPOLRound := rs.Proposal.GetData().GetPOLRound()
			if 0 <= pPOLRound {
				msg := &csspb.ProposalPOLMessage{
					Height:           rs.Height,
					ProposalPOLRound: pPOLRound,
					ProposalPOL:      csspb.TransferBitArray(rs.Votes.Prevotes(pPOLRound).BitArray()),
				}
				peer.SendBytes(DataChannel, csspb.MarshalDataToCssMsg(msg))
			}
			continue OUTER_LOOP
		}

		// Nothing to do. Sleep.
		time.Sleep(peerGossipSleepDuration)
		continue OUTER_LOOP
	}
}

func (conR *ConsensusReactor) gossipVotesRoutine(peer *p2p.Peer, ps *PeerState) {
OUTER_LOOP:
	for {
		if !peer.IsRunning() || !conR.IsRunning() {
			conR.slogger.Infof("Stopping gossipVotesRoutine for %v.", peer)
			return
		}

		rs := conR.conS.GetRoundState()
		if conR.gossipVotes(rs, ps) {
			continue OUTER_LOOP
		}
		time.Sleep(peerGossipSleepDuration)

		continue OUTER_LOOP
	}
}

func (conR *ConsensusReactor) gossipVotesToPeers(rs *RoundState) {
	peers := conR.Switch.Peers()
	for _, peer := range peers.List() {
		psData := peer.Data.Get(agtypes.PeerStateKey)
		ps := psData.(*PeerState)
		conR.gossipVotes(rs, ps)
	}
}

func (conR *ConsensusReactor) gossipVotes(rs *RoundState, ps *PeerState) bool {
	prs := ps.GetRoundState()
	// If height matches, then send LastCommit, Prevotes, Precommits.
	if rs.Height == prs.Height {
		// If there are lastCommits to send...
		if prs.Step == csspb.RoundStepType_NewHeight {
			if ps.PickSendVote(rs.LastCommit) {
				return true
			}
		}
		// If there are prevotes to send...
		if prs.Step <= csspb.RoundStepType_Prevote && prs.Round != -1 && prs.Round <= rs.Round {
			if ps.PickSendVote(rs.Votes.Prevotes(prs.Round)) {
				return true
			}
		}
		// If there are precommits to send...
		if prs.Step <= csspb.RoundStepType_Precommit && prs.Round != -1 && prs.Round <= rs.Round {
			if ps.PickSendVote(rs.Votes.Precommits(prs.Round)) {
				return true
			}
		}
		// If there are POLPrevotes to send...
		if prs.ProposalPOLRound != -1 {
			if polPrevotes := rs.Votes.Prevotes(prs.ProposalPOLRound); polPrevotes != nil {
				if ps.PickSendVote(polPrevotes) {
					return true
				}
			}
		}
	}

	// Special catchup logic.
	// If peer is lagging by height 1, send LastCommit.
	if prs.Height != 0 && rs.Height == prs.Height+1 {
		if ps.PickSendVote(rs.LastCommit) {
			return true
		}
	}

	// Catchup logic
	// If peer is lagging by more than 1, send Commit.
	if prs.Height != 0 && rs.Height >= prs.Height+2 {
		// Load the block commit for prs.Height,
		// which contains precommit signatures for prs.Height.
		commit := conR.conS.blockStore.LoadBlockCommit(prs.Height)
		conR.slogger.Infow("Loaded BlockCommit for catch-up", "height", prs.Height, "commit", commit)
		if ps.PickSendVote(commit) {
			return true
		}
	}

	return false
}

// NOTE: `queryMaj23Routine` has a simple crude design since it only comes
// into play for liveness when there's a signature DDoS attack happening.
func (conR *ConsensusReactor) queryMaj23Routine(peer *p2p.Peer, ps *PeerState) {
OUTER_LOOP:
	for {
		// Manage disconnects from self or peer.
		if !peer.IsRunning() || !conR.IsRunning() {
			conR.slogger.Infof("Stopping queryMaj23Routine for %v.", peer)
			return
		}

		// Maybe send Height/Round/Prevotes
		{
			rs := conR.conS.GetRoundState()
			prs := ps.GetRoundState()
			if rs.Height == prs.Height {
				if maj23, ok := rs.Votes.Prevotes(prs.Round).TwoThirdsMajority(); ok {
					peer.TrySendBytes(StateChannel, csspb.MarshalDataToCssMsg(&csspb.VoteSetMaj23Message{
						Height:  prs.Height,
						Round:   prs.Round,
						Type:    pbtypes.VoteType_Prevote,
						BlockID: &maj23,
					}))
					time.Sleep(peerQueryMaj23SleepDuration)
				}
			}
		}

		// Maybe send Height/Round/Precommits
		{
			rs := conR.conS.GetRoundState()
			prs := ps.GetRoundState()
			if rs.Height == prs.Height {
				if maj23, ok := rs.Votes.Precommits(prs.Round).TwoThirdsMajority(); ok {
					peer.TrySendBytes(StateChannel, csspb.MarshalDataToCssMsg(&csspb.VoteSetMaj23Message{
						Height:  prs.Height,
						Round:   prs.Round,
						Type:    pbtypes.VoteType_Precommit,
						BlockID: &maj23,
					}))
					time.Sleep(peerQueryMaj23SleepDuration)
				}
			}
		}

		// Maybe send Height/Round/ProposalPOL
		{
			rs := conR.conS.GetRoundState()
			prs := ps.GetRoundState()
			if rs.Height == prs.Height && prs.ProposalPOLRound >= 0 {
				if maj23, ok := rs.Votes.Prevotes(prs.ProposalPOLRound).TwoThirdsMajority(); ok {
					peer.TrySendBytes(StateChannel, csspb.MarshalDataToCssMsg(&csspb.VoteSetMaj23Message{
						Height:  prs.Height,
						Round:   prs.ProposalPOLRound,
						Type:    pbtypes.VoteType_Prevote,
						BlockID: &maj23,
					}))
					time.Sleep(peerQueryMaj23SleepDuration)
				}
			}
		}

		// Little point sending LastCommitRound/LastCommit,
		// These are fleeting and non-blocking.

		// Maybe send Height/CatchupCommitRound/CatchupCommit.
		{
			prs := ps.GetRoundState()
			if prs.CatchupCommitRound != -1 && 0 < prs.Height && prs.Height <= conR.conS.blockStore.Height() {
				commit := conR.conS.LoadCommit(prs.Height)
				peer.TrySendBytes(StateChannel, csspb.MarshalDataToCssMsg(&csspb.VoteSetMaj23Message{
					Height:  prs.Height,
					Round:   commit.Round(),
					Type:    pbtypes.VoteType_Precommit,
					BlockID: commit.BlockID,
				}))
				time.Sleep(peerQueryMaj23SleepDuration)
			}
		}

		time.Sleep(peerQueryMaj23SleepDuration)

		continue OUTER_LOOP
	}
}

func (conR *ConsensusReactor) String() string {
	// better not to access shared variables
	return "ConsensusReactor" // conR.StringIndented("")
}

func (conR *ConsensusReactor) StringIndented(indent string) string {
	s := "ConsensusReactor{\n"
	s += indent + "  " + conR.conS.StringIndented(indent+"  ") + "\n"
	for _, peer := range conR.Switch.Peers().List() {
		ps := peer.Data.Get(agtypes.PeerStateKey).(*PeerState)
		s += indent + "  " + ps.StringIndented(indent+"  ") + "\n"
	}
	s += indent + "}"
	return s
}

//-----------------------------------------------------------------------------

// Read only when returned by PeerState.GetRoundState().
type PeerRoundState struct {
	Height                   def.INT               // Height peer is at
	Round                    def.INT               // Round peer is at, -1 if unknown.
	Step                     csspb.RoundStepType   // Step peer is at
	StartTime                time.Time             // Estimated start of round 0 at this height
	Proposal                 bool                  // True if peer has proposal for this round
	ProposalBlockPartsHeader pbtypes.PartSetHeader //
	ProposalBlockParts       *BitArray             //
	ProposalPOLRound         def.INT               // Proposal's POL round. -1 if none.
	ProposalPOL              *BitArray             // nil until ProposalPOLMessage received.
	Prevotes                 *BitArray             // All votes peer has for this round
	Precommits               *BitArray             // All precommits peer has for this round
	LastCommitRound          def.INT               // Round of commit for last height. -1 if none.
	LastCommit               *BitArray             // All commit precommits of commit for last height.
	CatchupCommitRound       def.INT               // Round that we have commit for. Not necessarily unique. -1 if none.
	CatchupCommit            *BitArray             // All commit precommits peer has for this height & CatchupCommitRound
}

func (prs PeerRoundState) String() string {
	return prs.StringIndented("")
}

func (prs PeerRoundState) StringIndented(indent string) string {
	return fmt.Sprintf(`PeerRoundState{
%s  %v/%v/%v @%v
%s  Proposal %v -> %v
%s  POL      %v (round %v)
%s  Prevotes   %v
%s  Precommits %v
%s  LastCommit %v (round %v)
%s  Catchup    %v (round %v)
%s}`,
		indent, prs.Height, prs.Round, prs.Step, prs.StartTime,
		indent, prs.ProposalBlockPartsHeader, prs.ProposalBlockParts,
		indent, prs.ProposalPOL, prs.ProposalPOLRound,
		indent, prs.Prevotes,
		indent, prs.Precommits,
		indent, prs.LastCommit, prs.LastCommitRound,
		indent, prs.CatchupCommit, prs.CatchupCommitRound,
		indent)
}

//-----------------------------------------------------------------------------

var (
	ErrPeerStateHeightRegression = errors.New("Error peer state height regression")
	ErrPeerStateInvalidStartTime = errors.New("Error peer state invalid startTime")
)

type PeerState struct {
	Peer *p2p.Peer

	mtx sync.Mutex
	PeerRoundState

	slogger *zap.SugaredLogger
}

func NewPeerState(slogger *zap.SugaredLogger, peer *p2p.Peer) *PeerState {
	return &PeerState{
		Peer: peer,
		PeerRoundState: PeerRoundState{
			Round:              -1,
			ProposalPOLRound:   -1,
			LastCommitRound:    -1,
			CatchupCommitRound: -1,
		},
		slogger: slogger,
	}
}

// Returns an atomic snapshot of the PeerRoundState.
// There's no point in mutating it since it won't change PeerState.
func (ps *PeerState) GetRoundState() *PeerRoundState {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	prs := ps.PeerRoundState // copy
	return &prs
}

// Returns an atomic snapshot of the PeerRoundState's height
// used by the mempool to ensure peers are caught up before broadcasting new txs
func (ps *PeerState) GetHeight() def.INT {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.PeerRoundState.Height
}

func (ps *PeerState) SetHasProposal(proposal *pbtypes.Proposal) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	pdata := proposal.GetData()
	if ps.Height != pdata.Height || ps.Round != pdata.Round {
		return
	}
	if ps.Proposal {
		return
	}

	ps.Proposal = true
	ps.ProposalBlockPartsHeader = *(pdata.BlockPartsHeader)
	ps.ProposalBlockParts = NewBitArray(int(pdata.BlockPartsHeader.Total))
	ps.ProposalPOLRound = pdata.POLRound
	ps.ProposalPOL = nil // Nil until ProposalPOLMessage received.
}

func (ps *PeerState) SetHasProposalBlockPart(height, round def.INT, index int) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.Height != height || ps.Round != round {
		return
	}

	ps.ProposalBlockParts.SetIndex(index, true)
}

// Convenience function to send vote to peer.
// Returns true if vote was sent.
func (ps *PeerState) PickSendVote(votes agtypes.VoteSetReader) (ok bool) {
	if vote, ok := ps.PickVoteToSend(votes); ok {
		msg := &csspb.VoteMessage{vote}
		ps.Peer.SendBytes(VoteChannel, csspb.MarshalDataToCssMsg(msg))
		return true
	}
	return false
}

// votes: Must be the correct Size() for the Height().
func (ps *PeerState) PickVoteToSend(votes agtypes.VoteSetReader) (vote *pbtypes.Vote, ok bool) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if votes.Size() == 0 {
		return nil, false
	}

	height, round, type_, size := votes.Height(), votes.Round(), votes.Type(), votes.Size()

	// Lazily set data using 'votes'.
	if votes.IsCommit() {
		ps.ensureCatchupCommitRound(height, round, size)
	}
	ps.ensureVoteBitArrays(height, size)

	psVotes := ps.getVoteBitArray(height, round, type_)
	if psVotes == nil {
		return nil, false // Not something worth sending
	}
	if index, ok := votes.BitArray().Sub(psVotes).PickRandom(); ok {
		ps.setHasVote(height, round, type_, index)
		return votes.GetByIndex(index), true
	}
	return nil, false
}

func (ps *PeerState) getVoteBitArray(height, round def.INT, type_ pbtypes.VoteType) *BitArray {
	if !pbtypes.IsVoteTypeValid(type_) {
		PanicSanity("Invalid vote type")
	}

	if ps.Height == height {
		if ps.Round == round {
			switch type_ {
			case pbtypes.VoteType_Prevote:
				return ps.Prevotes
			case pbtypes.VoteType_Precommit:
				return ps.Precommits
			}
		}
		if ps.CatchupCommitRound == round {
			switch type_ {
			case pbtypes.VoteType_Prevote:
				return nil
			case pbtypes.VoteType_Precommit:
				return ps.CatchupCommit
			}
		}
		if ps.ProposalPOLRound == round {
			switch type_ {
			case pbtypes.VoteType_Prevote:
				return ps.ProposalPOL
			case pbtypes.VoteType_Precommit:
				return nil
			}
		}
		return nil
	}
	if ps.Height == height+1 {
		if ps.LastCommitRound == round {
			switch type_ {
			case pbtypes.VoteType_Prevote:
				return nil
			case pbtypes.VoteType_Precommit:
				return ps.LastCommit
			}
		}
		return nil
	}
	return nil
}

// 'round': A round for which we have a +2/3 commit.
func (ps *PeerState) ensureCatchupCommitRound(height, round def.INT, numValidators int) {
	if ps.Height != height {
		return
	}
	/*
		NOTE: This is wrong, 'round' could change.
		e.g. if orig round is not the same as block LastCommit round.
		if ps.CatchupCommitRound != -1 && ps.CatchupCommitRound != round {
			PanicSanity(Fmt("Conflicting CatchupCommitRound. Height: %v, Orig: %v, New: %v", height, ps.CatchupCommitRound, round))
		}
	*/
	if ps.CatchupCommitRound == round {
		return // Nothing to do!
	}
	ps.CatchupCommitRound = round
	if round == ps.Round {
		ps.CatchupCommit = ps.Precommits
	} else {
		ps.CatchupCommit = NewBitArray(numValidators)
	}
}

// NOTE: It's important to make sure that numValidators actually matches
// what the node sees as the number of validators for height.
func (ps *PeerState) EnsureVoteBitArrays(height def.INT, numValidators int) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	ps.ensureVoteBitArrays(height, numValidators)
}

func (ps *PeerState) ensureVoteBitArrays(height def.INT, numValidators int) {
	if ps.Height == height {
		if ps.Prevotes == nil {
			ps.Prevotes = NewBitArray(numValidators)
		}
		if ps.Precommits == nil {
			ps.Precommits = NewBitArray(numValidators)
		}
		if ps.CatchupCommit == nil {
			ps.CatchupCommit = NewBitArray(numValidators)
		}
		if ps.ProposalPOL == nil {
			ps.ProposalPOL = NewBitArray(numValidators)
		}
	} else if ps.Height == height+1 {
		if ps.LastCommit == nil {
			ps.LastCommit = NewBitArray(numValidators)
		}
	}
}

func (ps *PeerState) SetHasVote(vote *pbtypes.Vote) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	vdata := vote.GetData()
	ps.setHasVote(vdata.Height, vdata.Round, vdata.Type, int(vdata.ValidatorIndex))
}

func (ps *PeerState) setHasVote(height, round def.INT, type_ pbtypes.VoteType, index int) {
	//ps.slogger.Debugw("setHasVote(LastCommit)", "lastCommit", ps.LastCommit, "index", index)

	// NOTE: some may be nil BitArrays -> no side effects.
	ps.getVoteBitArray(height, round, type_).SetIndex(index, true)
}

func (ps *PeerState) ApplyNewRoundStepMessage(msg *csspb.NewRoundStepMessage) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	// Ignore duplicates or decreases
	if CompareHRS(msg.Height, msg.Round, msg.Step, ps.Height, ps.Round, ps.Step) <= 0 {
		return
	}

	// Just remember these values.
	psHeight := ps.Height
	psRound := ps.Round
	//psStep := ps.Step
	psCatchupCommitRound := ps.CatchupCommitRound
	psCatchupCommit := ps.CatchupCommit

	startTime := time.Now().Add(-1 * time.Duration(msg.SecondsSinceStartTime) * time.Second)
	ps.Height = msg.Height
	ps.Round = msg.Round
	ps.Step = msg.Step
	ps.StartTime = startTime
	if psHeight != msg.Height || psRound != msg.Round {
		ps.Proposal = false
		ps.ProposalBlockPartsHeader = pbtypes.PartSetHeader{}
		ps.ProposalBlockParts = nil
		ps.ProposalPOLRound = -1
		ps.ProposalPOL = nil
		// We'll update the BitArray capacity later.
		ps.Prevotes = nil
		ps.Precommits = nil
	}
	if psHeight == msg.Height && psRound != msg.Round && msg.Round == psCatchupCommitRound {
		// Peer caught up to CatchupCommitRound.
		// Preserve psCatchupCommit!
		// NOTE: We prefer to use prs.Precommits if
		// pr.Round matches pr.CatchupCommitRound.
		ps.Precommits = psCatchupCommit
	}
	if psHeight != msg.Height {
		// Shift Precommits to LastCommit.
		if psHeight+1 == msg.Height && psRound == msg.LastCommitRound {
			ps.LastCommitRound = msg.LastCommitRound
			ps.LastCommit = ps.Precommits
		} else {
			ps.LastCommitRound = msg.LastCommitRound
			ps.LastCommit = nil
		}
		// We'll update the BitArray capacity later.
		ps.CatchupCommitRound = -1
		ps.CatchupCommit = nil
	}
}

func (ps *PeerState) ApplyCommitStepMessage(msg *csspb.CommitStepMessage) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.Height != msg.Height {
		return
	}

	ps.ProposalBlockPartsHeader = *(msg.GetBlockPartsHeader())
	ps.ProposalBlockParts = csspb.TransferProtoBitArray(msg.GetBlockParts())
}

func (ps *PeerState) ApplyProposalPOLMessage(msg *csspb.ProposalPOLMessage) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.Height != msg.Height {
		return
	}
	if ps.ProposalPOLRound != msg.ProposalPOLRound {
		return
	}

	// TODO: Merge onto existing ps.ProposalPOL?
	// We might have sent some prevotes in the meantime.
	ps.ProposalPOL = csspb.TransferProtoBitArray(msg.ProposalPOL)
}

func (ps *PeerState) ApplyHasVoteMessage(msg *csspb.HasVoteMessage) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.Height != msg.Height {
		return
	}

	ps.setHasVote(msg.Height, msg.Round, msg.Type, int(msg.Index))
}

// The peer has responded with a bitarray of votes that it has
// of the corresponding BlockID.
// ourVotes: BitArray of votes we have for msg.BlockID
// NOTE: if ourVotes is nil (e.g. msg.Height < rs.Height),
// we conservatively overwrite ps's votes w/ msg.Votes.
func (ps *PeerState) ApplyVoteSetBitsMessage(msg *csspb.VoteSetBitsMessage, ourVotes *BitArray) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	votes := ps.getVoteBitArray(msg.Height, msg.Round, msg.Type)
	if votes != nil {
		msgVotes := csspb.TransferProtoBitArray(msg.Votes)
		if ourVotes == nil {
			votes.Update(msgVotes)
		} else {
			otherVotes := votes.Sub(ourVotes)
			hasVotes := otherVotes.Or(msgVotes)
			votes.Update(hasVotes)
		}
	}
}

func (ps *PeerState) String() string {
	return ps.StringIndented("")
}

func (ps *PeerState) StringIndented(indent string) string {
	return fmt.Sprintf(`PeerState{
%s  Key %v
%s  PRS %v
%s}`,
		indent, ps.Peer.Key,
		indent, ps.PeerRoundState.StringIndented(indent+"  "),
		indent)
}

//-------------------------------------
