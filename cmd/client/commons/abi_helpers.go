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
	case typeName == "bool":
		return utils.ParseBool(value)
	case strings.HasPrefix(typeName, "bool"):
		return utils.ParseBoolSlice(value)
	case typeName == "address":
		return utils.ParseAddress(value)
	case strings.HasPrefix(typeName, "address"):
		return utils.ParseAddressSlice(value)
	case typeName == "uint8":
		return utils.ParseUint8(value)
	case strings.HasPrefix(typeName, "uint8"):
		return utils.ParseUint8Slice(value)
	case typeName == "uint16":
		return utils.ParseUint16(value)
	case strings.HasPrefix(typeName, "uint16"):
		return utils.ParseUint16Slice(value)
	case typeName == "uint32":
		return utils.ParseUint32(value)
	case strings.HasPrefix(typeName, "uint32"):
		return utils.ParseUint32Slice(value)
	case typeName == "uint64":
		return utils.ParseUint64(value)
	case strings.HasPrefix(typeName, "uint64"):
		return utils.ParseUint64Slice(value)
	case typeName == "int8":
		return utils.ParseInt8(value)
	case strings.HasPrefix(typeName, "int8"):
		return utils.ParseInt8Slice(value)
	case typeName == "int16":
		return utils.ParseInt16(value)
	case strings.HasPrefix(typeName, "int16"):
		return utils.ParseInt16Slice(value)
	case typeName == "int32":
		return utils.ParseInt32(value)
	case strings.HasPrefix(typeName, "int32"):
		return utils.ParseInt32Slice(value)
	case typeName == "int64":
		return utils.ParseInt64(value)
	case strings.HasPrefix(typeName, "int64"):
		return utils.ParseInt64Slice(value)
	case typeName == "uint256" || typeName == "uint128" ||
		typeName == "int256" || typeName == "int128":
		return utils.ParseBigInt(value)
	case strings.HasPrefix(typeName, "uint256") ||
		strings.HasPrefix(typeName, "uint128") ||
		strings.HasPrefix(typeName, "int256") ||
		strings.HasPrefix(typeName, "int128"):
		return utils.ParseBigIntSlice(value)
	case typeName == "bytes8":
		return utils.ParseBytesM(value, 8)
	case strings.HasPrefix(typeName, "bytes8"):
		return utils.ParseBytesMSlice(value, 8)
	case typeName == "bytes16":
		return utils.ParseBytesM(value, 16)
	case strings.HasPrefix(typeName, "bytes16"):
		return utils.ParseBytesMSlice(value, 16)
	case typeName == "bytes32":
		return utils.ParseBytesM(value, 32)
	case strings.HasPrefix(typeName, "bytes32"):
		return utils.ParseBytesMSlice(value, 32)
	case typeName == "bytes64":
		return utils.ParseBytesM(value, 64)
	case strings.HasPrefix(typeName, "bytes64"):
		return utils.ParseBytesMSlice(value, 64)
	case typeName == "bytes":
		return utils.ParseBytes(value)
	case typeName == "string":
		return utils.ParseString(value)
	case strings.HasPrefix(typeName, "string"):
		return utils.ParseStringSlice(value)
	}
	return nil, fmt.Errorf("type %v of %v is unsupported", typeName, input.Name)
}
