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
	"bytes"
	"encoding/json"
	"time"
)

type SigInfo struct {
	PubKey    []byte `json:"pubkey"`
	Signature []byte `json:"signature"`
}

type AdminOPCmd struct {
	CmdType  string    `json:"cmdtype"` //type for what kind of adminOp
	Msg      []byte    `json:"msg"`
	SelfSign []byte    `json:"sigs"`
	Time     time.Time `json:"time"`
	Nonce    uint64    `json:"nonce"`
	SInfos   []SigInfo `json:"siginfos"`
}

func (cmd *AdminOPCmd) LoadMsg(o interface{}) error {
	var err error
	if cmd.Msg, err = json.Marshal(o); err != nil {
		return err
	}
	return nil
}

func (cmd *AdminOPCmd) ExtractMsg(o interface{}) (interface{}, error) {
	if err := json.Unmarshal(cmd.Msg, o); err != nil {
		return nil, err
	}
	return o, nil
}

type CmdType string

const (
	AdminOpChangeValidator = "changeValidator"
)

var (
	AdminTag = []byte("zaop")
)

func TagAdminOPTx(tx []byte) []byte {
	return WrapTx(AdminTag, tx)
}

func IsAdminOP(tx []byte) bool {
	return bytes.HasPrefix(tx, AdminTag)
}

type AdminVoteResult struct {
	Result    []byte
	PubKey    []byte
	Signature []byte
}
