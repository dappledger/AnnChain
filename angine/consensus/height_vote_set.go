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
	"strings"
	"sync"

	"github.com/pkg/errors"
	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	agtypes "github.com/dappledger/AnnChain/angine/types"

	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

type RoundVoteSet struct {
	Prevotes   *agtypes.VoteSet
	Precommits *agtypes.VoteSet
}

/*
Keeps track of all VoteSets from round 0 to round 'round'.

Also keeps track of up to one RoundVoteSet greater than
'round' from each peer, to facilitate catchup syncing of commits.

A commit is +2/3 precommits for a block at a round,
but which round is not known in advance, so when a peer
provides a precommit for a round greater than mtx.round,
we create a new entry in roundVoteSets but also remember the
peer to prevent abuse.
We let each peer provide us with up to 2 unexpected "catchup" rounds.
One for their LastCommit round, and another for the official commit round.
*/
type HeightVoteSet struct {
	chainID string
	height  def.INT
	valSet  *agtypes.ValidatorSet

	mtx               sync.Mutex
	round             def.INT                  // max tracked round
	roundVoteSets     map[def.INT]RoundVoteSet // keys: [0...round]
	peerCatchupRounds map[string][]def.INT     // keys: peer.Key; values: at most 2 rounds
}

func NewHeightVoteSet(chainID string, height def.INT, valSet *agtypes.ValidatorSet) *HeightVoteSet {
	hvs := &HeightVoteSet{
		chainID: chainID,
	}
	hvs.Reset(height, valSet)
	return hvs
}

func (hvs *HeightVoteSet) Reset(height def.INT, valSet *agtypes.ValidatorSet) {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()

	hvs.height = height
	hvs.valSet = valSet
	hvs.roundVoteSets = make(map[def.INT]RoundVoteSet)
	hvs.peerCatchupRounds = make(map[string][]def.INT)

	hvs.addRound(0)
	hvs.round = 0
}

func (hvs *HeightVoteSet) Height() def.INT {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	return hvs.height
}

func (hvs *HeightVoteSet) Round() def.INT {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	return hvs.round
}

// Create more RoundVoteSets up to round.
func (hvs *HeightVoteSet) SetRound(round def.INT) {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	if hvs.round != 0 && (round < hvs.round+1) {
		PanicSanity("SetRound() must increment hvs.round")
	}
	for r := hvs.round + 1; r <= round; r++ {
		if _, ok := hvs.roundVoteSets[r]; ok {
			continue // Already exists because peerCatchupRounds.
		}
		hvs.addRound(r)
	}
	hvs.round = round
}

func (hvs *HeightVoteSet) addRound(round def.INT) {
	if _, ok := hvs.roundVoteSets[round]; ok {
		PanicSanity("addRound() for an existing round")
	}
	prevotes := agtypes.NewVoteSet(hvs.chainID, hvs.height, round, pbtypes.VoteType_Prevote, hvs.valSet)
	precommits := agtypes.NewVoteSet(hvs.chainID, hvs.height, round, pbtypes.VoteType_Precommit, hvs.valSet)
	hvs.roundVoteSets[round] = RoundVoteSet{
		Prevotes:   prevotes,
		Precommits: precommits,
	}
}

// Duplicate votes return added=false, err=nil.
// By convention, peerKey is "" if origin is self.
func (hvs *HeightVoteSet) AddVote(vote *pbtypes.Vote, peerKey string) (added bool, err error) {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	vdata := vote.GetData()
	if !pbtypes.IsVoteTypeValid(vdata.Type) {
		return
	}
	voteSet := hvs.getVoteSet(vdata.Round, vdata.Type)
	if voteSet == nil {
		if rndz := hvs.peerCatchupRounds[peerKey]; len(rndz) < 2 {
			hvs.addRound(vdata.Round)
			voteSet = hvs.getVoteSet(vdata.Round, vdata.Type)
			hvs.peerCatchupRounds[peerKey] = append(rndz, vdata.Round)
		} else {
			// Peer has sent a vote that does not match our round,
			// for more than one round.  Bad peer!
			// TODO punish peer.
			// log.Warn("Deal with peer giving votes from unwanted rounds")
			return false, errors.Errorf("peer giving votes from unwanted rounds")
		}
	}
	added, err = voteSet.AddVote(vote)
	return
}

func (hvs *HeightVoteSet) Prevotes(round def.INT) *agtypes.VoteSet {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	return hvs.getVoteSet(round, pbtypes.VoteType_Prevote)
}

func (hvs *HeightVoteSet) Precommits(round def.INT) *agtypes.VoteSet {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	return hvs.getVoteSet(round, pbtypes.VoteType_Precommit)
}

// Last round and blockID that has +2/3 prevotes for a particular block or nil.
// Returns -1 if no such round exists.
func (hvs *HeightVoteSet) POLInfo() (polRound def.INT, polBlockID pbtypes.BlockID) {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	for r := hvs.round; r >= 0; r-- {
		rvs := hvs.getVoteSet(r, pbtypes.VoteType_Prevote)
		polBlockID, ok := rvs.TwoThirdsMajority()
		if ok {
			return r, polBlockID
		}
	}
	return -1, pbtypes.BlockID{}
}

func (hvs *HeightVoteSet) getVoteSet(round def.INT, type_ pbtypes.VoteType) *agtypes.VoteSet {
	rvs, ok := hvs.roundVoteSets[round]
	if !ok {
		return nil
	}
	switch type_ {
	case pbtypes.VoteType_Prevote:
		return rvs.Prevotes
	case pbtypes.VoteType_Precommit:
		return rvs.Precommits
	default:
		PanicSanity(Fmt("Unexpected vote type %X", type_))
		return nil
	}
}

func (hvs *HeightVoteSet) String() string {
	return hvs.StringIndented("")
}

func (hvs *HeightVoteSet) StringIndented(indent string) string {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	vsStrings := make([]string, 0, (len(hvs.roundVoteSets)+1)*2)
	// rounds 0 ~ hvs.round inclusive
	for round := def.INT(0); round <= hvs.round; round++ {
		voteSetString := hvs.roundVoteSets[round].Prevotes.StringShort()
		vsStrings = append(vsStrings, voteSetString)
		voteSetString = hvs.roundVoteSets[round].Precommits.StringShort()
		vsStrings = append(vsStrings, voteSetString)
	}
	// all other peer catchup rounds
	for round, roundVoteSet := range hvs.roundVoteSets {
		if round <= hvs.round {
			continue
		}
		voteSetString := roundVoteSet.Prevotes.StringShort()
		vsStrings = append(vsStrings, voteSetString)
		voteSetString = roundVoteSet.Precommits.StringShort()
		vsStrings = append(vsStrings, voteSetString)
	}
	return Fmt(`HeightVoteSet{H:%v R:0~%v
%s  %v
%s}`,
		hvs.height, hvs.round,
		indent, strings.Join(vsStrings, "\n"+indent+"  "),
		indent)
}

// If a peer claims that it has 2/3 majority for given blockKey, call this.
// NOTE: if there are too many peers, or too much peer churn,
// this can cause memory issues.
// TODO: implement ability to remove peers too
func (hvs *HeightVoteSet) SetPeerMaj23(round def.INT, type_ pbtypes.VoteType, peerID string, blockID *pbtypes.BlockID) {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	if !pbtypes.IsVoteTypeValid(type_) {
		return
	}
	voteSet := hvs.getVoteSet(round, type_)
	if voteSet == nil {
		return
	}
	voteSet.SetPeerMaj23(peerID, blockID)
}
