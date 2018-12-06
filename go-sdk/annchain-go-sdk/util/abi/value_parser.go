package abi

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	simplejson "github.com/bitly/go-simplejson"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"

	"errors"
	"strings"
)

func ParseUint8(value interface{}) (uint8, error) {
	if value == nil {
		return 0, fmt.Errorf("value cannot be nil")
	}
	v, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(v), nil
}

func ParseUint16(value interface{}) (uint16, error) {
	if value == nil {
		return 0, fmt.Errorf("value cannot be nil")
	}
	v, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(v), nil
}

func ParseUint32(value interface{}) (uint32, error) {
	if value == nil {
		return 0, fmt.Errorf("value cannot be nil")
	}
	v, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}

func ParseUint64(value interface{}) (uint64, error) {
	if value == nil {
		return 0, fmt.Errorf("value cannot be nil")
	}
	v, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func ParseInt8(value interface{}) (int8, error) {
	if value == nil {
		return 0, fmt.Errorf("value cannot be nil")
	}
	v, err := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(v), nil
}

func ParseInt16(value interface{}) (int16, error) {
	if value == nil {
		return 0, fmt.Errorf("value cannot be nil")
	}
	v, err := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(v), nil
}

func ParseInt32(value interface{}) (int32, error) {
	if value == nil {
		return 0, fmt.Errorf("value cannot be nil")
	}
	v, err := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func ParseInt64(value interface{}) (int64, error) {
	if value == nil {
		return 0, fmt.Errorf("value cannot be nil")
	}
	v, err := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func ParseBigInt(value interface{}) (*big.Int, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	strVal := fmt.Sprintf("%v", value)

	if strings.Index(strVal, "e") != -1 {
		s, err := parseScientificNotation(strVal)
		if err != nil {
			fmt.Println(fmt.Sprintf("fail to parse scientific notation %s: %v", strVal, err))
			return nil, err
		}
		bi, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return nil, fmt.Errorf("Fail to parse %v to big.Int", value)
		}
		return bi, nil
	}
	v, ok := new(big.Int).SetString(strVal, 10)
	if !ok {
		return nil, fmt.Errorf("Fail to convert %v to big.Int", value)
	}
	return v, nil
}

func parseScientificNotation(str string) (string, error) {
	eIndex := strings.Index(str, "e+")
	if eIndex == -1 || eIndex == 0 {
		return "", errors.New("invalid scientific notation number")
	}
	times, err := strconv.ParseInt(str[eIndex+2:], 10, 64)
	if err != nil {
		return "", err
	}
	if times == 0 {
		return str[:eIndex], nil
	}
	intTimes := int(times)
	pointIndex := strings.Index(str, ".")
	if pointIndex+1 == eIndex {
		return "", errors.New("invalid scientific notation number")
	}
	var withoutPoint string
	if pointIndex != -1 {
		withoutPoint = str[:pointIndex] + str[pointIndex+1:eIndex]
	} else {
		withoutPoint = str[:eIndex]
	}

	l := len(withoutPoint)
	if pointIndex == -1 {
		for i := 0; i < intTimes; i++ {
			withoutPoint += "0"
		}
		return deletePrefixZero(withoutPoint), nil
	} else if intTimes >= l-pointIndex {
		for i := 0; i < intTimes-(l-pointIndex); i++ {
			withoutPoint += "0"
		}
		return deletePrefixZero(withoutPoint), nil
	} else {
		withoutPoint = withoutPoint[:pointIndex+intTimes] + "." + withoutPoint[pointIndex+intTimes:]
		return deletePrefixZero(withoutPoint), nil
	}
}

func deletePrefixZero(s string) string {
	for strings.Index(s, "0") == 0 {
		s = s[1:]
	}
	return s
}

func ParseUint8Slice(value interface{}, size int) ([]uint8, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	retVals := []uint8{}
	for _, v := range values {
		retVal, err := ParseUint8(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseUint16Slice(value interface{}, size int) ([]uint16, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	retVals := []uint16{}
	for _, v := range values {
		retVal, err := ParseUint16(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseUint32Slice(value interface{}, size int) ([]uint32, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	retVals := []uint32{}
	for _, v := range values {
		retVal, err := ParseUint32(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseUint64Slice(value interface{}, size int) ([]uint64, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	retVals := []uint64{}
	for _, v := range values {
		retVal, err := ParseUint64(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseInt8Slice(value interface{}, size int) ([]int8, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	retVals := []int8{}
	for _, v := range values {
		retVal, err := ParseInt8(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseInt16Slice(value interface{}, size int) ([]int16, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	retVals := []int16{}
	for _, v := range values {
		retVal, err := ParseInt16(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseInt32Slice(value interface{}, size int) ([]int32, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	retVals := []int32{}
	for _, v := range values {
		retVal, err := ParseInt32(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseInt64Slice(value interface{}, size int) ([]int64, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	retVals := []int64{}
	for _, v := range values {
		retVal, err := ParseInt64(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseBigIntSlice(value interface{}, size int) ([]*big.Int, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	biValues := []*big.Int{}
	for _, v := range values {
		biValue, err := ParseBigInt(v)
		if err != nil {
			return nil, err
		}
		biValues = append(biValues, biValue)
	}
	return biValues, nil
}

func ParseAddress(value interface{}) (ethcmn.Address, error) {
	if value == nil {
		return ethcmn.Address{}, fmt.Errorf("value cannot be nil")
	}
	addrStr := fmt.Sprintf("%v", value)
	if IsAddressType(addrStr) {
		return ethcmn.HexToAddress(addrStr), nil
	} else {
		return ethcmn.Address{}, fmt.Errorf("connot convert %v to address", value)
	}
}

func ParseAddressSlice(value interface{}, size int) ([]ethcmn.Address, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	addrValues := []ethcmn.Address{}
	for _, v := range values {
		addrValue, err := ParseAddress(v)
		if err != nil {
			return nil, err
		}
		addrValues = append(addrValues, addrValue)
	}
	return addrValues, nil
}

func ParseBytesM(value interface{}, m int) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	b := []byte(fmt.Sprintf("%v", value))
	//if len(b) > 1 && b[0] == '"' && b[len(b)-1] == '"' {
	//	b = b[1 : len(b)-1]
	//}
	if len(b) > m {
		return nil, fmt.Errorf("%v is out of range: [%d]byte", value, m)
	}
	return ethcmn.RightPadBytes(b, m), nil
}

func ParseBytes(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	return []byte(fmt.Sprintf("%v", value)), nil
}

func ParseBytesMSlice(value interface{}, m, size int) ([][]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	retVals := [][]byte{}
	for _, v := range values {
		retVal, err := ParseBytesM(v, m)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseBool(value interface{}) (bool, error) {
	if value == nil {
		return false, fmt.Errorf("value cannot be nil")
	}
	b, err := strconv.ParseBool(fmt.Sprintf("%v", value))
	if err != nil {
		return false, err
	}
	return b, nil
}

func ParseBoolSlice(value interface{}, size int) ([]bool, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}
	if size != -1 && size != len(values) {
		return nil, fmt.Errorf("size of %v must be %d", value, size)
	}
	retVals := []bool{}
	for _, v := range values {
		retVal, err := ParseBool(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseString(value interface{}) (string, error) {
	b := []byte(fmt.Sprintf("%v", value))
	//	if len(b) > 1 && b[0] == '"' && b[len(b)-1] == '"' {
	//		b = b[1 : len(b)-1]
	//	}
	return string(b), nil
}

func IsAddressType(str string) bool {
	return strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X")
}

const TRIM_SET = " \n\t"

func JsonParamsToSlc(jsonParam json.RawMessage) ([]interface{}, error) {
	j, err := simplejson.NewJson(jsonParam)
	if err != nil {
		return nil, err
	}
	return j.MustArray(), nil
}

//  @Deprecated, has quotation trouble
func JsonParamsToSlcV1(jsonParam json.RawMessage) ([]interface{}, error) {

	//strJson := strings.TrimRight(strings.TrimLeft(string(jsonParam), "["), "]")
	strJson := strings.Trim(string(jsonParam), TRIM_SET)
	if len(strJson) == 0 || strJson == "[]" || strJson == "null" {
		return nil, nil
	}
	switch strJson[0] {
	case '{':
		// type object
		strJson = strJson[1 : len(strJson)-1]
		strSlc := strings.Split(strJson, ":")
		params := make([]interface{}, 0, len(strSlc))
		//str := ""
		return params, fmt.Errorf("invalid param type")
	case '[':
		// type set
		params, _, err := StringToSlc(strJson)
		return params, err
	default:
		return nil, fmt.Errorf("parse fail")
	}
}

// continue...TODO
func StringToSlc(jsonParam string) ([]interface{}, int, error) {
	strJson := strings.Trim(string(jsonParam), TRIM_SET)
	if len(strJson) < 2 {
		return nil, 0, nil
	}
	baseidx := strings.Index(jsonParam, strJson)
	var count int
	if strJson[0] == '[' {
		strJson = strJson[1:]
		count = 1
	}
	params := make([]interface{}, 0)
	byteStr := []byte(strJson)
	var begin int = -1
	for i := 0; i < len(byteStr); i++ {
		switch byteStr[i] {
		case '[':
			subParams, length, err := StringToSlc(string(strJson[i:]))
			if err != nil {
				return nil, 0, err
			}
			params = append(params, interface{}(subParams))
			i += length + 1
			begin = -1
		case ']':
			count--
			if count < 0 {
				return nil, 0, fmt.Errorf("\"[\" \"]\" mismatch")
			}
			if p := _checkAndTransferToParam(byteStr, begin, i); p != nil {
				params = append(params, p)
			}
			return params, baseidx + i + 1, nil

		case ',':
			if p := _checkAndTransferToParam(byteStr, begin, i); p != nil {
				params = append(params, p)
			}
			begin = -1
		default:
			if i == len(byteStr)-1 {
				params = append(params, _checkAndTransferToParam(byteStr, begin, i))
			}
			if begin < 0 {
				begin = i
			}
		}

	}
	return params, len(jsonParam), nil
}

func _checkAndTransferToParam(bts []byte, begin, end int) interface{} {
	if begin < 0 {
		return nil
	}
	// trust params begin & end
	ret := strings.Trim(string(bts[begin:end]), TRIM_SET)
	if strings.HasPrefix(ret, "\"0x") {
		// address drop '"'s
		return ret[1 : len(ret)-1]
	}
	return ret
}

func IntTo4Byte(num int) []byte {
	var buf bytes.Buffer
	BinWrite(&buf, uint32(num))
	return buf.Bytes()
}

func BinRead(buf *bytes.Buffer, data interface{}) {
	binary.Read(buf, binary.LittleEndian, data)
}

func BinWrite(buf *bytes.Buffer, data interface{}) {
	binary.Write(buf, binary.LittleEndian, data)
}
