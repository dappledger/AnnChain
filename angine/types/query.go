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
	"time"

	"github.com/dappledger/AnnChain/module/xlib/def"
)

const (
	// angine takes query id from 0x01 to 0x2F
	QueryTxExecution = 0x01
)

type TxExecutionResult struct {
	Height        def.INT   `json:"height"`
	BlockHash     []byte    `json:"blockhash"`
	BlockTime     time.Time `json:"blocktime"`
	ValidatorHash []byte    `json:"validatorhash"`
}

func (i *TxExecutionResult) ToBytes() ([]byte, error) {
	return json.Marshal(i)
}

func (i *TxExecutionResult) FromBytes(bytes []byte) error {
	return json.Unmarshal(bytes, i)
}
