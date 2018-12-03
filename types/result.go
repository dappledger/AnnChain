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
	"encoding/json"
)

// CONTRACT: a zero Result is OK.
type Result struct {
	Code CodeType `json:"Code"`
	Data []byte   `json:"Data"`
	Log  string   `json:"Log"` // Can be non-deterministic
}

type NewRPCResult struct {
	Code CodeType    `json:"Code"`
	Data interface{} `json:"Data"`
	Log  string      `json:"Log"` // Can be non-deterministic
}

type NewRoundResult struct {
}

type CommitResult struct {
	AppHash      []byte
	ReceiptsHash []byte
}

type ExecuteInvalidTx struct {
	Bytes []byte
	Error error
}

type ExecuteResult struct {
	ValidTxs   [][]byte
	InvalidTxs []ExecuteInvalidTx
	Error      error
}

func NewResult(code CodeType, data []byte, log string) Result {
	return Result{
		Code: code,
		Data: data,
		Log:  log,
	}
}

func (res Result) ToJSON() string {
	j, err := json.Marshal(res)
	if err != nil {
		return res.Log
	}
	return string(j)
}

func (res *Result) FromJSON(j string) *Result {
	err := json.Unmarshal([]byte(j), res)
	if err != nil {
		res.Code = CodeType_InternalError
		res.Log = j
	}
	return res
}

func (res Result) IsOK() bool {
	return res.Code == CodeType_OK
}

func (res Result) IsErr() bool {
	return res.Code != CodeType_OK
}

func (res Result) Error() string {
	// return fmt.Sprintf("{code:%v, data:%X, log:%v}", res.Code, res.Data, res.Log)
	return res.ToJSON()
}

func (res Result) String() string {
	// return fmt.Sprintf("{code:%v, data:%X, log:%v}", res.Code, res.Data, res.Log)
	return res.ToJSON()
}

func (res Result) PrependLog(log string) Result {
	return Result{
		Code: res.Code,
		Data: res.Data,
		Log:  log + ";" + res.Log,
	}
}

func (res Result) AppendLog(log string) Result {
	return Result{
		Code: res.Code,
		Data: res.Data,
		Log:  res.Log + ";" + log,
	}
}

func (res Result) SetLog(log string) Result {
	return Result{
		Code: res.Code,
		Data: res.Data,
		Log:  log,
	}
}

func (res Result) SetData(data []byte) Result {
	return Result{
		Code: res.Code,
		Data: data,
		Log:  res.Log,
	}
}

//----------------------------------------

// NOTE: if data == nil and log == "", same as zero Result.
func NewResultOK(data []byte, log string) Result {
	return Result{
		Code: CodeType_OK,
		Data: data,
		Log:  log,
	}
}

func NewRpcResultOK(data interface{}, log string) NewRPCResult {
	return NewRPCResult{
		Code: CodeType_OK,
		Data: data,
		Log:  log,
	}
}

func NewError(code CodeType, log string) Result {
	return Result{
		Code: code,
		Log:  log,
	}
}

func NewRpcError(code CodeType, log string) NewRPCResult {
	return NewRPCResult{
		Code: code,
		Log:  log,
	}
}
