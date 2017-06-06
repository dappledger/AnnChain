package refuse_list

import (
	"encoding/hex"
	"errors"
	"strings"

	dbm "gitlab.zhonganonline.com/ann/ann-module/lib/go-db"
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
