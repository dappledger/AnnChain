// Copyright © 2017 ZhongAn Technology
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

package evm

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/eth/common"
	etypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/gemmill/modules/go-clist"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/types"
)

const (
	txEvictInterval = 1 * time.Minute
	waitingLifeTime = 10 * time.Minute
)

var (
	errTxExist                  = errors.New("tx already exist in cache")
	errTxPoolWaitingQueueIsFull = errors.New("evm tx pool waiting queue is full")
)

type ethTxPool struct {
	pending         map[common.Address]*txSortedMap // All currently processable transactions
	waiting         map[common.Address]*txSortedMap // Queued but non-processable transactions
	waitingBeats    map[common.Address]time.Time    // Last heartbeat from each known address
	broadcastQueue  *clist.CList                    // list of txs to broadcast
	all             map[common.Hash]types.Tx        // tx cache for lookup
	extTxs          *clist.CList                    // extra transcations except Ethereum Transaction, eg. adminOP
	mtx             sync.Mutex
	app             *EVMApp
	waitingLifeTime time.Duration // Maximum amount of time non-executable transaction are queued
	waitingLimit    int           // waiting queue size limit
	pendingLimit    int           // pending queue size limit
	height          int64
	filter          []types.IFilter
}

func NewEthTxPool(app *EVMApp, conf *viper.Viper) *ethTxPool {
	return &ethTxPool{
		all:             make(map[common.Hash]types.Tx),
		waiting:         make(map[common.Address]*txSortedMap),
		waitingBeats:    make(map[common.Address]time.Time),
		pending:         make(map[common.Address]*txSortedMap),
		extTxs:          clist.New(),
		broadcastQueue:  clist.New(),
		waitingLimit:    conf.GetInt("block_size") * 10,
		pendingLimit:    conf.GetInt("block_size") * 10,
		waitingLifeTime: waitingLifeTime,
		app:             app,
	}
}

func (tp *ethTxPool) Start(height int64) {
	tp.setHeight(height)
	go tp.loop()
}

func (tp *ethTxPool) loop() {
	evict := time.NewTicker(txEvictInterval)
	defer evict.Stop()

	for {
		select {
		case <-evict.C:
			tp.Lock()
			for addr := range tp.waitingBeats {
				if time.Since(tp.waitingBeats[addr]) > tp.waitingLifeTime {
					if tp.waiting[addr].Get(tp.safeGetNonce(addr)) != nil {
						continue
					}

					// waiting queue of account does not have the pending nonce, delete all its waiting tx
					for _, tx := range tp.waiting[addr].Flatten() {
						delete(tp.all, tx.Hash())
					}
					delete(tp.waitingBeats, addr)
					delete(tp.waiting, addr)
				}
			}
			tp.Unlock()
		}
	}
}

func (tp *ethTxPool) Lock() {
	tp.mtx.Lock()
}

func (tp *ethTxPool) Unlock() {
	tp.mtx.Unlock()
}

func (tp *ethTxPool) setHeight(height int64) {
	atomic.StoreInt64(&tp.height, height)
}

func (tp *ethTxPool) RegisterFilter(filter types.IFilter) {
	tp.filter = append(tp.filter, filter)
}

// reap txpool txs to make block
func (tp *ethTxPool) Reap(maxTxs int) []types.Tx {
	tp.Lock()
	defer tp.Unlock()
	if maxTxs == 0 {
		return []types.Tx{}
	} else if maxTxs < 0 {
		maxTxs = tp.pendingLimit
	}

	allTxs := make([]types.Tx, 0, maxTxs)

	// reap adminOP txs
	if extTxs := tp.reapAdminOP(maxTxs); len(extTxs) == maxTxs {
		return extTxs
	} else {
		allTxs = append(allTxs, extTxs...)
	}

OUTLOOP: // reap normal txs
	for _, accountTxs := range tp.pending {
		for _, tx := range accountTxs.Flatten() {
			txBytes, exist := tp.all[tx.Hash()]
			if !exist {
				// cache miss
				txBytes, _ = rlp.EncodeToBytes(tx)
			}
			allTxs = append(allTxs, txBytes)
			if len(allTxs) == maxTxs {
				break OUTLOOP
			}
		}
	}
	log.Debug("reap return txs", zap.Int("count", len(allTxs)))
	return allTxs
}

