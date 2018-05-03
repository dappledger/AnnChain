/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package evm

import (
	"bytes"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/angine/types"
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
	abi      []byte // only used in contract creation tx
	rawbytes types.Tx
	tx       *ethtypes.Transaction
	ready    sync.WaitGroup
	status   int32
	err      error
}

func exeWithCPUParallelVeirfy(signer ethtypes.Signer, txs [][]byte, quit chan struct{},
	whenExec func(index int, raw []byte, tx *ethtypes.Transaction, abi []byte), whenError func(bs []byte, err error)) error {
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

	appTxQ := make([]appTx, len(txs))
	go makeTxQueue(txs, appTxQ, &exit)

	if validateRoutineCount < 1 || validateRoutineCount > 16 {
		validateRoutineCount = 8
	}
	for i := 0; i < validateRoutineCount; i++ {
		go validateRoutine(signer, appTxQ, &exit)
	}

	size := len(txs)
	for i := 0; i < size; i++ {
	INNERFOR:
		for {
			q := atomic.LoadInt32(&exit)
			if q == 1 {
				return errQuitExecute
			}

			status := atomic.LoadInt32(&appTxQ[i].status)
			switch status {
			case appTxStatusChecked:
				whenExec(i, appTxQ[i].rawbytes, appTxQ[i].tx, appTxQ[i].abi)
				break INNERFOR
			case appTxStatusFailed:
				whenError(appTxQ[i].rawbytes, appTxQ[i].err)
				break INNERFOR
			default:
				appTxQ[i].ready.Wait()
			}
		}
	}

	return nil
}

func makeTxQueue(txs [][]byte, apptxQ []appTx, exit *int32) {
	for i, raw := range txs {
		q := atomic.LoadInt32(exit)
		if q == 1 {
			return
		}

		apptxQ[i].rawbytes = raw
		apptxQ[i].ready.Add(1)

		// decode bytes
		var txBytes []byte
		txType := raw[:4]
		switch {
		case bytes.Equal(txType, EVMTxTag):
			txBytes = raw[4:]
		case bytes.Equal(txType, EVMCreateContractTxTag):
			if txCreate, err := DecodeCreateContract(types.UnwrapTx(raw)); err != nil {
				apptxQ[i].err = err
			} else {
				txBytes = txCreate.EthTx
				apptxQ[i].abi = txCreate.EthAbi
			}
		}
		if len(txBytes) > 0 {
			apptxQ[i].tx = new(ethtypes.Transaction)
			if err := rlp.DecodeBytes(txBytes, apptxQ[i].tx); err != nil {
				apptxQ[i].err = err
			}
		}

		atomic.StoreInt32(&apptxQ[i].status, appTxStatusInit)
	}
}

func validateRoutine(signer ethtypes.Signer, appTxQ []appTx, exit *int32) {
	size := len(appTxQ)
	for i := 0; i < size; i++ {
	INNERFOR:
		for {
			q := atomic.LoadInt32(exit)
			if q == 1 {
				return
			}

			status := atomic.LoadInt32(&appTxQ[i].status)
			switch status {
			case appTxStatusNone: // we can do nothing but waiting
				time.Sleep(time.Microsecond)
			case appTxStatusInit: // try validating
				tryValidate(signer, &appTxQ[i])
			default: // move to next
				break INNERFOR
			}
		}
	}
}

func tryValidate(signer ethtypes.Signer, tx *appTx) {
	swapped := atomic.CompareAndSwapInt32(&tx.status, appTxStatusInit, appTxStatusChecking)
	if !swapped {
		return
	}

	defer tx.ready.Done()

	//  when this tx exists errors
	if tx.err != nil {
		atomic.StoreInt32(&tx.status, appTxStatusFailed)
		return
	}

	// when this tx is not a evm-like tx
	if tx.tx == nil {
		atomic.StoreInt32(&tx.status, appTxStatusChecked)
		return
	}

	_, err := ethtypes.Sender(signer, tx.tx)
	if err != nil {
		atomic.StoreInt32(&tx.status, appTxStatusFailed)
		tx.err = err
		return
	}

	atomic.StoreInt32(&tx.status, appTxStatusChecked)
	return
}
