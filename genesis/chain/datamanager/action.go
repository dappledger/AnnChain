package datamanager

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/dappledger/AnnChain/genesis/chain/database"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
)

func (m *DataManager) PrepareAction() (*sql.Stmt, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}
	fields := []database.Feild{
		database.Feild{Name: "typei"},
		database.Feild{Name: "type"},
		database.Feild{Name: "height"},
		database.Feild{Name: "txhash"},
		database.Feild{Name: "fromaccount"},
		database.Feild{Name: "toaccount"},
		database.Feild{Name: "createat"},
		database.Feild{Name: "jdata"},
	}
	return m.qdb.Prepare(database.TableActions, fields)
}

func (m *DataManager) AddActionDataStmt(stmt *sql.Stmt, o types.ActionObject) (err error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	if o.GetActionBase().CreateAt == 0 {
		o.GetActionBase().CreateAt = uint64(time.Now().UnixNano())
	}

	jd, err := json.Marshal(o)
	if err != nil {
		return err
	}

	fields := []database.Feild{
		database.Feild{Name: "typei", Value: int(o.GetActionBase().Typei)},
		database.Feild{Name: "type", Value: o.GetActionBase().Type},
		database.Feild{Name: "height", Value: o.GetActionBase().Height.String()},
		database.Feild{Name: "txhash", Value: o.GetActionBase().TxHash.Hex()},
		database.Feild{Name: "fromaccount", Value: o.GetActionBase().FromAccount.Hex()},
		database.Feild{Name: "toaccount", Value: o.GetActionBase().ToAccount.Hex()},
		database.Feild{Name: "createat", Value: o.GetActionBase().CreateAt},
		database.Feild{Name: "jdata", Value: string(jd)},
	}
	_, err = m.qdb.Excute(stmt, fields)

	return err
}

// AddActionData add action record
func (m *DataManager) AddActionData(o types.ActionObject) (uint64, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	if o.GetActionBase().CreateAt == 0 {
		o.GetActionBase().CreateAt = uint64(time.Now().UnixNano())
	}

	jd, err := json.Marshal(o)
	if err != nil {
		return 0, err
	}

	fields := []database.Feild{
		database.Feild{Name: "typei", Value: int(o.GetActionBase().Typei)},
		database.Feild{Name: "type", Value: o.GetActionBase().Type},
		database.Feild{Name: "height", Value: o.GetActionBase().Height.String()},
		database.Feild{Name: "txhash", Value: o.GetActionBase().TxHash.Hex()},
		database.Feild{Name: "createat", Value: o.GetActionBase().CreateAt},
		database.Feild{Name: "jdata", Value: string(jd)},
	}
	if o.GetActionBase().FromAccount != (ethcmn.Address{}) {
		fields = append(fields, database.Feild{Name: "fromaccount", Value: o.GetActionBase().FromAccount.Hex()})
	}
	if o.GetActionBase().ToAccount != (ethcmn.Address{}) {
		fields = append(fields, database.Feild{Name: "toaccount", Value: o.GetActionBase().ToAccount.Hex()})
	}

	sqlRes, err := m.qdb.Insert(database.TableActions, fields)
	if err != nil {
		return 0, err
	}

	id, err := sqlRes.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

// QueryActionData query actions using paging pattern
func (m *DataManager) QueryActionData(q types.ActionsQuery) ([]types.ActionData, error) {
	// Query without account
	if q.Account == types.ZERO_ADDRESS {
		return m.queryActionDataCommon(q, nil, nil)
	}
	// Query with account
	return m.queryActionDataComonUnion(q, &q.Account, &q.Account)
}

func (m *DataManager) queryActionDataComonUnion(q types.ActionsQuery, from, to *ethcmn.Address) ([]types.ActionData, error) {
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
func (m *DataManager) queryActionDataCommon(q types.ActionsQuery, from, to *ethcmn.Address) ([]types.ActionData, error) {
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
