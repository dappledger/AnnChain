package datamanager

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/dappledger/AnnChain/genesis/chain/database"
	"github.com/dappledger/AnnChain/genesis/types"
)

func (m *DataManager) PrepareEffect() (*sql.Stmt, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}
	fields := []database.Feild{
		database.Feild{Name: "typei"},
		database.Feild{Name: "type"},
		database.Feild{Name: "height"},
		database.Feild{Name: "txhash"},
		database.Feild{Name: "actionid"},
		database.Feild{Name: "account"},
		database.Feild{Name: "createat"},
		database.Feild{Name: "jdata"},
	}
	return m.qdb.Prepare(database.TableEffects, fields)
}

func (m *DataManager) AddEffectDataStmt(stmt *sql.Stmt, o types.EffectObject) (err error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	if o.GetEffectBase().CreateAt == 0 {
		o.GetEffectBase().CreateAt = uint64(time.Now().UnixNano())
	}

	jd, err := json.Marshal(o)
	if err != nil {
		return err
	}

	fields := []database.Feild{
		database.Feild{Name: "typei", Value: int(o.GetEffectBase().Typei)},
		database.Feild{Name: "type", Value: o.GetEffectBase().Type},
		database.Feild{Name: "height", Value: o.GetEffectBase().Height.String()},
		database.Feild{Name: "txhash", Value: o.GetEffectBase().TxHash.Hex()},
		database.Feild{Name: "actionid", Value: o.GetEffectBase().ActionID},
		database.Feild{Name: "account", Value: o.GetEffectBase().Account.Hex()},
		database.Feild{Name: "createat", Value: o.GetEffectBase().CreateAt},
		database.Feild{Name: "jdata", Value: string(jd)},
	}

	_, err = m.qdb.Excute(stmt, fields)

	return err
}

// AddEffectData insert effect record
func (m *DataManager) AddEffectData(o types.EffectObject) (uint64, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	if o.GetEffectBase().CreateAt == 0 {
		o.GetEffectBase().CreateAt = uint64(time.Now().UnixNano())
	}

	jd, err := json.Marshal(o)
	if err != nil {
		return 0, err
	}

	fields := []database.Feild{
		database.Feild{Name: "typei", Value: int(o.GetEffectBase().Typei)},
		database.Feild{Name: "type", Value: o.GetEffectBase().Type},
		database.Feild{Name: "height", Value: o.GetEffectBase().Height.String()},
		database.Feild{Name: "txhash", Value: o.GetEffectBase().TxHash.Hex()},
		database.Feild{Name: "actionid", Value: o.GetEffectBase().ActionID},
		database.Feild{Name: "account", Value: o.GetEffectBase().Account.Hex()},
		database.Feild{Name: "createat", Value: o.GetEffectBase().CreateAt},
		database.Feild{Name: "jdata", Value: string(jd)},
	}

	sqlRes, err := m.qdb.Insert(database.TableEffects, fields)
	if err != nil {
		return 0, err
	}

	id, err := sqlRes.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

func (m *DataManager) QueryEffectData(q types.EffectsQuery) ([]types.EffectData, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "1", Value: 1},
	}

	if q.Account != types.ZERO_ADDRESS {
		where = append(where, database.Where{Name: "account", Value: q.Account.Hex()})
	}
	if q.TxHash != types.ZERO_HASH {
		where = append(where, database.Where{Name: "txhash", Value: q.TxHash.Hex()})
	}
	if q.Typei != types.TypeiUndefined {
		where = append(where, database.Where{Name: "typei", Value: q.Typei})
	}
	if q.Begin != 0 {
		where = append(where, database.Where{Name: "createat", Value: q.Begin, Op: ">="})
	}
	if q.End != 0 {
		where = append(where, database.Where{Name: "createat", Value: q.End, Op: "<="})
	}

	orderT, err := database.MakeOrder(q.Order, "effectid")
	if err != nil {
		return nil, err
	}
	paging := database.MakePaging("effectid", q.Cursor, q.Limit)

	var result []database.Effect

	err = m.qdb.SelectRows(database.TableEffects, where, orderT, paging, &result)

	if err != nil {
		return nil, err
	}

	var res []types.EffectData
	for _, r := range result {
		ad := types.EffectData{
			EffectID: r.EffectID,
			JSONData: r.JData,
		}
		res = append(res, ad)
	}

	return res, nil
}
