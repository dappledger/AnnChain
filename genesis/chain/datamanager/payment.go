package datamanager

import (
	//"encoding/json"
	//"time"

	"github.com/dappledger/AnnChain/genesis/chain/database"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
)

// QueryActionData query actions using paging pattern
func (m *DataManager) QueryPaymentData(q types.ActionsQuery) ([]types.ActionData, error) {
	// Query without account
	if q.Account == types.ZERO_ADDRESS {
		return m.queryPaymentDataCommon(q, nil, nil)
	}

	// Query with account
	return m.queryPaymentDataComonUnion(q, &q.Account, &q.Account)
}

func (m *DataManager) queryPaymentDataComonUnion(q types.ActionsQuery, from, to *ethcmn.Address) ([]types.ActionData, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	var wheres [][]database.Where

	where1 := []database.Where{
		database.Where{Name: "1", Value: 1},
	}
	if q.Typei != types.TypeiUndefined {
		where1 = append(where1, database.Where{Name: "typei", Value: q.Typei})
	}
	if from != nil && *from != types.ZERO_ADDRESS {
		where1 = append(where1, database.Where{Name: "fromaccount", Value: from.Hex()})
	}

	if q.TxHash != types.ZERO_HASH {
		where1 = append(where1, database.Where{Name: "txhash", Value: q.TxHash.Hex()})
	}
	if q.Begin != 0 {
		where1 = append(where1, database.Where{Name: "createat", Value: q.Begin, Op: ">="})
	}
	if q.End != 0 {
		where1 = append(where1, database.Where{Name: "createat", Value: q.End, Op: "<="})
	}

	wheres = append(wheres, where1)

	where2 := []database.Where{
		database.Where{Name: "1", Value: 1},
	}
	if q.Typei != types.TypeiUndefined {
		where2 = append(where2, database.Where{Name: "typei", Value: q.Typei})
	}

	if q.TxHash != types.ZERO_HASH {
		where2 = append(where2, database.Where{Name: "txhash", Value: q.TxHash.Hex()})
	}
	if q.Begin != 0 {
		where2 = append(where2, database.Where{Name: "createat", Value: q.Begin, Op: ">="})
	}
	if q.End != 0 {
		where2 = append(where2, database.Where{Name: "createat", Value: q.End, Op: "<="})
	}
	if to != nil && *to != types.ZERO_ADDRESS {
		where2 = append(where2, database.Where{Name: "toaccount", Value: to.Hex()})
	}

	wheres = append(wheres, where2)

	orderT, err := database.MakeOrder(q.Order, "actionid")
	if err != nil {
		return nil, err
	}
	paging := database.MakePaging("actionid", q.Cursor, q.Limit)

	var result []database.Action
	err = m.qdb.SelectRowsUnion(database.TableActions, wheres, orderT, paging, &result)
	if err != nil {
		return nil, err
	}

	var res []types.ActionData
	for _, r := range result {
		ad := types.ActionData{
			ActionID: r.ActionID,
			JSONData: r.JData,
		}
		res = append(res, ad)
	}

	return res, nil
}
func (m *DataManager) queryPaymentDataCommon(q types.ActionsQuery, from, to *ethcmn.Address) ([]types.ActionData, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "1", Value: 1},
	}
	if q.Typei != types.TypeiUndefined {
		where = append(where, database.Where{Name: "typei", Value: q.Typei})
	}
	if from != nil && *from != types.ZERO_ADDRESS {
		where = append(where, database.Where{Name: "fromaccount", Value: from.Hex()})
	}
	if to != nil && *to != types.ZERO_ADDRESS {
		where = append(where, database.Where{Name: "toaccount", Value: to.Hex()})
	}
	if q.TxHash != types.ZERO_HASH {
		where = append(where, database.Where{Name: "txhash", Value: q.TxHash.Hex()})
	}
	if q.Begin != 0 {
		where = append(where, database.Where{Name: "createat", Value: q.Begin, Op: ">="})
	}
	if q.End != 0 {
		where = append(where, database.Where{Name: "createat", Value: q.End, Op: "<="})
	}

	orderT, err := database.MakeOrder(q.Order, "actionid")
	if err != nil {
		return nil, err
	}
	paging := database.MakePaging("actionid", q.Cursor, q.Limit)

	var result []database.Action
	err = m.qdb.SelectRows(database.TableActions, where, orderT, paging, &result)
	if err != nil {
		return nil, err
	}

	var res []types.ActionData
	for _, r := range result {
		ad := types.ActionData{
			ActionID: r.ActionID,
			JSONData: r.JData,
		}
		res = append(res, ad)
	}

	return res, nil
}
