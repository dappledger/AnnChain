package abi

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoSuchMethod    = errors.New("no such method")
	ErrUnmatchedParams = errors.New("number of params is unmatched")
)

func ParseArgs(methodName string, abiDef ABI, params []interface{}) ([]interface{}, error) {
	var method Method
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
	args := []interface{}{}

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

func ParseArg(input Argument, value interface{}) (interface{}, error) {
	typeName := input.Type.String()
	switch {
	case strings.Index(typeName, "bool") == 0:
		if typeName == "bool" {
			return ParseBool(value)
		}
		return ParseBoolSlice(value, input.Type.SliceSize)
	case strings.Index(typeName, "address") == 0:
		if typeName == "address" {
			return ParseAddress(value)
		}
		return ParseAddressSlice(value, input.Type.SliceSize)
	case strings.Index(typeName, "uint8") == 0:
		if typeName == "uint8" {
			return ParseUint8(value)
		}
		return ParseUint8Slice(value, input.Type.SliceSize)
	case strings.Index(typeName, "uint16") == 0:
		if typeName == "uint16" {
			return ParseUint16(value)
		}
		return ParseUint16Slice(value, input.Type.SliceSize)
	case strings.Index(typeName, "uint32") == 0:
		if typeName == "uint32" {
			return ParseUint32(value)
		}
		return ParseUint32Slice(value, input.Type.SliceSize)
	case strings.Index(typeName, "uint64") == 0:
		if typeName == "uint64" {
			return ParseUint64(value)
		}
		return ParseUint64Slice(value, input.Type.SliceSize)
	case strings.Index(typeName, "int8") == 0:
		if typeName == "int8" {
			return ParseInt8(value)
		}
		return ParseInt8Slice(value, input.Type.SliceSize)
	case strings.Index(typeName, "int16") == 0:
		if typeName == "int16" {
			return ParseInt16(value)
		}
		return ParseInt16Slice(value, input.Type.SliceSize)
	case strings.Index(typeName, "int32") == 0:
		if typeName == "int32" {
			return ParseInt32(value)
		}
		return ParseInt32Slice(value, input.Type.SliceSize)
	case strings.Index(typeName, "int64") == 0:
		if typeName == "int64" {
			return ParseInt64(value)
		}
		return ParseInt64Slice(value, input.Type.SliceSize)
	case strings.Index(typeName, "uint256") == 0 ||
		strings.Index(typeName, "uint128") == 0 ||
		strings.Index(typeName, "int256") == 0 ||
		strings.Index(typeName, "int128") == 0:
		if typeName == "uint256" || typeName == "uint128" ||
			typeName == "int256" || typeName == "int128" {
			return ParseBigInt(value)
		}
		return ParseBigIntSlice(value, input.Type.SliceSize)
	case strings.Index(typeName, "bytes8") == 0:
		if typeName == "bytes8" {
			return ParseBytesM(value, 8)
		}
		return ParseBytesMSlice(value, 8, input.Type.SliceSize)
	case strings.Index(typeName, "bytes16") == 0:
		if typeName == "bytes16" {
			return ParseBytesM(value, 16)
		}
		return ParseBytesMSlice(value, 16, input.Type.SliceSize)
	case strings.Index(typeName, "bytes32") == 0:
		if typeName == "bytes32" {
			return ParseBytesM(value, 32)
		}
		return ParseBytesMSlice(value, 32, input.Type.SliceSize)
	case strings.Index(typeName, "bytes64") == 0:
		if typeName == "bytes64" {
			return ParseBytesM(value, 64)
		}
		return ParseBytesMSlice(value, 64, input.Type.SliceSize)
	case strings.Index(typeName, "bytes") == 0:
		if typeName == "bytes" {
			return ParseBytes(value)
		}
	case typeName == "string":
		return ParseString(value)
	}
	return nil, fmt.Errorf("type %v of %v is unsupported", typeName, input.Name)
}
