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
	CodeType_OK CodeType = 0
	// General response codes, 0 ~ 99
	CodeType_InternalError     CodeType = 1
	CodeType_EncodingError     CodeType = 2
	CodeType_BadNonce          CodeType = 3
	CodeType_Unauthorized      CodeType = 4
	CodeType_InsufficientFunds CodeType = 5
	CodeType_UnknownRequest    CodeType = 6
	CodeType_InvalidTx         CodeType = 7
	// Reserved for basecoin, 100 ~ 199
	CodeType_BaseDuplicateAddress     CodeType = 101
	CodeType_BaseEncodingError        CodeType = 102
	CodeType_BaseInsufficientFees     CodeType = 103
	CodeType_BaseInsufficientFunds    CodeType = 104
	CodeType_BaseInsufficientGasPrice CodeType = 105
	CodeType_BaseInvalidInput         CodeType = 106
	CodeType_BaseInvalidOutput        CodeType = 107
	CodeType_BaseInvalidPubKey        CodeType = 108
	CodeType_BaseInvalidSequence      CodeType = 109
	CodeType_BaseInvalidSignature     CodeType = 110
	CodeType_BaseUnknownAddress       CodeType = 111
	CodeType_BaseUnknownPubKey        CodeType = 112
	CodeType_BaseUnknownPlugin        CodeType = 113
	CodeType_WrongRLP                 CodeType = 114
	CodeType_SaveFailed               CodeType = 115
	// Reserved for governance, 200 ~ 299
	CodeType_GovUnknownEntity      CodeType = 201
	CodeType_GovUnknownGroup       CodeType = 202
	CodeType_GovUnknownProposal    CodeType = 203
	CodeType_GovDuplicateGroup     CodeType = 204
	CodeType_GovDuplicateMember    CodeType = 205
	CodeType_GovDuplicateProposal  CodeType = 206
	CodeType_GovDuplicateVote      CodeType = 207
	CodeType_GovInvalidMember      CodeType = 208
	CodeType_GovInvalidVote        CodeType = 209
	CodeType_GovInvalidVotingPower CodeType = 210
)

var CodeType_name = map[int32]string{
	0:   "OK",
	1:   "InternalError",
	2:   "EncodingError",
	3:   "BadNonce",
	4:   "Unauthorized",
	5:   "InsufficientFunds",
	6:   "UnknownRequest",
	101: "BaseDuplicateAddress",
	102: "BaseEncodingError",
	103: "BaseInsufficientFees",
	104: "BaseInsufficientFunds",
	105: "BaseInsufficientGasPrice",
	106: "BaseInvalidInput",
	107: "BaseInvalidOutput",
	108: "BaseInvalidPubKey",
	109: "BaseInvalidSequence",
	110: "BaseInvalidSignature",
	111: "BaseUnknownAddress",
	112: "BaseUnknownPubKey",
	113: "BaseUnknownPlugin",
	114: "WrongRLP",
	115: "SaveFailed",
	201: "GovUnknownEntity",
	202: "GovUnknownGroup",
	203: "GovUnknownProposal",
	204: "GovDuplicateGroup",
	205: "GovDuplicateMember",
	206: "GovDuplicateProposal",
	207: "GovDuplicateVote",
	208: "GovInvalidMember",
	209: "GovInvalidVote",
	210: "GovInvalidVotingPower",
}
var CodeType_value = map[string]int32{
	"OK":                       0,
	"InternalError":            1,
	"EncodingError":            2,
	"BadNonce":                 3,
	"Unauthorized":             4,
	"InsufficientFunds":        5,
	"UnknownRequest":           6,
	"BaseDuplicateAddress":     101,
	"BaseEncodingError":        102,
	"BaseInsufficientFees":     103,
	"BaseInsufficientFunds":    104,
	"BaseInsufficientGasPrice": 105,
	"BaseInvalidInput":         106,
	"BaseInvalidOutput":        107,
	"BaseInvalidPubKey":        108,
	"BaseInvalidSequence":      109,
	"BaseInvalidSignature":     110,
	"BaseUnknownAddress":       111,
	"BaseUnknownPubKey":        112,
	"BaseUnknownPlugin":        113,
	"WrongRLP":                 114,
	"SaveFailed":               115,
	"GovUnknownEntity":         201,
	"GovUnknownGroup":          202,
	"GovUnknownProposal":       203,
	"GovDuplicateGroup":        204,
	"GovDuplicateMember":       205,
	"GovDuplicateProposal":     206,
	"GovDuplicateVote":         207,
	"GovInvalidMember":         208,
	"GovInvalidVote":           209,
	"GovInvalidVotingPower":    210,
}

func (x CodeType) String() string {
	return CodeType_name[int32(x)]
}
