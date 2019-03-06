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
	"database/sql"
	"strconv"

	"github.com/dappledger/AnnChain/genesis/chain/database"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
)

func (m *DataManager) PrepareTransaction() (*sql.Stmt, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}
	fields := []database.Feild{
		database.Feild{Name: "txhash"},
		database.Feild{Name: "ledgerhash"},
		database.Feild{Name: "height"},
		database.Feild{Name: "createdat"},
		database.Feild{Name: "account"},
		database.Feild{Name: "target"},
		database.Feild{Name: "optype"},
		database.Feild{Name: "accountsequence"},
		database.Feild{Name: "feepaid"},
		database.Feild{Name: "resultcode"},
		database.Feild{Name: "resultcodes"},
		database.Feild{Name: "memo"},
	}

	return m.qdb.Prepare(database.TableTransactions, fields)
}

func (m *DataManager) AddTransactionStmt(stmt *sql.Stmt, data *types.TransactionData) (err error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	fields := []database.Feild{
		database.Feild{Name: "txhash", Value: data.TxHash.Hex()},
		database.Feild{Name: "ledgerhash", Value: data.LedgerHash.Hex()},
		database.Feild{Name: "height", Value: data.Height.String()},
		database.Feild{Name: "createdat", Value: data.CreateDate},
		database.Feild{Name: "account", Value: data.Account.Hex()},
		database.Feild{Name: "target", Value: data.Target.Hex()},
		database.Feild{Name: "optype", Value: data.OpType},
		database.Feild{Name: "accountsequence", Value: data.AccountSequence.String()},
		database.Feild{Name: "feepaid", Value: data.FeePaid.String()},
		database.Feild{Name: "resultcode", Value: data.ResultCode},
		database.Feild{Name: "resultcodes", Value: data.ResultCodes},
		database.Feild{Name: "memo", Value: data.Memo},
	}
	_, err = m.qdb.Excute(stmt, fields)
	return err
}

// AddTransaction insert a tx record
func (m *DataManager) AddTransaction(data *types.TransactionData) (uint64, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	fields := []database.Feild{
		database.Feild{Name: "txhash", Value: data.TxHash.Hex()},
		database.Feild{Name: "ledgerhash", Value: data.LedgerHash.Hex()},
		database.Feild{Name: "height", Value: data.Height.String()},
		database.Feild{Name: "createdat", Value: data.CreateDate},
		database.Feild{Name: "account", Value: data.Account.Hex()},
		database.Feild{Name: "target", Value: data.Target.Hex()},
		database.Feild{Name: "optype", Value: data.OpType},
		database.Feild{Name: "accountsequence", Value: data.AccountSequence.String()},
		database.Feild{Name: "feepaid", Value: data.FeePaid.String()},
		database.Feild{Name: "resultcode", Value: data.ResultCode},
		database.Feild{Name: "resultcodes", Value: data.ResultCodes},
		database.Feild{Name: "memo", Value: data.Memo},
	}

	sqlRes, err := m.qdb.Insert(database.TableTransactions, fields)
	if err != nil {
		return 0, err
	}

	id, err := sqlRes.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

// QuerySingleTx query single tx record
func (m *DataManager) QuerySingleTx(txhash *ethcmn.Hash) (*types.TransactionData, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "txhash", Value: txhash.Hex()},
	}

	var result []database.TxData
	err := m.qdb.SelectRows(database.TableTransactions, where, nil, nil, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	r := result[0]
	td := types.TransactionData{
		TxID:            r.TxID,
		TxHash:          ethcmn.HexToHash(r.TxHash),
		LedgerHash:      ethcmn.StringToLedgerHash(r.LedgerHash),
		Height:          ethcmn.Big(r.Height),
		CreateDate:      r.CreateDate,
		Account:         ethcmn.HexToAddress(r.Account),
		AccountSequence: ethcmn.Big(r.AccountSequence),
		FeePaid:         ethcmn.Big(r.FeePaid),
		ResultCode:      r.ResultCode,
		ResultCodes:     r.ResultCodes,
		Memo:            r.Memo,
	}
	return &td, nil
}

// QueryAccountTxs query account's tx records
func (m *DataManager) QueryAccountTxs(accid *ethcmn.Address, cursor, limit uint64, order string) ([]types.TransactionQueryData, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "1", Value: 1},
	}
	if accid != nil {
		where = append(where, database.Where{Name: "account", Value: accid.Hex()})
	}
	orderT, err := database.MakeOrder(order, "txid")
	if err != nil {
		return nil, err
	}
	paging := database.MakePaging("txid", cursor, limit)

	var result []database.TxData
	err = m.qdb.SelectRows(database.TableTransactions, where, orderT, paging, &result)
	if err != nil {
		return nil, err
	}

	var res []types.TransactionQueryData
	for _, r := range result {
		td := types.TransactionQueryData{
			Hash:     ethcmn.HexToHash(r.TxHash),
			Height:   ethcmn.Big(r.Height),
			CreateAt: r.CreateDate,
			From:     ethcmn.HexToAddress(r.Account),
			Target:   ethcmn.HexToAddress(r.Target),
			Nonce:    ethcmn.Big(r.AccountSequence),
			BaseFee:  ethcmn.Big(r.FeePaid),
			OpType:   r.OpType,
			Memo:     r.Memo,
		}

		res = append(res, td)
	}

	return res, nil
}

func (m *DataManager) QueryHeightTxs(height string, cursor, limit uint64, order string) ([]types.TransactionQueryData, error) {
	if m.qNeedLock {
		m.qLock.Lock()
		defer m.qLock.Unlock()
	}

	where := []database.Where{
		database.Where{Name: "1", Value: 1},
	}

	tmpheight, err := strconv.ParseUint(height, 10, 64)
	if err != nil {
		return nil, err
	}
	where = append(where, database.Where{Name: "height", Value: tmpheight})

	orderT, err := database.MakeOrder(order, "txid")
	if err != nil {
		return nil, err
	}
	paging := database.MakePaging("txid", cursor, limit)

	var result []database.TxData
	err = m.qdb.SelectRows(database.TableTransactions, where, orderT, paging, &result)
	if err != nil {
		return nil, err
	}

	var res []types.TransactionQueryData
	for _, r := range result {
		td := types.TransactionQueryData{
			Hash:     ethcmn.HexToHash(r.TxHash),
			Height:   ethcmn.Big(r.Height),
			CreateAt: r.CreateDate,
			From:     ethcmn.HexToAddress(r.Account),
			Target:   ethcmn.HexToAddress(r.Target),
			Nonce:    ethcmn.Big(r.AccountSequence),
			BaseFee:  ethcmn.Big(r.FeePaid),
			OpType:   r.OpType,
			Memo:     r.Memo,
		}

		res = append(res, td)
	}

	return res, nil
}

// QueryAllTxs query all tx records
func (m *DataManager) QueryAllTxs(cursor, limit uint64, order string) ([]types.TransactionQueryData, error) {
	return m.QueryAccountTxs(nil, cursor, limit, order)
}
