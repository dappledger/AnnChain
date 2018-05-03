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
	"strings"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/dappledger/AnnChain/module/lib/ed25519"
	"github.com/dappledger/AnnChain/module/lib/ed25519/extra25519"
	. "github.com/dappledger/AnnChain/module/lib/go-common"
)

type StPrivKey struct {
	PrivKey
}

func (p StPrivKey) MarshalJSON() ([]byte, error) {
	if p.PrivKey == nil {
		return json.Marshal(nil)
	}
	return p.PrivKey.MarshalJSON()
}

func (p *StPrivKey) UnmarshalJSON(data []byte) error {
	var dec []interface{}
	err := json.Unmarshal(data, &dec)
	if err != nil {
		return err
	}
	if len(dec) == 0 {
		return nil
	}
	if len(dec) < 2 {
		return errors.New("params missing at unmarshalJson privkey")
	}
	switch byte(dec[0].(float64)) {
	case PrivKeyTypeEd25519:
		p.PrivKey = &PrivKeyEd25519{}
	case PrivKeyTypeSecp256k1:
		p.PrivKey = &PrivKeySecp256k1{}
	default:
		return errors.New("wrong type of pubkey")
	}
	return p.PrivKey.UnmarshalJSON(data)
}

func (p *StPrivKey) String() string {
	if p == nil || p.PrivKey == nil {
		return ""
	}
	return p.PrivKey.String()
}

// PrivKey is part of PrivAccount and state.PrivValidator.
type PrivKey interface {
	Bytes() []byte
	Sign(msg []byte) Signature
	PubKey() PubKey
	Equals(PrivKey) bool
	KeyString() string
	String() string
	json.Marshaler
	json.Unmarshaler
}

// Types of PrivKey implementations
const (
	PrivKeyTypeEd25519   = byte(0x01)
	PrivKeyTypeSecp256k1 = byte(0x02)
)

//-------------------------------------

// Implements PrivKey
type PrivKeyEd25519 [64]byte

func (privKey *PrivKeyEd25519) Bytes() []byte {
	return (*privKey)[:]
}

func (privKey *PrivKeyEd25519) Sign(msg []byte) Signature {
	privKeyBytes := [64]byte(*privKey)
	signatureBytes := ed25519.Sign(&privKeyBytes, msg)
	bys := SignatureEd25519(*signatureBytes)
	return &bys
}

func (privKey *PrivKeyEd25519) PubKey() PubKey {
	privKeyBytes := [64]byte(*privKey)
	pubkey := PubKeyEd25519(*ed25519.MakePublicKey(&privKeyBytes))
	return &pubkey
}

func (privKey *PrivKeyEd25519) Equals(other PrivKey) bool {
	if otherEd, ok := other.(*PrivKeyEd25519); ok {
		return bytes.Equal((*privKey)[:], (*otherEd)[:])
	} else {
		return false
	}
}

func (privKey *PrivKeyEd25519) KeyString() string {
	return Fmt("%X", (*privKey)[:])
}

func (privKey *PrivKeyEd25519) ToCurve25519() *[32]byte {
	keyCurve25519 := new([32]byte)
	privKeyBytes := [64]byte(*privKey)
	extra25519.PrivateKeyToCurve25519(keyCurve25519, &privKeyBytes)
	return keyCurve25519
}

func (privKey *PrivKeyEd25519) String() string {
	return Fmt("PrivKeyEd25519{*****}")
}

func (privKey *PrivKeyEd25519) MarshalJSON() ([]byte, error) {
	hstr := strings.ToUpper(hex.EncodeToString((*privKey)[:64]))
	return json.Marshal([]interface{}{
		PrivKeyTypeEd25519, hstr,
	})
}

