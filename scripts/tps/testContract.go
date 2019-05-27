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

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
)

type res struct {
	id   int
	left int
	cost int64
}

var (
	threadCount   = 10
	sendPerThread = 100000000000000
	commit        = false
	tps           = 10
)

var (
	resq   = make(chan res, 1024)
	ops    int64
	resarr []int
)

func initStaVars() {
	resarr = make([]int, threadCount)
}

func randPrivkey() string {
	return fmt.Sprintf("%06xdafd3c0215b8526b26f8dbdb93242fc7dcfbdfa1000d93436d57%06x", rand.Int()%1000000, rand.Int()%1000000)
}

func testContractMultiCall() {
	assertContractExist(nil)
	rand.Seed(time.Now().UnixNano())

	fmt.Println("ThreadCount:", threadCount, "TPSPerThread:", tps)

	var wg sync.WaitGroup

	endFunc := resPrintRoutine()

	begin := time.Now().UnixNano()
	defer endFunc(time.Now().UnixNano() - begin)

	for i := 0; i < threadCount-1; i++ {
		go testContract(&wg, i, randPrivkey())
	}

	testContract(&wg, threadCount-1, randPrivkey()) // use to block routine

	wg.Wait()
}

func testContract(w *sync.WaitGroup, id int, privkey string) {
	log.DumpStack()
	// var err error
	if w != nil {
		w.Add(1)
	}

	if privkey == "" {
		privkey = defaultPrivKey
	}
	client := cl.NewClientJSONRPC(rpcTarget)

	callFunc := "transfer"
	args := []interface{}{"0fb229fa8e58308a0e1fc7c8a43a4f81fe7f43b8", 1}
	pk, err := crypto.ToECDSA(common.Hex2Bytes(privkey))
	panicErr(err)

	caller := crypto.PubkeyToAddress(pk.PublicKey)

	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	sleep := 1000 / tps
	for i := 0; i < sendPerThread; i++ {
		// nonce, err := getNonce(client, caller.Hex())
		// panicErr(err)

		// nonce := uint64(time.Now().UnixNano())
		start := time.Now().UnixNano()
		err := executeContract(client, privkey, defaultContractAddr, defaultAbis, callFunc, args, nonce, commit)
		panicErr(err)
		end := time.Now().UnixNano()

		// fmt.Printf("%d: %d\n", id, sendPerThread-i)
		resq <- res{id, sendPerThread - i, end - start}
		if i+1 == sendPerThread {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(sleep))

		nonce++
	}

	if w != nil {
		w.Done()
	}
}

func resPrintRoutine() func(int64) {
	var count, costsum, countsum int64
	go func() {
		timet := time.NewTicker(time.Second)
		for {
			select {
			case r := <-resq:
				resarr[r.id] = r.left

				func() {
					s := fmt.Sprintf("[%4f s/op] ", float32(r.cost)/float32(time.Second))
					//for i := 0; i < threadCount; i++ {
					//	s += fmt.Sprintf("[%02d:%4d,cost:%d]", i+1, resarr[i], r.cost)
					//}
					fmt.Println(s)
				}()

				count++
				countsum++
				costsum += r.cost
			case <-timet.C:
				ops = count
				count = 0
			}
		}
	}()
	return func(costall int64) {
		// maybe concurent read
		if countsum == 0 {
			fmt.Println("=====nonce op======")
			return
		}
		costsum_t := atomic.LoadInt64(&costsum)
		count_t := atomic.LoadInt64(&countsum)
		perTx := float32(costsum_t) / float32(count_t) / float32(time.Second)
		fmt.Println("====total:", count_t, ",per tx:", perTx, "s,per second:", float32(count_t)/float32(costall)*float32(time.Second), "======")
	}
}

func testCreateContract() {
	privkey := defaultPrivKey
	client := cl.NewClientJSONRPC(rpcTarget)

	pk, err := crypto.ToECDSA(common.Hex2Bytes(privkey))
	panicErr(err)

	caller := crypto.PubkeyToAddress(pk.PublicKey)
	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	fmt.Println("nonce", nonce)

	caddr, err := createContract(client, privkey, defaultBytecode, nonce)
	panicErr(err)

	time.Sleep(time.Second * 3)

	exist := existContract(client, defaultPrivKey, caddr, defaultBytecode)
	fmt.Println("Contract exist:", exist)
}

func testExistContract() {
	client := cl.NewClientJSONRPC(rpcTarget)

	exist := existContract(client, defaultPrivKey, defaultContractAddr, defaultBytecode)
	fmt.Println("Contract exist:", exist)
}

func testReadContract() {
	var err error
	privkey := defaultPrivKey
	client := cl.NewClientJSONRPC(rpcTarget)

	assertContractExist(client)

	callFunc := "get"
	args := []interface{}{}
	pk, err := crypto.ToECDSA(common.Hex2Bytes(privkey))
	panicErr(err)

	caller := crypto.PubkeyToAddress(pk.PublicKey)
	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	// fmt.Println("nonce", nonce)
	err = readContract(client, privkey, defaultContractAddr, defaultAbis, callFunc, args, nonce)
	panicErr(err)
}

func testCallContract() {
	assertContractExist(nil)

	fmt.Println("ThreadCount:", threadCount, "SendPerThread:", sendPerThread)

	var wg sync.WaitGroup

	endFunc := resPrintRoutine()

	begin := time.Now().UnixNano()
	defer endFunc(time.Now().UnixNano() - begin)

	for i := 0; i < threadCount-1; i++ {
		go contractTest(&wg, i, fmt.Sprintf("%06xdafd3c0215b8526b26f8dbdb93242fc7dcfbdfa1000d93436d57%06x", rand.Int()%1000000, rand.Int()%1000000))
		// go testContract(&wg,i, "")
	}

	contractTest(&wg, threadCount-1, "") // use to block routine

	wg.Wait()
}

func contractTest(w *sync.WaitGroup, id int, privkey string) {
	log.DumpStack()
	// var err error
	if w != nil {
		w.Add(1)
	}

	if privkey == "" {
		privkey = defaultPrivKey
	}
	client := cl.NewClientJSONRPC(rpcTarget)

	callFunc := "transfer"
	args := []interface{}{"0fb229fa8e58308a0e1fc7c8a43a4f81fe7f43b8", 0}
	pk, err := crypto.ToECDSA(common.Hex2Bytes(privkey))
	panicErr(err)

	caller := crypto.PubkeyToAddress(pk.PublicKey)

	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	sleep := 1000 / tps
	for i := 0; i < sendPerThread; i++ {
		// nonce, err := getNonce(client, caller.Hex())
		// panicErr(err)

		// nonce := uint64(time.Now().UnixNano())
		start := time.Now().UnixNano()
		err := contractExecute(privkey, defaultContractAddr, defaultAbis, callFunc, args, nonce, commit)
		panicErr(err)
		end := time.Now().UnixNano()

		// fmt.Printf("%d: %d\n", id, sendPerThread-i)
		resq <- res{id, sendPerThread - i, end - start}
		if i+1 == sendPerThread {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(sleep))

		nonce++
	}

	if w != nil {
		w.Done()
	}
}
