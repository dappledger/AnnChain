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
	"sync/atomic"

	"golang.org/x/crypto/sha3"

	ethcmn "github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/rlp"
)

// Transaction tx
type Transaction struct {
	Data TxData

	// caches
	hash  atomic.Value
	valid atomic.Value
}

type TxData struct {
	Caller     ethcmn.Address
	PublicKey  []byte
	Signature  []byte
	TimeStamp  uint64
	CryptoType string
}

func (tx *Transaction) Hash() (h ethcmn.Hash) {
	if hash := tx.hash.Load(); hash != nil {
		h = hash.(ethcmn.Hash)
		return
	}

	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, tx.Data)
	hw.Sum(h[:0])

	tx.hash.Store(h)
	return
}

func (tx *Transaction) SigHash() (h ethcmn.Hash) {
	return rlpHash([]interface{}{
		tx.Data.Caller,
		tx.Data.TimeStamp,
	})
}

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP() ([]byte, error) {
	return rlp.EncodeToBytes(tx.Data)
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(b []byte) error {
	return rlp.DecodeBytes(b, &tx.Data)
}

func rlpHash(x interface{}) (h ethcmn.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
