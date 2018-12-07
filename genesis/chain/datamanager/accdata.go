package datamanager

import (
	"errors"

	"github.com/dappledger/AnnChain/genesis/chain/database"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
)

func (m *DataManager) AddAccData(acct ethcmn.Address, k, v string) (uint64, error) {
	if m.opNeedLock {
		m.opLock.Lock()
		defer m.opLock.Unlock()
	}

	fields := []database.Feild{
		database.Feild{Name: "accountid", Value: acct.Hex()},
		database.Feild{Name: "datakey", Value: k},
		database.Feild{Name: "datavalue", Value: v},
	}

	sqlRes, err := m.opdb.Insert(database.TableAccData, fields)
	if err != nil {
		return 0, err
	}

	id, err := sqlRes.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

// UpdateAccData update
func (m *DataManager) UpdateAccData(acct ethcmn.Address, k, v string, isPub bool) error {
	if m.opNeedLock {
		m.opLock.Lock()
		defer m.opLock.Unlock()
	}

	fields := []database.Feild{
		database.Feild{Name: "datavalue", Value: v},
	}

	where := []database.Where{
		database.Where{Name: "accountid", Value: acct.Hex()},
		database.Where{Name: "datakey", Value: k},
	}

	_, err := m.opdb.Update(database.TableAccData, fields, where)
	if err != nil {
		return err
	}

	return nil
}

// query account's all managedata
func (m *DataManager) QueryAccData(acc ethcmn.Address, order string) (datas []map[string]string, err error) {
	if m.opNeedLock {
		m.opLock.Lock()
		defer m.opLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "accountid", Value: acc.Hex()},
	}

	orderT, err := database.MakeOrder(order, "dataid")
	if err != nil {
		return nil, err
	}
	paging := database.MakePaging("dataid", 0, 200)

	var result []database.AccData
	err = m.opdb.SelectRows(database.TableAccData, where, orderT, paging, &result)
	if err != nil {
		return nil, err
	}

	datas = make([]map[string]string, len(result))

	for i, r := range result {
		datas[i] = make(map[string]string)
		datas[i]["name"] = r.DataKey
		datas[i]["value"] = r.DataValue
	}

	return
}

// QueryManageData query all recores of a specific account
func (m *DataManager) QueryAccountManagedata(accid ethcmn.Address, name string, cursor, limit uint64, order string) (datas []map[string]string, err error) {
	if m.opNeedLock {
		m.opLock.Lock()
		defer m.opLock.Unlock()
	}

	var wheres []database.Where

	if name != "" {
		wheres = append(wheres, database.Where{Name: "accountid", Value: accid.Hex()})
		wheres = append(wheres, database.Where{Name: "datakey", Value: name})
	} else {
		wheres = append(wheres, database.Where{Name: "accountid", Value: accid.Hex()})
	}

	orderT, err := database.MakeOrder(order, "dataid")
	if err != nil {
		return nil, err
	}
	paging := database.MakePaging("dataid", cursor, limit)

	var result []database.AccData
	err = m.opdb.SelectRows(database.TableAccData, wheres, orderT, paging, &result)
	if err != nil {
		return nil, err
	}

	datas = make([]map[string]string, len(result))

	for i, r := range result {
		datas[i] = make(map[string]string)
		datas[i]["name"] = r.DataKey
		datas[i]["value"] = r.DataValue
	}
	return
}

func (m *DataManager) QuerySingleManageData(accid ethcmn.Address, keys string) (datas map[string]string, err error) {

	if m.opNeedLock {
		m.opLock.Lock()
		defer m.opLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "accountid", Value: accid.Hex()},
		database.Where{Name: "datakey", Value: keys},
	}

	var result []database.AccData
	err = m.opdb.SelectRows(database.TableAccData, where, nil, nil, &result)
	if err != nil {
		return nil, err
	}

	datas = make(map[string]string, len(result))

	for _, r := range result {
		datas["name"] = r.DataKey
		datas["value"] = r.DataValue
	}
	return
}

// QueryAccData query all recores of a specific account
func (m *DataManager) QueryOneAccData(acc ethcmn.Address, key string) (err error) {
	if m.opNeedLock {
		m.opLock.Lock()
		defer m.opLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "accountid", Value: acc.Hex()},
		database.Where{Name: "datakey", Value: key},
	}

	var result []database.AccData
	err = m.opdb.SelectRows(database.TableAccData, where, nil, nil, &result)
	if err != nil {
		return err
	}

	if len(result) == 0 {
		return nil
	} else {
		return errors.New("key already have")
	}
}
