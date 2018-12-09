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
	"strings"

	"github.com/dappledger/AnnChain/eth/accounts/abi"
	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/angine/types"
	ac "github.com/dappledger/AnnChain/module/lib/go-common"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getNonce(client *cl.ClientJSONRPC, address string) (uint64, error) {
	tmResult := new(types.RPCResult)

	addrHex := ac.SanitizeHex(address)
	addr := common.Hex2Bytes(addrHex)
	query := append([]byte{1}, addr...)

	if client == nil {
		client = cl.NewClientJSONRPC(logger, rpcTarget)
	}
	_, err := client.Call("query", []interface{}{defaultChainID, query}, tmResult)
	if err != nil {
		return 0, err
	}

	res := (*tmResult).(*types.ResultQuery)
	if res.Result.IsErr() {
		fmt.Println(res.Result.Code, res.Result.Log)
		return 0, errors.New(res.Result.Error())
	}
	nonce := new(uint64)
	err = rlp.DecodeBytes(res.Result.Data, nonce)
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

func assertContractExist(client *cl.ClientJSONRPC) {
	if client == nil {
		client = cl.NewClientJSONRPC(logger, rpcTarget)
	}
	exist := existContract(client, defaultPrivKey, defaultContractAddr, defaultBytecode)
	if !exist {
		panic("contract not exists")
	}
}
