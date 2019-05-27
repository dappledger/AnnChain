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
	"github.com/dappledger/AnnChain/gemmill/modules/go-db"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/dappledger/AnnChain/gemmill/refuse_list"
	agtypes "github.com/dappledger/AnnChain/gemmill/types"
)

type Specialop struct {
	ChangedValidators []*agtypes.ValidatorAttr
	DisconnectedPeers []*p2p.Peer
	AddRefuseKeys     []crypto.PubKey
	DeleteRefuseKeys  []crypto.PubKey

	validators **agtypes.ValidatorSet
	sw         *p2p.Switch
	privkey    crypto.PrivKey
	//stateDB     db.DB
	refuselist  *refuse_list.RefuseList
	eventSwitch agtypes.EventSwitch
}

func NewSpecialop(statedb db.DB) *Specialop {
	s := Specialop{
		ChangedValidators: make([]*agtypes.ValidatorAttr, 0),
		DisconnectedPeers: make([]*p2p.Peer, 0),
		AddRefuseKeys:     make([]crypto.PubKey, 0),
		DeleteRefuseKeys:  make([]crypto.PubKey, 0),
		//stateDB:           statedb,
	}
	//TODO InitParams
	return &s
}

func (s *Specialop) Init(p *InitParams) {
	s.ChangedValidators = make([]*agtypes.ValidatorAttr, 0)
	s.DisconnectedPeers = make([]*p2p.Peer, 0)
	s.AddRefuseKeys = make([]crypto.PubKey, 0)
	s.DeleteRefuseKeys = make([]crypto.PubKey, 0)

	s.sw = p.Switch
	s.validators = p.Validators // get initial validatorset from switch, then no more updates from it
	s.privkey = p.PrivKey
	s.refuselist = p.RefuseList
}

func (s *Specialop) Reload(p *ReloadParams) {
	s.sw = p.Switch
	s.validators = p.Validators // get initial validatorset from switch, then no more updates from it
	s.privkey = p.PrivKey
	s.refuselist = p.RefuseList
	//s.stateDB = p.StateDB
}

func (s *Specialop) Stop() {

}

func (s *Specialop) CheckTx(tx []byte) (bool, error) {
	if !agtypes.IsSpecialOP(tx) {
		return true, nil
	}
	cmd := &agtypes.SpecialOPCmd{}
	if err := json.Unmarshal(agtypes.UnwrapTx(tx), cmd); err != nil {
		return true, err
	}
	return false, nil
}

func (s *Specialop) DeliverTx(tx []byte, i int) (bool, error) {
	if !agtypes.IsSpecialOP(tx) {
		return true, nil
	}
	cmd := &agtypes.SpecialOPCmd{}
	if err := json.Unmarshal(agtypes.UnwrapTx(tx), cmd); err != nil {
		return true, err
	}
	if i >= 0 { //vote channel needn't to check signs again
		if !s.CheckMajor23(cmd) {
			log.Error("need more than 2/3 total voting power")
			return false, nil
		}
	}
	return false, s.ProcessSpecialOP(cmd)
}

func (s *Specialop) ExecBlock(p *ExecBlockParams) (*ExecBlockReturns, error) {
	// Run ExTxs of block
	for i, tx := range p.Block.Data.ExTxs {
		if _, err := s.DeliverTx(tx, i); err != nil {
			return nil, fmt.Errorf("[Plugin Specialop ExecBlock]:%s", zap.Error(err))
		}
	}

	return nil, nil
}

func (s *Specialop) BeginBlock(p *BeginBlockParams) (*BeginBlockReturns, error) {
	return nil, nil
}

