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

package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-utils"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"go.uber.org/zap"
)

const (
	stepNone      = 0 // Used to distinguish the initial state
	stepPropose   = 1
	stepPrevote   = 2
	stepPrecommit = 3
)

func voteToStep(vote *Vote) int8 {
	switch vote.Type {
	case VoteTypePrevote:
		return stepPrevote
	case VoteTypePrecommit:
		return stepPrecommit
	default:
		gcmn.PanicSanity("Unknown vote type")
		return 0
	}
}

type PrivValidator struct {
	Address       []byte           `json:"address"`
	PubKey        crypto.PubKey    `json:"pub_key"`
	LastHeight    int64            `json:"last_height"`
	LastRound     int64            `json:"last_round"`
	LastStep      int8             `json:"last_step"`
	LastSignature crypto.Signature `json:"last_signature"` // so we dont lose signatures
	LastSignBytes []byte           `json:"last_signbytes"` // so we dont lose signatures

	// PrivKey should be empty if a Signer other than the default is being used.
	PrivKey crypto.PrivKey `json:"priv_key"`
	Signer  `json:"-"`

	// For persistence.
	// Overloaded for testing.
	filePath string
	mtx      sync.Mutex
}

// This is used to sign votes.
// It is the caller's duty to verify the msg before calling Sign,
// eg. to avoid double signing.
// Currently, the only callers are SignVote and SignProposal
type Signer interface {
	Sign(msg []byte) crypto.Signature
}

// Implements Signer
type DefaultSigner struct {
	priv crypto.PrivKey
}

func NewDefaultSigner(priv crypto.PrivKey) *DefaultSigner {
	return &DefaultSigner{priv: priv}
}

// Implements Signer
func (ds *DefaultSigner) Sign(msg []byte) crypto.Signature {
	return ds.priv.Sign(msg)
}

func (privVal *PrivValidator) SetSigner(s Signer) {
	privVal.Signer = s
}

// Generates a new validator with private key.
func GenPrivValidator(cryptoType string, privkey crypto.PrivKey) (*PrivValidator, error) {
	privKey := privkey
	if utils.CheckItfcNil(privkey) {
		var err error
		privKey, err = crypto.GenPrivkeyByType(cryptoType)
		if err != nil {
			return nil, err
		}
	}

	pubKey := privKey.PubKey()
	return &PrivValidator{
		Address:       pubKey.Address(),
		PubKey:        pubKey,
		PrivKey:       privKey,
		LastHeight:    0,
		LastRound:     0,
		LastStep:      stepNone,
		LastSignature: nil,
		LastSignBytes: nil,
		filePath:      "",
		Signer:        NewDefaultSigner(privKey),
	}, nil
}

func LoadPrivValidator(filePath string) (*PrivValidator, error) {
	privValJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	privVal := wire.ReadJSON(&PrivValidator{}, privValJSONBytes, &err).(*PrivValidator)
	if err != nil {
		return nil, err
	}
	privVal.filePath = filePath
	privVal.Signer = NewDefaultSigner(privVal.PrivKey)
	return privVal, nil
}

func LoadOrGenPrivValidator(filePath string) *PrivValidator {
	var privValidator *PrivValidator
	if _, err := os.Stat(filePath); err == nil {
		privValidator, _ = LoadPrivValidator(filePath)
		log.Infow("Loaded PrivValidator", "file", filePath, "privValidator", privValidator)
	} else {
		privValidator, _ = GenPrivValidator("", nil)
		privValidator.SetFile(filePath)
		privValidator.Save()
		log.Info("Generated PrivValidator", zap.String("file", filePath))
	}
	return privValidator
}

func (pv *PrivValidator) CopyReset() (cp PrivValidator) {
	cp.PubKey = pv.PubKey
	cp.PrivKey = pv.PrivKey
	cp.Signer = NewDefaultSigner(pv.GetPrivKey())
	return
}

func (pv *PrivValidator) GetPubKey() crypto.PubKey {
	return pv.PubKey
}

func (pv *PrivValidator) GetPrivKey() crypto.PrivKey {
	return pv.PrivKey
}

func (pv *PrivValidator) GetLastSignature() crypto.Signature {
	return pv.LastSignature
}

func (pv *PrivValidator) UnmarshalJSON(data []byte) error {
	st := struct {
		PubKey        crypto.PubKey    `json:"pub_key"`
		LastHeight    int64            `json:"last_height"`
		LastRound     int64            `json:"last_round"`
		LastStep      int8             `json:"last_step"`
		LastSignature crypto.Signature `json:"last_signature"`
		LastSignBytes []byte           `json:"last_signbytes"`
		PrivKey       crypto.PrivKey   `json:"priv_key"`
	}{}
	if err := json.Unmarshal(data, &st); err != nil {
		return err
	}
	pv.PubKey = st.PubKey
	pv.LastHeight = st.LastHeight
	pv.LastStep = st.LastStep
	pv.LastRound = st.LastRound
	pv.LastSignature = st.LastSignature
	pv.LastSignBytes = st.LastSignBytes
	pv.PrivKey = st.PrivKey
	return nil
}

