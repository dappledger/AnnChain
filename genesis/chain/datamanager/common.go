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

package datamanager

import (
	"github.com/dappledger/AnnChain/genesis/chain/database"
	"github.com/dappledger/AnnChain/genesis/types"
)

func processQueryBase(qb *types.QueryBase, where []database.Where,
	timeField, orderField, pagingField string) (order *database.Order, paging *database.Paging, err error) {
	if qb.Begin != 0 {
		where = append(where, database.Where{Name: timeField, Value: qb.Begin, Op: ">="})
	}
	if qb.End != 0 {
		where = append(where, database.Where{Name: timeField, Value: qb.End, Op: "<="})
	}
	order, err = database.MakeOrder(qb.Order, orderField)
	if err != nil {
		return nil, nil, err
	}
	paging = database.MakePaging(pagingField, qb.Cursor, qb.Limit)

	return
}
