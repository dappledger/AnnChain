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
	"errors"

	"github.com/dappledger/AnnChain/genesis/chain/database"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
)

var managedatacategory types.ManageDataCategory

func (m *DataManager) AddAccData(acct ethcmn.Address, k, v, category string) (uint64, error) {
	if m.opNeedLock {
		m.opLock.Lock()
		defer m.opLock.Unlock()
	}

	fields := []database.Feild{
		database.Feild{Name: "accountid", Value: acct.Hex()},
		database.Feild{Name: "datakey", Value: k},
		database.Feild{Name: "datavalue", Value: v},
		database.Feild{Name: "category", Value: category},
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
func (m *DataManager) UpdateAccData(acct ethcmn.Address, k, v, category string) error {
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
		database.Where{Name: "category", Value: category},
	}

	_, err := m.opdb.Update(database.TableAccData, fields, where)
	if err != nil {
		return err
	}

	return nil
}

// query account's all managedata
func (m *DataManager) QueryAccData(acc ethcmn.Address, order string) (datas map[string]types.ManageDataCategory, err error) {
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

	datas = make(map[string]types.ManageDataCategory, len(result))

	for _, r := range result {
		managedatacategory.Category = r.Category
		managedatacategory.Value = r.DataValue
		datas[r.DataKey] = managedatacategory
	}
	return
}

// QueryManageData query all recores of a specific account
func (m *DataManager) QueryAccountManagedata(accid ethcmn.Address, category string, name string, cursor, limit uint64, order string) (datas map[string]types.ManageDataCategory, err error) {
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

	if category != "" {
		wheres = append(wheres, database.Where{Name: "accountid", Value: accid.Hex()})
		wheres = append(wheres, database.Where{Name: "category", Value: category})
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

	datas = make(map[string]types.ManageDataCategory, len(result))

	for _, r := range result {
		managedatacategory.Category = r.Category
		managedatacategory.Value = r.DataValue
		datas[r.DataKey] = managedatacategory
	}
	return
}

func (m *DataManager) QuerySingleManageData(accid ethcmn.Address, keys string) (datas map[string]types.ManageDataCategory, err error) {

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

	datas = make(map[string]types.ManageDataCategory, len(result))

	for _, r := range result {
		managedatacategory.Category = r.Category
		managedatacategory.Value = r.DataValue
		datas[r.DataKey] = managedatacategory
	}
	return
}

func (m *DataManager) QueryCategoryManageData(accid ethcmn.Address, category string) (datas map[string]types.ManageDataCategory, err error) {

	if m.opNeedLock {
		m.opLock.Lock()
		defer m.opLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "accountid", Value: accid.Hex()},
		database.Where{Name: "category", Value: category},
	}

	var result []database.AccData
	err = m.opdb.SelectRows(database.TableAccData, where, nil, nil, &result)
	if err != nil {
		return nil, err
	}

	datas = make(map[string]types.ManageDataCategory, len(result))

	for _, r := range result {
		managedatacategory.Category = r.Category
		managedatacategory.Value = r.DataValue
		datas[r.DataKey] = managedatacategory
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
