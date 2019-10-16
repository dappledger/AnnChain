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

package commons

import (
	"fmt"
	"strings"

	"github.com/dappledger/AnnChain/cmd/client/utils"
	"github.com/dappledger/AnnChain/eth/accounts/abi"
)

func ParseArgs(methodName string, abiDef abi.ABI, params []interface{}) ([]interface{}, error) {
	var method abi.Method
	if methodName == "" {
		method = abiDef.Constructor
	} else {
		var ok bool
		method, ok = abiDef.Methods[methodName]
		if !ok {
			return nil, ErrNoSuchMethod
		}
	}

	if params == nil {
		params = []interface{}{}
	}
	if len(params) != len(method.Inputs) {
		return nil, ErrUnmatchedParams
	}
	var args []interface{}

	for i := range params {
		a, err := ParseArg(method.Inputs[i], params[i])
		if err != nil {
			fmt.Println(fmt.Sprintf("fail to parse args %v into %s: %v ", params[i], method.Inputs[i].Name, err))
			return nil, err
		}
		args = append(args, a)
	}
	return args, nil
}

func ParseArg(input abi.Argument, value interface{}) (interface{}, error) {
	typeName := input.Type.String()
	switch {
	case strings.Index(typeName, "bool") == 0:
		if typeName == "bool" {
			return utils.ParseBool(value)
		}
		return utils.ParseBoolSlice(value)
	case strings.Index(typeName, "address") == 0:
		if typeName == "address" {
			return utils.ParseAddress(value)
		}
		return utils.ParseAddressSlice(value)
	case strings.Index(typeName, "uint8") == 0:
		if typeName == "uint8" {
			return utils.ParseUint8(value)
		}
		return utils.ParseUint8Slice(value)
	case strings.Index(typeName, "uint16") == 0:
		if typeName == "uint16" {
			return utils.ParseUint16(value)
		}
		return utils.ParseUint16Slice(value)
	case strings.Index(typeName, "uint32") == 0:
		if typeName == "uint32" {
			return utils.ParseUint32(value)
		}
		return utils.ParseUint32Slice(value)
	case strings.Index(typeName, "uint64") == 0:
		if typeName == "uint64" {
			return utils.ParseUint64(value)
		}
		return utils.ParseUint64Slice(value)
	case strings.Index(typeName, "int8") == 0:
		if typeName == "int8" {
			return utils.ParseInt8(value)
		}
		return utils.ParseInt8Slice(value)
	case strings.Index(typeName, "int16") == 0:
		if typeName == "int16" {
			return utils.ParseInt16(value)
		}
		return utils.ParseInt16Slice(value)
	case strings.Index(typeName, "int32") == 0:
		if typeName == "int32" {
			return utils.ParseInt32(value)
		}
		return utils.ParseInt32Slice(value)
	case strings.Index(typeName, "int64") == 0:
		if typeName == "int64" {
			return utils.ParseInt64(value)
		}
		return utils.ParseInt64Slice(value)
	case strings.Index(typeName, "uint256") == 0 ||
		strings.Index(typeName, "uint128") == 0 ||
		strings.Index(typeName, "int256") == 0 ||
		strings.Index(typeName, "int128") == 0:
		if typeName == "uint256" || typeName == "uint128" ||
			typeName == "int256" || typeName == "int128" {
			return utils.ParseBigInt(value)
		}
		return utils.ParseBigIntSlice(value)
	case strings.Index(typeName, "bytes8") == 0:
		if typeName == "bytes8" {
			return utils.ParseBytesM(value, 8)
		}
		return utils.ParseBytesMSlice(value, 8)
	case strings.Index(typeName, "bytes16") == 0:
		if typeName == "bytes16" {
			return utils.ParseBytesM(value, 16)
		}
		return utils.ParseBytesMSlice(value, 16)
	case strings.Index(typeName, "bytes32") == 0:
		if typeName == "bytes32" {
			return utils.ParseBytesM(value, 32)
		}
		return utils.ParseBytesMSlice(value, 32)
	case strings.Index(typeName, "bytes64") == 0:
		if typeName == "bytes64" {
			return utils.ParseBytesM(value, 64)
		}
		return utils.ParseBytesMSlice(value, 64)
	case strings.Index(typeName, "bytes") == 0:
		if typeName == "bytes" {
			return utils.ParseBytes(value)
		}
	case strings.Index(typeName, "string") == 0:
		if typeName == "string" {
			return utils.ParseString(value)
		}
		return utils.ParseStringSlice(value)
	}
	return nil, fmt.Errorf("type %v of %v is unsupported", typeName, input.Name)
}