func (s *Specialop) EndBlock(p *EndBlockParams) (*EndBlockReturns, error) {
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

func (s *Specialop) Reset() {
	s.ChangedValidators = s.ChangedValidators[:0]
	s.DisconnectedPeers = s.DisconnectedPeers[:0]
	s.AddRefuseKeys = s.AddRefuseKeys[:0]
	s.DeleteRefuseKeys = s.DeleteRefuseKeys[:0]
}

func (s *Specialop) SignSpecialOP(cmd *agtypes.SpecialOPCmd) (sigdata []byte, res error) {
	nodePubKey := crypto.SetNodePubkey(cmd.PubKey)
	switch cmd.CmdType {
	case agtypes.SpecialOP_ChangeValidator:
		if !s.isCA(nodePubKey) {
			err := errors.New("[SignSpecialOP] only CA can issue special op")
			return []byte{}, err
		}
		_, err := s.ParseValidator(cmd)
		if err != nil {
			return []byte{}, err
		}
		return crypto.GetNodeSigBytes(s.privkey.Sign(cmd.Msg)), nil
	case agtypes.SpecialOP_Disconnect,
		agtypes.SpecialOP_AddRefuseKey,
		agtypes.SpecialOP_DeleteRefuseKey:
		if !s.isCA(nodePubKey) {
			err := errors.New("[SignSpecialOP] only CA can issue special op")
			return []byte{}, err
		}
		return crypto.GetNodeSigBytes(s.privkey.Sign(cmd.Msg)), nil
	default:
		err := errors.New("unknown special op")
		return []byte{}, err
	}
}

func (s *Specialop) ProcessSpecialOP(cmd *agtypes.SpecialOPCmd) error {
	switch cmd.CmdType {
	case agtypes.SpecialOP_ChangeValidator:
		validator, err := s.ParseValidator(cmd)
		if err != nil {
			return err
		}
		s.ChangedValidators = append(s.ChangedValidators, validator)

		msgPubKey := crypto.SetNodePubkey(validator.PubKey)
		s.DeleteRefuseKeys = append(s.DeleteRefuseKeys, msgPubKey)
	case agtypes.SpecialOP_Disconnect:
		sw := *(s.sw)
		peers := sw.Peers().List()

		msgPubKey := crypto.SetNodePubkey(cmd.Msg)
		if (*s.validators).HasAddress(msgPubKey.Address()) {
			_, v := (*s.validators).GetByAddress(msgPubKey.Address())
			pk := crypto.GetNodePubkeyBytes(v.PubKey)
			s.ChangedValidators = append(s.ChangedValidators, &agtypes.ValidatorAttr{Power: 0, IsCA: v.IsCA, PubKey: pk[:]})
		}
		for _, peer := range peers {
			if peer.NodeInfo.PubKey == msgPubKey {
				s.DisconnectedPeers = append(s.DisconnectedPeers, peer)
				break
			}
		}
		s.AddRefuseKeys = append(s.AddRefuseKeys, (msgPubKey))
		return nil
	case agtypes.SpecialOP_AddRefuseKey:

		msgPubKey := crypto.SetNodePubkey(cmd.Msg)
		s.AddRefuseKeys = append(s.AddRefuseKeys, (msgPubKey))
	case agtypes.SpecialOP_DeleteRefuseKey:
		bys := []byte{}
		if _, err := cmd.ExtractMsg(&bys); err != nil {
			return err
		}

		msgPubKey := crypto.SetNodePubkey(bys)
		s.DeleteRefuseKeys = append(s.DeleteRefuseKeys, (msgPubKey))
	default:
		return errors.New("unsupported special operation")
	}

	return nil
}

func (s *Specialop) CheckMajor23(cmd *agtypes.SpecialOPCmd) bool {
	var major23 int64
	for _, sig := range cmd.Sigs {
		sigPubKey := crypto.SetNodePubkey(sig)
		publen := crypto.NodePubkeyLen()
		if (*s.validators).HasAddress(sigPubKey.Address()) {
			_, validator := (*s.validators).GetByAddress(sigPubKey.Address())
			sig64 := crypto.SetNodeSignature(sig[publen:])
			if sigPubKey.VerifyBytes(cmd.Msg, sig64) {
				major23 += validator.VotingPower
			} else {
				log.Info("check major 2/3", zap.String("vote nil", fmt.Sprintf("%X", sig[:publen])))
			}
		}
	}

	return major23 > (*s.validators).TotalVotingPower()*2/3
}

func (s *Specialop) ParseValidator(cmd *agtypes.SpecialOPCmd) (*agtypes.ValidatorAttr, error) {
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

func (s *Specialop) isValidatorPubKey(pubkey crypto.PubKey) bool {
	return (*s.validators).HasAddress(pubkey.Address())
}

func (s *Specialop) isCA(pubkey crypto.PubKey) bool {
	_, v := (*s.validators).GetByAddress(pubkey.Address())
	return v != nil && v.IsCA
}

func (s *Specialop) updateValidators(validators *agtypes.ValidatorSet, changedValidators []*agtypes.ValidatorAttr) error {
	// TODO: prevent change of 1/3+ at once
	for _, v := range changedValidators {
		pubkey := crypto.SetNodePubkey(v.PubKey)
		address := pubkey.Address()
		power := int64(v.Power)
		// mind the overflow from uint64
		if power < 0 {
			return fmt.Errorf("Power (%d) overflows int64", v.Power)
		}

		_, val := validators.GetByAddress(address)
		if val == nil {
			// add val
			// TODO: check if validator node really exists
			added := validators.Add(agtypes.NewValidator(pubkey, power, v.IsCA))
			if !added {
				return fmt.Errorf("Failed to add new validator %X with voting power %d", address, power)
			}
		} else if v.Power == 0 {
			// remove val
			_, removed := validators.Remove(address)
			if !removed {
				return fmt.Errorf("Failed to remove validator %X", address)
			}
		} else {
			if val.VotingPower != power || val.IsCA != v.IsCA {
				// update val
				val.VotingPower = power
				val.IsCA = v.IsCA
				updated := validators.Update(val)
				if !updated {
					return fmt.Errorf("Failed to update validator %X with voting power %d", address, power)
				}
			}
		}
	}
	return nil
}

func (s *Specialop) SetEventSwitch(sw agtypes.EventSwitch) {
	s.eventSwitch = sw
}
