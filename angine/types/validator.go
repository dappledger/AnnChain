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
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/ripemd160"

	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

// Volatile state for each Validator
// TODO: make non-volatile identity
// 	- Remove Accum - it can be computed, and now valset becomes identifying
type Validator struct {
	Address     []byte          `json:"address"`
	PubKey      crypto.StPubKey `json:"pub_key"`
	VotingPower def.INT         `json:"voting_power"`
	Accum       def.INT         `json:"accum"`
	IsCA        bool            `json:"is_ca"`
}

func (v *Validator) GetPubKey() crypto.PubKey {
	return v.PubKey.PubKey
}

type ValidatorAttr struct {
	PubKey []byte `json:"pubKey,omitempty"`
	Power  uint64 `json:"power,omitempty"`
	IsCA   bool   `json:"isCA,omitempty"`
}

func (m *ValidatorAttr) Reset() { *m = ValidatorAttr{} }

func (m *ValidatorAttr) String() string {
	return fmt.Sprintf("[%X,%v,%v]", m.PubKey, m.Power, m.IsCA)
}

func (m *ValidatorAttr) GetPubKey() []byte {
	if m != nil {
		return m.PubKey
	}
	return nil
}

func (m *ValidatorAttr) GetPower() uint64 {
	if m != nil {
		return m.Power
	}
	return 0
}

func (m *ValidatorAttr) GetIsCA() bool {
	if m != nil {
		return m.IsCA
	}
	return false
}

func NewValidator(pubKey crypto.PubKey, votingPower def.INT, isCA bool) *Validator {
	return &Validator{
		Address:     pubKey.Address(),
		PubKey:      crypto.StPubKey{pubKey},
		VotingPower: votingPower,
		Accum:       0,
		IsCA:        isCA,
	}
}

// Creates a new copy of the validator so we can mutate accum.
// Panics if the validator is nil.
func (v *Validator) Copy() *Validator {
	vCopy := *v
	return &vCopy
}

// Returns the one with higher Accum.
func (v *Validator) CompareAccum(other *Validator) *Validator {
	if v == nil {
		return other
	}
	if v.Accum > other.Accum {
		return v
	} else if v.Accum < other.Accum {
		return other
	} else {
		if bytes.Compare(v.Address, other.Address) < 0 {
			return v
		} else if bytes.Compare(v.Address, other.Address) > 0 {
			return other
		} else {
			PanicSanity("Cannot compare identical validators")
			return nil
		}
	}
}

func (v *Validator) String() string {
	if v == nil {
		return "nil-Validator"
	}
	return fmt.Sprintf("Validator{%X %v VP:%v A:%v CA:%v}",
		v.Address,
		v.PubKey,
		v.VotingPower,
		v.Accum,
		v.IsCA)
}

func (v *Validator) Bytes() []byte {
	bys, err := json.Marshal(v)
	if err != nil {
		fmt.Println("debug:json marshal validator wrong", err)
	}
	return bys
}

func (v *Validator) Hash() []byte {
	hasher := ripemd160.New()
	hasher.Write(v.Bytes())
	return hasher.Sum(nil)
}

//--------------------------------------------------------------------------------
// For testing...

func RandValidator(logger *zap.Logger, randPower bool, minPower def.INT) (*Validator, *PrivValidator) {
	privVal := GenPrivValidator(logger, nil)
	_, tempFilePath := Tempfile("priv_validator_")
	privVal.SetFile(tempFilePath)
	votePower := minPower
	if randPower {
		votePower += def.INT(RandUint32())
	}
	val := NewValidator(privVal.GetPubKey(), votePower, true)
	return val, privVal
}
