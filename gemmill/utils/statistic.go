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

package utils

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

/*
*
*简易统计并行/串行代码 在周期(FLUSH_INTERVAL)内执行次数,每次最大/最小执行时间,以及未完成的并发个数
*一个周期后重新统计次数及执行时间,未完成的并发个数保留
*项目需要xlog支持,见xlog-->README.md
*
*=======================Using Example:========================================
*First Register In func init():
*
*const(
*	MEMPOOL_TX_QUEST = iota
*	MEMPOOL_TX_SYNCOTHER
*)
*
*count.data[MEMPOOL_TX_QUEST].Init("mempool_tx_request", time.Second)
*count.data[MEMPOOL_TX_SYNCOTHER].Init("mempool_tx_sync_to_other", time.Second)
*
*--------------------------Caller:--------------------------------------------
*
*func (mem *Mempool) CheckTx(tx types.Tx, cb func(*abci.Response)) (err error) {
*
*	endFunc := utils.CountSlc(utils.MEMPOOL_TX_QUEST).Begin()
*	utils.CountSlc(utils.MEMPOOL_TX_QUEST).Add(1)
*	defer endFunc()
*
*	mem.questMtx.Lock()
*	defer mem.questMtx.Unlock()
*	.......
*}
*
*--------------------------Or:-----------------------------------------------
*
*endFunc := utils.CountSlc(utils.MEMPOOL_TX_SYNCOTHER).Begin()
*utils.CountSlc(utils.MEMPOOL_TX_SYNCOTHER).Add(1)
*
*success := peer.Send(MempoolChannel, struct{ MempoolMessage }{msg})
*
*endFunc()
*
*Call utils.Start()
*
*==========================<xlog dir>/<ann>.dbg=============================================
*
*[D]20170614 17:23:17.654095,statistic.go:125,[statistic],mempool_tx_request,undone:28187
*[D]20170614 17:23:17.654133,statistic.go:125,[statistic],mempool_tx_sync_to_other,count:5,minPass:5121,maxPass:15191,undone:0
*
 */

const (
	EXECUTE_EVM_TX = iota
	FUNC_MAX_NUM
)

const FLUSH_INTERVAL = 1 * time.Second // statistic cycle
var started bool

type StStatistic struct {
	count   int64 // exec count, call function Add()
	maxPass int64 // max passing time
	minPass int64 // min passing time
	total   int64 // all passing time,unprint yet,maybe overflow
	undone  int64 // undone(called begin(), but haven't called end()) number

	name      string
	threshold time.Duration // unused now,use as a refrence number

	mtx sync.Mutex
}

func (ss *StStatistic) Init(name string, threshold time.Duration) {
	if !started {
		return
	}
	ss.name = name
	ss.threshold = threshold
	ss.reset()
}

func (ss *StStatistic) reset() {
	ss.mtx.Lock()
	defer ss.mtx.Unlock()
	ss.count = 0
	ss.maxPass = 0
	ss.minPass = 0
	ss.total = 0
}

func (ss *StStatistic) Add(num int64) {
	if ss == nil || !started {
		return
	}
	atomic.AddInt64(&ss.count, num)
}

// return endFunc
// if dont call endFunc(), goroutine will leak
func (ss *StStatistic) Begin() func() {
	if ss == nil || !started {
		return func() {}
	}
	atomic.AddInt64(&ss.undone, 1)
	begin := time.Now().UnixNano()
	bariar := make(chan int64)
	go func() {
		pass := <-bariar
		//xlog.Dbgf("[count per time],%v,const,%v,count,%v", ss.name, pass, ss.count)
		ss.mtx.Lock()
		defer ss.mtx.Unlock()
		if ss.maxPass < pass {
			ss.maxPass = pass
		}
		if ss.minPass == 0 || ss.minPass > pass {
			ss.minPass = pass
		}
		ss.total += pass
	}()
	return func() {
		bariar <- time.Now().UnixNano() - begin
		atomic.AddInt64(&ss.undone, -1)
	}
}

func (ss *StStatistic) End(endFunc func()) {
	if ss == nil || !started {
		return
	}
	endFunc()
}

func (ss *StStatistic) String() string {
	if ss.count == 0 {
		return fmt.Sprintf("[statistic],%v,undone:%v", ss.name, ss.undone)
	}

	return fmt.Sprintf("[statistic],%v,count:%v,minPass:%v,maxPass:%v,undone:%v", ss.name, ss.count,
		transferInt(ss.minPass), transferInt(ss.maxPass),
		ss.undone)
	return fmt.Sprintf("[statistic],%v,count:%v,minPass:%v,maxPass:%v,undone:%v", ss.name, ss.count,
		transferFloat(ss.minPass), transferFloat(ss.maxPass),
		ss.undone)
}

func transferInt(tm int64) int64 {
	return tm
}

func transferFloat(tm int64) float32 {
	return float32(tm) / float32(time.Millisecond)
}

type CountTime struct {
	data []StStatistic
}

func (ct *CountTime) init(len int) {
	count.data = make([]StStatistic, len)
}

func (ct *CountTime) start() {
	if started {
		return
	}
	started = true
	go ct.logRountine()
}

func (ct *CountTime) logRountine() {
	for _ = range time.NewTicker(FLUSH_INTERVAL).C {
		for i := range ct.data {
			fmt.Println(ct.data[i].String())
			ct.data[i].reset()
		}
		fmt.Println("==========================")
	}
}

func CountSlc(index int) *StStatistic {
	if !started {
		return nil
	}
	if index < 0 || index > len(count.data)-1 {
		fmt.Println("[statistic],not registered yet:", index)
		return &StStatistic{}
	}
	return &count.data[index]
}

var count CountTime

func initStat() {
	count.init(FUNC_MAX_NUM)

	count.data[EXECUTE_EVM_TX].Init("execute_evm_tx", time.Second)
}

func StartStat() {
	initStat()
	count.start()
}
