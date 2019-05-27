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
	"fmt"
	"strings"

	dbm "github.com/dappledger/AnnChain/gemmill/modules/go-db"
	log "github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"go.uber.org/zap"
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

func (rl *RefuseList) QueryRefuseKey(pubKey []byte) (keyExist bool) {
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

func (rl *RefuseList) AddRefuseKey(pubKey []byte) {
	rl.db.SetSync(pubKey[:], keyExistValue)
	log.Info("[refuse list],add key", zap.String("pubkey", fmt.Sprintf("%x", pubKey)))
}

func (rl *RefuseList) DeleteRefuseKey(pubKey []byte) (err error) {
	if rl.QueryRefuseKey(pubKey) {
		rl.db.DeleteSync(pubKey[:])
		log.Info("[refuse list],del key", zap.String("pubkey", fmt.Sprintf("%x", pubKey)))
	} else {
		err = errors.New("pubKey not exist")
		log.Warn("[refuse list],del key", zap.String("pubkey", fmt.Sprintf("%x", pubKey)), zap.Error(err))
	}
	return
}
