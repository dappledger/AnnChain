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
	"strings"

	"github.com/dappledger/AnnChain/eth/accounts/abi"
	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/rlp"
	ac "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
	"github.com/dappledger/AnnChain/gemmill/types"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getNonce(client *cl.ClientJSONRPC, address string) (uint64, error) {
	rpcResult := new(types.ResultQuery)

	addrHex := ac.SanitizeHex(address)
	addr := common.Hex2Bytes(addrHex)
	query := append([]byte{1}, addr...)

	if client == nil {
		client = cl.NewClientJSONRPC(rpcTarget)
	}
	_, err := client.Call("query", []interface{}{query}, rpcResult)
	if err != nil {
		return 0, err
	}

	if rpcResult.Result.IsErr() {
		fmt.Println(rpcResult.Result.Code, rpcResult.Result.Log)
		return 0, errors.New(rpcResult.Result.Error())
	}
	nonce := new(uint64)
	err = rlp.DecodeBytes(rpcResult.Result.Data, nonce)
	panicErr(err)

	return *nonce, nil
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
		parsedData := ParseData(output)
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
	d := ParseData(output)
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

func assertContractExist(client *cl.ClientJSONRPC) {
	if client == nil {
		client = cl.NewClientJSONRPC(rpcTarget)
	}
	exist := existContract(client, defaultPrivKey, defaultContractAddr, defaultBytecode)
	if !exist {
		panic("contract not exists")
	}
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