// Try a new transaction in the tx pool. Tx may come from local rpc or remote node broadcast.
func (tp *ethTxPool) ReceiveTx(rawTx types.Tx) error {
	if types.IsAdminOP(rawTx) {
		return tp.handleAdminOP(rawTx)
	}

	tx := &etypes.Transaction{}
	if err := rlp.DecodeBytes(rawTx, tx); err != nil {
		return err
	}
	if err := tp.CheckAndAdd(tx, rawTx); err != nil {
		return err
	}
	tp.Lock()
	tp.broadcastNewTx(rawTx)
	tp.Unlock()
	return nil
}

// receive and handle adminOP txs
func (tp *ethTxPool) handleAdminOP(tx types.Tx) error {
	tp.Lock()
	defer tp.Unlock()

	for e := tp.extTxs.Front(); e != nil; e = e.Next() {
		extTx := e.Value.(types.Tx)
		if bytes.Equal(tx, extTx) {
			return errTxExist
		}
	}

	// remove oldest extTx if reach size limit
	if tp.extTxs.Len() >= tp.pendingLimit {
		front := tp.extTxs.Front()
		tp.extTxs.Remove(front)
		front.DetachPrev()
	}
	tp.extTxs.PushBack(tx)
	tp.broadcastNewTx(tx)
	return nil
}

// reap extTxs to make block
func (tp *ethTxPool) reapAdminOP(maxTxs int) []types.Tx {
	if tp.extTxs.Len() == 0 {
		return []types.Tx{}
	}
	txs := make([]types.Tx, 0, maxTxs)
	for e := tp.extTxs.Front(); e != nil && len(txs) < maxTxs; e = e.Next() {
		extTx := e.Value.(types.Tx)
		txs = append(txs, extTx)
	}
	return txs
}

// remove extTxs already involved in block
func (tp *ethTxPool) refreshAdminOP(txsMap map[string]struct{}) {
	if tp.extTxs.Len() == 0 {
		return
	}
	txsLen := tp.extTxs.Len()
	index := 0
	for e := tp.extTxs.Front(); e != nil && index < txsLen; e = e.Next() {
		index++
		extTx := e.Value.(types.Tx)
		// Remove the tx if it's already in a block
		if _, ok := txsMap[string(extTx)]; ok {
			tp.extTxs.Remove(e)
			e.DetachPrev()
		}
	}
}

// get account nonce from app.state
func (tp *ethTxPool) safeGetNonce(addr common.Address) uint64 {
	tp.app.stateMtx.Lock()
	nonce := tp.app.state.GetNonce(addr)
	tp.app.stateMtx.Unlock()
	return nonce
}

func (tp *ethTxPool) CheckAndAdd(tx *etypes.Transaction, rawTx types.Tx) error {
	tp.Lock()
	defer tp.Unlock()
	if _, exist := tp.all[tx.Hash()]; exist {
		return errTxExist
	}

	from, _ := etypes.Sender(tp.app.Signer, tx)
	currentNonce := tp.safeGetNonce(from)
	if currentNonce > tx.Nonce() {
		return fmt.Errorf("nonce(%d) different with getNonce(%d)", tx.Nonce(), currentNonce)
	}

	if err := tp.addWaiting(tx, from); err != nil {
		return err
	}
	tp.all[tx.Hash()] = rawTx
	if currentNonce == tx.Nonce() {
		tp.promoteExecutables([]common.Address{from})
	}
	return nil
}

