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

package plugin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/dappledger/AnnChain/gemmill/refuse_list"
	agtypes "github.com/dappledger/AnnChain/gemmill/types"
)

type AdminApp interface {
	GetNonce() uint64
	From() []byte
}

type AdminOp struct {
	ChangedValidators []*agtypes.ValidatorAttr
	DisconnectedPeers []*p2p.Peer
	AddRefuseKeys     []crypto.PubKey
	DeleteRefuseKeys  []crypto.PubKey
	validators        **agtypes.ValidatorSet
	sw                *p2p.Switch
	privkey           crypto.PrivKey
	refuselist        *refuse_list.RefuseList
	eventSwitch       agtypes.EventSwitch
}

func (s *AdminOp) Init(p *InitParams) {
	s.ChangedValidators = make([]*agtypes.ValidatorAttr, 0)
	s.DisconnectedPeers = make([]*p2p.Peer, 0)
	s.AddRefuseKeys = make([]crypto.PubKey, 0)
	s.DeleteRefuseKeys = make([]crypto.PubKey, 0)

	s.sw = p.Switch
	s.validators = p.Validators // get initial validatorset from switch, then no more updates from it
	s.privkey = p.PrivKey
	s.refuselist = p.RefuseList
}

func (s *AdminOp) Reload(p *ReloadParams) {
	s.sw = p.Switch
	s.validators = p.Validators // get initial validatorset from switch, then no more updates from it
	s.privkey = p.PrivKey
	s.refuselist = p.RefuseList
	//s.stateDB = p.StateDB
}

func (s *AdminOp) Stop() {

}

func (s *AdminOp) CheckTx(tx []byte) (bool, error) {
	if !agtypes.IsAdminOP(tx) {
		return true, nil
	}
	cmd := &agtypes.AdminOPCmd{}
	if err := json.Unmarshal(agtypes.UnwrapTx(tx), cmd); err != nil {
		return true, err
	}
	return false, nil
}

func (s *AdminOp) DeliverTx(tx []byte, i int) (bool, error) {
	if !agtypes.IsAdminOP(tx) {
		return true, nil
	}
	cmd := &agtypes.AdminOPCmd{}
	if err := json.Unmarshal(agtypes.UnwrapTx(tx), cmd); err != nil {
		return true, err
	}
	if i >= 0 { //vote channel needn't to check signs again
		if !s.CheckMajor23(cmd) {
			log.Error("need more than 2/3 total voting power")
			return false, nil
		}
	}
	return false, s.ProcessAdminOP(cmd, nil)
}

func (s *AdminOp) ExecTX(app AdminApp, tx []byte) error {
	data := agtypes.UnwrapTx(tx)
	cmd := &agtypes.AdminOPCmd{}
	if err := json.Unmarshal(data, cmd); err != nil {
		log.Error(fmt.Sprintf("AdminOPCmd Unmarshal :%s\ndata=%x", err.Error(), data))
		return err
	}
	if !s.CheckMajor23(cmd) {
		log.Error("need more than 2/3 total voting power")
		return fmt.Errorf("need more than 2/3 total voting power")
	}
	return s.ProcessAdminOP(cmd, app)
}

func (s *AdminOp) ExecBlock(p *ExecBlockParams) (*ExecBlockReturns, error) {
	// Run ExTxs of block
	for i, tx := range p.Block.Data.ExTxs {
		if _, err := s.DeliverTx(tx, i); err != nil {
			return nil, fmt.Errorf("[Plugin AdminOp ExecBlock]:%s", err.Error())
		}
	}

	return nil, nil
}

func (s *AdminOp) BeginBlock(p *BeginBlockParams) (*BeginBlockReturns, error) {
	return nil, nil
}

