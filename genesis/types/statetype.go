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
	"math/big"

	"github.com/dappledger/AnnChain/ann-module/xlib"

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
)

// Account represents a contract or basic ethereum account.
type Account interface {
	SubBalance(amount *big.Int, log string)
	AddBalance(amount *big.Int, log string)
	SetBalance(*big.Int, string)
	SetNonce(uint64)
	Balance() *big.Int
	Address() ethcmn.Address
	ReturnGas(*big.Int)
	ForEachStorage(cb func(key, value ethcmn.Hash) bool)
	Value() *big.Int
	SetCode(ethcmn.Hash, []byte)
}

type InflationVotes struct {
	Votes *big.Int
	Dest  ethcmn.Address
}

func (i *InflationVotes) Key() xlib.SortKey {
	return &i.Dest
}

func (i *InflationVotes) Less(data xlib.Sortable) bool {
	if iv, ok := data.(*InflationVotes); ok {
		return i.Votes.Cmp(iv.Votes) > 0
	}
	return false
}
