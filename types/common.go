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

package types

import (
	"encoding/hex"
)

func StringTo32byte(key string) ([32]byte, error) {
	var byte32 [32]byte
	sec, err := hex.DecodeString(key)
	if err != nil {
		return byte32, err
	}
	copy(byte32[:], sec)
	return byte32, nil
}

func StringTo64byte(key string) ([64]byte, error) {
	var byte64 [64]byte
	seckey, err := hex.DecodeString(key)
	if err != nil {
		return byte64, err
	}
	copy(byte64[:], seckey)
	return byte64, nil
}

func StringToAnybyte(key string) ([]byte, error) {
	seckey, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	b := make([]byte, len(seckey))
	copy(b, seckey)
	return b, nil
}

func PrivKeyByteToByte64(bytes []byte) (byte64 [64]byte) {
	if len(bytes) == 0 {
		return
	}
	pkb := bytes[1:]
	copy(byte64[:], pkb)
	return
}

func Byte64Tobyte(bytes64 [64]byte) (bytes []byte) {
	bytes = make([]byte, 64)
	copy(bytes, bytes64[:])
	return
}

func BytesToByte64(bytes []byte) (b64 [64]byte) {
	copy(b64[:], bytes)
	return
}

func Substr(str string, start int, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

func StrToSmallStr(str string) (smallstr string) {
	bytes, _ := hex.DecodeString(str)
	smallstr = hex.EncodeToString(bytes)
	return
}
