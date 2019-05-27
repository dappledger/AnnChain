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
	"time"

	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
)

func testTxCall() {
	fmt.Println("ThreadCount:", threadCount, "SendPerThread:", sendPerThread)

	var wg sync.WaitGroup

	endFunc := resPrintRoutine()

	begin := time.Now().UnixNano()
	defer endFunc(time.Now().UnixNano() - begin)

	for i := 0; i < threadCount-1; i++ {
		go testTx(&wg, i)
	}

	testTx(&wg, threadCount-1) // use to block routine

	wg.Wait()
}

func testTx(w *sync.WaitGroup, id int) {
	log.DumpStack()
	// var err error
	if w != nil {
		w.Add(1)
	}

	key, _ := crypto.GenerateKey()

	toKey, _ := crypto.GenerateKey()
	toAddr := crypto.PubkeyToAddress(toKey.PublicKey)

	num := rand.Intn(255)
	payload := fmt.Sprint(num)
	data := []byte(payload)

	client := cl.NewClientJSONRPC(rpcTarget)

	caller := crypto.PubkeyToAddress(key.PublicKey)

	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	sleep := 1000 / tps
	for i := 0; i < sendPerThread; i++ {
		start := time.Now().UnixNano()
		err := sendTx(key, nonce, toAddr, data, commit)
		panicErr(err)
		end := time.Now().UnixNano()

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
