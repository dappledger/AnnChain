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
	"encoding/hex"
	"errors"

	ed "github.com/dappledger/AnnChain/module/lib/ed25519"
	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-merkle"
)

// Signable is an interface for all signable things.
// It typically removes signatures before serializing.
type Signable interface {
	// TODO rectify to buffer
	GetBytes(chainID string) ([]byte, error)
}

// SignBytes is a convenience method for getting the bytes to sign of a Signable.
func SignBytes(chainID string, o Signable) []byte {
	bys, err := o.GetBytes(chainID)
	if err != nil {
		PanicCrisis(err)
	}
	return bys
}

// HashSignBytes is a convenience method for getting the hash of the bytes of a signable
func HashSignBytes(chainID string, o Signable) []byte {
	return merkle.SimpleHashFromBinary(SignBytes(chainID, o))
}

func SignCA(secbytes []byte, plainTxt []byte) (string, error) {
	if len(plainTxt) < 32 {
		return "", errors.New("size of pubkey too short")
	}
	var secbyte64 [64]byte
	copy(secbyte64[:], secbytes)
	pubBytes := plainTxt[:32]
	chainID := []byte(plainTxt[32:])
	signature := ed.Sign(&secbyte64, append(pubBytes, chainID...))
	ss := hex.EncodeToString((*signature)[:])
	return ss, nil
}
