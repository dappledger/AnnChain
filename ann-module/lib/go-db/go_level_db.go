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

package db

import (
	"fmt"
	"path"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"

	. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
)

func init() {
	dbCreator := func(name string, dir string) (DB, error) {
		return NewGoLevelDB(name, dir)
	}
	registerDBCreator(LevelDBBackendStr, dbCreator, false)
	registerDBCreator(GoLevelDBBackendStr, dbCreator, false)
}

type GoLevelDB struct {
	db *leveldb.DB
}

func NewGoLevelDB(name string, dir string) (*GoLevelDB, error) {
	dbPath := path.Join(dir, name+".db")
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}
	database := &GoLevelDB{db: db}
	return database, nil
}

func (db *GoLevelDB) Get(key []byte) []byte {
	res, err := db.db.Get(key, nil)
	if err != nil {
		if err == errors.ErrNotFound {
			return nil
		} else {
			PanicCrisis(err)
		}
	}
	return res
}

func (db *GoLevelDB) Set(key []byte, value []byte) {
	err := db.db.Put(key, value, nil)
	if err != nil {
		PanicCrisis(err)
	}
}

func (db *GoLevelDB) SetSync(key []byte, value []byte) {
	err := db.db.Put(key, value, &opt.WriteOptions{Sync: true})
	if err != nil {
		PanicCrisis(err)
	}
}

func (db *GoLevelDB) Delete(key []byte) {
	err := db.db.Delete(key, nil)
	if err != nil {
		PanicCrisis(err)
	}
}

func (db *GoLevelDB) DeleteSync(key []byte) {
	err := db.db.Delete(key, &opt.WriteOptions{Sync: true})
	if err != nil {
		PanicCrisis(err)
	}
}

func (db *GoLevelDB) DB() *leveldb.DB {
	return db.db
}

func (db *GoLevelDB) Close() {
	db.db.Close()
}

func (db *GoLevelDB) Print() {
	iter := db.db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}
}

func (db *GoLevelDB) Iterator() Iterator {
	return db.db.NewIterator(nil, nil)
}

func (db *GoLevelDB) NewBatch() Batch {
	batch := new(leveldb.Batch)
	return &goLevelDBBatch{db, batch}
}

//--------------------------------------------------------------------------------

type goLevelDBBatch struct {
	db    *GoLevelDB
	batch *leveldb.Batch
}

func (mBatch *goLevelDBBatch) Set(key, value []byte) {
	mBatch.batch.Put(key, value)
}

func (mBatch *goLevelDBBatch) Delete(key []byte) {
	mBatch.batch.Delete(key)
}

func (mBatch *goLevelDBBatch) Write() {
	err := mBatch.db.db.Write(mBatch.batch, nil)
	if err != nil {
		PanicCrisis(err)
	}
}
