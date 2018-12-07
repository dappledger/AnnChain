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

package angine

import (
	"encoding/json"
	"fmt"

	"github.com/dappledger/AnnChain/angine/plugin"
	"github.com/dappledger/AnnChain/angine/types"
)

func (e *Angine) ProcessSpecialOP(tx []byte) error {
	if !types.IsSpecialOP(tx) {
		return fmt.Errorf("tx is not a specialop: %v", tx)
	}

	var cmd types.SpecialOPCmd
	err := json.Unmarshal(types.UnwrapTx(tx), &cmd)
	if err != nil {
		return err
	}

	cmd.ExCmd = tx
	cmd.NodePubKey = e.privValidator.PubKey.Bytes()

	sigbytes, err := e.CheckSpecialOp(&cmd)
	if err != nil {
		return err
	}

	cmd.NodeSig = sigbytes

	cmdbyte, _ := json.Marshal(cmd)
	sptx := append([]byte("zaop"), cmdbyte...)
	return e.BroadcastTx(sptx)
}

func (e *Angine) CheckSpecialOp(cmd *types.SpecialOPCmd) ([]byte, error) {
	switch cmd.CmdType {
	case types.SpecialOP_ChangeValidator,
		types.SpecialOP_Disconnect,
		types.SpecialOP_AddRefuseKey,
		types.SpecialOP_DeleteRefuseKey:
		var spPlug *plugin.Specialop

		for _, p := range e.stateMachine.Plugins {
			if ps, ok := p.(*plugin.Specialop); ok {
				spPlug = ps
				break
			}
		}
		if spPlug != nil {
			err, sig := spPlug.CheckSpecialOP(cmd)
			if err == nil {
				return sig.Bytes(), nil
			}
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unimplemented: %s", cmd.CmdType)
	}
	return nil, nil
}

// type voteResult struct {
// 	Result    []byte
// 	Validator *types.Validator
// }

// // CollectSpecialVotes returns nil means the vote passed
// func (e *Angine) CollectSpecialVotes(cmd *types.SpecialOPCmd, tx []byte) error {
// 	var votedAny, major23VotingPower int64
// 	totalVotingPower := e.consensus.GetTotalVotingPower()
// 	_, validators := e.GetValidators()
// 	votes := make(chan *voteResult, len(validators))
// 	defer close(votes)
// 	pubkey := e.PrivValidator().PubKey
// 	for _, validator := range validators {
// 		if !validator.PubKey.Equals(pubkey) {
// 			go func(data []byte, v *types.Validator, votes chan *voteResult) {
// 				if e.getSpecialVote == nil {
// 					votes <- nil
// 					e.logger.Warn("incomplete specialop support: getSpecialVote is nil")
// 					return
// 				}
// 				if res, err := e.getSpecialVote(data, v); err != nil {
// 					e.logger.Info("get special vote error", zap.Error(err))
// 					votes <- nil
// 				} else {
// 					votes <- &voteResult{
// 						Result:    res,
// 						Validator: v,
// 					}
// 				}
// 			}(tx, validator, votes)
// 		} else {
// 			votedAny += validator.VotingPower
// 			major23VotingPower += validator.VotingPower
// 		}
// 	}
// COLLECT:
// 	for {
// 		select {
// 		case res := <-votes:
// 			if res != nil {
// 				votedAny += res.Validator.VotingPower
// 				if err := CheckSpecialOPVoteSig(cmd, res.Validator.PubKey, res.Result); err != nil {
// 					e.logger.Info("check speci vote signature error", zap.Error(err))
// 				} else {
// 					major23VotingPower += res.Validator.VotingPower
// 					cmd.Sigs = append(cmd.Sigs, append(res.Validator.PubKey.Bytes(), res.Result...))
// 				}
// 			}
// 			if major23VotingPower > totalVotingPower*2/3 || votedAny == totalVotingPower {
// 				break COLLECT
// 			}
// 		case <-time.After(60 * time.Second):
// 			break COLLECT
// 		}
// 	}
// 	if major23VotingPower <= totalVotingPower*2/3 {
// 		return fmt.Errorf("need more than 2/3 total voting power, total:%d, got:%d", totalVotingPower, major23VotingPower)
// 	}
// 	return nil
// }

// func CheckSpecialOPVoteSig(cmd *types.SpecialOPCmd, pk crypto.PubKey, sigData []byte) error {
// 	pk32 := [32]byte(pk.(crypto.PubKeyEd25519))
// 	signature, err := crypto.SignatureFromBytes(sigData)
// 	if err != nil {
// 		return fmt.Errorf("fail to get signature from sigs")
// 	}
// 	sig64 := [64]byte(signature.(crypto.SignatureEd25519))
// 	if !ed25519.Verify(&pk32, cmd.ExCmd, &sig64) {
// 		return fmt.Errorf("signature verification failed: %v", pk32)
// 	}

// 	return nil
// }
