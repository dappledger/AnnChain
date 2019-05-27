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
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/dappledger/AnnChain/cmd/client/commons"
	"github.com/dappledger/AnnChain/eth/accounts/abi"
	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	ac "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
	"github.com/dappledger/AnnChain/gemmill/types"
)

func createContract(client *cl.ClientJSONRPC, privkey, bytecode string, nonce uint64) (string, error) {

	data := common.Hex2Bytes(bytecode)
	tx := ethtypes.NewContractCreation(nonce, big.NewInt(0), gasLimit, big.NewInt(0), data)

	privkey = ac.SanitizeHex(privkey)
	key, err := crypto.HexToECDSA(privkey)
	panicErr(err)
	sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), key)
	panicErr(err)
	sigTx, err := tx.WithSignature(ethSigner, sig)
	panicErr(err)
	b, err := rlp.EncodeToBytes(sigTx)
	panicErr(err)

	// rpcResult := new(types.rpcResult)
	rpcResult := new(types.ResultBroadcastTx)
	if client == nil {
		client = cl.NewClientJSONRPC(rpcTarget)
	}
	_, err = client.Call("broadcast_tx_async", []interface{}{b}, rpcResult)
	panicErr(err)

	if rpcResult.Code != types.CodeType_OK {
		fmt.Println(rpcResult.Code, string(rpcResult.Data))
		return "", errors.New(string(rpcResult.Data))
	}

	fmt.Println(rpcResult.Code, string(rpcResult.Data))
	priv, err := crypto.ToECDSA(common.Hex2Bytes(privkey))
	panicErr(err)

	caller := crypto.PubkeyToAddress(priv.PublicKey)
	addr := crypto.CreateAddress(caller, nonce)
	fmt.Println("contract address:", addr.Hex())

	return addr.Hex(), nil
}

func executeContract(client *cl.ClientJSONRPC, privkey, contract, abijson, callfunc string, args []interface{}, nonce uint64, commit bool) error {
	aabbii, err := abi.JSON(strings.NewReader(abijson))
	panicErr(err)
	abiArgs, err := commons.ParseArgs(callfunc, aabbii, args)
	data, err := aabbii.Pack(callfunc, abiArgs...)
	panicErr(err)

	// nonce := uint64(time.Now().UnixNano())
	to := common.HexToAddress(contract)
	privkey = ac.SanitizeHex(privkey)

	tx := ethtypes.NewTransaction(nonce, to, big.NewInt(0), gasLimit, big.NewInt(0), data)
	key, err := crypto.HexToECDSA(privkey)
	panicErr(err)
	sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), key)
	panicErr(err)
	sigTx, err := tx.WithSignature(ethSigner, sig)
	panicErr(err)
	b, err := rlp.EncodeToBytes(sigTx)
	panicErr(err)

	if client == nil {
		client = cl.NewClientJSONRPC(rpcTarget)
	}
	if !commit {
		rpcResult := new(types.ResultBroadcastTx)
		_, err = client.Call("broadcast_tx_async", []interface{}{b}, rpcResult)
		panicErr(err)

		if rpcResult.Code != types.CodeType_OK {
			fmt.Println(rpcResult.Code, string(rpcResult.Data), rpcResult.Log)
			return errors.New(string(rpcResult.Data))
		}
		return nil
	}

	rpcResult := new(types.ResultBroadcastTxCommit)
	_, err = client.Call("broadcast_tx_commit", []interface{}{b}, rpcResult)
	panicErr(err)

	if rpcResult.Code != types.CodeType_OK {
		fmt.Println(rpcResult.Code, string(rpcResult.Data), rpcResult.Log)
		return errors.New(string(rpcResult.Data))
	}
	// fmt.Println(res.Code)

	return nil
}

func readContract(client *cl.ClientJSONRPC, privkey, contract, abijson, callfunc string, args []interface{}, nonce uint64) error {
	aabbii, err := abi.JSON(strings.NewReader(abijson))
	panicErr(err)
	data, err := aabbii.Pack(callfunc, args...)
	panicErr(err)

	// nonce := uint64(time.Now().UnixNano())
	to := common.HexToAddress(contract)
	privkey = ac.SanitizeHex(privkey)

	tx := ethtypes.NewTransaction(nonce, to, big.NewInt(0), gasLimit, big.NewInt(0), data)
	key, err := crypto.HexToECDSA(privkey)
	panicErr(err)
	sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), key)
	panicErr(err)
	sigTx, err := tx.WithSignature(ethSigner, sig)
	panicErr(err)
	b, err := rlp.EncodeToBytes(sigTx)
	panicErr(err)

	query := append([]byte{0}, b...)
	rpcResult := new(types.RPCResult)
	if client == nil {
		client = cl.NewClientJSONRPC(rpcTarget)
	}
	_, err = client.Call("query", []interface{}{query}, rpcResult)
	panicErr(err)

	res := (*rpcResult).(*types.ResultQuery)
	fmt.Println("query result:", common.Bytes2Hex(res.Result.Data))
	parseResult, _ := unpackResult(callfunc, aabbii, string(res.Result.Data))
	fmt.Println("parse result:", reflect.TypeOf(parseResult), parseResult)

	return nil
}

func existContract(client *cl.ClientJSONRPC, privkey, contract, bytecode string) bool {
	if strings.Contains(bytecode, "f300") {
		bytecode = strings.Split(bytecode, "f300")[1]
	}

	data := common.Hex2Bytes(bytecode)
	privkey = ac.SanitizeHex(privkey)
	to := common.HexToAddress(ac.SanitizeHex(contract))

	tx := ethtypes.NewTransaction(0, to, big.NewInt(0), gasLimit, big.NewInt(0), crypto.Keccak256(data))
	key, err := crypto.HexToECDSA(privkey)
	panicErr(err)
	sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), key)
	panicErr(err)
	sigTx, err := tx.WithSignature(ethSigner, sig)
	panicErr(err)
	b, err := rlp.EncodeToBytes(sigTx)
	panicErr(err)

	query := append([]byte{4}, b...)
	rpcResult := new(types.ResultQuery)
	if client == nil {
		client = cl.NewClientJSONRPC(rpcTarget)
	}
	_, err = client.Call("query", []interface{}{query}, rpcResult)
	panicErr(err)

	hex := common.Bytes2Hex(rpcResult.Result.Data)
	if hex == "01" {
		return true
	}
	return false
}