func (s *AdminOp) EndBlock(p *EndBlockParams) (*EndBlockReturns, error) {
	defer s.Reset()
	changedValidators := make([]*agtypes.ValidatorAttr, 0, len(s.ChangedValidators)+len(p.ChangedValidators))
	copy(changedValidators, p.ChangedValidators)
	for _, v := range s.ChangedValidators {
		overrideByApp := false
		for _, vv := range p.ChangedValidators {
			if bytes.Equal(v.GetPubKey(), vv.GetPubKey()) {
				overrideByApp = true
				break
			}
		}
		if !overrideByApp {
			changedValidators = append(changedValidators, v)
		}
	}
	err := s.updateValidators(p.NextValidatorSet, changedValidators)
	if err != nil {
		return &EndBlockReturns{NextValidatorSet: p.NextValidatorSet}, err
	}

	// s.validators is a ** pointing to *(state.validators)
	// update validatorset in out plugin & switch
	if s.validators != nil {
		*s.validators = p.NextValidatorSet
	}

	for _, peer := range s.DisconnectedPeers {
		s.sw.StopPeerGracefully(peer)
	}

	if len(s.AddRefuseKeys) > 0 {
		for _, k := range s.AddRefuseKeys {
			s.refuselist.AddRefuseKey(k.Bytes())
		}
	}
	if len(s.DeleteRefuseKeys) > 0 {
		for _, k := range s.DeleteRefuseKeys {
			s.refuselist.DeleteRefuseKey(k.Bytes())
		}
	}
	return &EndBlockReturns{NextValidatorSet: p.NextValidatorSet}, nil
}

func (s *AdminOp) Reset() {
	s.ChangedValidators = s.ChangedValidators[:0]
	s.DisconnectedPeers = s.DisconnectedPeers[:0]
	s.AddRefuseKeys = s.AddRefuseKeys[:0]
	s.DeleteRefuseKeys = s.DeleteRefuseKeys[:0]
}

// if power <=0, node is a peer, else if power > 0, node is a validator
// addnode: change node to peer
// update: change power value so that node to be validator
// remove: remove node from validators
func (s *AdminOp) ProcessAdminOP(cmd *agtypes.AdminOPCmd, app AdminApp) error {
	if cmd.CmdType != agtypes.AdminOpChangeValidator {
		return errors.New("unsupported admin operation")
	}
	vAttr, err := s.ParseValidator(cmd)
	if err != nil {
		return fmt.Errorf("parse validator err:%v", err)
	}
	if !bytes.Equal(app.From(), vAttr.Addr) {
		return fmt.Errorf("verify nonce err")
	}
	nonce := app.GetNonce()
	if vAttr.Nonce+1 != nonce {
		err = fmt.Errorf("admin nonce error:need(%d) gived(%d)", nonce, vAttr.Nonce)
		return err
	}

	msgPubKey := crypto.SetNodePubkey(vAttr.PubKey)
	switch vAttr.Cmd {
	case agtypes.ValidatorCmdAddPeer:
		nodePub := crypto.SetNodePubkey(vAttr.PubKey)
		nodeSig := crypto.SetNodeSignature(cmd.SelfSign)
		if !nodePub.VerifyBytes(cmd.Msg, nodeSig) {
			log.Error(fmt.Sprintf("node(%s) self-verify faield", nodePub.KeyString()))
			return fmt.Errorf("self verify failed")
		}
		_, val := (*s.validators).GetByAddress(msgPubKey.Address())
		if val != nil {
			log.Warn(fmt.Sprintf("node(%s) has in chain;", msgPubKey.KeyString()))
			return nil
		}
		s.ChangedValidators = append(s.ChangedValidators, vAttr)
		s.DeleteRefuseKeys = append(s.DeleteRefuseKeys, msgPubKey)
		return nil
	case agtypes.ValidatorCmdUpdateNode:
		_, val := (*s.validators).GetByAddress(msgPubKey.Address())
		if val == nil {
			log.Warn(fmt.Sprintf("please add node(%s) first;", msgPubKey.KeyString()))
			return fmt.Errorf("not add into chain")
		}
		if val.VotingPower == vAttr.Power {
			log.Warn(fmt.Sprintf("node(%s) has the same state;", msgPubKey.KeyString()))
			return nil
		}
		s.ChangedValidators = append(s.ChangedValidators, vAttr)
		s.DeleteRefuseKeys = append(s.DeleteRefuseKeys, msgPubKey)
		return nil
	case agtypes.ValidatorCmdRemoveNode:
		if !(*s.validators).HasAddress(msgPubKey.Address()) {
			return nil
		}
		s.ChangedValidators = append(s.ChangedValidators, vAttr)
		//disconnect;
		sw := *(s.sw)
		peers := sw.Peers().List()
		for _, peer := range peers {
			if peer.NodeInfo.PubKey == msgPubKey {
				s.DisconnectedPeers = append(s.DisconnectedPeers, peer)
				break
			}
		}
		s.AddRefuseKeys = append(s.AddRefuseKeys, msgPubKey)
		return nil
	default:
		return errors.New("unsupported admin operation:" + string(vAttr.Cmd))
	}

	return nil
}

