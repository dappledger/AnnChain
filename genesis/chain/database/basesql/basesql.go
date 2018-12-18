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

package basesql

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"path"

	"github.com/dappledger/AnnChain/ann-module/lib/go-config"
	"github.com/dappledger/AnnChain/genesis/chain/database"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Basesql sql-like database is not goroutine-safe
type Basesql struct {
	conn   *sqlx.DB
	tx     *sqlx.Tx
	logger *zap.Logger
}

// Init initialization
// init db connection
// create tables if not exist
func (bs *Basesql) Init(dbname string, cfg config.Config, logger *zap.Logger) error {
	dbDriver := cfg.GetString("db_type")
	var dbConn string
	switch dbDriver {
	case database.DBTypeSQLite3:
		dbConn = path.Join(cfg.GetString("db_dir"), dbname)
	}
	conn, err := sqlx.Connect(dbDriver, dbConn)
	if err != nil {
		return err
	}

	bs.conn = conn

	_, err = bs.conn.Exec("PRAGMA cache_size = 8000;")
	_, err = bs.conn.Exec("PRAGMA synchronous = OFF;")
	_, err = bs.conn.Exec("PRAGMA temp_store = MEMORY;")

	bs.logger = logger

	return nil
}

// Close close conn
func (bs *Basesql) Close() {
	if bs.tx != nil {
		bs.tx.Rollback()
		bs.tx = nil
	}

	if bs.conn != nil {
		bs.conn.Close()
		bs.conn = nil
	}
}

// PrepareTables create tables if not exists
func (bs *Basesql) PrepareTables(ctsqls, cisqls []string) error {
	for _, ctsql := range ctsqls {
		_, err := bs.conn.Exec(ctsql)
		if err != nil {
			return err
		}
	}
	for _, cisql := range cisqls {
		_, err := bs.conn.Exec(cisql)
		if err != nil {
			// index may already exists, so just warn here
			bs.logger.Warn("create sql index failed:" + err.Error())
		}
	}

	return nil
}

func (bs *Basesql) Excute(stmt *sql.Stmt, fields []database.Feild) (sql.Result, error) {
	values := make([]interface{}, len(fields))
	for i, v := range fields {
		values[i] = v.Value
	}
	return stmt.Exec(values...)
}

func (bs *Basesql) Prepare(table string, fields []database.Feild) (*sql.Stmt, error) {

	var sqlBuff bytes.Buffer

	sqlBuff.WriteString(fmt.Sprintf("insert into %s (", table))
	for i := 0; i < len(fields)-1; i++ {
		sqlBuff.WriteString(fmt.Sprintf("%s,", fields[i].Name))
	}
	sqlBuff.WriteString(fmt.Sprintf("%s) values (", fields[len(fields)-1].Name))

	for i := 1; i < len(fields); i++ {
		sqlBuff.WriteString(fmt.Sprintf("?,"))
	}
	sqlBuff.WriteString(fmt.Sprintf("?);"))

	if bs.tx != nil {
		return bs.tx.Prepare(sqlBuff.String())
	} else {
		return bs.conn.Prepare(sqlBuff.String())
	}
}

// Insert insert a record
func (bs *Basesql) Insert(table string, fields []database.Feild) (sql.Result, error) {
	if table == "" || len(fields) == 0 {
		return nil, errors.New("nothing to insert")
	}

	var sqlBuff bytes.Buffer

	// fill field name
	sqlBuff.WriteString(fmt.Sprintf("insert into %s (", table))
	for i := 0; i < len(fields)-1; i++ {
		sqlBuff.WriteString(fmt.Sprintf("%s,", fields[i].Name))
	}
	sqlBuff.WriteString(fmt.Sprintf("%s) values (", fields[len(fields)-1].Name))

	// fill field value
	for i := 0; i < len(fields)-1; i++ {
		sqlBuff.WriteString("?,")
	}
	sqlBuff.WriteString("?);")

	// execute
	values := make([]interface{}, len(fields))
	for i, v := range fields {
		values[i] = v.Value
	}
	var res sql.Result
	var err error

	if bs.tx != nil {
		res, err = bs.tx.Exec(sqlBuff.String(), values...)
	} else {
		res, err = bs.conn.Exec(sqlBuff.String(), values...)
	}

	return res, err
}

