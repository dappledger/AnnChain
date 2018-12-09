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

package crypto

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	. "github.com/dappledger/AnnChain/module/lib/go-common"
)

type StSignature struct {
	Signature
}

func (s StSignature) MarshalJSON() ([]byte, error) {
	if s.Signature == nil {
		return json.Marshal(nil)
	}
	return s.Signature.MarshalJSON()
}

func (s *StSignature) UnmarshalJSON(data []byte) error {
	var dec []interface{}
	err := json.Unmarshal(data, &dec)
	if err != nil {
		return err
	}
	if len(dec) == 0 {
		return nil
	}
	if len(dec) < 2 {
		return errors.New("params missing when unmarshalJson PubKey")
	}
	switch byte(dec[0].(float64)) {
	case SignatureTypeEd25519:
		s.Signature = &SignatureEd25519{}
	case SignatureTypeSecp256k1:
		s.Signature = &SignatureSecp256k1{}
	default:
		return errors.New("unknown signature type")
	}
	return s.Signature.UnmarshalJSON(data)
}

func (s *StSignature) String() string {
	if s == nil || s.Signature == nil {
		return ""
	}
	return s.Signature.String()
}

// Signature is a part of Txs and consensus Votes.
type Signature interface {
	Bytes() []byte
	IsZero() bool
	String() string
	Equals(Signature) bool
	json.Marshaler
	json.Unmarshaler
}

// Types of Signature implementations
const (
	SignatureTypeEd25519    = byte(0x01)
	SignatureTypeSecp256k1  = byte(0x02)
	PSignatureTypeEd25519   = byte(0x03)
	PSignatureTypeSecp256k1 = byte(0x04)
)

func SignatureFromBytes(tp byte, sigBytes []byte) (sig Signature, err error) {
	switch tp {
	case SignatureTypeEd25519:
		sigt := [64]byte{}
		copy(sigt[:], sigBytes)
		sigarr := SignatureEd25519(sigt)
		return &sigarr, nil
	case SignatureTypeSecp256k1:
		sigt := SignatureSecp256k1(sigBytes)
		return &sigt, nil
	}
	return nil, errors.New("undefined signature type")
}

//-------------------------------------

// Implements Signature
type SignatureEd25519 [64]byte

func (sig *SignatureEd25519) Bytes() []byte {
	//	return wire.BinaryBytes(struct{ Signature }{sig})
	return (*sig)[:]
}

func (sig *SignatureEd25519) IsZero() bool { return sig == nil || len(*sig) == 0 }

func (sig *SignatureEd25519) String() string {
	if sig == nil {
		return ""
	}
	return fmt.Sprintf("/%X.../", Fingerprint((*sig)[:]))
}

func (sig *SignatureEd25519) Equals(other Signature) bool {
	if sig == nil {
		return other == nil
	}
	if otherEd, ok := other.(*SignatureEd25519); ok {
		return bytes.Equal((*sig)[:], (*otherEd)[:])
	}
	return false
}

func (sig *SignatureEd25519) MarshalJSON() ([]byte, error) {
	if sig == nil {
		return json.Marshal(nil)
	}
	hstr := strings.ToUpper(hex.EncodeToString((*sig)[:64]))
	return json.Marshal([]interface{}{
		SignatureTypeEd25519, hstr,
	})
}

func (sig *SignatureEd25519) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || sig == nil {
		return nil
	}
	var dec []interface{}
	err := json.Unmarshal(data, &dec)
	if err != nil {
		return err
	}
	if len(dec) < 2 {
		return errors.New("params missing when unmarshalJson SignatureEd25519")
	}
	if byte(dec[0].(float64)) != SignatureTypeEd25519 {
		return errors.New("wrong marshal result for SignatureEd25519")
	}
	hstr := dec[1].(string)
	bytes, err := hex.DecodeString(hstr)
	if err != nil {
		return err
	}
	if len(bytes) < 64 {
		return errors.New("bytes shorter than 64, for unmarshal SignatureEd25519")
	}
	copy((*sig)[:64], bytes[:64])
	return nil
}

//-------------------------------------

// Implements Signature
type SignatureSecp256k1 []byte

func (sig *SignatureSecp256k1) Bytes() []byte {
	if sig == nil {
		return nil
	}
	return *sig
}

func (sig *SignatureSecp256k1) IsZero() bool { return sig == nil || len(*sig) == 0 }

func (sig *SignatureSecp256k1) String() string {
	if sig == nil {
		return ""
	}
	return fmt.Sprintf("/%X.../", Fingerprint((*sig)[:]))
}

func (sig *SignatureSecp256k1) Equals(other Signature) bool {
	if sig == nil {
		return other == nil
	}
	if otherEd, ok := other.(*SignatureSecp256k1); ok {
		return bytes.Equal((*sig)[:], (*otherEd)[:])
	}
	return false
}

func (sig *SignatureSecp256k1) MarshalJSON() ([]byte, error) {
	if sig == nil {
		return json.Marshal(nil)
	}
	hstr := strings.ToUpper(hex.EncodeToString((*sig)[:64]))
	return json.Marshal([]interface{}{
		SignatureTypeSecp256k1, hstr,
	})
}

func (sig *SignatureSecp256k1) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || sig == nil {
		return nil
	}
	var dec []interface{}
	err := json.Unmarshal(data, &dec)
	if err != nil {
		return err
	}
	if len(dec) < 2 {
		return errors.New("params missing when unmarshalJson SignatureSecp256k1")
	}
	if byte(dec[0].(float64)) != SignatureTypeSecp256k1 {
		return errors.New("wrong marshal result for SignatureSecp256k1")
	}
	hstr := dec[1].(string)
	bytes, err := hex.DecodeString(hstr)
	if err != nil {
		return err
	}
	if len(bytes) < 64 {
		return errors.New("bytes shorter than 64, for unmarshal SignatureSecp256k1")
	}
	copy((*sig)[:64], bytes[:64])
	return nil
}
