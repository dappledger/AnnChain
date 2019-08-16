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

package hash

import (
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

type HashType string

const (
	HashTypeSha256    HashType = "sha256"
	HashTypeRipemd160          = "ripemd160"
)

var DoHash = ripemd160Func
var DoHashName string

func ConfigHasher(typ HashType) {
	switch typ {
	case HashTypeSha256:
		DoHash = sha256Func
	default:
		DoHash = ripemd160Func
	}
}

func sha256Func(bytes []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(bytes)
	return hasher.Sum(nil)

}

func ripemd160Func(bytes []byte) []byte {
	hasher := ripemd160.New()
	hasher.Write(bytes)
	return hasher.Sum(nil)
}

//Workrand
func Keccak256Func(bytes []byte) []byte {
	return sha256Func(bytes)
}
