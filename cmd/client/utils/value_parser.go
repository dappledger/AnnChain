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

package utils

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/dappledger/AnnChain/eth/common"
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
			return nil, err
		}
		bi, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return nil, fmt.Errorf("fail to parse %v to big.Int", value)
		}
		return bi, nil
	}
	v, ok := new(big.Int).SetString(strVal, 10)
	if !ok {
		return nil, fmt.Errorf("fail to convert %v to big.Int", value)
	}
	return v, nil
}

func ParseUint8Slice(value interface{}) ([]uint8, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var retVals []uint8
	for _, v := range values {
		retVal, err := ParseUint8(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseUint16Slice(value interface{}) ([]uint16, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var retVals []uint16
	for _, v := range values {
		retVal, err := ParseUint16(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseUint32Slice(value interface{}) ([]uint32, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var retVals []uint32
	for _, v := range values {
		retVal, err := ParseUint32(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseUint64Slice(value interface{}) ([]uint64, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var retVals []uint64
	for _, v := range values {
		retVal, err := ParseUint64(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseInt8Slice(value interface{}) ([]int8, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var retVals []int8
	for _, v := range values {
		retVal, err := ParseInt8(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseInt16Slice(value interface{}) ([]int16, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var retVals []int16
	for _, v := range values {
		retVal, err := ParseInt16(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseInt32Slice(value interface{}) ([]int32, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var retVals []int32
	for _, v := range values {
		retVal, err := ParseInt32(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseInt64Slice(value interface{}) ([]int64, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var retVals []int64
	for _, v := range values {
		retVal, err := ParseInt64(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
}

func ParseBigIntSlice(value interface{}) ([]*big.Int, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	if values, ok := value.([]*big.Int); ok {
		return values, nil
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var biValues []*big.Int
	for _, v := range values {
		biValue, err := ParseBigInt(v)
		if err != nil {
			return nil, err
		}
		biValues = append(biValues, biValue)
	}
	return biValues, nil
}

func ParseAddress(value interface{}) (common.Address, error) {
	if value == nil {
		return common.Address{}, fmt.Errorf("value cannot be nil")
	}
	if addr, ok := value.(common.Address); ok {
		return addr, nil
	}
	return common.HexToAddress(fmt.Sprintf("%v", value)), nil
}

func ParseAddressSlice(value interface{}) ([]common.Address, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var addrValues []common.Address
	for _, v := range values {
		addrValue, err := ParseAddress(v)
		if err != nil {
			return nil, err
		}
		addrValues = append(addrValues, addrValue)
	}
	return addrValues, nil
}

func ParseBytesM(value interface{}, m int) (interface{}, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	b, err := ParseBytes(value)
	if err != nil {
		return b, err
	}
	if len(b) > m {
		return nil, fmt.Errorf("%v is out of range: [%d]byte", value, m)
	}
	b = common.LeftPadBytes(b, m)
	switch m {
	case 8:
		var buf [8]byte
		copy(buf[:], b)
		return buf, nil
	case 16:
		var buf [16]byte
		copy(buf[:], b)
		return buf, nil
	case 32:
		var buf [32]byte
		copy(buf[:], b)
		return buf, nil
	case 64:
		var buf [64]byte
		copy(buf[:], b)
		return buf, nil
	}
	return nil, fmt.Errorf("type(bytes%d) not support", m)
}

func ParseBytes(value interface{}) ([]byte, error) {

	switch reflect.TypeOf(value).Kind() {
	case reflect.String:
		return []byte(common.FromHex(value.(string))), nil
	case reflect.Slice:
		return value.([]byte), nil
	case reflect.Array:
		dstr := fmt.Sprintf("%x", value)
		return []byte(common.FromHex(dstr)), nil
	default:
		return nil, fmt.Errorf("unsupoort value type")
	}
}

func ParseBytesMSlice(value interface{}, m int) (interface{}, error) {

	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}

	values, ok := value.([]interface{})
	if !ok {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice, reflect.Array:
			return value, nil
		default:
			return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
		}
	}

	retVals := make([]interface{}, 0, len(values))
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

func ParseBoolSlice(value interface{}) ([]bool, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	var retVals []bool
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
	return fmt.Sprintf("%s", value), nil
}

func ParseStringSlice(value interface{}) ([]string, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %v to []interface{}", value)
	}

	retVals := []string{}
	for _, v := range values {
		retVal, err := ParseString(v)
		if err != nil {
			return nil, err
		}
		retVals = append(retVals, retVal)
	}
	return retVals, nil
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
