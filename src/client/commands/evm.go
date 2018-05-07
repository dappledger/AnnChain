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

package commands

import (
	_ "encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/dappledger/AnnChain/eth/abi"
	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/angine/types"
	ac "github.com/dappledger/AnnChain/module/lib/go-common"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/src/chain/app/evm"
	"github.com/dappledger/AnnChain/src/client/commons"
)

var (
	// this signer appears to be a must in evm 1.5.9
	ethSigner = ethtypes.HomesteadSigner{}

	// gasLimit is the amount we will borrow from gas pool
	gasLimit = big.NewInt(90000000000)

	//ContractCommands defines a more git-like subcommand system
	EVMCommands = cli.Command{
		Name:     "evm",
		Usage:    "operations for evm contract",
		Category: "Contract",
		Subcommands: []cli.Command{
			{
				Name:   "create",
				Usage:  "create a new contract",
				Action: createContract,
				Flags: []cli.Flag{
					anntoolFlags.addr,
					anntoolFlags.pwd,
					anntoolFlags.bytecode,
					anntoolFlags.privkey,
					anntoolFlags.nonce,
					anntoolFlags.abif,
					anntoolFlags.callf,
				},
			},
			{
				Name:   "execute",
				Usage:  "execute a new contract",
				Action: executeContract,
				Flags: []cli.Flag{
					anntoolFlags.addr,
					anntoolFlags.pwd,
					anntoolFlags.payload,
					anntoolFlags.privkey,
					anntoolFlags.nonce,
					anntoolFlags.abistr,
					anntoolFlags.callstr,
					anntoolFlags.to,
					anntoolFlags.abif,
					anntoolFlags.callf,
				},
			},
			{
				Name:   "read",
				Usage:  "read a contract",
				Action: readContract,
				Flags: []cli.Flag{
					anntoolFlags.addr,
					anntoolFlags.pwd,
					anntoolFlags.payload,
					anntoolFlags.privkey,
					anntoolFlags.nonce,
					anntoolFlags.abistr,
					anntoolFlags.callstr,
					anntoolFlags.to,
					anntoolFlags.abif,
					anntoolFlags.callf,
				},
			},
			{
				Name:   "exist",
				Usage:  "check if a contract has been deployed",
				Action: existContract,
				Flags: []cli.Flag{
					anntoolFlags.callf,
				},
			},
		},
	}
)

func readContract(ctx *cli.Context) error {
	chainID := ctx.GlobalString("target")
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	aabbii, _, err := getAbiJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	function := json.Get("function").MustString()
	//if !aabbii.Methods[function].Const {
	//	fmt.Printf("we can only read constant method, %s is not! Any consequence is on you.\n", function)
	//}
	params := json.Get("params").MustArray()
	contractAddress := ac.SanitizeHex(json.Get("contract").MustString())
	args, err := commons.ParseArgs(function, *aabbii, params)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	data, err := aabbii.Pack(function, args...)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	privkey := ctx.String("privkey")
	address := ac.SanitizeHex(ctx.String("address"))
	passwd := ctx.String("passwd")
	nonce := ctx.Uint64("nonce")
	to := common.HexToAddress(contractAddress)

	if privkey == "" {
		privkey = json.Get("privkey").MustString()
	}
	if privkey == "" && (address == "" || passwd == "") {
		panic("should provide privkey or address-passwd pair.")
	}
	privkey = ac.SanitizeHex(privkey)
	tx := ethtypes.NewTransaction(nonce, to, big.NewInt(0), gasLimit, big.NewInt(0), data)

	if privkey != "" {
		key, err := crypto.HexToECDSA(privkey)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}

		sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}
		sigTx, err := tx.WithSignature(ethSigner, sig)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}
		b, err := rlp.EncodeToBytes(sigTx)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}
		query := append([]byte{0}, b...)
		clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
		tmResult := new(types.RPCResult)
		_, err = clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}

		res := (*tmResult).(*types.ResultQuery)
		parseResult, _ := unpackResult(function, *aabbii, string(res.Result.Data))
		fmt.Println("parse result:", reflect.TypeOf(parseResult), parseResult)
	}

	return nil
}

