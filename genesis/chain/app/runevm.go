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

package app

import (
	"math/big"
	"time"

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	ethcore "github.com/dappledger/AnnChain/genesis/eth/core"
	ethstate "github.com/dappledger/AnnChain/genesis/eth/core/state"
	ethtypes "github.com/dappledger/AnnChain/genesis/eth/core/types"
	ethvm "github.com/dappledger/AnnChain/genesis/eth/core/vm"
	ethparams "github.com/dappledger/AnnChain/genesis/eth/params"
)

var (
	chainConfig = &ethparams.ChainConfig{}
	evmConfig   = ethvm.Config{}
	ethSigner   = ethtypes.HomesteadSigner{}
)

func NewContractCreation(nonce uint64, from ethcmn.Address, amount, gasLimit, gasPrice *big.Int, data []byte) *ethtypes.Transaction {
	return ethtypes.NewContractCreation(nonce, from, amount, gasLimit, gasPrice, data)
}

func NewContractTransaction(nonce uint64, from, to ethcmn.Address, amount, gasLimit, gasPrice *big.Int, data []byte) *ethtypes.Transaction {
	return ethtypes.NewTransaction(nonce, from, to, amount, gasLimit, gasPrice, data)
}

func RunEvm(curHeader *ethtypes.Header, state *ethstate.StateDB, tx *ethtypes.Transaction) (receipt *ethtypes.Receipt, gas *big.Int, err error) {

	mLog := ethvm.NewStructLogger(&ethvm.LogConfig{})

	evmConfig.Tracer = mLog

	gp := new(ethcore.GasPool).AddGas(ethcmn.MaxBig) //GasPool 是一个块中执行事务所需的gas消耗。

	receipt, gas, err = ethcore.ApplyTransaction(
		chainConfig,
		nil,
		gp,
		state,
		curHeader,
		tx,
		big.NewInt(0),
		evmConfig)

	return
}

func QueryContractExcute(curHeader *ethtypes.Header, state *ethstate.StateDB, tx *ethtypes.Transaction) (res []byte, gas *big.Int, err error) {

	mLog := ethvm.NewStructLogger(&ethvm.LogConfig{})

	evmConfig.Tracer = mLog

	fakeHeader := &ethtypes.Header{
		ParentHash: ethcmn.HexToHash("0x00"),
		Difficulty: big.NewInt(0),
		GasLimit:   ethcmn.MaxBig,
		Number:     curHeader.Number,
		Time:       big.NewInt(time.Now().Unix()),
	}

	txMsg, _ := tx.AsMessage(ethtypes.HomesteadSigner{})

	envCxt := ethcore.NewEVMContext(txMsg, fakeHeader, nil)

	vmEnv := ethvm.NewEVM(envCxt, state, chainConfig, evmConfig)

	gpl := new(ethcore.GasPool).AddGas(ethcmn.MaxBig)

	res, gas, _, err = ethcore.ApplyMessage(vmEnv, txMsg, gpl)

	return

}
