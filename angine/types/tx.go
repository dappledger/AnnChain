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
	"unsafe"

	"github.com/dappledger/AnnChain/module/lib/go-merkle"
)

type Tx []byte

func IsExtendedTx(tx []byte) bool {
	return IsSpecialOP(tx) || IsSuspectTx(tx) || IsVoteChannel(tx)
}

// NOTE: this is the hash of the encoded Tx.
// Tx has no types at this level, so just length-prefixed.
// Alternatively, it may make sense to add types here and let
// []byte be type 0x1 so we can have versioned txs if need be in the future.
func (tx Tx) Hash() []byte {
	return merkle.SimpleHashFromBinary(tx)
}

type Txs []Tx

func (txs Txs) Hash() []byte {
	// Recursive impl.
	// Copied from go-merkle to avoid allocations
	switch len(txs) {
	case 0:
		return nil
	case 1:
		return txs[0].Hash()
	default:
		left := Txs(txs[:(len(txs)+1)/2]).Hash()
		right := Txs(txs[(len(txs)+1)/2:]).Hash()
		return merkle.SimpleHashFromTwoHashes(left, right)
	}
}

func (txs Txs) ToBytes() [][]byte {
	return *((*[][]byte)(unsafe.Pointer(&txs)))
}

func BytesToTxSlc(txs [][]byte) []Tx {
	return *((*[]Tx)(unsafe.Pointer(&txs)))
}

func WrapTx(prefix []byte, tx []byte) []byte {
	return append(prefix, tx...)
}

func UnwrapTx(tx []byte) []byte {
	if len(tx) > 4 {
		return tx[4:]
	}
	return tx
}
