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
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/dappledger/AnnChain/module/lib/go-config"
	"github.com/dappledger/AnnChain/advertise/chain/log"
)

func getJSON(o interface{}) string {
	j, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return string(j)
}

func getConn(dir, dbname string) *Basesql {
	cfg := config.NewMapConfig(nil)
	if dir == "" {
		dir = "/Users/caojingqi/temp/sqlitedb/"
	}
	if dbname == "" {
		dbname = "test.db"
	}

	cfg.Set("db_type", "sqlite3")
	cfg.Set("db_dir", dir)

	bs := &Basesql{}
	logger := log.Initialize("", "output.log", "err.log")
	err := bs.Init(dbname, cfg, logger)
	if err != nil {
		panic(err)
	}

	return bs
}

func TestInit(t *testing.T) {
	_ = getConn("", "")
}

func TestInsert(t *testing.T) {

}

func TestDelete(t *testing.T) {

}

func TestUpdate(t *testing.T) {

}

func TestSelectRows(t *testing.T) {

}

func TestInsertReplace(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	dbName := fmt.Sprintf("%d.db", rand.Int63())
	bs := getConn("/Users/caojingqi/temp/sqlitedb/", dbName)
	defer bs.Close()

	createTable := `
	CREATE TABLE IF NOT EXISTS innerdata
	(
		datakey			TEXT		NOT NULL,
		name			TEXT		NOT NULL,
		datavalue		TEXT		NOT NULL,
		PRIMARY KEY		(datakey)
	);`

	type InnerData struct {
		Key   string `db:"datakey"`
		Name  string `db:"name"`
		Value string `db:"datavalue"`
	}
	var rows []InnerData

	_, err := bs.conn.Exec(createTable)
	PanicError(t, err)

	rs, err := bs.conn.Exec("replace into innerdata(datakey, name, datavalue) values('abc', 'name', 'hello')")
	PanicError(t, err)
	fmt.Println(rs.LastInsertId())
	err = bs.conn.Select(&rows, "select * from innerdata where datakey='abc'")
	PanicError(t, err)
	for _, v := range rows {
		fmt.Println(v)
	}

	rs, err = bs.conn.Exec("replace into innerdata(datakey, name, datavalue) values('abc', 'name', 'hello world')")
	PanicError(t, err)
	fmt.Println(rs.LastInsertId())
	rows = rows[:0]
	err = bs.conn.Select(&rows, "select * from innerdata where datakey='abc'")
	PanicError(t, err)
	for _, v := range rows {
		fmt.Println(v)
	}
}

func TestSql(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	dbName := fmt.Sprintf("%d.db", rand.Int63())
	bs := getConn("/Users/caojingqi/temp/sqlitedb/", dbName)
	defer bs.Close()

	createTable := `
	CREATE TABLE IF NOT EXISTS innerdata
	(
		datakey			VARCHAR(66)		NOT NULL,
		datavalue		TEXT			NOT NULL,
		PRIMARY KEY		(datakey)
	);`

	_, err := bs.conn.Exec(createTable)
	PanicError(t, err)

}

func PanicError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
