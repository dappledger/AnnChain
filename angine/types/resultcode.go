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

type CodeType int32

const (
	// General response codes, 0 ~ 99
	CodeType_OK                     CodeType = 0
	CodeType_InternalError          CodeType = 1
	CodeType_BadNonce               CodeType = 2
	CodeType_InvalidTx              CodeType = 3
	CodeType_LowBalance             CodeType = 4
	CodeType_Timeout                CodeType = 5
	CodeType_NullData               CodeType = 6
	CodeType_DecodingError          CodeType = 7
	CodeType_EncodingError          CodeType = 8
	CodeType_AccDataLengthError     CodeType = 9
	CodeType_AccCategoryLengthError CodeType = 10
	CodeType_JsonError              CodeType = 11

	// Reserved for basecoin, 100 ~ 199
	CodeType_BaseInsufficientFunds CodeType = 101
	CodeType_BaseInvalidInput      CodeType = 102
	CodeType_BaseInvalidSignature  CodeType = 103
	CodeType_BaseUnknownAddress    CodeType = 104
	CodeType_WrongRLP              CodeType = 105
	CodeType_SaveFailed            CodeType = 106

	CodeType_BadLimit  CodeType = 401
	CodeType_BadPrice  CodeType = 402
	CodeType_BadAmount CodeType = 403
)

var CodeType_name = map[int32]string{
	0:   "OK",
	1:   "InternalError",
	2:   "BadNonce",
	3:   "InvalidTx",
	4:   "LowBalance",
	5:   "RequestTimeout",
	6:   "EmptyData",
	7:   "DecodingError",
	8:   "EncodingError",
	9:   "AccDataLengthError",
	10:  "AccCategoryLengthError",
	11:  "JsonError",
	101: "BaseInsufficientFunds",
	102: "BaseInvalidInput",
	103: "BaseInvalidSignature",
	104: "BaseUnknownAddress",
	105: "WrongRLP",
	106: "SaveFailed",
	401: "CodeType_BadLimit",
	402: "CodeType_BadPrice",
	403: "CodeType_BadAmount",
}
var CodeType_value = map[string]int32{
	"OK":                     0,
	"InternalError":          1,
	"BadNonce":               2,
	"InvalidTx":              3,
	"LowBalance":             4,
	"RequestTimeout":         5,
	"EmptyData":              6,
	"DecodingError":          7,
	"EncodingError":          8,
	"AccDataLengthError":     9,
	"AccCategoryLengthError": 10,
	"JsonError":              11,
	"BaseInsufficientFunds":  101,
	"BaseInvalidInput":       102,
	"BaseInvalidSignature":   103,
	"BaseUnknownAddress":     104,
	"WrongRLP":               105,
	"SaveFailed":             106,
	"CodeType_BadLimit":      401,
	"CodeType_BadPrice":      402,
	"CodeType_BadAmount":     403,
}

func (x CodeType) String() string {
	return CodeType_name[int32(x)]
}
