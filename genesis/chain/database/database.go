package database

import (
	"database/sql"
	"errors"

	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/ann-module/lib/go-config"
)

const (
	TableEffects       = "effects"
	TableLedgerheaders = "ledgerheaders"
	TableActions       = "actions"
	TableTransactions  = "transactions"
	TableAccData       = "accdata"
)

const (
	DBTypeSQLite3 = "sqlite3"
)

// Feild database field
type Feild struct {
	Name  string
	Value interface{}
}

// Where query field
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
	Limit       uint64 // limit
}

// Database interface for delos app database-operation
type Database interface {
	Init(dbname string, cfg config.Config, logger *zap.Logger) error
	Close()
	GetInitSQLs() (opt, opi, qt, qi []string)
	PrepareTables(ctsqls, cisqls []string) error

	Insert(table string, fields []Feild) (sql.Result, error)
	Delete(table string, where []Where) (sql.Result, error)
	Update(table string, toupdate []Feild, where []Where) (sql.Result, error)
	SelectRows(table string, where []Where, order *Order, paging *Paging, result interface{}) error
	SelectRowsOffset(table string, where []Where, order *Order, offset, limit uint64, result interface{}) error
	SelectRawSQL(table string, sqlStr string, values []interface{}, result interface{}) error
	SelectRowsUnion(table string, wheres [][]Where, order *Order, paging *Paging, result interface{}) error
	Excute(stmt *sql.Stmt, fields []Feild) (sql.Result, error)
	Prepare(table string, fields []Feild) (*sql.Stmt, error)

	Begin() error
	Commit() error
	Rollback() error
}

//MakeOrder make a order object
func MakeOrder(ordertype string, fields ...string) (*Order, error) {
	if ordertype == "" {
		ordertype = "desc"
	}

	if ordertype != "asc" && ordertype != "ASC" && ordertype != "desc" && ordertype != "DESC" {
		return nil, errors.New("invalid order type :" + ordertype)
	}

	return &Order{
		Type:   ordertype,
		Feilds: fields,
	}, nil
}

// MakePaging make a paging object
func MakePaging(colName string, colValue uint64, limit uint64) *Paging {
	if limit == 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}
	if colValue < 0 {
		colValue = 0
	}

	return &Paging{
		CursorName:  colName,
		CursorValue: colValue * limit,
		Limit:       limit,
	}
}