func unpackResult(method string, abiDef abi.ABI, output string) (interface{}, error) {
	m, ok := abiDef.Methods[method]
	if !ok {
		return nil, errors.New("No such method")
	}
	if len(m.Outputs) == 0 {
		return nil, errors.New("method " + m.Name + " doesn't have any returns")
	}
	if len(m.Outputs) == 1 {
		var result interface{}
		parsedData := common.ParseData(output)
		if err := abiDef.Unpack(&result, method, parsedData); err != nil {
			fmt.Println("error:", err)
			return nil, err
		}
		if strings.Index(m.Outputs[0].Type.String(), "bytes") == 0 {
			b := result.([]byte)
			idx := 0
			for i := 0; i < len(b); i++ {
				if b[i] == 0 {
					idx = i
				} else {
					break
				}
			}
			b = b[idx+1:]
			return fmt.Sprintf("%s", b), nil
		}
		return result, nil
	}
	d := common.ParseData(output)
	var result []interface{}
	if err := abiDef.Unpack(&result, method, d); err != nil {
		fmt.Println("fail to unpack outpus:", err)
		return nil, err
	}

	retVal := map[string]interface{}{}
	for i, output := range m.Outputs {
		if strings.Index(output.Type.String(), "bytes") == 0 {
			b := result[i].([]byte)
			idx := 0
			for i := 0; i < len(b); i++ {
				if b[i] == 0 {
					idx = i
				} else {
					break
				}
			}
			b = b[idx+1:]
			retVal[output.Name] = fmt.Sprintf("%s", b)
		} else {
			retVal[output.Name] = result[i]
		}
	}
	return retVal, nil
}

func executeContract(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chain is required", 127)
	}
	chainID := ctx.GlobalString("target")
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 110)
	}
	aabbii, _, err := getAbiJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 110)
	}
	function := json.Get("function").MustString()
	params := json.Get("params").MustArray()
	contractAddress := ac.SanitizeHex(json.Get("contract").MustString())
	args, err := commons.ParseArgs(function, *aabbii, params)
	if err != nil {
		return cli.NewExitError(err.Error(), 110)
	}

	data, err := aabbii.Pack(function, args...)
	if err != nil {
		return cli.NewExitError(err.Error(), 110)
	}

	privkey := ctx.String("privkey")
	if privkey == "" {
		privkey = json.Get("privkey").MustString()
	}
	address := ac.SanitizeHex(ctx.String("address"))
	passwd := ctx.String("passwd")
	nonce := ctx.Uint64("nonce")
	to := common.HexToAddress(contractAddress)

	if privkey == "" && (address == "" || passwd == "") {
		panic("should provide privkey or address-passwd pair.")
	}
	privkey = ac.SanitizeHex(privkey)
	tx := ethtypes.NewTransaction(nonce, to, big.NewInt(0), gasLimit, big.NewInt(0), data)

	if privkey != "" {
		key, err := crypto.HexToECDSA(privkey)
		if err != nil {
			return cli.NewExitError(err.Error(), 123)
		}
		sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
		if err != nil {
			return cli.NewExitError(err.Error(), 123)
		}
		sigTx, err := tx.WithSignature(ethSigner, sig)
		if err != nil {
			return cli.NewExitError(err.Error(), 123)
		}
		b, err := rlp.EncodeToBytes(sigTx)
		if err != nil {
			return cli.NewExitError(err.Error(), 123)
		}

		tmResult := new(types.RPCResult)
		clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
		_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{chainID, types.WrapTx(evm.EVMTxTag, b)}, tmResult)
		if err != nil {
			return cli.NewExitError(err.Error(), 123)
		}

		//tmRes := tmResult.(*ctypes.RPCResult)
		//res := (*tmResult).(*ctypes.ResultBroadcastTx)

		fmt.Println("tx result:", sigTx.Hash().Hex())
	}

	return nil
}

