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

	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/dappledger/AnnChain/module/lib/ed25519"
	"github.com/dappledger/AnnChain/module/lib/ed25519/extra25519"
	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/xlib"
	"golang.org/x/crypto/ripemd160"
)

type StPubKey struct {
	PubKey
}

func (p StPubKey) MarshalJSON() ([]byte, error) {
	if p.PubKey == nil {
		return json.Marshal(nil)
	}
	return p.PubKey.MarshalJSON()
}

func (p *StPubKey) UnmarshalJSON(data []byte) error {
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
	case PubKeyTypeEd25519:
		p.PubKey = &PubKeyEd25519{}
	case PubKeyTypeSecp256k1:
		p.PubKey = &PubKeySecp256k1{}
	default:
		return errors.New("wrong type of pubkey")
	}
	return p.PubKey.UnmarshalJSON(data)
}

func (p *StPubKey) String() string {
	if p == nil || p.PubKey == nil {
		return ""
	}
	return p.PubKey.String()
}

// PubKey is part of Account and Validator.
type PubKey interface {
	Address() []byte
	Bytes() []byte
	KeyString() string
	VerifyBytes(msg []byte, sig Signature) bool
	Equals(PubKey) bool
	String() string
	json.Marshaler
	json.Unmarshaler
}

// Types of PubKey implementations
const (
	PubKeyTypeEd25519   = byte(0x01)
	PubKeyTypeSecp256k1 = byte(0x02)
)

func PubKeyFromBytes(tp byte, pubKeyBytes []byte) (pubKey PubKey, err error) {
	switch tp {
	case PubKeyTypeEd25519:
		pub := [32]byte{}
		copy(pub[:], pubKeyBytes)
		pk := PubKeyEd25519(pub)
		return &pk, nil
	case PubKeyTypeSecp256k1:
		pub := [64]byte{}
		copy(pub[:], pubKeyBytes)
		pk := PubKeySecp256k1(pub)
		return &pk, nil
	}
	return nil, errors.New("undefined pubkey type")
}

//-------------------------------------

// Implements PubKey
type PubKeyEd25519 [32]byte

func (pubKey *PubKeyEd25519) Address() []byte {
	var w bytes.Buffer
	if err := xlib.WriteBytes(&w, (*pubKey)[:]); err != nil {
		PanicCrisis(err)
	}
	encodedPubkey := append([]byte{PubKeyTypeEd25519}, w.Bytes()...)
	hasher := ripemd160.New()
	hasher.Write(encodedPubkey) // does not error
	return hasher.Sum(nil)
}

func (pubKey *PubKeyEd25519) Bytes() []byte {
	//return wire.BinaryBytes(struct{ PubKey }{pubKey})
	return (*pubKey)[:]
}

func (pubKey *PubKeyEd25519) VerifyBytes(msg []byte, sig_ Signature) bool {
	sig, ok := sig_.(*SignatureEd25519)
	if !ok {
		fmt.Println("to ed25519 failed")
		return false
	}
	pubKeyBytes := [32]byte(*pubKey)
	sigBytes := [64]byte(*sig)
	return ed25519.Verify(&pubKeyBytes, msg, &sigBytes)
}

// For use with golang/crypto/nacl/box
// If error, returns nil.
func (pubKey *PubKeyEd25519) ToCurve25519() *[32]byte {
	keyCurve25519, pubKeyBytes := new([32]byte), [32]byte(*pubKey)
	ok := extra25519.PublicKeyToCurve25519(keyCurve25519, &pubKeyBytes)
	if !ok {
		return nil
	}
	return keyCurve25519
}

func (pubKey *PubKeyEd25519) String() string {
	return Fmt("PubKeyEd25519{%X}", (*pubKey)[:])
}

// Must return the full bytes in hex.
// Used for map keying, etc.
func (pubKey *PubKeyEd25519) KeyString() string {
	return Fmt("%X", (*pubKey)[:])
}

func (pubKey *PubKeyEd25519) Equals(other PubKey) bool {
	if otherEd, ok := other.(*PubKeyEd25519); ok {
		return bytes.Equal((*pubKey)[:], (*otherEd)[:])
	} else {
		return false
	}
}

