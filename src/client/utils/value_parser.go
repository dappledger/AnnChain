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


package utils

import (
	"fmt"
	"math/big"
	"strconv"

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
	v, ok := new(big.Int).SetString(fmt.Sprintf("%v", value), 10)
	if !ok {
		return nil, fmt.Errorf("Fail to convert %v to big.Int", value)
	}
	return v, nil
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

func ParseAddress(value interface{}) (common.Address, error) {
	if value == nil {
		return common.Address{}, fmt.Errorf("value cannot be nil")
	}
	return common.HexToAddress(fmt.Sprintf("%v", value)), nil
}

func ParseAddressSlice(value interface{}, size int) ([]common.Address, error) {
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
	addrValues := []common.Address{}
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
	if len(b) > m {
		return nil, fmt.Errorf("%v is out of range: [%d]byte", value, m)
	}
	return common.LeftPadBytes(b, m), nil
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
	return fmt.Sprintf("%s", value), nil
}
