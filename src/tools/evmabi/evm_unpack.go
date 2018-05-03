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


package evmabi

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"

	"github.com/dappledger/AnnChain/eth/abi"
	"github.com/dappledger/AnnChain/eth/common"
	"github.com/pkg/errors"
)

func LocateMethod(abiInst *abi.ABI, methodID []byte) (*abi.Method, error) {
	for _, m := range abiInst.Methods {
		if bytes.Equal(m.Id(), methodID) {
			return &m, nil
		}
	}
	return nil, errors.New("method id doesn't exist")
}

// reads the integer based on its kind
func ReadInteger(kind reflect.Kind, b []byte) interface{} {
	switch kind {
	case reflect.Uint8:
		return b[len(b)-1]
	case reflect.Uint16:
		return binary.BigEndian.Uint16(b[len(b)-2:])
	case reflect.Uint32:
		return binary.BigEndian.Uint32(b[len(b)-4:])
	case reflect.Uint64:
		return binary.BigEndian.Uint64(b[len(b)-8:])
	case reflect.Int8:
		return int8(b[len(b)-1])
	case reflect.Int16:
		return int16(binary.BigEndian.Uint16(b[len(b)-2:]))
	case reflect.Int32:
		return int32(binary.BigEndian.Uint32(b[len(b)-4:]))
	case reflect.Int64:
		return int64(binary.BigEndian.Uint64(b[len(b)-8:]))
	default:
		return new(big.Int).SetBytes(b)
	}
}

// reads a bool
func ReadBool(word []byte) (bool, error) {
	for _, b := range word[:31] {
		if b != 0 {
			return false, errBadBool
		}
	}
	switch word[31] {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, errBadBool
	}
}

// A function type is simply the address with the function selection signature at the end.
// This enforces that standard by always presenting it as a 24-array (address + sig = 24 bytes)
func ReadFunctionType(t abi.Type, word []byte) (funcTy [24]byte, err error) {
	if t.T != abi.FunctionTy {
		return [24]byte{}, fmt.Errorf("abi: invalid type in call to make function type byte array")
	}
	if garbage := binary.BigEndian.Uint64(word[24:32]); garbage != 0 {
		err = fmt.Errorf("abi: got improperly encoded function type, got %v", word)
	} else {
		copy(funcTy[:], word[0:24])
	}
	return
}

// through reflection, creates a fixed array to be read from
func ReadFixedBytes(t abi.Type, word []byte) (interface{}, error) {
	if t.T != abi.FixedBytesTy {
		return nil, fmt.Errorf("abi: invalid type in call to make fixed byte array")
	}
	// convert
	array := reflect.New(t.Type).Elem()

	reflect.Copy(array, reflect.ValueOf(word[0:t.Size]))
	return array.Interface(), nil

}

// iteratively unpack elements
func ForEachUnpack(t abi.Type, output []byte, start, size int) (interface{}, error) {
	if start+32*size > len(output) {
		return nil, fmt.Errorf("abi: cannot marshal in to go array: offset %d would go over slice boundary (len=%d)", len(output), start+32*size)
	}

	// this value will become our slice or our array, depending on the type
	var refSlice reflect.Value
	slice := output[start : start+size*32]

	if t.IsSlice {
		// declare our slice

	} else if t.IsArray {
		// declare our array
		refSlice = reflect.New(t.Type).Elem()
	} else {
		return nil, fmt.Errorf("abi: invalid type in array/slice unpacking stage")
	}

	for i, j := start, 0; j*32 < len(slice); i, j = i+32, j+1 {
		// this corrects the arrangement so that we get all the underlying array values
		if t.Elem.IsArray && j != 0 {
			i = start + t.Elem.Size*32*j
		}
		inter, err := ToGoType(i, *t.Elem, output)
		if err != nil {
			return nil, err
		}
		// append the item to our reflect slice
		refSlice.Index(j).Set(reflect.ValueOf(inter))
	}

	// return the interface
	return refSlice.Interface(), nil
}

func RequireLengthPrefix(t *abi.Type) bool {
	return t.T != abi.FixedBytesTy && (t.T == abi.StringTy || t.T == abi.BytesTy || t.IsSlice)
}

func ToGoType(index int, t abi.Type, data []byte) (interface{}, error) {
	var (
		begin, length int
		err           error
		packed        []byte

		cursor = index * 32
	)

	if cursor+32 > len(data) {
		return nil, errors.Wrap(errors.New("length insufficient"), "[ToGoType]")
	}

	// check for a slice type, copied from Method.pack
	if RequireLengthPrefix(&t) {
		begin, length, err = LengthPrefixPointsTo(cursor, data)
		if err != nil {
			return nil, err
		}

		if t.T == abi.StringTy {
			return string(data[begin : begin+length]), nil
		} else if t.T == abi.BytesTy {
			return data[begin : begin+length], nil
		} else {
			return ToGoSlice(index, t, data)
		}
	} else {
		packed = data[cursor : cursor+32]
		switch t.T {
		case abi.IntTy, abi.UintTy:
			return ReadInteger(t.Kind, packed), nil
		case abi.BoolTy:
			return ReadBool(packed)
		case abi.AddressTy:
			return common.BytesToAddress(packed), nil
		case abi.HashTy:
			return common.BytesToHash(packed), nil
		case abi.FixedBytesTy:
			return ReadFixedBytes(t, packed)
		case abi.FunctionTy:
			return ReadFunctionType(t, packed)
		default:
			return nil, errors.Wrap(errors.New("unsupported abi.Type"), "[upackElement]")
		}
	}
}

