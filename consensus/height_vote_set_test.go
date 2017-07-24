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
	"testing"

	. "github.com/tendermint/go-common"
	"github.com/tendermint/tendermint/config/tendermint_test"
	"github.com/tendermint/tendermint/types"
)

func init() {
	config = tendermint_test.ResetConfig("consensus_height_vote_set_test")
}

func TestPeerCatchupRounds(t *testing.T) {
	valSet, privVals := types.RandValidatorSet(10, 1)

	hvs := NewHeightVoteSet(config.GetString("chain_id"), 1, valSet)

	vote999_0 := makeVoteHR(t, 1, 999, privVals, 0)
	added, err := hvs.AddVote(vote999_0, "peer1")
	if !added || err != nil {
		t.Error("Expected to successfully add vote from peer", added, err)
	}

	vote1000_0 := makeVoteHR(t, 1, 1000, privVals, 0)
	added, err = hvs.AddVote(vote1000_0, "peer1")
	if !added || err != nil {
		t.Error("Expected to successfully add vote from peer", added, err)
	}

	vote1001_0 := makeVoteHR(t, 1, 1001, privVals, 0)
	added, err = hvs.AddVote(vote1001_0, "peer1")
	if added {
		t.Error("Expected to *not* add vote from peer, too many catchup rounds.")
	}

	added, err = hvs.AddVote(vote1001_0, "peer2")
	if !added || err != nil {
		t.Error("Expected to successfully add vote from another peer")
	}

}

func makeVoteHR(t *testing.T, height, round int, privVals []*types.PrivValidator, valIndex int) *types.Vote {
	privVal := privVals[valIndex]
	vote := &types.Vote{
		ValidatorAddress: privVal.Address,
		ValidatorIndex:   valIndex,
		Height:           height,
		Round:            round,
		Type:             types.VoteTypePrecommit,
		BlockID:          types.BlockID{[]byte("fakehash"), types.PartSetHeader{}},
	}
	chainID := config.GetString("chain_id")
	err := privVal.SignVote(chainID, vote)
	if err != nil {
		panic(Fmt("Error signing vote: %v", err))
		return nil
	}
	return vote
}
