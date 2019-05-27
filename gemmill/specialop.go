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

package gemmill

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	crypto "github.com/dappledger/AnnChain/gemmill/go-crypto"
	log "github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/plugin"
	"github.com/dappledger/AnnChain/gemmill/types"
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

	publicKey := crypto.SetNodePubkey(cmd.PubKey)
	odlsigdata := cmd.Signature
	signature := crypto.SetNodeSignature(odlsigdata)
	cmd.Signature = nil
	signMessage, _ := json.Marshal(cmd)

	if !publicKey.VerifyBytes(signMessage, signature) {
		return errors.Errorf("invalid signature")
	}
	cmd.Signature = odlsigdata

	_, validators := e.consensus.GetValidators()
	myPubKey := e.privValidator.PubKey
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
	cmd.Sigs = append(cmd.Sigs, append(crypto.GetNodePubkeyBytes(myPubKey), mySigbytes...))

	if len(validators) > 1 {
		if err := e.CollectSpecialVotes(cmd); err != nil {
			log.Error("collect special votes", zap.Error(err))
			return err
		}
	}

	cmdBytes, _ := json.Marshal(cmd)
	return e.BroadcastTx(types.WrapTx(types.SpecialTag, cmdBytes))
}

// AppendSignatureToSpecialCmd appends signature onto cmd.Sigs
func (e *Angine) AppendSignatureToSpecialCmd(cmd *types.SpecialOPCmd, pubkey crypto.PubKey, sig crypto.Signature) {
	pkey := crypto.GetNodePubkeyBytes(pubkey)
	sigbytes := crypto.GetNodeSigBytes(sig)
	cmd.Sigs = append(cmd.Sigs, append(pkey, sigbytes...))
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
			log.Warn("specialvote, collecting votes timeout")
			break COLLECT
		case v := <-votesCh:
			voteresult := &types.SpecialVoteResult{}
			if err := json.Unmarshal(v, voteresult); err != nil {
				log.Warn("specialvote", zap.Error(err))
				continue
			}

			publicKey := crypto.SetNodePubkey(voteresult.PubKey)
			sinagure := crypto.SetNodeSignature(voteresult.Signature)
			voteresult.Signature = nil
			signMessage, _ := json.Marshal(voteresult)
			if !publicKey.VerifyBytes(signMessage, sinagure) {
				log.Warn("specialvote, signature is invalid")
				continue
			}

			if !validators.HasAddress(publicKey.Address()) {
				log.Warn("specialvote, non-validator blended in", zap.String("pubkey", publicKey.KeyString()))
				continue
			}

			if _, ok := votedPubKeys[publicKey.KeyString()]; ok {
				log.Warn("specialvote, adversory validator with double vote: %s", zap.String("pubkey", publicKey.KeyString()))
				continue
			}
			votedPubKeys[publicKey.KeyString()] = struct{}{}

			_, val := validators.GetByAddress(publicKey.Address())

			votedAny += val.VotingPower
			sig := crypto.SetNodeSignature(voteresult.Result)
			if publicKey.VerifyBytes(cmd.Msg, sig) {
				major23VotingPower += val.VotingPower
				e.AppendSignatureToSpecialCmd(cmd, val.PubKey, sig)
			} else {
				sig2 := crypto.SetNodeSignature([]byte{})
				e.AppendSignatureToSpecialCmd(cmd, val.PubKey, sig2)
			}

			if major23VotingPower > (e.consensus.GetTotalVotingPower() * 2 / 3) {
				cancelCollect()
				return nil
			}
			if votedAny == e.consensus.GetTotalVotingPower() {
				cancelCollect()
				log.Error("specialvote, insufficient votes")
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
		log.Error("collect votes", zap.Error(err))
		return nil
	}

	txPubkey := crypto.SetNodePubkey(cmd.PubKey)
	txSig := crypto.SetNodeSignature(cmd.Signature)
	oldSig := cmd.Signature

	cmd.Signature = nil
	sigsBak := cmd.Sigs
	cmd.Sigs = nil
	signMessage, _ := json.Marshal(cmd)

	cmd.Signature = oldSig
	cmd.Sigs = sigsBak

	// vote and sign
	var res *types.SpecialVoteResult
	pubbytes := crypto.GetNodePubkeyBytes(e.PrivValidator().PubKey)
	if txPubkey.VerifyBytes(signMessage, txSig) {
		sig, err := e.SignSpecialOP(cmd)
		if err != nil {
			log.Error("sign special error", zap.Error(err))
			return nil
		}
		res = &types.SpecialVoteResult{
			Result: sig,
			PubKey: pubbytes,
		}
	} else {
		res = &types.SpecialVoteResult{
			Result: nil,
			PubKey: pubbytes,
		}
	}

	resBytes, _ := json.Marshal(res)
	signature := e.PrivValidator().Sign(resBytes)
	res.Signature = crypto.GetNodeSigBytes(signature)
	resBytes, _ = json.Marshal(res)

	return resBytes
}

// CheckSpecialOPVoteSig just wraps the action on how to check the votes we got.
func CheckSpecialOPVoteSig(cmd *types.SpecialOPCmd, pk crypto.PubKey, sigData []byte) error {
	if len(sigData) != 64 {
		return errors.Errorf("sigData shoud be 64-byte long, got %d", len(sigData))
	}

	sig := crypto.SetNodeSignature(sigData)
	if !pk.VerifyBytes(cmd.Msg, sig) {
		return fmt.Errorf("signature verification failed: %x", crypto.GetNodePubkeyBytes(pk))
	}

	return nil
}
