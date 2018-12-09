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

package refuse_list

import (
	"encoding/hex"
	"errors"
	"strings"

	dbm "github.com/dappledger/AnnChain/ann-module/lib/go-db"
)

type RefuseList struct {
	db dbm.DB
}

var (
	dbName        = "refuse_list"
	keyExistValue = []byte{'Y'}
)

func NewRefuseList(dbBackend, dbDir string) *RefuseList {
	refuseListDB := dbm.NewDB(dbName, dbBackend, dbDir)
	return &RefuseList{refuseListDB}
}

func (rl *RefuseList) Stop() {
	rl.db.Close()
}

func (rl *RefuseList) QueryRefuseKey(pubKey [32]byte) (keyExist bool) {
	ret := rl.db.Get(pubKey[:])
	if len(ret) == 1 {
		keyExist = true
	}
	return
}

func (rl *RefuseList) ListAllKey() (keyList []string) {
	iter := rl.db.Iterator()
	for iter.Next() {
		key := iter.Key()
		str := hex.EncodeToString(key)
		keyList = append(keyList, strings.ToUpper(str))
	}
	return
}

func (rl *RefuseList) AddRefuseKey(pubKey [32]byte) {
	rl.db.SetSync(pubKey[:], keyExistValue)
}

func (rl *RefuseList) DeleteRefuseKey(pubKey [32]byte) (err error) {
	if rl.QueryRefuseKey(pubKey) {
		rl.db.DeleteSync(pubKey[:])
	} else {
		err = errors.New("pubKey not exist")
	}
	return
}
