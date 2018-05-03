/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package basesql

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"path"

	"github.com/jmoiron/sqlx"
	//"github.com/dappledger/AnnChain/module/lib/go-config"
	"github.com/spf13/viper"
	"github.com/dappledger/AnnChain/src/chain/database"
	"go.uber.org/zap"
)

// Basesql sql-like database
//	is not goroutine-safe
type Basesql struct {
	conn   *sqlx.DB
	tx     *sqlx.Tx
	logger *zap.Logger
}

// Init initialization
//	init db connection
func (bs *Basesql) Init(dbname string, cfg *viper.Viper, logger *zap.Logger) error {
	dbDriver := cfg.GetString("db_type")
	var dbConn string
	switch dbDriver {
	case database.DBTypeSQLite3:
		dbConn = path.Join(cfg.GetString("db_dir"), dbname)
	case database.DBTypeMySQL:
		// who knows
	}
	conn, err := sqlx.Connect(dbDriver, dbConn)
	if err != nil {
		return err
	}

	bs.conn = conn
	err = bs.conn.Ping()
	if err != nil {
		return err
	}

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

// InitTables create tables and indexs if not exists
func (bs *Basesql) InitTables(tables, indexs []string) error {
	for _, ctsql := range tables {
		_, err := bs.conn.Exec(ctsql)
		if err != nil {
			return err
		}
	}
	for _, cisql := range indexs {
		_, err := bs.conn.Exec(cisql)
		if err != nil {
			// index may already exists, so just warn here
			bs.logger.Warn("create sql index failed:" + err.Error())
			fmt.Println("create sql index failed:", err.Error())
		}
	}

	return nil
}

// Count count records
func (bs *Basesql) Count(table string, where []database.Where) (int, error) {
	if table == "" {
		return 0, errors.New("params invalid")
	}

	var sqlBuff bytes.Buffer
	sqlBuff.WriteString(fmt.Sprintf("select count(1) from %s where 1 = 1", table))
	for i := 0; i < len(where); i++ {
		sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
	}

	values := make([]interface{}, len(where))
	for i, v := range where {
		values[i] = v.Value
	}

	// execute
	var err error
	var count int
	if bs.tx != nil {
		err = bs.tx.Get(&count, sqlBuff.String(), values...)
	} else {
		err = bs.conn.Get(&count, sqlBuff.String(), values...)
	}

	return count, err
}

// Insert insert a record
func (bs *Basesql) Insert(table string, feilds []database.Feild) (sql.Result, error) {
	if table == "" || len(feilds) == 0 {
		return nil, errors.New("nothing to insert")
	}

	var sqlBuff bytes.Buffer

	// fill feild name
	sqlBuff.WriteString(fmt.Sprintf("insert into %s (", table))
	for i := 0; i < len(feilds)-1; i++ {
		sqlBuff.WriteString(fmt.Sprintf("%s,", feilds[i].Name))
	}
	sqlBuff.WriteString(fmt.Sprintf("%s) values (", feilds[len(feilds)-1].Name))

	// fill feild value
	for i := 0; i < len(feilds)-1; i++ {
		sqlBuff.WriteString("?,")
	}
	sqlBuff.WriteString("?);")

	bs.logger.Debug(sqlBuff.String())

	// execute
	values := make([]interface{}, len(feilds))
	for i, v := range feilds {
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

// Replace insert or replace a record
func (bs *Basesql) Replace(table string, feilds []database.Feild) (sql.Result, error) {
	if table == "" || len(feilds) == 0 {
		return nil, errors.New("nothing to replace")
	}

	var sqlBuff bytes.Buffer

	// fill feild name
	sqlBuff.WriteString(fmt.Sprintf("replace into %s (", table))
	for i := 0; i < len(feilds)-1; i++ {
		sqlBuff.WriteString(fmt.Sprintf("%s,", feilds[i].Name))
	}
	sqlBuff.WriteString(fmt.Sprintf("%s) values (", feilds[len(feilds)-1].Name))

	// fill feild value
	for i := 0; i < len(feilds)-1; i++ {
		sqlBuff.WriteString("?,")
	}
	sqlBuff.WriteString("?);")

	bs.logger.Debug(sqlBuff.String())

	// execute
	values := make([]interface{}, len(feilds))
	for i, v := range feilds {
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

// SelectRows select rows to struct slice, using "xxx > xxx" as paging method
func (bs *Basesql) SelectRows(table string, where []database.Where, order *database.Order, paging *database.Paging, result interface{}) error {
	if table == "" {
		return errors.New("table name is required")
	}
	if len(where) == 0 {
		return errors.New("full-table-select is not allowed")
	}
	if order != nil && (len(order.Feilds) == 0 || order.Type == "") {
		return errors.New("order type and feilds is required")
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
		// append where clause for paging
		if paging != nil && paging.CursorValue != 0 {
			sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", paging.CursorName, order.GetOp()))
			values = append(values, paging.CursorValue)
		}

		// append order by clause for ordering
		sqlBuff.WriteString(fmt.Sprintf(" order by %s ", order.Feilds[0]))
		for i := 1; i < len(order.Feilds); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" , %s ", order.Feilds[i]))
		}
		sqlBuff.WriteString(order.Type)

		// append limit clause for paging
		if paging != nil {
			sqlBuff.WriteString(" limit ? ")
			values = append(values, paging.Limit)
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

// SelectRowsOffset select rows to struct slice, using "limit xxx offset xxx" as paging method
func (bs *Basesql) SelectRowsOffset(table string, where []database.Where, order *database.Order, offset, limit uint64, result interface{}) error {
	if table == "" {
		return errors.New("table name is required")
	}
	if len(where) == 0 {
		return errors.New("full-table-select is not allowed")
	}
	if order != nil && (len(order.Feilds) == 0 || order.Type == "") {
		return errors.New("order type and feilds is required")
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

// Begin begin a new transaction
func (bs *Basesql) Begin() error {
	if bs.tx != nil {
		bs.tx.Rollback()
		bs.tx = nil
	}

	tx, err := bs.conn.Beginx()
	if err != nil {
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
			return err
		}
	}

	bs.tx = nil
	return nil
}
