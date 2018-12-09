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


package database

import (
	"database/sql"
	"errors"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	//"github.com/dappledger/AnnChain/module/lib/go-config"
)

const (
	DBTypeSQLite3 = "sqlite3"
	DBTypeMySQL   = "mysql" // TODO
)

// Feild database feild
type Feild struct {
	Name  string
	Value interface{}
}

// Where query feild
type Where struct {
	Name  string
	Value interface{}
	Op    string // can be =、>、<、<> and any operator supported by sql-database
}

// GetOp get operator of current where clause, default =
func (w *Where) GetOp() string {
	if w.Op == "" {
		return "="
	}
	return w.Op
}

// Order  used to identify query order
type Order struct {
	Type   string   // "asc" or "desc"
	Feilds []string // order by x
}

// GetOp used in sql
func (o *Order) GetOp() string {
	if o != nil && o.Type == "desc" {
		return "<="
	}

	return ">="
}

type Paging struct {
	CursorName  string // cursor column
	CursorValue uint64 // cursor column
	Limit       uint   // limit
}

// Database interface for  app database-operation
type Database interface {
	Init(dbname string, cfg *viper.Viper, logger *zap.Logger) error
	Close()
	InitTables(ctsqls, cisqls []string) error

	Count(table string, where []Where) (int, error)

	Insert(table string, feilds []Feild) (sql.Result, error)
	Replace(table string, feilds []Feild) (sql.Result, error) // insert or replace
	Delete(table string, where []Where) (sql.Result, error)
	Update(table string, toupdate []Feild, where []Where) (sql.Result, error)
	SelectRows(table string, where []Where, order *Order, paging *Paging, result interface{}) error
	SelectRowsOffset(table string, where []Where, order *Order, offset, limit uint64, result interface{}) error

	Begin() error
	Commit() error
	Rollback() error
}

//MakeOrder make a order object
func MakeOrder(ordertype string, feilds ...string) (*Order, error) {
	if ordertype == "" {
		ordertype = "asc"
	}

	if ordertype != "asc" && ordertype != "ASC" && ordertype != "desc" && ordertype != "DESC" {
		return nil, errors.New("invalid order type :" + ordertype)
	}

	return &Order{
		Type:   ordertype,
		Feilds: feilds,
	}, nil
}

// MakePaging make a paging object
func MakePaging(colName string, colValue uint64, limit uint) *Paging {
	if limit == 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}

	return &Paging{
		CursorName:  colName,
		CursorValue: colValue,
		Limit:       limit,
	}
}
