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
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	etypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
)

var (
	validateRoutineCount = runtime.NumCPU()
	// validateRoutineCount = 1
)

const (
	appTxStatusNone     int32 = 0 // new
	appTxStatusInit     int32 = 1 // decoded from bytes 2 tx
	appTxStatusChecking int32 = 2 // validating by one routine
	appTxStatusChecked  int32 = 3 // validated
	appTxStatusFailed   int32 = 4 // tx is invalid
)

type appTx struct {
	rawbytes gtypes.Tx
	oribys   gtypes.Tx
	tx       *etypes.Transaction
	ready    sync.WaitGroup
	status   int32
	err      error
}

type BeginExecFunc func() (ExecFunc, EndExecFunc)
type ExecFunc func(index int, raw []byte, tx *etypes.Transaction) error
type EndExecFunc func(bs []byte, err error) bool

func exeWithCPUParallelVeirfy(signer etypes.Signer, txs gtypes.Txs,
	quit chan struct{}, beginExec BeginExecFunc) error {
	var exit int32
	go func() {
		if quit == nil {
			return
		}
		select {
		case <-quit:
			atomic.StoreInt32(&exit, 1)
		case <-time.NewTimer(time.Second * 60).C:
			return
		}
	}()

	appTxQ := makeTxQueue(txs)
	go initTxQueue(txs, appTxQ, &exit)

	if validateRoutineCount < 1 || validateRoutineCount > 16 {
		validateRoutineCount = 8
	}
	for i := 0; i < validateRoutineCount; i++ {
		go validateRoutine(signer, appTxQ, &exit)
	}

	size := len(txs)
	for i := 0; i < size; i++ {
		lsize := len(appTxQ[i])
		var oriBytes []byte
		var err error
		exec, end := beginExec()
		for j := 0; j < lsize; j++ {
			pcur := &appTxQ[i][j]
		INNERFOR:
			for {
				q := atomic.LoadInt32(&exit)
				if q == 1 {
					return errQuitExecute
				}

				status := atomic.LoadInt32(&pcur.status)
				switch status {
				case appTxStatusChecked:
					//err = whenExec(i, pcur.rawbytes, pcur.tx)
					pcur.err = exec(i, pcur.rawbytes, pcur.tx)
					break INNERFOR
				case appTxStatusFailed:
					//whenError(pcur.rawbytes, pcur.err)
					break INNERFOR
				default:
					pcur.ready.Wait()
				}
			}
			if pcur.err != nil {
				err = pcur.err
				break
			}
		}
		if lsize > 0 {
			oriBytes = appTxQ[i][0].oribys
		}
		if !end(oriBytes, err) {
			break
		}
	}

	return nil
}

func makeTxQueue(txs gtypes.Txs) [][]appTx {
	q := make([][]appTx, len(txs))
	for i := range txs {
		tx := gtypes.Tx(txs[i])
		l := tx.Size()
		q[i] = make([]appTx, l)
	}
	return q
}

func initTxQueue(txs gtypes.Txs, apptxQ [][]appTx, exit *int32) {
	var j int
	for i := range apptxQ {
		q := atomic.LoadInt32(exit)
		if q == 1 {
			return
		}
		tx := gtypes.Tx(txs[i])
		j = 0
		if err := txQueue(tx, apptxQ, i, j); err != nil {
			log.Error("interpret txs err", zap.Error(err))
			// TODO
		}
	}
}

func txQueue(tptx gtypes.Tx, apptxQ [][]appTx, i, j int) error {
	cur := &apptxQ[i][j]
	cur.rawbytes = tptx
	cur.ready.Add(1)

	// decode bytes
	if len(tptx) > 0 {
		cur.tx = new(etypes.Transaction)
		if err := rlp.DecodeBytes(tptx, cur.tx); err != nil {
			cur.err = err
		}
	}

	atomic.StoreInt32(&cur.status, appTxStatusInit)
	if j == 0 {
		apptxQ[i][j].oribys = tptx
	}
	j++
	return nil
}

func validateRoutine(signer etypes.Signer, appTxQ [][]appTx, exit *int32) {
	size := len(appTxQ)
	for i := 0; i < size; i++ {
		lsize := len(appTxQ[i])
	OUTERFOR:
		for j := 0; j < lsize; j++ {
			pcur := &appTxQ[i][j]
		INNERFOR:
			for {
				q := atomic.LoadInt32(exit)
				if q == 1 {
					return
				}

				status := atomic.LoadInt32(&pcur.status)
				switch status {
				case appTxStatusNone: // we can do nothing but waiting
					time.Sleep(time.Microsecond)
				case appTxStatusInit: // try validating
					if err := tryValidate(signer, pcur); err != nil {
						break OUTERFOR
					}
				default: // move to next
					break INNERFOR
				}
			}
		}
	}
}

func tryValidate(signer etypes.Signer, tx *appTx) error {
	swapped := atomic.CompareAndSwapInt32(&tx.status, appTxStatusInit, appTxStatusChecking)
	if !swapped {
		return nil
	}

	defer tx.ready.Done()

	//  when this tx exists errors
	if tx.err != nil {
		atomic.StoreInt32(&tx.status, appTxStatusFailed)
		return nil
	}

	// when this tx is not a evm-like tx
	if tx.tx == nil {
		atomic.StoreInt32(&tx.status, appTxStatusChecked)
		return nil
	}

	_, err := etypes.Sender(signer, tx.tx)
	if err != nil {
		atomic.StoreInt32(&tx.status, appTxStatusFailed)
		tx.err = err
		return err
	}

	atomic.StoreInt32(&tx.status, appTxStatusChecked)
	return nil
}
