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
	"github.com/dappledger/AnnChain/gemmill/go-hash"
)

var (
	node_crypto_type = CryptoTypeZhongAn //default value;
	CryptoType       = CryptoTypeZhongAn
)

func GetNodeCryptoType() string {
	return node_crypto_type
}

func NodeInit(crypto string) {
	switch crypto {
	case CryptoTypeZhongAn:
		node_crypto_type = crypto
		hash.ConfigHasher(hash.HashTypeRipemd160)
	default:
		hash.ConfigHasher(hash.HashTypeRipemd160)
	}
}

//--------------------------------ed25519-----------------------------
func setNodePubkey_ed25519(data []byte) PubKey {
	msgPubKey := PubKeyEd25519{}
	copy(msgPubKey[:], data)
	return msgPubKey
}

func setNodePrivKey_ed25519(data []byte) PrivKey {
	pk := PrivKeyEd25519{}
	copy(pk[:], data)
	return pk
}

func setNodeSignature_ed25519(data []byte) Signature {
	pk := SignatureEd25519{}
	copy(pk[:], data)
	return pk
}

//--------------------------------------------------
func NodePubkeyLen() int {
	return PubKeyLenEd25519
}

func GenNodePrivKey() PrivKey {
	return GenPrivKeyEd25519()
}

func SetNodePubkey(data []byte) PubKey {
	return setNodePubkey_ed25519(data)
}

func SetNodePrivKey(data []byte) PrivKey {
	return setNodePrivKey_ed25519(data)
}

func SetNodeSignature(data []byte) Signature {
	return setNodeSignature_ed25519(data)
}

func GetNodePubkeyBytes(pkey PubKey) []byte {
	gpkey := pkey.(PubKeyEd25519)
	return gpkey[:]
}

func GetNodePrivKeyBytes(pkey PrivKey) []byte {
	gpkey := pkey.(PrivKeyEd25519)
	return gpkey[:]
}

func GetNodeSigBytes(sig Signature) []byte {
	gsig := sig.(SignatureEd25519)
	return gsig[:]
}