// Delete delete records
func (bs *Basesql) Delete(table string, where []database.Where) (sql.Result, error) {
	if table == "" {
		return nil, errors.New("table name is required")
	}
	if len(where) == 0 {
		return nil, errors.New("table-clearing is not allowed")
	}

	var sqlBuff bytes.Buffer
	sqlBuff.WriteString(fmt.Sprintf("delete from %s where 1 = 1", table))
	for i := 0; i < len(where); i++ {
		sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
	}

	// execute
	values := make([]interface{}, len(where))
	for i, v := range where {
		values[i] = v.Value
	}
	var res sql.Result
	var err error
	if bs.tx != nil {
		res, err = bs.tx.Exec(sqlBuff.String(), values...)
	} else {
		res, err = bs.conn.Exec(sqlBuff.String(), values...)
	}

	return res, err
}

// Update update records
func (bs *Basesql) Update(table string, toupdate []database.Feild, where []database.Where) (sql.Result, error) {
	if table == "" {
		return nil, errors.New("table name is required")
	}
	if len(where) == 0 {
		return nil, errors.New("full-table-update is not allowed")
	}
	if len(toupdate) == 0 {
		return nil, errors.New("to-update-nothing is not allowed")
	}

	var sqlBuff bytes.Buffer
	sqlBuff.WriteString(fmt.Sprintf(" update %s set %s = ? ", table, toupdate[0].Name))
	for i := 1; i < len(toupdate); i++ {
		sqlBuff.WriteString(fmt.Sprintf(", %s = ? ", toupdate[i].Name))
	}

	sqlBuff.WriteString(fmt.Sprintf(" where 1 = 1 "))
	for i := 0; i < len(where); i++ {
		sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
	}

	// execute
	values := make([]interface{}, len(toupdate)+len(where))
	for i, v := range toupdate {
		values[i] = v.Value
	}
	for i, v := range where {
		values[len(toupdate)+i] = v.Value
	}

	var res sql.Result
	var err error
	if bs.tx != nil {
		res, err = bs.tx.Exec(sqlBuff.String(), values...)
	} else {
		res, err = bs.conn.Exec(sqlBuff.String(), values...)
	}

	return res, err
}
func (bs *Basesql) SelectRowsUnion(table string, wheres [][]database.Where, order *database.Order, paging *database.Paging, result interface{}) error {
	if table == "" {
		return errors.New("table name is required")
	}
	if len(wheres) == 0 {
		return errors.New("full-table-select is not allowed")
	}
	if order != nil && (len(order.Feilds) == 0 || order.Type == "") {
		return errors.New("order type and fields is required")
	}

	var values []interface{}

	wheresLen := len(wheres)

	var sqlBuff bytes.Buffer

	for _, where := range wheres {

		wheresLen--

		for _, v := range where {
			values = append(values, v.Value)
		}

		sqlBuff.WriteString(fmt.Sprintf("select * from %s where 1 = 1", table))
		for i := 0; i < len(where); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
		}
		if wheresLen > 0 {
			sqlBuff.WriteString(" union ")
		}
	}

	if order != nil {
		// append order by clause for ordering
		sqlBuff.WriteString(fmt.Sprintf(" order by %s ", order.Feilds[0]))
		for i := 1; i < len(order.Feilds); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" , %s ", order.Feilds[i]))
		}
		sqlBuff.WriteString(order.Type)

		sqlBuff.WriteString(fmt.Sprintf(" limit %d offset %d ", paging.Limit, paging.CursorValue))
	}

	// execute
	var err error
	if bs.tx != nil {
		err = bs.tx.Select(result, sqlBuff.String(), values...)
	} else {
		err = bs.conn.Select(result, sqlBuff.String(), values...)
	}

	return err
}