func createContract(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chain is required", 127)
	}
	chainID := ctx.GlobalString("target")
	nonce := ctx.Uint64("nonce")
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	aabbii, abiBytes, err := getAbiJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	params := json.Get("params").MustArray()
	privkey := json.Get("privkey").MustString()
	bytecode := common.Hex2Bytes(json.Get("bytecode").MustString())
	if len(bytecode) == 0 {
		return cli.NewExitError("please give me the bytecode the contract", 127)
	}
	if len(params) > 0 {
		args, err := commons.ParseArgs("", *aabbii, params)
		if err != nil {
			cli.NewExitError(err.Error(), 127)
		}
		data, err := aabbii.Pack("", args...)
		if err != nil {
			cli.NewExitError(err.Error(), 127)
		}
		bytecode = append(bytecode, data...)
	}

	if privkey == "" {
		if privkey = ctx.String("privkey"); privkey == "" {
			return errors.New("should provide privkey or address-passwd pair")
		}
	}

	tx := ethtypes.NewContractCreation(nonce, big.NewInt(0), gasLimit, big.NewInt(0), bytecode)
	privkey = ac.SanitizeHex(privkey)
	if privkey != "" {
		key, err := crypto.HexToECDSA(privkey)
		if err != nil {
			return cli.NewExitError(err.Error(), 110)
		}
		sig, _ := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
		signedTx, _ := tx.WithSignature(ethSigner, sig)
		b, err := rlp.EncodeToBytes(signedTx)
		if err != nil {
			return cli.NewExitError(err.Error(), 110)
		}

		cctx := evm.CreateContractTx{
			EthTx:  b,
			EthAbi: abiBytes,
		}
		cctxBytes, err := evm.EncodeCreateContract(cctx)
		if err != nil {
			return cli.NewExitError(err.Error(), 110)
		}
		bytesLoad := types.WrapTx(evm.EVMCreateContractTxTag, cctxBytes)
		tmResult := new(types.RPCResult)
		clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
		_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{chainID, bytesLoad}, tmResult)
		if err != nil {
			return cli.NewExitError(err.Error(), 110)
		}

		//tmRes := tmResult.(*ctypes.RPCResult)
		//res := (*tmResult).(*ctypes.ResultBroadcastTx)
		fmt.Println("tx result:", signedTx.Hash().Hex())
		sender, _ := ethtypes.Sender(ethSigner, signedTx)
		contractAddr := crypto.CreateAddress(sender, signedTx.Nonce())
		fmt.Println("contract address:", contractAddr.Hex())
	}

	return nil
}

func existContract(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chain is required", 127)
	}
	chainID := ctx.GlobalString("target")
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	bytecode := json.Get("bytecode").MustString()
	contractAddress := json.Get("contract").MustString()
	privkey := json.Get("privkey").MustString()
	if privkey == "" || contractAddress == "" || bytecode == "" {
		return cli.NewExitError("missing params", 127)
	}
	if strings.Contains(bytecode, "f300") {
		bytecode = strings.Split(bytecode, "f300")[1]
	}
	data := common.Hex2Bytes(bytecode)
	privkey = ac.SanitizeHex(privkey)
	contractAddr := common.HexToAddress(ac.SanitizeHex(contractAddress))
	tx := ethtypes.NewTransaction(0, contractAddr, big.NewInt(0), gasLimit, big.NewInt(0), crypto.Keccak256(data))
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	signedTx, err := tx.WithSignature(ethSigner, sig)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	query := append([]byte{4}, txBytes...)
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	_, err = clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)
	hex := common.Bytes2Hex(res.Result.Data)
	if hex == "01" {
		fmt.Println("Yes!!!")
	} else {
		fmt.Println("No!!!")
	}
	return nil
}

func getCallParamsJSON(ctx *cli.Context) (*simplejson.Json, error) {
	var calljson string
	if ctx.String("callf") != "" {
		dat, err := fileData(ctx.String("callf"))
		if err != nil {
			return nil, err
		}
		calljson = string(dat)
	} else {
		calljson = ctx.String("calljson")
	}
	return simplejson.NewJson([]byte(calljson))
}

func getAbiJSON(ctx *cli.Context) (*abi.ABI, []byte, error) {
	var abijson string
	if ctx.String("abif") == "" {
		abijson = ctx.String("abi")
	} else {
		dat, err := fileData(ctx.String("abif"))
		if err != nil {
			return nil, nil, err
		}
		abijson = string(dat)
	}
	jAbi, err := abi.JSON(strings.NewReader(abijson))
	if err != nil {
		return nil, nil, err
	}

	return &jAbi, []byte(abijson), nil
}
