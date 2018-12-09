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

	"github.com/dappledger/AnnChain/angine/blockchain/refuse_list"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/ed25519"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/lib/go-db"
	"github.com/dappledger/AnnChain/module/lib/go-p2p"
)

type Specialop struct {
	ChangedValidators []*agtypes.ValidatorAttr
	DisconnectedPeers []*p2p.Peer
	AddRefuseKeys     [][32]byte
	DeleteRefuseKeys  [][32]byte

	validators  **agtypes.ValidatorSet
	sw          *p2p.Switch
	privkey     crypto.PrivKeyEd25519
	db          db.DB
	logger      *zap.Logger
	refuselist  *refuse_list.RefuseList
	eventSwitch agtypes.EventSwitch
}

func NewSpecialop(logger *zap.Logger, statedb db.DB) *Specialop {
	s := Specialop{
		ChangedValidators: make([]*agtypes.ValidatorAttr, 0),
		DisconnectedPeers: make([]*p2p.Peer, 0),
		AddRefuseKeys:     make([][32]byte, 0),
		DeleteRefuseKeys:  make([][32]byte, 0),
		logger:            logger,
		db:                statedb,
	}
	//TODO InitParams
	return &s
}

func (s *Specialop) Init(p *InitParams) {
	s.ChangedValidators = make([]*agtypes.ValidatorAttr, 0)
	s.DisconnectedPeers = make([]*p2p.Peer, 0)
	s.AddRefuseKeys = make([][32]byte, 0)
	s.DeleteRefuseKeys = make([][32]byte, 0)

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
	s.logger = p.Logger
	s.db = p.DB
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
			return false, errors.New("need more than 2/3 total voting power")
		}
	}
	return false, s.ProcessSpecialOP(cmd)
}

func (s *Specialop) ExecBlock(p *ExecBlockParams) (*ExecBlockReturns, error) {
	// Run ExTxs of block
	for i, tx := range p.Block.Data.ExTxs {
		if _, err := s.DeliverTx(tx, i); err != nil {
			s.logger.Error("[Plugin Specialop ExecBlock]", zap.Error(err))
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

func (s *Specialop) SignSpecialOP(cmd *agtypes.SpecialOPCmd) (sig crypto.SignatureEd25519, res error) {
	nodePubKey := crypto.PubKeyEd25519{}
	copy(nodePubKey[:], cmd.PubKey)
	if !s.isCA(&nodePubKey) {
		err := errors.New("[SignSpecialOP] only CA can issue special op")
		return crypto.SignatureEd25519{}, err
	}

	// verify all the signatures from cmd.sigs, return error if anything fails
	// for _, sig := range cmd.Sigs {
	// 	pk32 := [32]byte{}
	// 	copy(pk32[:], sig[:32])
	// 	sig64 := [64]byte{}
	// 	copy(sig64[:], sig[32:])
	// 	if !ed25519.Verify(&pk32, cmd.Msg, &sig64) {
	// 		err := errors.New("signature verification failed")
	// 		return crypto.SignatureEd25519{}, err
	// 	}
	// }

	switch cmd.CmdType {
	case agtypes.SpecialOP_ChangeValidator:
		_, err := s.ParseValidator(cmd)
		if err != nil {
			return crypto.SignatureEd25519{}, err
		}
		return *(s.privkey.Sign(cmd.Msg).(*crypto.SignatureEd25519)), nil
	case agtypes.SpecialOP_Disconnect,
		agtypes.SpecialOP_AddRefuseKey,
		agtypes.SpecialOP_DeleteRefuseKey:
		return *(s.privkey.Sign(cmd.Msg).(*crypto.SignatureEd25519)), nil
	default:
		err := errors.New("unknown special op")
		return crypto.SignatureEd25519{}, err
	}
}

func (s *Specialop) ProcessSpecialOP(cmd *agtypes.SpecialOPCmd) error {
	nodePubKey := crypto.PubKeyEd25519{}
	copy(nodePubKey[:], cmd.PubKey)

	if !s.isCA(&nodePubKey) {
		return errors.New("[ProcessSpecialOP] only CA can issue special op")
	}
	//if !s.CheckMajor23(cmd) {
	//	fmt.Println("err check Major 2222")
	//	return errors.New("need more than 2/3 total voting power")
	//}
	switch cmd.CmdType {
	case agtypes.SpecialOP_ChangeValidator:
		validator, err := s.ParseValidator(cmd)
		if err != nil {
			return err
		}
		s.ChangedValidators = append(s.ChangedValidators, validator)
	case agtypes.SpecialOP_Disconnect:
		sw := *(s.sw)
		peers := sw.Peers().List()
		msgPubKey := crypto.PubKeyEd25519{}
		copy(msgPubKey[:], cmd.Msg)
		if (*s.validators).HasAddress(msgPubKey.Address()) {
			_, v := (*s.validators).GetByAddress(msgPubKey.Address())
			pk := v.GetPubKey().(*crypto.PubKeyEd25519)
			s.ChangedValidators = append(s.ChangedValidators, &agtypes.ValidatorAttr{Power: 0, IsCA: v.IsCA, PubKey: pk[:]})
		}
		for _, peer := range peers {
			if peer.NodeInfo.PubKey == msgPubKey {
				s.DisconnectedPeers = append(s.DisconnectedPeers, peer)
				break
			}
		}
		s.AddRefuseKeys = append(s.AddRefuseKeys, [32]byte(msgPubKey))
		return nil
	case agtypes.SpecialOP_AddRefuseKey:
		msgPubKey := crypto.PubKeyEd25519{}
		copy(msgPubKey[:], cmd.Msg)
		s.AddRefuseKeys = append(s.AddRefuseKeys, [32]byte(msgPubKey))
	case agtypes.SpecialOP_DeleteRefuseKey:
		msgPubKey := crypto.PubKeyEd25519{}
		copy(msgPubKey[:], cmd.Msg)
		s.DeleteRefuseKeys = append(s.DeleteRefuseKeys, [32]byte(msgPubKey))
	default:
		return errors.New("unsupported special operation")
	}

	return nil
}

func (s *Specialop) CheckMajor23(cmd *agtypes.SpecialOPCmd) bool {
	var major23 int64
	for _, sig := range cmd.Sigs {
		sigPubKey := crypto.PubKeyEd25519{}
		copy(sigPubKey[:], sig[:32])
		if (*s.validators).HasAddress(sigPubKey.Address()) {
			_, validator := (*s.validators).GetByAddress(sigPubKey.Address())
			pubKey32 := [32]byte(sigPubKey)
			sig64 := [64]byte{}
			copy(sig64[:], sig[32:])
			if ed25519.Verify(&pubKey32, cmd.Msg, &sig64) {
				major23 += validator.VotingPower
			} else {
				s.logger.Info("check major 2/3", zap.String("vote nil", fmt.Sprintf("%X", pubKey32)))
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
		pubkey := crypto.PubKeyEd25519{}
		copy(pubkey[:], v.PubKey)
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
			added := validators.Add(agtypes.NewValidator(&pubkey, power, v.IsCA))
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