// // interprets a 32 byte slice as an offset and then determines which indice to look to decode the type.
func LengthPrefixPointsTo(cursor int, output []byte) (start int, length int, err error) {
	offset := int(binary.BigEndian.Uint64(output[cursor+24 : cursor+32]))
	if offset+32 > len(output) {
		return 0, 0, fmt.Errorf("abi: cannot marshal in to go slice: offset %d would go over slice boundary (len=%d)", len(output), offset+32)
	}
	length = int(binary.BigEndian.Uint64(output[offset+24:offset+32])) * 32
	if offset+32+length > len(output) {
		return 0, 0, fmt.Errorf("abi: cannot marshal in to go type: length insufficient %d require %d", len(output), offset+32+length)
	}
	start = offset + 32

	//fmt.Printf("LENGTH PREFIX INFO: \nsize: %v\noffset: %v\nstart: %v\n", length, offset, start)
	return
}

func ToGoSlice(i int, t abi.Type, output []byte) (interface{}, error) {
	index := i * 32
	// The slice must, at very least be large enough for the index+32 which is exactly the size required
	// for the [offset in output, size of offset].
	if index+32 > len(output) {
		return nil, fmt.Errorf("abi: cannot marshal in to go slice: insufficient size output %d require %d", len(output), index+32)
	}
	elem := t.Elem

	// first we need to create a slice of the type
	var refSlice reflect.Value
	switch elem.T {
	case abi.IntTy, abi.UintTy, abi.BoolTy:
		// create a new reference slice matching the element type
		switch t.Kind {
		case reflect.Bool:
			refSlice = reflect.ValueOf([]bool(nil))
		case reflect.Uint8:
			refSlice = reflect.ValueOf([]uint8(nil))
		case reflect.Uint16:
			refSlice = reflect.ValueOf([]uint16(nil))
		case reflect.Uint32:
			refSlice = reflect.ValueOf([]uint32(nil))
		case reflect.Uint64:
			refSlice = reflect.ValueOf([]uint64(nil))
		case reflect.Int8:
			refSlice = reflect.ValueOf([]int8(nil))
		case reflect.Int16:
			refSlice = reflect.ValueOf([]int16(nil))
		case reflect.Int32:
			refSlice = reflect.ValueOf([]int32(nil))
		case reflect.Int64:
			refSlice = reflect.ValueOf([]int64(nil))
		default:
			refSlice = reflect.ValueOf([]*big.Int(nil))
		}
	case abi.AddressTy: // address must be of slice Address
		refSlice = reflect.ValueOf([]common.Address(nil))
	case abi.HashTy: // hash must be of slice hash
		refSlice = reflect.ValueOf([]common.Hash(nil))
	case abi.FixedBytesTy:
		refSlice = reflect.ValueOf([][]byte(nil))
	default: // no other types are supported
		return nil, fmt.Errorf("abi: unsupported slice type %v", elem.T)
	}

	var slice []byte
	var size int
	var offset int
	if t.IsSlice {
		// get the offset which determines the start of this array ...
		offset = int(binary.BigEndian.Uint64(output[index+24 : index+32]))
		if offset+32 > len(output) {
			return nil, fmt.Errorf("abi: cannot marshal in to go slice: offset %d would go over slice boundary (len=%d)", len(output), offset+32)
		}

		slice = output[offset:]
		// ... starting with the size of the array in elements ...
		size = int(binary.BigEndian.Uint64(slice[24:32]))
		slice = slice[32:]
		// ... and make sure that we've at the very least the amount of bytes
		// available in the buffer.
		if size*32 > len(slice) {
			return nil, fmt.Errorf("abi: cannot marshal in to go slice: insufficient size output %d require %d", len(output), offset+32+size*32)
		}

		// reslice to match the required size
		slice = slice[:size*32]
	} else if t.IsArray {
		//get the number of elements in the array
		size = t.SliceSize

		//check to make sure array size matches up
		if index+32*size > len(output) {
			return nil, fmt.Errorf("abi: cannot marshal in to go array: offset %d would go over slice boundary (len=%d)", len(output), index+32*size)
		}
		//slice is there for a fixed amount of times
		slice = output[index : index+size*32]
	}

	for i := 0; i < size; i++ {
		var (
			inter        interface{}             // interface type
			returnOutput = slice[i*32 : i*32+32] // the return output
			err          error
		)
		// set inter to the correct type (cast)
		switch elem.T {
		case abi.IntTy, abi.UintTy:
			inter = ReadInteger(t.Kind, returnOutput)
		case abi.BoolTy:
			inter, err = ReadBool(returnOutput)
			if err != nil {
				return nil, err
			}
		case abi.AddressTy:
			inter = common.BytesToAddress(returnOutput)
		case abi.HashTy:
			inter = common.BytesToHash(returnOutput)
		case abi.FixedBytesTy:
			inter = returnOutput
		}
		// append the item to our reflect slice
		refSlice = reflect.Append(refSlice, reflect.ValueOf(inter))
	}

	// return the interface
	return refSlice.Interface(), nil
}
