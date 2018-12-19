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

	"github.com/dappledger/AnnChain/angine/refuse_list"
	"github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/ann-module/lib/ed25519"
	"github.com/dappledger/AnnChain/ann-module/lib/go-crypto"
	"github.com/dappledger/AnnChain/ann-module/lib/go-db"
	"github.com/dappledger/AnnChain/ann-module/lib/go-p2p"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
)

type Specialop struct {
	ChangedValidators []*types.ValidatorAttr
	DisconnectedPeers []*p2p.Peer
	AddRefuseKeys     [][32]byte
	DeleteRefuseKeys  [][32]byte

	validators **types.ValidatorSet
	sw         *p2p.Switch
	privkey    crypto.PrivKeyEd25519
	db         *db.DB

	refuselist *refuse_list.RefuseList
}

func NewSpecialop(statedb *db.DB) *Specialop {
	s := Specialop{
		ChangedValidators: make([]*types.ValidatorAttr, 0),
		DisconnectedPeers: make([]*p2p.Peer, 0),
		AddRefuseKeys:     make([][32]byte, 0),
		DeleteRefuseKeys:  make([][32]byte, 0),

		db: statedb,
	}

	return &s
}

func (s *Specialop) InitPlugin(p *InitPluginParams) {
	s.sw = p.Switch
	s.validators = p.Validators // get initial validatorset from switch, then no more updates from it
	s.privkey = p.PrivKey
	s.refuselist = p.RefuseList
}

func (s *Specialop) CheckTx(tx []byte) (bool, error) {
	if !types.IsSpecialOP(tx) {
		return true, nil
	}
	var cmd types.SpecialOPCmd
	err := json.Unmarshal(types.UnwrapTx(tx), &cmd)
	if err != nil || cmd.CmdCode != types.SpecialOP {
		return true, err
	}
	return false, nil
}

func (s *Specialop) DeliverTx(tx []byte, i int) (bool, error) {
	if !types.IsSpecialOP(tx) {
		return true, nil
	}
	var cmd types.SpecialOPCmd
	err := json.Unmarshal(types.UnwrapTx(tx), &cmd)
	if err != nil || cmd.CmdCode != types.SpecialOP {
		return true, err
	}
	err = s.ProcessSpecialOP(&cmd)
	return false, err
}

func (s *Specialop) BeginBlock(p *BeginBlockParams) (*BeginBlockReturns, error) {
	return nil, nil
}

