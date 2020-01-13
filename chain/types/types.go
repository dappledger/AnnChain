// Copyright © 2017 ZhongAn Technology
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

import "github.com/dappledger/AnnChain/eth/common"

type (
	// LastBlockInfo used for crash recover
	LastBlockInfo struct {
		Height  int64  `json:"height"`
		AppHash []byte `json:"apphash"`
	}

	// Receipt used to record tx execute result
	Receipt struct {
		TxHash  common.Hash
		Height  uint64
		Success bool
		Message string
	}

	KV struct {
		Key   []byte
		Value []byte
	}

	KVs []*KV

	QueryType = byte
)

const (
	APIQueryTx                          = iota
	QueryType_Contract        QueryType = 0
	QueryType_Nonce           QueryType = 1
	QueryType_Balance         QueryType = 2
	QueryType_Receipt         QueryType = 3
	QueryType_Existence       QueryType = 4
	QueryType_PayLoad         QueryType = 5
	QueryType_TxRaw           QueryType = 6
	QueryTxLimit              QueryType = 9
	QueryTypeContractByHeight QueryType = 10
	QueryType_Key             QueryType = 11
	QueryType_Key_Prefix      QueryType = 12
	QueryType_Pending_Nonce   QueryType = 13
)

var KVTxType = []byte("kvTx-")
