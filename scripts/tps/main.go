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
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	log "github.com/dappledger/AnnChain/gemmill/modules/go-log"
)

var (
	gasLimit = uint64(90000000000)

	logger *zap.Logger
)

var (
	//rpcTarget = "tcp://172.28.162.134:2657"
	rpcTarget           = "tcp://0.0.0.0:46657"
	defaultAbis         = "[{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"
	defaultBytecode     = bc
	defaultContractAddr = "0x0f3200148775219ead5ba8d2f2bf9f0ae1f76ea9"
	defaultPrivKey      = "7d73c3dafd3c0215b8526b26f8dbdb93242fc7dcfbdfa1000d93436d577c3b94"
	defaultAddress      = "aafbb065a30878528b214807863541fb2d15c555"
	defaultChainID      = ""
)

func main() {
	if len(os.Args) < 2 {
		panic("usage: test op [commit]")
	}
	logger, _ = log.Initialize("", "")
	prepare()
	start := time.Now()

	op := os.Args[1]
	switch op {
	case "create":
		testCreateContract()
	case "read":
		testReadContract()
	case "call":
		testContractMultiCall()
	case "exist":
		testExistContract()
	case "apiCall":
		testCallContract()
	case "tx":
		testTxCall()
	default:
		panic("unsupport op:" + op)
	}

	end := time.Now()
	fmt.Println("time used:", end.Sub(start).Seconds(), "s")
}

func prepare() {
	rpct := os.Getenv("rpc")
	if rpct != "" {
		rpcTarget = rpct
	}

	tc := os.Getenv("tc")
	if tc != "" {
		tci, _ := strconv.Atoi(tc)
		threadCount = tci
	}

	sc := os.Getenv("sc")
	if sc != "" {
		sci, _ := strconv.Atoi(sc)
		sendPerThread = sci
	}

	t := os.Getenv("tps")
	if t != "" {
		ti, _ := strconv.Atoi(t)
		tps = ti
	}

	if len(os.Args) > 2 {
		commit = os.Args[2] == "commit"
	}
	initStaVars()
}