// SelectRows select rows to struct slice
func (bs *Basesql) SelectRows(table string, where []database.Where, order *database.Order, paging *database.Paging, result interface{}) error {
	if table == "" {
		return errors.New("table name is required")
	}
	if len(where) == 0 {
		return errors.New("full-table-select is not allowed")
	}
	if order != nil && (len(order.Feilds) == 0 || order.Type == "") {
		return errors.New("order type and fields is required")
	}

	values := make([]interface{}, len(where))
	for i, v := range where {
		values[i] = v.Value
	}

	var sqlBuff bytes.Buffer
	sqlBuff.WriteString(fmt.Sprintf("select * from %s where 1 = 1", table))
	for i := 0; i < len(where); i++ {
		sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
	}
	if order != nil {
		// append order by clause for ordering
		sqlBuff.WriteString(fmt.Sprintf(" order by %s ", order.Feilds[0]))
		for i := 1; i < len(order.Feilds); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" , %s ", order.Feilds[i]))
		}
		sqlBuff.WriteString(order.Type)

		// append limit clause for paging
		if paging != nil {
			sqlBuff.WriteString(fmt.Sprintf(" limit %d offset %d ", paging.Limit, paging.CursorValue))
		}
	}

	// execute
	var err error
	if bs.tx != nil {
		err = bs.tx.Select(result, sqlBuff.String(), values...)
	} else {
		err = bs.conn.Select(result, sqlBuff.String(), values...)
	}

	return err
}

// SelectRowsOffset select rows to struct slice
func (bs *Basesql) SelectRowsOffset(table string, where []database.Where, order *database.Order, offset, limit uint64, result interface{}) error {
	if table == "" {
		return errors.New("table name is required")
	}
	if len(where) == 0 {
		return errors.New("full-table-select is not allowed")
	}
	if order != nil && (len(order.Feilds) == 0 || order.Type == "") {
		return errors.New("order type and fields is required")
	}

	values := make([]interface{}, len(where))
	for i, v := range where {
		values[i] = v.Value
	}

	var sqlBuff bytes.Buffer
	sqlBuff.WriteString(fmt.Sprintf("select * from %s where 1 = 1", table))
	for i := 0; i < len(where); i++ {
		sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
	}
	if order != nil {
		// append order by clause for ordering
		sqlBuff.WriteString(fmt.Sprintf(" order by %s ", order.Feilds[0]))
		for i := 1; i < len(order.Feilds); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" , %s ", order.Feilds[i]))
		}
		sqlBuff.WriteString(order.Type)

		// append limit clause for paging
		sqlBuff.WriteString(fmt.Sprintf(" limit %d offset %d ", limit, offset))
	}

	// execute
	var err error
	if bs.tx != nil {
		err = bs.tx.Select(result, sqlBuff.String(), values...)
	} else {
		err = bs.conn.Select(result, sqlBuff.String(), values...)
	}

	return err
}

// SelectRawSQL query useing raw sql
func (bs *Basesql) SelectRawSQL(table string, sqlStr string, values []interface{}, result interface{}) error {
	if table == "" {
		return errors.New("table name is required")
	}

	// execute
	var err error
	if bs.tx != nil {
		err = bs.tx.Select(result, sqlStr, values...)
	} else {
		err = bs.conn.Select(result, sqlStr, values...)
	}

	return err
}

// Begin begin a new transaction
func (bs *Basesql) Begin() error {
	tx, err := bs.conn.Beginx()
	if err != nil {
		bs.tx = nil
		return err
	}

	bs.tx = tx
	return nil
}

// Commit commit current transaction
func (bs *Basesql) Commit() error {
	if bs.tx != nil {
		err := bs.tx.Commit()
		if err != nil {
			bs.tx = nil
			return err
		}
	}

	bs.tx = nil
	return nil
}

// Rollback rollback current transaction
func (bs *Basesql) Rollback() error {
	if bs.tx != nil {
		err := bs.tx.Rollback()
		if err != nil {
			bs.tx = nil
			return err
		}
	}
	bs.tx = nil
	return nil
}
