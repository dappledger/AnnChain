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
	"bytes"
	"encoding/hex"
	_ "encoding/json"
	"fmt"
	// "io/ioutil"
	// "math/big"
	//"reflect"
	// "strings"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/dappledger/AnnChain/eth/accounts/abi"
	// ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	// "github.com/dappledger/AnnChain/eth/crypto"
	// "github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/src/client/commons"
	"github.com/dappledger/AnnChain/src/tools/evmabi"
)

var (
	calljson = `{
    "privkey": "a8971729fbc199fb3459529cebcd8704791fc699d88ac89284f23ff8e7fca7d6",
    "contract": "0x341efb295051fa28c5fc31a4aea21a53128cd496",
    "function": "alltypes",
    "params": [[true, false, false, true],8,256,1,343,"0x341efb295051fa28c5fc31a4aea21a53128cd496", [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15]]
}"`
	contractABI = `[
	{
		"constant": false,
		"inputs": [
{
"name": "blslice",
"type": "bool[]"
},
			{
				"name": "i8",
				"type": "int8"
			},
			{
				"name": "i256",
				"type": "int256"
			},
			{
				"name": "u8",
				"type": "uint8"
			},
			{
				"name": "u",
				"type": "uint256"
			},
			{
				"name": "addr",
				"type": "address"
			},
			{
				"name": "islice",
				"type": "int256[]"
			}
		],
		"name": "alltypes",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
)

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestEchoAddress(t *testing.T) {
	fmt.Println(hex.DecodeString("341efb295051fa28c5fc31a4aea21a53128cd496"))
}

func TestUnpacking(t *testing.T) {
	abiInst, err := abi.JSON(bytes.NewBufferString(contractABI))
	checkError(t, err)

	sj, err := simplejson.NewFromReader(bytes.NewBufferString(calljson))
	checkError(t, err)

	fnName, err := sj.Get("function").String()
	checkError(t, err)

	params, err := sj.Get("params").Array()
	checkError(t, err)

	args, err := commons.ParseArgs(fnName, abiInst, params)
	checkError(t, err)

	data, err := abiInst.Pack(fnName, args...)
	checkError(t, err)

	abiMethod, err := locateMethod(&abiInst, data[:4])
	checkError(t, err)

	checkError(t, unpackMethodInputs(abiMethod, data[4:]))
}

func unpackMethodInputs(m *abi.Method, data []byte) error {
	for i, a := range m.Inputs {
		if a.Indexed {
			continue
		}

		res, err := evmabi.ToGoType(i, a.Type, data)
		if err != nil {
			return err
		}
		_ = res
		fmt.Printf("%v::%v = %+v\n", a.Name, a.Type.String(), res)
	}
	return nil
}
