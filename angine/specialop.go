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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine/plugin"
	"github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/ed25519"
	crypto "github.com/dappledger/AnnChain/module/lib/go-crypto"
)

// ProcessSpecialOP is the ability of Angine.
// It is called when a specialop tx comes in but before going into mempool.
// We do a specialvotes collection here.
// Only with 2/3+ voting power can the tx go into mempool.
func (e *Angine) ProcessSpecialOP(tx []byte) error {
	if !types.IsSpecialOP(tx) {
		return fmt.Errorf("tx is not a specialop: %v", tx)
	}
	cmd := &types.SpecialOPCmd{}
	if err := json.Unmarshal(types.UnwrapTx(tx), cmd); err != nil {
		return errors.Wrap(err, "fail parsing SpecialOPCmd")
	}
	signature := [64]byte{}
	publicKey := [32]byte{}
	copy(signature[:], cmd.Signature)
	copy(publicKey[:], cmd.PubKey)
	cmd.Signature = nil
	signMessage, _ := json.Marshal(cmd)
	if !ed25519.Verify(&publicKey, signMessage, &signature) {
		return errors.Errorf("invalid signature")
	}
	cmd.Signature = signature[:]

	_, validators := e.consensus.GetValidators()
	myPubKey := e.privValidator.GetPubKey().(*crypto.PubKeyEd25519)
	var myVotingPower int64
	for _, val := range validators {
		if val.PubKey.KeyString() == myPubKey.KeyString() && val.IsCA {
			myVotingPower = val.VotingPower
			break
		}
	}
	if myVotingPower == 0 {
		return fmt.Errorf("only CA can do specialOP")
	}

	mySigbytes, err := e.SignSpecialOP(cmd)
	if err != nil {
		return err
	}
	// append our own signature
	cmd.Sigs = append(cmd.Sigs, append(myPubKey[:], mySigbytes...))

	if len(validators) > 1 {
		if err := e.CollectSpecialVotes(cmd); err != nil {
			e.logger.Error("collect special votes", zap.Error(err))
			return err
		}
	}

	cmdBytes, _ := json.Marshal(cmd)
	return e.BroadcastTx(types.WrapTx(types.SpecialTag, cmdBytes))
}

// AppendSignatureToSpecialCmd appends signature onto cmd.Sigs
func (e *Angine) AppendSignatureToSpecialCmd(cmd *types.SpecialOPCmd, pubkey crypto.PubKey, sig crypto.Signature) {
	pk := pubkey.(*crypto.PubKeyEd25519)
	s := sig.(*crypto.SignatureEd25519)
	cmd.Sigs = append(cmd.Sigs, append(pk[:], s[:]...))
}

// CollectSpecialVotes collects special votes.
// Communications are made on p2p port by a dedicated reactor.
// Within tracerouter_msg_ttl timeout, see if we can get more than 2/3 voting power.
func (e *Angine) CollectSpecialVotes(cmd *types.SpecialOPCmd) error {
	var votedAny, major23VotingPower int64
	cmdBytes, _ := json.Marshal(cmd)
	_, validators := e.GetValidators()
	candidatesNum := validators.Size() - 1
	votesCh := make(chan []byte, candidatesNum)
	defer close(votesCh)

	// this timeout has nothing to do with consensus, it happens before the tx is accepted
	spCtx, cancelCollect := context.WithTimeout(context.Background(), time.Duration(e.tune.Conf.GetInt("tracerouter_msg_ttl"))*time.Second)
	e.traceRouter.Broadcast(spCtx, cmdBytes, votesCh)

	_, myVal := validators.GetByAddress(e.PrivValidator().GetAddress())
	votedAny = myVal.VotingPower
	major23VotingPower = myVal.VotingPower
	votedPubKeys := make(map[string]struct{})

COLLECT:
	for {
		select {
		case <-spCtx.Done():
			cancelCollect()
			e.logger.Warn("specialvote, collecting votes timeout")
			break COLLECT
		case v := <-votesCh:
			voteresult := &types.SpecialVoteResult{}
			if err := json.Unmarshal(v, voteresult); err != nil {
				e.logger.Warn("specialvote", zap.Error(err))
				continue
			}
			publicKey := [32]byte{}
			signature := [64]byte{}
			copy(signature[:], voteresult.Signature)
			copy(publicKey[:], voteresult.PubKey)
			voteresult.Signature = nil
			signMessage, _ := json.Marshal(voteresult)
			if !ed25519.Verify(&publicKey, signMessage, &signature) {
				e.logger.Warn("specialvote, signature is invalid")
				continue
			}

			pked := crypto.PubKeyEd25519(publicKey)
			if !validators.HasAddress(pked.Address()) {
				e.logger.Warn("specialvote, non-validator blended in", zap.String("pubkey", pked.KeyString()))
				continue
			}

			if _, ok := votedPubKeys[pked.KeyString()]; ok {
				e.logger.Warn("specialvote, adversory validator with double vote: %s", zap.String("pubkey", pked.KeyString()))
				continue
			}
			votedPubKeys[pked.KeyString()] = struct{}{}

			_, val := validators.GetByAddress(pked.Address())
			voteSig := [64]byte{}
			copy(voteSig[:], voteresult.Result)

			votedAny += val.VotingPower
			if ed25519.Verify(&publicKey, cmd.Msg, &voteSig) {
				major23VotingPower += val.VotingPower
				sig := crypto.SignatureEd25519(voteSig)
				e.AppendSignatureToSpecialCmd(cmd, val.GetPubKey(), &sig)
			} else {
				sig := crypto.SignatureEd25519{}
				e.AppendSignatureToSpecialCmd(cmd, val.GetPubKey(), &sig)
			}
			if major23VotingPower > (e.consensus.GetTotalVotingPower() * 2 / 3) {
				cancelCollect()
				return nil
			}
			if votedAny == e.consensus.GetTotalVotingPower() {
				cancelCollect()
				e.logger.Error("specialvote, insufficient votes")
				return errors.Errorf("insufficient votes")
			}
		}
	}

	if major23VotingPower > (e.consensus.GetTotalVotingPower() * 2 / 3) {
		return nil
	}

	return errors.Errorf("insufficient votes")
}

