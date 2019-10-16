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

package commands

import (
	json2 "encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"

	"github.com/bitly/go-simplejson"
	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/cmd/client/commons"
	"github.com/dappledger/AnnChain/eth/accounts/abi"
	"github.com/dappledger/AnnChain/eth/common"
	etypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	client "github.com/dappledger/AnnChain/gemmill/rpc/client"
	"github.com/dappledger/AnnChain/gemmill/types"
)

var (
	// gasLimit is the amount we will borrow from gas pool
	gasLimit = uint64(90000000000)

	//ContractCommands defines a more git-like subcommand system
	EVMCommands = cli.Command{
		Name:     "evm",
		Usage:    "operations for evm",
		Category: "Contract",
		Subcommands: []cli.Command{
			{
				Name:   "create",
				Usage:  "create a new contract",
				Action: createContract,
				Flags: []cli.Flag{
					anntoolFlags.bytecode,
					anntoolFlags.nonce,
					anntoolFlags.abif,
					anntoolFlags.callf,
				},
			}, {
				Name:   "call",
				Usage:  "execute a new contract",
				Action: callContract,
				Flags: []cli.Flag{
					anntoolFlags.payload,
					anntoolFlags.nonce,
					anntoolFlags.abistr,
					anntoolFlags.callstr,
					anntoolFlags.to,
					anntoolFlags.abif,
					anntoolFlags.callf,
				},
			}, {
				Name:   "read",
				Usage:  "read a contract",
				Action: readContract,
				Flags: []cli.Flag{
					anntoolFlags.payload,
					anntoolFlags.nonce,
					anntoolFlags.abistr,
					anntoolFlags.callstr,
					anntoolFlags.to,
					anntoolFlags.abif,
					anntoolFlags.callf,
				},
			}, {
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

func createContract(ctx *cli.Context) error {
	nonce := ctx.Uint64("nonce")
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	aabbii, err := getAbiJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	params := json.Get("params").MustArray()
	bytecode := common.Hex2Bytes(json.Get("bytecode").MustString())
	if len(bytecode) == 0 {
		return cli.NewExitError("please give me the bytecode the contract", 127)
	}
	if len(params) > 0 {
		args, err := commons.ParseArgs("", *aabbii, params)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}
		data, err := aabbii.Pack("", args...)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}
		bytecode = append(bytecode, data...)
	}

	tx := etypes.NewContractCreation(nonce, big.NewInt(0), gasLimit, big.NewInt(0), bytecode)

	key, err := requireAccPrivky(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	privBytes := common.Hex2Bytes(key)
	addrBytes, err := getAddrBytes(privBytes)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	signer, sig, err := SignTx(privBytes, tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	signedTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	b, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return cli.NewExitError(err.Error(), 110)
	}

	rpcResult := new(types.ResultBroadcastTxCommit)
	clientJSON := client.NewClientJSONRPC(commons.QueryServer)
	_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{b}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 110)
	}

	hash := rpcResult.TxHash

	contractAddr := crypto.CreateAddress(common.BytesToAddress(addrBytes), signedTx.Nonce())
	fmt.Println("contract address:", contractAddr.Hex())
	fmt.Println("tx result:", hash)

	return nil
}

func callContract(ctx *cli.Context) error {
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return err
	}
	aabbii, err := getAbiJSON(ctx)
	if err != nil {
		return err
	}
	function := json.Get("function").MustString()
	params := json.Get("params").MustArray()
	contractAddress := gcmn.SanitizeHex(json.Get("contract").MustString())
	args, err := commons.ParseArgs(function, *aabbii, params)
	if err != nil {
		panic(err)
	}

	data, err := aabbii.Pack(function, args...)
	if err != nil {
		return err
	}

	nonce := ctx.Uint64("nonce")
	to := common.HexToAddress(contractAddress)

	tx := etypes.NewTransaction(nonce, to, big.NewInt(0), gasLimit, big.NewInt(0), data)

	key, err := requireAccPrivky(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	privBytes := common.Hex2Bytes(key)

	signer, sig, err := SignTx(privBytes, tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	sigTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	b, err := rlp.EncodeToBytes(sigTx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	rpcResult := new(types.ResultBroadcastTxCommit)
	clientJSON := client.NewClientJSONRPC(commons.QueryServer)
	_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{b}, rpcResult)
	if err != nil {
		return err
	}

	hash := rpcResult.TxHash
	fmt.Println("tx result:", hash)

	return nil
}

func readContract(ctx *cli.Context) error {
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	aabbii, err := getAbiJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	function := json.Get("function").MustString()
	if !aabbii.Methods[function].Const {
		fmt.Printf("we can only read constant method, %s is not! Any consequence is on you.\n", function)
	}
	params := json.Get("params").MustArray()
	contractAddress := gcmn.SanitizeHex(json.Get("contract").MustString())
	args, err := commons.ParseArgs(function, *aabbii, params)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	data, err := aabbii.Pack(function, args...)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	nonce := ctx.Uint64("nonce")
	to := common.HexToAddress(contractAddress)

	tx := etypes.NewTransaction(nonce, to, big.NewInt(0), gasLimit, big.NewInt(0), data)

	key, err := requireAccPrivky(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	privBytes := common.Hex2Bytes(key)

	signer, sig, err := SignTx(privBytes, tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	sigTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	b, err := rlp.EncodeToBytes(sigTx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	query := append([]byte{0}, b...)
	clientJSON := client.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(types.ResultQuery)
	_, err = clientJSON.Call("query", []interface{}{query}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	parseResult, _ := UnpackResult(function, *aabbii, string(rpcResult.Result.Data))
	responseJSON, err := json2.Marshal(parseResult)

	fmt.Println("result:", string(responseJSON))

	return nil
}

func UnpackResult(method string, abiDef abi.ABI, output string) (interface{}, error) {
	m, ok := abiDef.Methods[method]
	if !ok {
		return nil, fmt.Errorf("No such method")
	}
	if len(m.Outputs) == 1 {
		var result interface{}
		parsedData := ParseData(output)
		if err := abiDef.Unpack(&result, method, parsedData); err != nil {
			fmt.Println("error:", err)
			return nil, err
		}
		if strings.Index(m.Outputs[0].Type.String(), "bytes") == 0 {
			b, err := bytesN2Slice(result, m.Outputs[0].Type.Size)
			if err != nil {
				return nil, err
			}
			idx := 0
			for idx = 0; idx < len(b); idx++ {
				if b[idx] != 0 {
					break
				}
			}
			b = b[idx:]
			return fmt.Sprintf("0x%x", b), nil
		}
		return result, nil
	}

	d := ParseData(output)
	result := make([]interface{}, m.Outputs.LengthNonIndexed())
	if err := abiDef.Unpack(&result, method, d); err != nil {
		fmt.Println("fail to unpack outpus:", err)
		return nil, err
	}

	retVal := map[string]interface{}{}
	for i, output := range m.Outputs {
		var value interface{}
		if strings.Index(output.Type.String(), "bytes") == 0 {
			b, err := bytesN2Slice(result[i], m.Outputs[0].Type.Size)
			if err != nil {
				return nil, err
			}
			idx := 0
			for idx = 0; idx < len(b); idx++ {
				if b[idx] != 0 {
					break
				}
			}
			b = b[idx:]
			value = fmt.Sprintf("0x%x", b)
		} else {
			value = result[i]
		}
		if len(output.Name) == 0 {
			retVal[fmt.Sprintf("%v", i)] = value
		} else {
			retVal[output.Name] = value
		}
	}
	return retVal, nil
}

func existContract(ctx *cli.Context) error {
	json, err := getCallParamsJSON(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	bytecode := json.Get("bytecode").MustString()
	contractAddress := json.Get("contract").MustString()
	if contractAddress == "" || bytecode == "" {
		return cli.NewExitError("missing params", 127)
	}
	if strings.Contains(bytecode, "f300") {
		bytecode = strings.Split(bytecode, "f300")[1]
	}
	data := common.Hex2Bytes(bytecode)
	contractAddr := common.HexToAddress(gcmn.SanitizeHex(contractAddress))

	tx := etypes.NewTransaction(0, contractAddr, big.NewInt(0), gasLimit, big.NewInt(0), crypto.Keccak256(data))

	key, err := requireAccPrivky(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	privBytes := common.Hex2Bytes(key)

	signer, sig, err := SignTx(privBytes, tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	signedTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	query := append([]byte{4}, txBytes...)
	clientJSON := client.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(types.ResultQuery)
	_, err = clientJSON.Call("query", []interface{}{query}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	hex := common.Bytes2Hex(rpcResult.Result.Data)
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

func getAbiJSON(ctx *cli.Context) (*abi.ABI, error) {
	var abijson string
	if ctx.String("abif") == "" {
		abijson = ctx.String("abi")
	} else {
		dat, err := fileData(ctx.String("abif"))
		if err != nil {
			return nil, err
		}
		abijson = string(dat)
	}
	jAbi, err := abi.JSON(strings.NewReader(abijson))
	if err != nil {
		return nil, err
	}
	return &jAbi, nil
}

func ParseData(data ...interface{}) (ret []byte) {
	for _, item := range data {
		switch t := item.(type) {
		case string:
			var str []byte
			if IsHex(t) {
				str = common.Hex2Bytes(t[2:])
			} else {
				str = []byte(t)
			}

			ret = append(ret, common.RightPadBytes(str, 32)...)
		case []byte:
			ret = append(ret, common.LeftPadBytes(t, 32)...)
		}
	}

	return
}

func IsHex(str string) bool {
	l := len(str)
	return l >= 4 && l%2 == 0 && str[0:2] == "0x"
}

func bytesN2Slice(value interface{}, m int) ([]byte, error) {
	switch m {
	case 0:
		v := value.([]byte)
		return v, nil
	case 8:
		v := value.([8]byte)
		return v[:], nil
	case 16:
		v := value.([16]byte)
		return v[:], nil
	case 32:
		v := value.([32]byte)
		return v[:], nil
	case 64:
		v := value.([64]byte)
		return v[:], nil
	}
	return nil, fmt.Errorf("type(bytes%d) not support", m)
}

func SignTx(privBytes []byte, tx *etypes.Transaction) (signer etypes.Signer, sig []byte, err error) {
	signer = new(etypes.HomesteadSigner)

	privkey, err := crypto.ToECDSA(privBytes)
	if err != nil {
		return nil, nil, err
	}

	sig, err = crypto.Sign(signer.Hash(tx).Bytes(), privkey)

	return signer, sig, nil
}

func getAddrBytes(privBytes []byte) (addrBytes []byte, err error) {
	privkey, err := crypto.ToECDSA(privBytes)
	if err != nil {
		return nil, err
	}
	addr := crypto.PubkeyToAddress(privkey.PublicKey)
	addrBytes = addr[:]

	return addrBytes, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func fileData(str string) ([]byte, error) {
	path, _ := pathExists(str)
	if !path {
		fstr := strings.Replace(str, "\\\r\n", "\r\n", -1)
		fstr = strings.Replace(fstr, "\\\"", "\"", -1)
		return []byte(fstr), nil
	}
	return ioutil.ReadFile(str)
}
