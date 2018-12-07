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

	"go.uber.org/zap"

	. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	"github.com/dappledger/AnnChain/ann-module/lib/go-crypto"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
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
	RPCAddress  string        `json:"rpc"`
}

type ValidatorAttr struct {
	PubKey     []byte `json:"pubKey,omitempty"`
	Power      uint64 `json:"power,omitempty"`
	RPCAddress string `json:"rPCAddress,omitempty"`
	IsCA       bool   `json:"isCA,omitempty"`
}

func (m *ValidatorAttr) Reset() { *m = ValidatorAttr{} }
func (m *ValidatorAttr) String() string {
	return fmt.Sprintf("[%s,%s,%s,%s]", m.PubKey, m.Power, m.IsCA, m.RPCAddress)
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
func (m *ValidatorAttr) GetRPCAddress() string {
	if m != nil {
		return m.RPCAddress
	}
	return ""
}
func (m *ValidatorAttr) GetIsCA() bool {
	if m != nil {
		return m.IsCA
	}
	return false
}

func NewValidator(pubKey crypto.PubKey, votingPower int64, isCA bool, rpcaddress string) *Validator {
	return &Validator{
		Address:     pubKey.Address(),
		PubKey:      pubKey,
		VotingPower: votingPower,
		Accum:       0,
		IsCA:        isCA,
		RPCAddress:  rpcaddress,
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
	return fmt.Sprintf("Validator{%X %v VP:%v A:%v CA:%v RPC:%v}",
		v.Address,
		v.PubKey,
		v.VotingPower,
		v.Accum,
		v.IsCA,
		v.RPCAddress)
}

func (v *Validator) Hash() []byte {
	return wire.BinaryRipemd160(v)
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
	PanicSanity("ValidatorCodec.Compare not implemented")
	return 0
}

//--------------------------------------------------------------------------------
// For testing...

func RandValidator(logger *zap.Logger, randPower bool, minPower int64) (*Validator, *PrivValidator) {
	privVal := GenPrivValidator(logger)
	_, tempFilePath := Tempfile("priv_validator_")
	privVal.SetFile(tempFilePath)
	votePower := minPower
	if randPower {
		votePower += int64(RandUint32())
	}
	val := NewValidator(privVal.PubKey, votePower, true, "tcp://0.0.0.0:46657")
	return val, privVal
}
