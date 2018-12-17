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
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/dappledger/AnnChain/ann-module/xlib"

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
)

const TEST_NUM = 10

func TestSortList(t *testing.T) {
	var sl xlib.SortList
	sl.Init(TEST_NUM)
	vslc := make([]InflationVotes, TEST_NUM)
	var i int64
	for i = 0; i < TEST_NUM; i++ {
		vslc[i].Dest = ethcmn.StringToAddress(fmt.Sprintf("%v", i))
		rand.Seed(i)
		vslc[i].Votes = big.NewInt(i + rand.Int63())
		fmt.Println("origin:", vslc[i].Dest, vslc[i].Votes)
		sl.Put(&vslc[i])
	}

	sl.Exec(func(data xlib.Sortable) bool {
		if iv, ok := data.(*InflationVotes); ok {
			fmt.Println(iv.Dest, iv.Votes)
		}
		return true
	})
}
