package ethtools

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

var ethSigner = ethtypes.HomesteadSigner{}

func ParseData(methodName string, abiDef abi.ABI, params []interface{}) (string, error) {
	args, err := ParseArgs(methodName, &abiDef, params)
	if err != nil {
		return "", err
	}
	data, err := abiDef.Pack(methodName, args...)
	if err != nil {
		return "", err
	}

	var hexData string
	for _, b := range data {
		hexDataP := strconv.FormatInt(int64(b), 16)
		if len(hexDataP) == 1 {
			hexDataP = "0" + hexDataP
		}
		hexData += hexDataP
	}
	return hexData, nil
}

func ParseArgs(methodName string, abiDef *abi.ABI, params []interface{}) ([]interface{}, error) {
	var method abi.Method
	if methodName == "" {
		method = abiDef.Constructor
	} else {
		var ok bool
		method, ok = abiDef.Methods[methodName]
		if !ok {
			return nil, fmt.Errorf("no such method")
		}
	}

	if params == nil {
		params = []interface{}{}
	}
	if len(params) != len(method.Inputs) {
		return nil, fmt.Errorf("unmatched params")
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

func ParseArg(input abi.Argument, value interface{}) (interface{}, error) {
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

func unpackResult(method string, abiDef abi.ABI, output string) (interface{}, error) {
	m, ok := abiDef.Methods[method]
	if !ok {
		return nil, fmt.Errorf("No such method")
	}
	if len(m.Outputs) == 1 {
		var result interface{}
		parsedData := common.ParseData(output)
		if err := abiDef.Unpack(&result, method, parsedData); err != nil {
			fmt.Println("error:", err)
			return nil, err
		}
		outtype := m.Outputs[0].Type.String()
		if strings.Index(outtype, "bytes") == 0 {
			return parseByteNSlc(outtype, result), nil
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
		if output.Name == "" {
			return result, nil
		}
		outtype := output.Type.String()
		if strings.Index(outtype, "bytes") == 0 {
			retVal[output.Name] = parseByteNSlc(outtype, result[i])
		} else {
			retVal[output.Name] = result[i]
		}
	}
	return retVal, nil
}

func parseByteNSlc(outtype string, result interface{}) string {
	if outtype != "bytes[]" && outtype != "bytes" {
		// TODO just 1D-array now
		b := result.([][]byte)
		idx := 0
		for i := 0; i < len(b); i++ {
			for j := 0; j < len(b[i]); j++ {
				if b[i][j] == 0 {
					idx = j
				} else {
					break
				}
			}
			b[i] = b[i][idx+1:]
		}
		return fmt.Sprintf("%s", b)
	} else {
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
		return fmt.Sprintf("%s", b)
	}

}

func SignTx(tx *ethtypes.Transaction, privkey *ecdsa.PrivateKey) (*ethtypes.Transaction, error) {
	sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), privkey)
	if err != nil {
		return nil, err
	}
	sigTx, err := tx.WithSignature(ethSigner, sig)
	if err != nil {
		return nil, err
	}
	return sigTx, nil
}

// TxSignToBytes returns txBytes,txHashBytes,error
func TxSignToBytes(tx *ethtypes.Transaction, privkey *ecdsa.PrivateKey) ([]byte, common.Hash, error) {
	signedTx, err := SignTx(tx, privkey)
	if err != nil {
		return nil, common.Hash{}, err
	}
	txBytes, err := rlp.EncodeToBytes(signedTx)
	txHash := signedTx.Hash()
	return txBytes, txHash, err
}