// Tell tx pool that these txs were committed.
func (tp *ethTxPool) Update(height int64, txs []types.Tx) {
	log.Debug("update tx pool txs", zap.Int64("height", height))
	tp.Lock()
	defer tp.Unlock()
	tp.setHeight(height)
	if len(txs) == 0 {
		return
	}

	txsMap := make(map[string]struct{})
	for _, tx := range txs {
		txsMap[string(tx)] = struct{}{}
	}
	tp.refreshBroadcastList(txsMap)
	tp.refreshAdminOP(txsMap)

	return
}

// update pool txs after evm state updated
func (tp *ethTxPool) updateToState() {
	log.Debug("update txpool to evm app state")
	tp.Lock()
	defer tp.Unlock()
	// validate the pool of pending transactions, this will remove
	// any transactions that have been included in the block or
	// have been invalidated because of another transaction (e.g.
	// higher gas price)
	tp.demoteUnexecutables()
	// Check the queue and move transactions over to the pending if possible
	// or remove those that have become invalid
	tp.promoteExecutables(nil)
}

func (tp *ethTxPool) Size() int {
	tp.Lock()
	defer tp.Unlock()
	return tp.extTxs.Len() + len(tp.all)
}

// blocking get first element of broadcast queue
func (tp *ethTxPool) TxsFrontWait() *clist.CElement {
	return tp.broadcastQueue.FrontWait()
}

// Remove all transactions from tx and cache
func (tp *ethTxPool) Flush() {
	tp.Lock()
	tp.waiting = make(map[common.Address]*txSortedMap)
	tp.pending = make(map[common.Address]*txSortedMap)
	tp.waitingBeats = make(map[common.Address]time.Time)
	tp.all = make(map[common.Hash]types.Tx)
	tp.broadcastQueue = clist.New()
	tp.extTxs = clist.New()
	tp.Unlock()
}

// promoteExecutables moves transactions that have become processable from the
// waiting queue to the set of pending transactions. During this process, all
// invalidated transactions (low nonce, low balance) are deleted.
func (tp *ethTxPool) promoteExecutables(addrs []common.Address) {
	// If the pending limit is overflown, start equalizing allowances
	pendingTxCount := 0
	for _, accountTxs := range tp.pending {
		pendingTxCount += accountTxs.Len()
	}

	if pendingTxCount >= tp.pendingLimit {
		// pending queue is full, do not promote
		return
	}

	// Gather all the accounts potentially needing updates
	if addrs == nil {
		addrs = make([]common.Address, 0, len(tp.waiting))
		for addr := range tp.waiting {
			addrs = append(addrs, addr)
		}
	}

	for _, addr := range addrs {
		if pendingTxCount >= tp.pendingLimit {
			// pending queue is full, do not promote
			return
		}
		nonce := tp.safeGetNonce(addr)
		waiting := tp.waiting[addr]

		// Drop all transactions that are deemed too old (low nonce)
		oldTxs := waiting.Forward(nonce)
		for _, otx := range oldTxs {
			delete(tp.all, otx.Hash())
		}

		// Gather up to N executable transactions and promote them
		txs := waiting.ReadyN(nonce, tp.pendingLimit-pendingTxCount)

		// Delete the entire queue entry if it became empty.
		if waiting.Len() == 0 {
			delete(tp.waiting, addr)
			delete(tp.waitingBeats, addr)
		}

		if len(txs) == 0 {
			continue
		}

		if tp.pending[addr] == nil {
			tp.pending[addr] = newTxSortedMap()
		}
		for _, tx := range txs {
			// pending is not full, add
			if err := tp.pending[addr].Add(tx); err == nil {
				pendingTxCount++
			}
		}
	}
}