func (pubKey *PubKeyEd25519) MarshalJSON() ([]byte, error) {
	hstr := strings.ToUpper(hex.EncodeToString((*pubKey)[:32]))
	return json.Marshal([]interface{}{
		PubKeyTypeEd25519, hstr,
	})
}

func (pubKey *PubKeyEd25519) UnmarshalJSON(data []byte) error {
	var dec []interface{}
	err := json.Unmarshal(data, &dec)
	if err != nil {
		return err
	}
	if len(dec) < 2 {
		return errors.New("params missing when unmarshalJson PubKeyEd25519")
	}
	if byte(dec[0].(float64)) != PubKeyTypeEd25519 {
		return errors.New("wrong marshal result for PubKeyTypeEd25519")
	}
	hstr := dec[1].(string)
	bytes, err := hex.DecodeString(hstr)
	if err != nil {
		return err
	}
	if len(bytes) < 32 {
		return errors.New("bytes shorter than 32")
	}
	copy(pubKey[:32], bytes[:32])
	return nil
}

//-------------------------------------

// Implements PubKey
type PubKeySecp256k1 [64]byte

func (pubKey *PubKeySecp256k1) Address() []byte {
	var w bytes.Buffer
	err := xlib.WriteBytes(&w, (*pubKey)[:])
	if err != nil {
		PanicCrisis(err)
	}
	// append type byte
	encodedPubkey := append([]byte{PubKeyTypeSecp256k1}, w.Bytes()...)
	hasher := ripemd160.New()
	hasher.Write(encodedPubkey) // does not error
	return hasher.Sum(nil)
}

func (pubKey *PubKeySecp256k1) Bytes() []byte {
	//return wire.BinaryBytes(struct{ PubKey }{pubKey})
	return (*pubKey)[:]
}

func (pubKey *PubKeySecp256k1) VerifyBytes(msg []byte, sig_ Signature) bool {
	pub__, err := secp256k1.ParsePubKey(append([]byte{0x04}, (*pubKey)[:]...), secp256k1.S256())
	if err != nil {
		return false
	}
	sig, ok := sig_.(*SignatureSecp256k1)
	if !ok {
		return false
	}
	sig__, err := secp256k1.ParseDERSignature((*sig)[:], secp256k1.S256())
	if err != nil {
		return false
	}
	return sig__.Verify(Sha256(msg), pub__)
}

func (pubKey *PubKeySecp256k1) String() string {
	return Fmt("PubKeySecp256k1{%X}", (*pubKey)[:])
}

// Must return the full bytes in hex.
// Used for map keying, etc.
func (pubKey *PubKeySecp256k1) KeyString() string {
	return Fmt("%X", (*pubKey)[:])
}

func (pubKey *PubKeySecp256k1) Equals(other PubKey) bool {
	if otherSecp, ok := other.(*PubKeySecp256k1); ok {
		return bytes.Equal((*pubKey)[:], (*otherSecp)[:])
	} else {
		return false
	}
}

func (pubKey *PubKeySecp256k1) MarshalJSON() ([]byte, error) {
	hstr := strings.ToUpper(hex.EncodeToString((*pubKey)[:64]))
	return json.Marshal([]interface{}{
		PubKeyTypeSecp256k1, hstr,
	})
}

func (pubKey *PubKeySecp256k1) UnmarshalJSON(data []byte) error {
	var dec []interface{}
	err := json.Unmarshal(data, &dec)
	if err != nil {
		return err
	}
	if len(dec) < 2 {
		return errors.New("params missing when unmarshalJson PubKeyEd25519")
	}
	if byte(dec[0].(float64)) != PubKeyTypeSecp256k1 {
		return errors.New("wrong marshal result for PubKeyTypeSecp256k1")
	}
	hstr := dec[1].(string)
	bytes, err := hex.DecodeString(hstr)
	if err != nil {
		return err
	}
	if len(bytes) < 64 {
		return errors.New("bytes shorter than 64")
	}
	copy((*pubKey)[:64], bytes[:64])
	return nil
}
