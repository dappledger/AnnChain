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
	"sync"
	"time"

	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
)

const (
	t = 200
	c = 100
)

var (
	wg  sync.WaitGroup
	arr = make([][]uint64, t)
)

func testGetNonce() {
	begin := time.Now()

	for i := 0; i < t-1; i++ {
		go getNonceRoutine(i)
	}

	getNonceRoutine(t - 1) // use to block routine

	wg.Wait()

	var count, zcount int
	m := make(map[uint64]struct{}, t*c)
	for i := 0; i < t; i++ {
		for j := 0; j < c; j++ {
			if _, exist := m[arr[i][j]]; exist && arr[i][j] != 0 {
				fmt.Println("duplicated:", arr[i][j])
				count++
			} else if arr[i][j] == 0 {
				zcount++
			}
			m[arr[i][j]] = struct{}{}
		}
	}

	end := time.Now()

	fmt.Println("totally:", t*c, " duplicated:", count, "zero:", zcount)
	fmt.Println(int64(time.Second)/(end.Sub(begin).Nanoseconds()/t/c), "op/s")
}

func getNonceRoutine(id int) {
	client := cl.NewClientJSONRPC(rpcTarget)

	wg.Add(1)
	arr[id] = make([]uint64, c)

	for i := 0; i < c; i++ {
		nonce, err := getNonce(client, "")
		if err != nil {
			fmt.Println(err)
			continue
		}
		if nonce == 0 {
			continue
		}

		arr[id][i] = nonce
		// fmt.Printf("id:%v left:%v\n", id, c-i)
	}
	wg.Done()
}
