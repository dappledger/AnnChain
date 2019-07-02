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