// add validates a transaction and inserts it into the waiting queue for
// later pending promotion and execution.
func (tp *ethTxPool) addWaiting(tx *etypes.Transaction, address common.Address) error {
	waitingTxCount := 0
	for _, txs := range tp.waiting {
		waitingTxCount += txs.Len()
	}
	if waitingTxCount >= tp.waitingLimit {
		// waiting queue is full, try replace or return err
		if tp.waiting[address] == nil || !tp.waiting[address].TryReplace(tx) {
			return errTxPoolWaitingQueueIsFull
		}
	} else {
		if tp.waiting[address] == nil {
			tp.waiting[address] = newTxSortedMap()
		}
		err := tp.waiting[address].Add(tx)
		if err != nil {
			return err
		}
	}
	tp.waitingBeats[address] = time.Now()
	return nil
}

// demoteUnexecutables removes invalid and processed transactions from the pools
// executable/pending queue and any subsequent transactions that become unexecutable
// are moved back into the waiting queue.
func (tp *ethTxPool) demoteUnexecutables() {
	for addr, accountTxs := range tp.pending {
		nonce := tp.safeGetNonce(addr)

		// Drop all transactions that are deemed too old (low nonce)
		for _, tx := range accountTxs.Forward(nonce) {
			hash := tx.Hash()
			delete(tp.all, hash)
		}

		if accountTxs.Len() == 0 {
			delete(tp.pending, addr)
		} else if accountTxs.Len() > 0 && accountTxs.Get(nonce) == nil {
			// If there's a gap in front, alert (should never happen) and postpone all transactions
			for _, tx := range accountTxs.items {
				log.Warn("Demoting invalidated transaction", zap.String("hash", tx.Hash().Hex()))
				if err := tp.addWaiting(tx, addr); err != nil {
					// demote pending to waiting failed, waiting queue maybe full, delete tx
					delete(tp.all, tx.Hash())
				}
			}
			// Delete the entire queue entry if it became empty.
			delete(tp.pending, addr)
		}
	}
}

// add one tx to broadcast list
func (tp *ethTxPool) broadcastNewTx(tx types.Tx) {
	if tp.broadcastQueue.Len() >= tp.waitingLimit+tp.pendingLimit {
		e := tp.broadcastQueue.Front()
		tp.broadcastQueue.Remove(e)
		e.DetachPrev()
	}
	memTx := &types.TxInPool{
		Height: atomic.LoadInt64(&tp.height),
		Tx:     tx,
	}
	tp.broadcastQueue.PushBack(memTx)
}

// Remove broadcast list transactions that are already in txs
func (tp *ethTxPool) refreshBroadcastList(txsMap map[string]struct{}) {
	txsLen := tp.broadcastQueue.Len()
	index := 0
	for e := tp.broadcastQueue.Front(); e != nil && index < txsLen; e = e.Next() {
		index++
		memTx := e.Value.(*types.TxInPool)
		// Remove the tx if it's already in a block
		if _, ok := txsMap[string(memTx.Tx)]; ok {
			tp.broadcastQueue.Remove(e)
			e.DetachPrev()
		}
	}
	log.Debug("broadcast list refresh length", zap.Int("before:", txsLen), zap.Int("after:", tp.broadcastQueue.Len()))
}

// print state of tx pool for debug
func (tp *ethTxPool) state() string {
	var (
		waitingCount int
		pendingCount int
		waitingTxs   string
		pendingTxs   string
	)
	for addr, txs := range tp.waiting {
		waitingCount += txs.Len()
		waitingTxs += fmt.Sprintf("address: %s, txs: %v, ", addr.Hex(), txs)
	}

	for addr, txs := range tp.pending {
		pendingCount += txs.Len()
		pendingTxs += fmt.Sprintf("address: %s, txs: %v, ", addr.Hex(), txs)
	}

	state := fmt.Sprintf("Waiting tx size: %v n"+
		"Waiting Txs:%s,"+
		"Pending tx size: %v, "+
		"Pending Txs:%s,"+
		"All tx size: %v "+
		"BrodcastQueue size: %v",
		waitingCount,
		waitingTxs,
		pendingCount,
		pendingTxs,
		len(tp.all),
		tp.broadcastQueue.Len(),
	)
	return state
}