func (privVal *PrivValidator) SetFile(filePath string) {
	privVal.mtx.Lock()
	defer privVal.mtx.Unlock()
	privVal.filePath = filePath
}

func (privVal *PrivValidator) Save() error {
	privVal.mtx.Lock()
	defer privVal.mtx.Unlock()
	return privVal.save()
}

func (privVal *PrivValidator) save() error {
	if privVal.filePath == "" {
		// gcmn.PanicSanity("Cannot save PrivValidator: filePath not set")
		return fmt.Errorf("Cannot save PrivValidator: filePath is empty")
	}
	jsonBytes := wire.JSONBytesPretty(privVal)
	err := gcmn.WriteFileAtomic(privVal.filePath, jsonBytes, 0600)
	if err != nil {
		// `@; BOOM!!!
		// gcmn.PanicCrisis(err)
		return err
	}

	return nil
}

// NOTE: Unsafe!
func (privVal *PrivValidator) Reset() {
	privVal.LastHeight = 0
	privVal.LastRound = 0
	privVal.LastStep = 0
	privVal.LastSignature = nil
	privVal.LastSignBytes = nil
	privVal.Save()
}

func (privVal *PrivValidator) GetAddress() []byte {
	return privVal.Address
}

func (privVal *PrivValidator) GetPrivateKey() crypto.PrivKey {
	return privVal.PrivKey
}

func (privVal *PrivValidator) SignVote(chainID string, vote *Vote) error {
	privVal.mtx.Lock()
	defer privVal.mtx.Unlock()
	signature, err := privVal.signBytesHRS(vote.Height, vote.Round, voteToStep(vote), SignBytes(chainID, vote))
	if err != nil {
		return errors.New(gcmn.Fmt("Error signing vote: %v", err))
	}
	vote.Signature = signature
	return nil
}

func (privVal *PrivValidator) SignProposal(chainID string, proposal *Proposal) error {
	privVal.mtx.Lock()
	defer privVal.mtx.Unlock()
	signature, err := privVal.signBytesHRS(proposal.Height, proposal.Round, stepPropose, SignBytes(chainID, proposal))
	if err != nil {
		return errors.New(gcmn.Fmt("Error signing proposal: %v", err))
	}
	proposal.Signature = signature
	return nil
}

// check if there's a regression. Else sign and write the hrs+signature to disk
func (privVal *PrivValidator) signBytesHRS(height, round int64, step int8, signBytes []byte) (crypto.Signature, error) {
	// If height regression, err
	if privVal.LastHeight > height {
		return nil, errors.New("Height regression")
	}
	// More cases for when the height matches
	if privVal.LastHeight == height {
		// If round regression, err
		if privVal.LastRound > round {
			return nil, errors.New("Round regression")
		}
		// If step regression, err
		if privVal.LastRound == round {
			if privVal.LastStep > step {
				return nil, errors.New("Step regression")
			} else if privVal.LastStep == step {
				if privVal.LastSignBytes != nil {
					if privVal.LastSignature == nil {
						gcmn.PanicSanity("privVal: LastSignature is nil but LastSignBytes is not!")
					}
					// so we dont sign a conflicting vote or proposal
					// NOTE: proposals are non-deterministic (include time),
					// so we can actually lose them, but will still never sign conflicting ones
					if bytes.Equal(privVal.LastSignBytes, signBytes) {
						log.Infof("Using privVal.LastSignature: %X", privVal.LastSignature)
						return privVal.LastSignature, nil
					}
				}
				return nil, errors.New("Step regression")
			}
		}
	}

	// Sign
	signature := privVal.Sign(signBytes)

	// Persist height/round/step
	privVal.LastHeight = height
	privVal.LastRound = round
	privVal.LastStep = step
	privVal.LastSignature = signature
	privVal.LastSignBytes = signBytes
	privVal.save()

	return signature, nil

}

func (privVal *PrivValidator) String() string {
	return fmt.Sprintf("PrivValidator{%X LH:%v, LR:%v, LS:%v}", privVal.Address, privVal.LastHeight, privVal.LastRound, privVal.LastStep)
}

//-------------------------------------

type PrivValidatorsByAddress []*PrivValidator

func (pvs PrivValidatorsByAddress) Len() int {
	return len(pvs)
}

func (pvs PrivValidatorsByAddress) Less(i, j int) bool {
	return bytes.Compare(pvs[i].Address, pvs[j].Address) == -1
}

func (pvs PrivValidatorsByAddress) Swap(i, j int) {
	it := pvs[i]
	pvs[i] = pvs[j]
	pvs[j] = it
}
