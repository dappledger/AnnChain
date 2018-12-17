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
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
)

const AccDataLength = 1000

// TypeiUndefined when typei undefined, use 127
const TypeiUndefined = 127

type ShowAccount struct {
	Address string                        `json:"address"`
	Balance string                        `json:"balance"`
	Data    map[string]ManageDataCategory `json:"data"`
}

type QueryExRequest struct {
	Data   []byte
	Cursor uint64
	Limit  uint
	Order  string
}

type QueryTxRequest struct {
	TxHash *ethcmn.Hash
	Cursor uint64
	Limit  uint
	Order  string
}

type QueryActionsRequest struct {
	Cursor uint64
	Limit  uint
	Order  string
}

type QueryContractExist struct {
	ByteCode string `json:"byte_code"`
	CodeHash string `json:"code_hash"`
	IsExist  bool   `json:"is_exist"`
}
