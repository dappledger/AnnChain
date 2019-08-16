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
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/dappledger/AnnChain/eth/common"
	etypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
)

const (
	txCount = 10
)

var (
	EthSigner = etypes.HomesteadSigner{}
)

func TestExe(t *testing.T) {
	validateRoutineCount = 8

	fmt.Println("CPU:", validateRoutineCount)
	fmt.Println("making random txs...")
	txs := randomTxs(txCount)

	fmt.Println("begin test...")

	err := RunCPUSerialVerifyTest(txs)
	if err != nil {
		t.Error(err)
	}

	err = RunCPUParallelVerifyTest(txs)
	if err != nil {
		t.Error(err)
	}

	err = RunGPUVerifyTest(txs)
	if err != nil {
		t.Error(err)
	}

	txs10 := randomTxs(10)
	err = RunCPUSerialVerifyFailTest(txs10, t)
	if err != nil {
		t.Error(err)
	}

	err = RunCPUParallelFailVerifyTest(txs10, t)
	if err != nil {
		t.Error(err)
	}
}

func beginTestFunc() (ExecFunc, EndExecFunc) {
	exec := func(index int, raw []byte, tx *etypes.Transaction) error {
		if tx == nil {
			panicErr(errors.New("tx is nil"))
		}
		_, err := etypes.Sender(EthSigner, tx)
		panicErr(err)
		return nil
	}
	end := func(bs []byte, err error) bool {
		if err != nil {
			fmt.Println(err)
			return false
		}
		return true
	}
	return exec, end
}

func beginTestFailFunc(t *testing.T) func() (ExecFunc, EndExecFunc) {
	return func() (ExecFunc, EndExecFunc) {
		var preindex, count int
		exec := func(index int, raw []byte, tx *etypes.Transaction) error {
			if tx == nil {
				panicErr(errors.New("tx is nil"))
			}
			if index == preindex {
				count++
				if count > 5 {
					if count == 7 {
						return errors.New("make [fake] serial error")
					}
					if count > 7 {
						return errors.New("index err after serial error")
					}
				}
			}
			preindex = index
			return nil
		}
		end := func(bs []byte, err error) bool {
			if err != nil {
				if !strings.Contains(err.Error(), "[fake]") {
					t.Error("err", err)
				}
				fmt.Println(err)
				return false
			}
			return true
		}
		return exec, end
	}
}

func RunCPUSerialVerifyTest(txs gtypes.Txs) error {
	begin := time.Now()

	exeWithCPUSerialVeirfy(txs, beginTestFunc)

	fmt.Println("serial   use time", time.Now().Sub(begin).Seconds())

	return nil
}

func RunCPUSerialVerifyFailTest(txs gtypes.Txs, t *testing.T) error {
	begin := time.Now()

	exeWithCPUSerialVeirfy(txs, beginTestFailFunc(t))

	fmt.Println("serial failed  use time", time.Now().Sub(begin).Seconds())

	return nil
}

func RunCPUParallelVerifyTest(txs gtypes.Txs) error {
	begin := time.Now()

	err := exeWithCPUParallelVeirfy(EthSigner, txs, nil, beginTestFunc)
	if err != nil {
		return err
	}

	fmt.Println("parallel use time", time.Now().Sub(begin).Seconds())

	return nil
}

func RunCPUParallelFailVerifyTest(txs gtypes.Txs, t *testing.T) error {
	begin := time.Now()

	err := exeWithCPUParallelVeirfy(EthSigner, txs, nil, beginTestFailFunc(t))
	if err != nil {
		return err
	}

	fmt.Println("parallel failed use time", time.Now().Sub(begin).Seconds())

	return nil
}

func RunGPUVerifyTest(txs gtypes.Txs) error {
	begin := time.Now()

	// exeWithGPUVeirfy(EthSigner, txs, nil, func(index int, raw []byte, tx *etypes.Transaction, abi []byte) {
	// }, func(raw []byte, err error) {
	// 	fmt.Println(err)
	// })

	fmt.Println("gpu     use time", time.Now().Sub(begin).Seconds())

	return nil
}

func makeTxsBys(i int, privkey string) []byte {
	tx := etypes.NewTransaction(uint64(i), common.Address{}, big.NewInt(0), 90000000000, big.NewInt(0), []byte("dsfsdflsjflsajfldsjflasjfljflsjflksjflkjioejfoijoijelkfno34534n5,34n5k34n5,34n,mn,ren,mn43kn,mrne,mrn,rwnfrewrne4j443i534h543tnkjrenknd"))
	key, err := crypto.HexToECDSA(privkey)
	panicErr(err)
	sig, err := crypto.Sign(tx.Hash().Bytes(), key)
	panicErr(err)
	sigTx, err := tx.WithSignature(EthSigner, sig)
	panicErr(err)
	b, err := rlp.EncodeToBytes(sigTx)
	panicErr(err)
	return b
}

func randomTxs(count int) (txs gtypes.Txs) {
	privkey := "7d73c3dafd3c0215b8526b26f8dbdb93242fc7dcfbdfa1000d93436d577c3b94"
	for i := 0; i < count; i++ {
		var b []byte
		b = makeTxsBys(i, privkey)

		txs = append(txs, b)
	}

	return
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func TestNothing(t *testing.T) {}
