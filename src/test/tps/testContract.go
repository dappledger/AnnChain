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


package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/crypto"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
)

type res struct {
	id   int
	left int
}

var (
	threadCount   = 10
	sendPerThread = 10000
)

var (
	resq   = make(chan res, 1024)
	resarr = make([]int, threadCount)

	ops int
)

func testContractMultiCall() {
	assertContractExist(nil)

	fmt.Println("ThreadCount:", threadCount, "SendPerThread:", sendPerThread)

	var wg sync.WaitGroup

	go resPrintRoutine()

	for i := 0; i < threadCount-1; i++ {
		go testContract(&wg, i, fmt.Sprintf("7d73c3dafd3c0215b8526b26f8dbdb93242fc7dcfbdfa1000d93436d577c30%02d", i))
		// go testContract(&wg,i, "")
	}

	testContract(&wg, threadCount-1, "") // use to block routine

	wg.Wait()
}

func testContract(w *sync.WaitGroup, id int, privkey string) {
	// var err error
	if w != nil {
		w.Add(1)
	}

	if privkey == "" {
		privkey = defaultPrivKey
	}
	client := cl.NewClientJSONRPC(logger, rpcTarget)

	callFunc := "add"
	args := []interface{}{}
	pk := crypto.ToECDSA(common.Hex2Bytes(privkey))
	caller := crypto.PubkeyToAddress(pk.PublicKey)

	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	for i := 0; i < sendPerThread; i++ {
		// nonce, err := getNonce(client, caller.Hex())
		// panicErr(err)

		// nonce := uint64(time.Now().UnixNano())
		err := executeContract(client, privkey, defaultContractAddr, defaultAbis, callFunc, args, nonce)
		panicErr(err)

		// fmt.Printf("%d: %d\n", id, sendPerThread-i)
		resq <- res{id, sendPerThread - i}
		time.Sleep(time.Millisecond * 1)

		nonce++
	}

	if w != nil {
		w.Done()
	}
}

func resPrintRoutine() {
	count := 0
	timet := time.NewTicker(time.Second)
	for {
		select {
		case r := <-resq:
			resarr[r.id] = r.left

			func() {
				s := fmt.Sprintf("[%4d op/s] ", ops)
				for i := 0; i < threadCount; i++ {
					s += fmt.Sprintf("[%02d:%4d]", i+1, resarr[i])
				}
				fmt.Println(s)
			}()

			count++
		case <-timet.C:
			ops = count
			count = 0
		}
	}
}

func testCreateContract() {
	privkey := defaultPrivKey
	client := cl.NewClientJSONRPC(logger, rpcTarget)

	pk := crypto.ToECDSA(common.Hex2Bytes(privkey))
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
	client := cl.NewClientJSONRPC(logger, rpcTarget)

	exist := existContract(client, defaultPrivKey, defaultContractAddr, defaultBytecode)
	fmt.Println("Contract exist:", exist)
}

func testReadContract() {
	var err error
	privkey := defaultPrivKey
	client := cl.NewClientJSONRPC(logger, rpcTarget)

	assertContractExist(client)

	callFunc := "get"
	args := []interface{}{}
	pk := crypto.ToECDSA(common.Hex2Bytes(privkey))
	caller := crypto.PubkeyToAddress(pk.PublicKey)
	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)

	// fmt.Println("nonce", nonce)
	err = readContract(client, privkey, defaultContractAddr, defaultAbis, callFunc, args, nonce)
	panicErr(err)
}
