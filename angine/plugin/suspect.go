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
	"encoding/hex"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine/consensus"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	// "github.com/dappledger/AnnChain/module/lib/ed25519"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	dbm "github.com/dappledger/AnnChain/module/lib/go-db"
	"github.com/dappledger/AnnChain/module/lib/go-p2p"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

type SuspectPlugin struct {
	logger     *zap.Logger
	db         dbm.DB
	privkey    crypto.PrivKeyEd25519
	sw         *p2p.Switch
	validators **agtypes.ValidatorSet

	validatorsContainer IValidatorsContainer
	eventSwitch         agtypes.EventSwitch
	angine              IBroadcastable
	punishable          func() IPunishable

	hypoMtx    sync.Mutex
	Hypocrites Hypocrites
	Suspects   Suspects

	suspectsAccuseCounter map[string]int
}

type IPunishable interface {
	SuspectValidator(pubkey []byte, reason string)
}

type IValidatorsContainer interface {
	GetValidators() (*agtypes.ValidatorSet, *agtypes.ValidatorSet)
}

type IBroadcastable interface {
	BroadcastTx([]byte) error
	BroadcastTxCommit([]byte) error
}

type Hypocrites map[string][]*agtypes.Hypocrite
type Suspects map[string][]*agtypes.Hypocrite

func (sp *SuspectPlugin) Init(p *InitParams) {
	sp.sw = p.Switch
	sp.validators = p.Validators // get initial validatorset from switch, then no more updates from it
	sp.privkey = p.PrivKey
	sp.logger = p.Logger
	sp.db = p.DB

	sp.Suspects = make(Suspects)
	sp.Hypocrites = make(Hypocrites)
	sp.suspectsAccuseCounter = make(map[string]int)

	agtypes.AddListenerForEvent(sp.eventSwitch, "SuspectPlugin", agtypes.EventStringTimeoutPropose(), func(data agtypes.TMEventData) {
		ed := data.(agtypes.EventDataRoundState)
		rs := ed.RoundState.(*consensus.RoundState)

		proof := agtypes.HypoProposalTimeoutEvidence{
			Proposal: rs.Proposal,
			Height:   ed.Height,
			Round:    ed.Round,
		}

		if rs.ProposalBlockParts != nil {
			proof.BlockPartsBA = rs.ProposalBlockParts.BitArray()
		}

		pubkeyStr := rs.Validators.Proposer().PubKey.KeyString()
		hypo := agtypes.NewHypocrite(pubkeyStr, agtypes.REASON_PROPOSETO, proof)
		sp.AddHypocrite(pubkeyStr, hypo)

		suspectKey := sp.suspectKey(pubkeyStr, proof.Height, proof.Round)
		if _, ok := sp.suspectsAccuseCounter[suspectKey]; !ok {
			sp.suspectsAccuseCounter[suspectKey] = rs.Validators.Size()
		}
	})
}

func (sp *SuspectPlugin) SetPunishable(p func() IPunishable) {
	sp.punishable = p
}

func (sp *SuspectPlugin) SetBroadcastable(a IBroadcastable) {
	sp.angine = a
}

func (sp *SuspectPlugin) SetEventSwitch(sw agtypes.EventSwitch) {
	sp.eventSwitch = sw
}

func (sp *SuspectPlugin) SetValidatorsContainer(c IValidatorsContainer) {
	sp.validatorsContainer = c
}

func (sp *SuspectPlugin) AddHypocrite(pk string, hypo *agtypes.Hypocrite) {
	sp.hypoMtx.Lock()
	sp.Hypocrites[pk] = append(sp.Hypocrites[pk], hypo)
	sp.hypoMtx.Unlock()
}

func (sp *SuspectPlugin) Reload(p *ReloadParams) {
	sp.logger = p.Logger
	sp.db = p.DB
	sp.sw = p.Switch
	sp.validators = p.Validators // get initial validatorset from switch, then no more updates from it
	sp.privkey = p.PrivKey
}

func (sp *SuspectPlugin) CheckTx(tx []byte) (bool, error) {
	if !agtypes.IsSuspectTx(tx) {
		return true, nil
	}
	susTx := &agtypes.SuspectTx{}
	return false, susTx.FromBytes(agtypes.UnwrapTx(tx))
}