func (s *Specialop) EndBlock(p *EndBlockParams) (*EndBlockReturns, error) {
	defer s.Reset()

	changedValidators := make([]*types.ValidatorAttr, 0, len(s.ChangedValidators)+len(p.ChangedValidators))
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
			s.refuselist.AddRefuseKey(k)
		}
	}
	if len(s.DeleteRefuseKeys) > 0 {
		for _, k := range s.DeleteRefuseKeys {
			s.refuselist.DeleteRefuseKey(k)
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

func (s *Specialop) CheckSpecialOP(cmd *types.SpecialOPCmd) (res error, sig crypto.Signature) {
	nodePubKey, err := crypto.PubKeyFromBytes(cmd.NodePubKey)
	if err != nil {
		return err, s.privkey.Sign([]byte(err.Error()))
	}
	if !s.isValidatorPubKey(nodePubKey) {
		err := errors.New("[CheckSpecialOP] only validators can issue special op")
		return err, s.privkey.Sign([]byte(err.Error()))
	}

	switch cmd.CmdType {
	case types.SpecialOP_ChangeValidator:
		_, err := s.ParseValidator(cmd.Msg)
		if err != nil {
			return err, s.privkey.Sign([]byte(err.Error()))
		}
		return nil, s.privkey.Sign(cmd.ExCmd)
	case types.SpecialOP_Disconnect,
		types.SpecialOP_AddRefuseKey,
		types.SpecialOP_DeleteRefuseKey:
		return nil, s.privkey.Sign(cmd.ExCmd)
	default:
		err := errors.New("unknown special op")
		return err, s.privkey.Sign([]byte(err.Error()))
	}
}

func (s *Specialop) ProcessSpecialOP(cmd *types.SpecialOPCmd) error {
	nodePubKey, err := crypto.PubKeyFromBytes(cmd.NodePubKey)
	if err != nil {
		return err
	}
	if !s.isValidatorPubKey(nodePubKey) {
		return errors.New("[ProcessSpecialOP] only validators can issue special op")
	}

	// TODO
	// Check nonce of nodePubKey

	switch cmd.CmdType {
	case types.SpecialOP_ChangeValidator:
		validator, err := s.ParseValidator(cmd.Msg)
		if err != nil {
			return err
		}
		if validator == nil {
			return errors.New("change validator nil")
		}

		if err := s.ValidateChangeValidator(cmd, validator); err != nil {
			return err
		}

		for index, value := range s.ChangedValidators {
			if value.Power == 0 {
				s.ChangedValidators = append(s.ChangedValidators[:index], s.ChangedValidators[index+1:]...)
			}
		}

		s.ChangedValidators = append(s.ChangedValidators, validator)

	case types.SpecialOP_Disconnect:
		if !s.CheckMajor23(cmd) {
			return errors.New("need more than 2/3 total voting power")
		}

		sw := *(s.sw)
		peers := sw.Peers().List()
		msgPubKey, err := crypto.PubKeyFromBytes(cmd.Msg)
		if err != nil {
			return errors.New("disconnect msg should contain the target peer's pubkey")
		}
		if (*s.validators).HasAddress(msgPubKey.Address()) {
			_, v := (*s.validators).GetByAddress(msgPubKey.Address())
			s.ChangedValidators = append(s.ChangedValidators, &types.ValidatorAttr{Power: 0, IsCA: v.IsCA, RPCAddress: v.RPCAddress, PubKey: v.PubKey.Bytes()})
		}
		for _, peer := range peers {
			if peer.NodeInfo.PubKey == msgPubKey.(crypto.PubKeyEd25519) {
				s.DisconnectedPeers = append(s.DisconnectedPeers, peer)
				break
			}
		}
		s.AddRefuseKeys = append(s.AddRefuseKeys, msgPubKey.(crypto.PubKeyEd25519))
		return nil
	case types.SpecialOP_AddRefuseKey:
		if !s.CheckMajor23(cmd) {
			return errors.New("need more than 2/3 total voting power")
		}

		msgPubKey, err := crypto.PubKeyFromBytes(cmd.Msg)
		if err != nil {
			return errors.New("invalid peer pubkey")
		}
		s.AddRefuseKeys = append(s.AddRefuseKeys, msgPubKey.(crypto.PubKeyEd25519))
	case types.SpecialOP_DeleteRefuseKey:
		if !s.CheckMajor23(cmd) {
			return errors.New("need more than 2/3 total voting power")
		}

		msgPubKey, err := crypto.PubKeyFromBytes(cmd.Msg)
		if err != nil {
			return errors.New("invalid peer pubkey")
		}
		s.DeleteRefuseKeys = append(s.DeleteRefuseKeys, msgPubKey.(crypto.PubKeyEd25519))
	default:
		return errors.New("unsupported special operation")
	}

	return nil
}

func (s *Specialop) ValidateChangeValidator(cmd *types.SpecialOPCmd, toAdd *types.ValidatorAttr) error {
	var major23 int64
	pubToAdd, err := crypto.PubKeyFromBytes(toAdd.PubKey)
	if err != nil {
		return err
	}
	pubToAddEd := pubToAdd.(crypto.PubKeyEd25519)

	for _, v := range (*s.validators).Validators {
		for _, s := range cmd.Sigs {
			signedPkByte64 := types.BytesToByte64(s)

			valPk := [32]byte(v.PubKey.(crypto.PubKeyEd25519))

			if ed25519.Verify(&valPk, pubToAddEd[:], &signedPkByte64) {
				major23 += v.VotingPower

				// We need only one signature of all validators
				return nil
			}
		}
	}

	return fmt.Errorf("ChangeValidator tx failed of insufficient auth")
}

func (s *Specialop) CheckMajor23(cmd *types.SpecialOPCmd) bool {
	var major23 int64
	for _, validator := range (*s.validators).Validators {
		for _, sig := range cmd.Sigs {
			sigPubKey, err := crypto.PubKeyFromBytes(sig[:33])
			if err == nil && validator.PubKey.Equals(sigPubKey) {
				valPubKey := [32]byte(validator.PubKey.(crypto.PubKeyEd25519))
				signature, err := crypto.SignatureFromBytes(sig[33:])
				if err != nil {
					fmt.Println(err)
				}
				sigByte64 := [64]byte(signature.(crypto.SignatureEd25519))
				if ed25519.Verify(&valPubKey, cmd.ExCmd, &sigByte64) {
					major23 += validator.VotingPower
				}
				break
			}
		}
	}
	return major23 > (*s.validators).TotalVotingPower()*2/3
}

func (s *Specialop) ParseValidator(msg []byte) (*types.ValidatorAttr, error) {
	var validator = new(types.ValidatorAttr)
	if err := wire.ReadJSONBytes(msg, validator); err != nil {
		return nil, err
	}
	return validator, nil
}

func (s *Specialop) isValidatorPubKey(pubkey crypto.PubKey) bool {
	isV := false
	for _, v := range (*s.validators).Validators {
		if pubkey.Equals(v.PubKey) {
			isV = true
			break
		}
	}
	return isV
}

func (s *Specialop) updateValidators(validators *types.ValidatorSet, changedValidators []*types.ValidatorAttr) error {
	// TODO: prevent change of 1/3+ at once
	for _, v := range changedValidators {
		pubkey, err := crypto.PubKeyFromBytes(v.PubKey) // NOTE: expects go-wire encoded pubkey
		if err != nil {
			return err
		}
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
			added := validators.Add(types.NewValidator(pubkey, power, v.IsCA, v.RPCAddress))
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