func (privKey *PrivKeyEd25519) UnmarshalJSON(data []byte) error {
	var dec []interface{}
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}
	if len(dec) < 2 {
		return errors.New("params missing when unmarshalJson PrivKeyTypeEd25519")
	}
	if byte(dec[0].(float64)) != PrivKeyTypeEd25519 {
		return errors.New("wrong marshal result for PrivKeyTypeEd25519")
	}
	hstr := dec[1].(string)
	bytes, err := hex.DecodeString(hstr)
	if err != nil {
		return err
	}
	if len(bytes) < 64 {
		return errors.New("bytes shorter than 64")
	}
	copy((*privKey)[:64], bytes[:64])
	return nil
}

func GenPrivKeyEd25519() PrivKeyEd25519 {
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], CRandBytes(32))
	ed25519.MakePublicKey(privKeyBytes)
	return PrivKeyEd25519(*privKeyBytes)
}

// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeyEd25519FromSecret(secret []byte) PrivKeyEd25519 {
	privKey32 := Sha256(secret) // Not Ripemd160 because we want 32 bytes.
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], privKey32)
	ed25519.MakePublicKey(privKeyBytes)
	return PrivKeyEd25519(*privKeyBytes)
}

//-------------------------------------

// Implements PrivKey
type PrivKeySecp256k1 [32]byte

func (privKey *PrivKeySecp256k1) Bytes() []byte {
	return (*privKey)[:]
}

func (privKey *PrivKeySecp256k1) Sign(msg []byte) Signature {
	priv__, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), (*privKey)[:])
	sig__, err := priv__.Sign(Sha256(msg))
	if err != nil {
		PanicSanity(err)
	}
	bys := SignatureSecp256k1(sig__.Serialize())
	return &bys
}

func (privKey *PrivKeySecp256k1) PubKey() PubKey {
	_, pub__ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), (*privKey)[:])
	pub := [64]byte{}
	copy(pub[:], pub__.SerializeUncompressed()[1:])
	pubkey := PubKeySecp256k1(pub)
	return &pubkey
}

func (privKey *PrivKeySecp256k1) Equals(other PrivKey) bool {
	if otherSecp, ok := other.(*PrivKeySecp256k1); ok {
		return bytes.Equal((*privKey)[:], (*otherSecp)[:])
	}
	return false
}

func (privKey *PrivKeySecp256k1) String() string {
	return Fmt("PrivKeySecp256k1{*****}")
}

func (privKey *PrivKeySecp256k1) KeyString() string {
	return Fmt("%X", (*privKey)[:])
}

func (privKey *PrivKeySecp256k1) MarshalJSON() ([]byte, error) {
	hstr := strings.ToUpper(hex.EncodeToString((*privKey)[:32]))
	return json.Marshal([]interface{}{
		PrivKeyTypeSecp256k1, hstr,
	})
}

func (privKey *PrivKeySecp256k1) UnmarshalJSON(data []byte) error {
	var dec []interface{}
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}
	if len(dec) < 2 {
		return errors.New("params missing when unmarshalJson PrivKeyTypeSecp256k1")
	}
	if byte(dec[0].(float64)) != PrivKeyTypeSecp256k1 {
		return errors.New("wrong marshal result for PrivKeyTypeSecp256k1")
	}
	hstr := dec[1].(string)
	bytes, err := hex.DecodeString(hstr)
	if err != nil {
		return err
	}
	if len(bytes) < 32 {
		return errors.New("bytes shorter than 64")
	}
	copy((*privKey)[:32], bytes[:32])
	return nil
}

func GenPrivKeySecp256k1() PrivKeySecp256k1 {
	privKeyBytes := [32]byte{}
	copy(privKeyBytes[:], CRandBytes(32))
	priv, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privKeyBytes[:])
	copy(privKeyBytes[:], priv.Serialize())
	return PrivKeySecp256k1(privKeyBytes)
}

// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeySecp256k1FromSecret(secret []byte) PrivKeySecp256k1 {
	privKey32 := Sha256(secret) // Not Ripemd160 because we want 32 bytes.
	priv, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privKey32)
	privKeyBytes := [32]byte{}
	copy(privKeyBytes[:], priv.Serialize())
	return PrivKeySecp256k1(privKeyBytes)
}
