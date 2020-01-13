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

	"github.com/dappledger/AnnChain/gemmill/modules/go-clist"
)

type TxPool interface {
	Lock()
	Unlock()
	Reap(count int) []Tx
	ReceiveTx(tx Tx) error
	Update(height int64, txs []Tx)
	Size() int
	TxsFrontWait() *clist.CElement
	Flush()
	RegisterFilter(filter IFilter)
	GetPendingMaxNonce([]byte) (uint64, error)
}

// A transaction that successfully ran
type TxInPool struct {
	Counter int64 // a simple incrementing counter
	Height  int64 // height that this tx had been validated in
	Tx      Tx
}

func (memTx *TxInPool) GetHeight() int64 {
	return atomic.LoadInt64(&memTx.Height)
}

// anyone impplements IFilter can be register as a tx filter
type IFilter interface {
	CheckTx(Tx) (bool, error)
}

type TxpoolFilter struct {
	cb func([]byte) (bool, error)
}

func (m TxpoolFilter) CheckTx(tx Tx) (bool, error) {
	return m.cb(tx)
}

func NewTxpoolFilter(f func([]byte) (bool, error)) TxpoolFilter {
	return TxpoolFilter{cb: f}
}
