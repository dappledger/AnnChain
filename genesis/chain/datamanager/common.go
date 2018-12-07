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