//SignSpecialOP wraps around plugin.Specialop
func (e *Angine) SignSpecialOP(cmd *types.SpecialOPCmd) ([]byte, error) {
	switch cmd.CmdType {
	case types.SpecialOP_ChangeValidator,
		types.SpecialOP_Disconnect,
		types.SpecialOP_AddRefuseKey,
		types.SpecialOP_DeleteRefuseKey:
		var spPlug *plugin.Specialop

		for _, p := range e.plugins {
			if ps, ok := p.(*plugin.Specialop); ok {
				spPlug = ps
				break
			}
		}
		if spPlug != nil {
			sig, err := spPlug.SignSpecialOP(cmd)
			if err != nil {
				return nil, err
			}
			return sig[:], nil
		}
	default:
		return nil, fmt.Errorf("unimplemented: %s", cmd.CmdType)
	}
	return nil, nil
}

// SpecialOPResponseHandler defines what we do when we get a specialop request from a peer.
// This is mainly used as a callback within TraceRouter.
func (e *Angine) SpecialOPResponseHandler(data []byte) []byte {
	cmd := &types.SpecialOPCmd{}
	if err := json.Unmarshal(data, cmd); err != nil {
		e.logger.Error("collect votes", zap.Error(err))
		return nil
	}

	// verify original tx signature
	txSig := [64]byte{}
	txPubkey := [32]byte{}
	copy(txSig[:], cmd.Signature)
	copy(txPubkey[:], cmd.PubKey)
	cmd.Signature = nil
	sigsBak := cmd.Sigs
	cmd.Sigs = nil
	signMessage, _ := json.Marshal(cmd)
	verifyResult := ed25519.Verify(&txPubkey, signMessage, &txSig)
	cmd.Signature = txSig[:]
	cmd.Sigs = sigsBak

	// vote and sign
	var res *types.SpecialVoteResult
	localPub := e.PrivValidator().GetPubKey().(*crypto.PubKeyEd25519)
	if verifyResult {
		sig, err := e.SignSpecialOP(cmd)
		if err != nil {
			e.logger.Error("sign special error", zap.Error(err))
			return nil
		}
		res = &types.SpecialVoteResult{
			Result: sig,
			PubKey: (*localPub)[:],
		}
	} else {
		res = &types.SpecialVoteResult{
			Result: nil,
			PubKey: (*localPub)[:],
		}
	}

	resBytes, _ := json.Marshal(res)
	signature := e.PrivValidator().Sign(resBytes).(*crypto.SignatureEd25519)
	res.Signature = (*signature)[:]
	resBytes, _ = json.Marshal(res)

	return resBytes
}

// CheckSpecialOPVoteSig just wraps the action on how to check the votes we got.
func CheckSpecialOPVoteSig(cmd *types.SpecialOPCmd, pk crypto.PubKey, sigData []byte) error {
	if len(sigData) != 64 {
		return errors.Errorf("sigData shoud be 64-byte long, got %d", len(sigData))
	}
	pk32 := [32]byte(*(pk.(*crypto.PubKeyEd25519)))
	sig64 := [64]byte{}
	copy(sig64[:], sigData)
	if !ed25519.Verify(&pk32, cmd.Msg, &sig64) {
		return fmt.Errorf("signature verification failed: %v", pk32)
	}

	return nil
}
