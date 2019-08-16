// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
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
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sort"
	"unsafe"
)

func CheckItfcNil(itfc interface{}) bool {
	d := (*struct {
		itab uintptr
		data uintptr
	})(unsafe.Pointer(&itfc))
	return d.data == 0
}

func BinRead(reader io.Reader, data interface{}) error {
	return binary.Read(reader, binary.BigEndian, data)
}

func BinWrite(writer io.Writer, data interface{}) error {
	return binary.Write(writer, binary.BigEndian, data)
}

func Uint32Bytes(num uint32) []byte {
	var bb bytes.Buffer
	BinWrite(&bb, num)
	return bb.Bytes()
}

// ErrUnexpectedEOF when it is the end
func ReadBytes(reader io.Reader) ([]byte, error) {
	byLen, err := ReadVarint(reader)
	if err != nil {
		return nil, err
	}
	bys := make([]byte, byLen)
	err = BinRead(reader, &bys)
	return bys, err
}

func WriteBytes(writer io.Writer, bys []byte) error {
	err := WriteVarint(writer, len(bys))
	if err != nil {
		return err
	}
	return BinWrite(writer, bys)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src) //
}

func uvarintSize(i uint64) int {
	if i == 0 {
		return 0
	}
	if i < 1<<8 {
		return 1
	}
	if i < 1<<16 {
		return 2
	}
	if i < 1<<24 {
		return 3
	}
	if i < 1<<32 {
		return 4
	}
	if i < 1<<40 {
		return 5
	}
	if i < 1<<48 {
		return 6
	}
	if i < 1<<56 {
		return 7
	}
	return 8
}

func WriteVarint(w io.Writer, i int) error {
	var negate = false
	if i < 0 {
		negate = true
		i = -i
	}
	var size = uvarintSize(uint64(i))
	var err error
	if negate {
		// e.g. 0xF1 for a single negative byte
		err = BinWrite(w, uint8(size+0xF0))
	} else {
		err = BinWrite(w, uint8(size))
	}
	if err != nil {
		return err
	}
	if size > 0 {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		err = BinWrite(w, buf[(8-size):])
	}
	return err
}

func ReadByte(r io.Reader) (byte, error) {
	var buf [1]byte
	_, err := io.ReadFull(r, buf[:])
	return buf[0], err
}

func ReadUint8(r io.Reader) (uint8, error) {
	bys, err := ReadByte(r)
	if err != nil {
		return 0, err
	}
	return uint8(bys), nil
}

func ReadVarint(r io.Reader) (int, error) {
	ui8, err := ReadUint8(r)
	if err != nil {
		return 0, err
	}
	var negate = false
	if (ui8 >> 4) == 0xF {
		negate = true
		ui8 = ui8 & 0x0F
	}
	if ui8 > 8 {
		return 0, errors.New("Varint overflow")
	}
	if ui8 == 0 {
		if negate {
			err = errors.New("Varint does not allow negative zero")
		}
		return 0, err
	}
	var buf [8]byte
	_, err = io.ReadFull(r, buf[(8-ui8):])
	if err != nil {
		return 0, err
	}
	var i = int(binary.BigEndian.Uint64(buf[:]))
	if negate {
		return -i, nil
	}
	return i, nil
}

type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Int64Slice) Sort() { sort.Sort(p) }
