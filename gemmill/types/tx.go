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
	"unsafe"

	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/gemmill/go-hash"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/modules/go-merkle"
	"github.com/dappledger/AnnChain/gemmill/utils"
)

type Tx []byte

// NOTE: this is the hash of the go-wire encoded Tx.
// Tx has no types at this level, so just length-prefixed.
// Alternatively, it may make sense to add types here and let
// []byte be type 0x1 so we can have versioned txs if need be in the future.

// ethereum transaction hash
func (tx Tx) Hash() []byte {
	if IsSerialTx(tx) {
		buf := new(bytes.Buffer)
		tx.Deal(func(t Tx) error {
			buf.Write(t.Hash())
			return nil
		})

		return buf.Bytes()

	} else {
		return hash.Keccak256OrSM3Func(tx)
	}
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

func (tx Tx) Size() int {
	if !IsSerialTx(tx) {
		return 1
	}
	buf := bytes.NewBuffer(SerialTxBody(tx))
	var count uint32
	var err error
	if err = utils.BinRead(buf, &count); err != nil {
		log.Error("tx wrap error", zap.Error(err))
		return 0
	}
	return int(count)
}

func (tx Tx) Deal(exec func(Tx) error) error {
	if !IsSerialTx(tx) {
		return exec(tx)
	}
	buf := bytes.NewBuffer(SerialTxBody(tx))
	var count uint32
	var err error
	if err = utils.BinRead(buf, &count); err != nil {
		return err
	}
	for i := uint32(0); i < count; i++ {
		var temTx []byte
		temTx, err = utils.ReadBytes(buf)
		if err != nil {
			return err
		}
		if err = exec(temTx); err != nil {
			return err
		}
	}
	return nil
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

func (txs Txs) ToBytes() [][]byte {
	return *((*[][]byte)(unsafe.Pointer(&txs)))
}