func (sp *SuspectPlugin) DeliverTx(tx []byte, i int) (bool, error) {
	if !agtypes.IsSuspectTx(tx) {
		return true, nil
	}
	susTx := &agtypes.SuspectTx{}
	if err := susTx.FromBytes(agtypes.UnwrapTx(tx)); err != nil {
		return false, err
	}

	pk := &crypto.PubKeyEd25519{}
	sg := &crypto.SignatureEd25519{}
	copy(pk[:], susTx.PubKey)
	copy(sg[:], susTx.Signature)
	susTx.Signature = nil
	bs, _ := susTx.ToBytes()
	if !pk.VerifyBytes(bs, sg) {
		return false, errors.Wrap(errors.Errorf("invalid signature, pk: %X, sig: %X", pk[:], sg[:]), "[SuspectPlugin DeliverTx]")
	}

	switch susTx.Suspect.Reason {
	case agtypes.REASON_PROPOSETO:
		evd := susTx.Suspect.Evidence.(*agtypes.HypoProposalTimeoutEvidence)
		suspectKey := sp.suspectKey(susTx.Suspect.PubKey, evd.Height, evd.Round)
		sp.Suspects[suspectKey] = append(sp.Suspects[suspectKey], susTx.Suspect)
		if len(sp.Suspects[suspectKey]) > 2*sp.suspectsAccuseCounter[suspectKey]/3 {
			susBytes, _ := hex.DecodeString(susTx.Suspect.PubKey)
			sp.punishable().SuspectValidator(susBytes, "proposal timeout")
		}
	case agtypes.REASON_BADVOTE:
		evd := susTx.Suspect.Evidence.(*agtypes.HypoBadVoteEvidence)
		suspectKey := sp.suspectKey(susTx.Suspect.PubKey, evd.Height, evd.Round)
		sp.Suspects[suspectKey] = append(sp.Suspects[suspectKey], susTx.Suspect)
		if len(sp.Suspects[suspectKey]) > 2*sp.suspectsAccuseCounter[suspectKey]/3 {
			susBytes, _ := hex.DecodeString(susTx.Suspect.PubKey)
			sp.punishable().SuspectValidator(susBytes, "bad vote")
		}
	}

	return false, nil
}

func (sp *SuspectPlugin) BeginBlock(p *BeginBlockParams) (*BeginBlockReturns, error) {
	pk := sp.privkey.PubKey().(*crypto.PubKeyEd25519)
	sp.hypoMtx.Lock()
	for k, hypos := range sp.Hypocrites {
		for _, h := range hypos {
			tx := agtypes.SuspectTx{
				Suspect: h,
				PubKey:  pk[:],
			}
			toSignBytes, _ := tx.ToBytes()
			sig := sp.privkey.Sign(toSignBytes).(*crypto.SignatureEd25519)
			tx.Signature = sig[:]
			txBytes, _ := tx.ToBytes()
			sp.angine.BroadcastTx(agtypes.WrapTx(agtypes.SuspectTxTag, txBytes))
		}
		delete(sp.Hypocrites, k)
	}
	sp.Hypocrites = make(Hypocrites)
	sp.hypoMtx.Unlock()
	return nil, nil
}

func (sp *SuspectPlugin) EndBlock(p *EndBlockParams) (*EndBlockReturns, error) {
	return nil, nil
}

func (sp *SuspectPlugin) ExecBlock(p *ExecBlockParams) (*ExecBlockReturns, error) {
	for i, tx := range p.Block.Data.ExTxs {
		if _, err := sp.DeliverTx(tx, i); err != nil {
			sp.logger.Error("[Plugin SuspectPlugin ExecBlock]", zap.Error(err))
		}
	}

	return nil, nil
}

func (sp *SuspectPlugin) ReportBadVote(pk crypto.PubKey, evidence interface{}) {
	hypo := agtypes.NewHypocrite(pk.KeyString(), agtypes.REASON_BADVOTE, evidence)
	sp.AddHypocrite(pk.KeyString(), hypo)
}

func (sp *SuspectPlugin) ReportPeerError(pubkey crypto.PubKey, reason interface{}) {
	{
		lastVSet, vSet := sp.validatorsContainer.GetValidators()
		_, lv := lastVSet.GetByAddress(pubkey.Address())
		_, v := vSet.GetByAddress(pubkey.Address())

		// peerError only cares validators
		if lv == nil && v == nil {
			return
		}
	}

	// hypo := agtypes.NewHypocrite(pubkey.KeyString(), agtypes.REASON_CONNECTION, reason)
	// sp.AddHypocrite(pubkey.KeyString(), hypo)
}

func (sp *SuspectPlugin) Reset() {}

func (sp *SuspectPlugin) Stop() {
	sp.eventSwitch.RemoveListener("SuspectPlugin")
}

func (sp *SuspectPlugin) suspectKey(pk string, height, round def.INT) string {
	return agtypes.HeightToA(height) + pk + agtypes.HeightToA(round)
}