func (s *AdminOp) CheckMajor23(cmd *agtypes.AdminOPCmd) bool {
	msg := cmd.Msg
	var major23 int64
	for _, sig := range cmd.SInfos {
		sigPubKey := crypto.SetNodePubkey(sig.PubKey)
		_, validator := (*s.validators).GetByAddress(sigPubKey.Address())
		if validator != nil && validator.VotingPower > 0 {
			sig64 := crypto.SetNodeSignature(sig.Signature)
			if sigPubKey.VerifyBytes(msg, sig64) {
				major23 += validator.VotingPower
			} else {
				log.Info("check major 2/3", zap.String("vote nil", fmt.Sprintf("sig=%X;pubkey=%X", sig.Signature, sigPubKey.KeyString())))
			}
		} else {
			log.Warn(fmt.Sprintf("node(%s) is not validator", sigPubKey.KeyString()))
		}
	}
	return major23 > (*s.validators).TotalVotingPower()*2/3
}

func (s *AdminOp) ParseValidator(cmd *agtypes.AdminOPCmd) (*agtypes.ValidatorAttr, error) {
	validator := &agtypes.ValidatorAttr{}
	data, err := cmd.ExtractMsg(validator)
	if err != nil {
		return nil, err
	}
	validator, ok := data.(*agtypes.ValidatorAttr)
	if !ok {
		return nil, errors.New("change validator nil")
	}
	return validator, nil
}

func (s *AdminOp) isValidatorPubKey(pubkey crypto.PubKey) bool {
	return (*s.validators).HasAddress(pubkey.Address())
}

func (s *AdminOp) updateValidators(validators *agtypes.ValidatorSet, changedValidators []*agtypes.ValidatorAttr) error {
	// TODO: prevent change of 1/3+ at once
	for _, vAttr := range changedValidators {
		pubkey := crypto.SetNodePubkey(vAttr.PubKey)
		address := pubkey.Address()
		switch vAttr.Cmd {
		case agtypes.ValidatorCmdAddPeer, agtypes.ValidatorCmdUpdateNode:
			_, val := validators.GetByAddress(address)
			if val == nil {
				// TODO: check if validator node really exists
				added := validators.Add(agtypes.NewValidator(pubkey, vAttr.GetPower(), vAttr.GetIsCA()))
				if !added {
					return fmt.Errorf("failed to add new validator %X with voting power %d", address, vAttr.GetPower())
				}
			} else {
				if val.VotingPower != vAttr.GetPower() {
					val.VotingPower = vAttr.GetPower()
					val.IsCA = vAttr.GetIsCA()
					updated := validators.Update(val)
					if !updated {
						return fmt.Errorf("failed to update validator %X with voting power %d", address, vAttr.GetPower())
					}
				}
			}
		case agtypes.ValidatorCmdRemoveNode:
			_, removed := validators.Remove(address)
			if !removed {
				return fmt.Errorf("Failed to remove validator %X", address)
			}
		default:
			log.Warn("unsupported admin operation:" + string(vAttr.Cmd))
		}
	}
	return nil
}

func (s *AdminOp) SetEventSwitch(sw agtypes.EventSwitch) {
	s.eventSwitch = sw
}
