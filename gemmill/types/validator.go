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
	"fmt"
	"io"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
)

// Volatile state for each Validator
// TODO: make non-volatile identity
// 	- Remove Accum - it can be computed, and now valset becomes identifying
type Validator struct {
	Address     []byte        `json:"address"`
	PubKey      crypto.PubKey `json:"pub_key"`
	VotingPower int64         `json:"voting_power"`
	Accum       int64         `json:"accum"`
	IsCA        bool          `json:"is_ca"`
}

type ValidatorCmd string

const (
	ValidatorCmdAddPeer    ValidatorCmd = "add_peer"
	ValidatorCmdUpdateNode ValidatorCmd = "update_node"
	ValidatorCmdRemoveNode ValidatorCmd = "remove_node"
)

type ValidatorAttr struct {
	PubKey []byte       `json:"pubKey,omitempty"` //hex of pubkey;
	Power  int64        `json:"power,omitempty"`
	Cmd    ValidatorCmd `json:"cmd"`
	Addr   []byte       `json:"addr"`
	Nonce  uint64       `json:"nonce"`
}

func (m *ValidatorAttr) Reset() { *m = ValidatorAttr{} }
func (m *ValidatorAttr) String() string {
	return fmt.Sprintf("[%s,%x,%d]", m.Cmd, m.GetPubKey(), m.GetPower())
}
func (m *ValidatorAttr) GetPubKey() []byte {
	if m != nil {
		return m.PubKey
	}
	return nil
}

func (m *ValidatorAttr) GetPower() int64 {
	if m != nil {
		return m.Power
	}
	return 0
}

func (m *ValidatorAttr) GetIsCA() bool {
	if m != nil {
		return m.GetPower() > 0
	}
	return false
}

func NewValidator(pubKey crypto.PubKey, votingPower int64, isCA bool) *Validator {
	return &Validator{
		Address:     pubKey.Address(),
		PubKey:      pubKey,
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
			gcmn.PanicSanity("Cannot compare identical validators")
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

func (v *Validator) Hash() []byte {
	return wire.BinaryHash(v)
}

//-------------------------------------

var ValidatorCodec = validatorCodec{}

type validatorCodec struct{}

func (vc validatorCodec) Encode(o interface{}, w io.Writer, n *int, err *error) {
	wire.WriteBinary(o.(*Validator), w, n, err)
}

func (vc validatorCodec) Decode(r io.Reader, n *int, err *error) interface{} {
	return wire.ReadBinary(&Validator{}, r, 0, n, err)
}

func (vc validatorCodec) Compare(o1 interface{}, o2 interface{}) int {
	gcmn.PanicSanity("ValidatorCodec.Compare not implemented")
	return 0
}

//--------------------------------------------------------------------------------
// For testing...

func RandValidator(randPower bool, minPower int64) (*Validator, *PrivValidator) {
	privVal, err := GenPrivValidator(crypto.CryptoTypeZhongAn, nil)
	if err != nil {
		gcmn.PanicSanity("Failed to generate PrivValidator")
	}
	_, tempFilePath := gcmn.Tempfile("priv_validator_")
	privVal.SetFile(tempFilePath)
	votePower := minPower
	if randPower {
		votePower += int64(gcmn.RandUint32())
	}
	val := NewValidator(privVal.PubKey, votePower, true)
	return val, privVal
}
