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
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/dappledger/AnnChain/eth/accounts/abi"
	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/angine/types"
	ac "github.com/dappledger/AnnChain/module/lib/go-common"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/src/chain/app/evm"
	"github.com/dappledger/AnnChain/src/client/commons"
)

func createContract(client *cl.ClientJSONRPC, privkey, bytecode string, nonce uint64) (string, error) {
	data := common.Hex2Bytes(bytecode)
	tx := ethtypes.NewContractCreation(nonce, big.NewInt(0), gasLimit, big.NewInt(0), data)

	privkey = ac.SanitizeHex(privkey)
	key, err := crypto.HexToECDSA(privkey)
	panicErr(err)
	sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
	panicErr(err)
	signedTx, err := tx.WithSignature(ethSigner, sig)
	panicErr(err)
	b, err := rlp.EncodeToBytes(signedTx)
	panicErr(err)

	cctx := evm.CreateContractTx{
		EthTx:  b,
		EthAbi: []byte(defaultAbis),
	}
	cctxBytes, err := evm.EncodeCreateContract(cctx)
	panicErr(err)

	bytesLoad := types.WrapTx(evm.EVMCreateContractTxTag, cctxBytes)
	tmResult := new(types.RPCResult)
	if client == nil {
		client = cl.NewClientJSONRPC(logger, rpcTarget)
	}
	_, err = client.Call("broadcast_tx_sync", []interface{}{defaultChainID, bytesLoad}, tmResult)
	panicErr(err)

	res := (*tmResult).(*types.ResultBroadcastTx)
	if res.Code != 0 {
		fmt.Println("============Error", res.Code, string(res.Data))
		return "", errors.New(string(res.Data))
	}

	fmt.Println("tx result:", signedTx.Hash().Hex())
	sender, _ := ethtypes.Sender(ethSigner, signedTx)
	contractAddr := crypto.CreateAddress(sender, signedTx.Nonce())
	fmt.Println("contract address:", contractAddr.Hex())

	return contractAddr.Hex(), nil
}

func executeContract(client *cl.ClientJSONRPC, privkey, contract, abijson, callfunc string, args []interface{}, nonce uint64) error {
	aabbii, err := abi.JSON(strings.NewReader(abijson))
	panicErr(err)
	data, err := aabbii.Pack(callfunc, args...)
	panicErr(err)

	to := common.HexToAddress(contract)
	privkey = ac.SanitizeHex(privkey)

	tx := ethtypes.NewTransaction(nonce, to, big.NewInt(0), gasLimit, big.NewInt(0), data)
	key, err := crypto.HexToECDSA(privkey)
	panicErr(err)
	sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
	panicErr(err)
	sigTx, err := tx.WithSignature(ethSigner, sig)
	panicErr(err)
	b, err := rlp.EncodeToBytes(sigTx)
	panicErr(err)

	tmResult := new(types.RPCResult)
	if client == nil {
		client = cl.NewClientJSONRPC(logger, commons.QueryServer)
	}
	_, err = client.Call("broadcast_tx_sync", []interface{}{defaultChainID, types.WrapTx(evm.EVMTxTag, b)}, tmResult)
	panicErr(err)

	res := (*tmResult).(*types.ResultBroadcastTx)
	if res.Code != 0 {
		fmt.Println("============Error", res.Code, string(res.Data))
		return errors.New(string(res.Data))
	}

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
	sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
	panicErr(err)
	sigTx, err := tx.WithSignature(ethSigner, sig)
	panicErr(err)
	b, err := rlp.EncodeToBytes(sigTx)
	panicErr(err)

	query := append([]byte{0}, b...)
	tmResult := new(types.RPCResult)
	if client == nil {
		client = cl.NewClientJSONRPC(logger, rpcTarget)
	}
	_, err = client.Call("query", []interface{}{defaultChainID, query}, tmResult)
	panicErr(err)

	res := (*tmResult).(*types.ResultQuery)
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
	sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
	panicErr(err)
	sigTx, err := tx.WithSignature(ethSigner, sig)
	panicErr(err)
	b, err := rlp.EncodeToBytes(sigTx)
	panicErr(err)

	query := append([]byte{4}, b...)
	tmResult := new(types.RPCResult)
	if client == nil {
		client = cl.NewClientJSONRPC(logger, rpcTarget)
	}
	_, err = client.Call("query", []interface{}{defaultChainID, query}, tmResult)
	panicErr(err)

	res := (*tmResult).(*types.ResultQuery)
	hex := common.Bytes2Hex(res.Result.Data)
	if hex == "01" {
		return true
	}
	return false
}
