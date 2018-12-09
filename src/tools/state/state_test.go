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


package state

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	db "github.com/tendermint/tmlibs/db"
	"github.com/dappledger/AnnChain/module/xlib/mlist"
)

const (
	TEST_NUM    = 100 // better >= 100
	STORE_PRE   = "store_num_%v"
	STORE_TAIL  = "store_to_state_db_%v"
	MODIFY_TAIL = "modified_to_state_db_%v"
)

func initStateRunData() (state *State, storeMap *mlist.MapList) {
	//godb, err := db.NewGoLevelDB("remoteStatus", "./testdb")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//state = NewState(godb)
	memDB := db.NewMemDB()
	state = NewState(memDB)
	storeMap = mlist.NewMapList()
	for i := 0; i < TEST_NUM; i++ {
		var data StateKvData
		data.Init(fmt.Sprintf(STORE_PRE, i), []byte(fmt.Sprintf(STORE_TAIL, i)))
		storeMap.Set(data.Key(), &data)
	}
	return
}

func execMemDB(d db.DB) {
	itr := d.Iterator()
	for itr.Next() {
		fmt.Printf("memDB,key:%v,value:%v\n", hex.EncodeToString(itr.Key()), string(itr.Value()))
	}
}

func stateCreateRunData(t *testing.T, state *State, storeMap *mlist.MapList) ([]byte, error) {
	storeMap.Exec(func(k string, vl interface{}) {
		v := vl.(*StateKvData)
		if err := state.CreateData(v); err != nil {
			t.Error("create data err:", err, ",key:", k)
			return
		}
	})
	return state.Commit()
}

func testStoreModifyOrDel(i int) (modify, del bool) {
	if randInt := i % 2; randInt == 0 && i < 50 {
		if i < 30 {
			modify = true
		} else {
			del = true
		}
	}
	return
}

func stateCheckKvStored(t *testing.T, state *State, storeMap *mlist.MapList, changedMap *mlist.MapList) {
	var i int
	storeMap.ExecBreak(func(k string, vl interface{}) bool {
		i++
		var vbytes []byte
		if changedMap != nil {
			modify, del := testStoreModifyOrDel(i)
			if modify {
				var ok bool
				vl, ok = changedMap.Get(k)
				if !ok {
					t.Error("changedMap,data not found,key:", k)
					return false
				}
			} else if del {
				if len(vbytes) != 0 {
					t.Error("data not deleted,k:", k, ",data:", string(vbytes))
					return false
				}
				return true
			}
		}
		sv, _ := state.GetData(k, KVDataFromBytes)
		kvdata := sv.(*StateKvData)
		kdbytes, _ := kvdata.Bytes()
		vbytes, _ = vl.(*StateKvData).Bytes()
		if !bytes.Equal(kdbytes, vbytes) {
			t.Error("get value from state changed:", string(kdbytes),
				", should be:", string(vbytes))
			return false
		}
		return true
	})
}

func stateModifyDelKvData(t *testing.T, state *State, storeMap *mlist.MapList) ([]byte, error, *mlist.MapList) {
	changeMap := mlist.NewMapList()
	var i int
	storeMap.ExecBreak(func(k string, vl interface{}) bool {
		i++
		sv, err := state.GetData(k, KVDataFromBytes)
		kvdata := sv.(*StateKvData)
		if err != nil {
			t.Error("get data from state err:", err)
			return false
		}
		modify, del := testStoreModifyOrDel(i)
		if modify {
			kvdata.Reset(k, []byte(fmt.Sprintf(MODIFY_TAIL, i)))
		} else if del {
			kvdata = nil
		}
		changeMap.Set(k, kvdata)
		state.ModifyData(k, kvdata)
		return true
	})
	bys, err := state.Commit()
	return bys, err, changeMap
}

func TestManageData(t *testing.T) {
	state, storeMap := initStateRunData()

	root, err := stateCreateRunData(t, state, storeMap)
	if err != nil {
		t.Error("commit err:", err)
		return
	}
	t.Log("commit root:", hex.EncodeToString(root))

	dupState := state.Copy()
	dupState.Reload(root)
	stateCheckKvStored(t, dupState, storeMap, nil)

	root2, err2, changedMap := stateModifyDelKvData(t, dupState, storeMap)
	if err2 != nil {
		t.Error("commit err:", err2)
		return
	}
	t.Log("commit root:", hex.EncodeToString(root2))
	if bytes.Equal(root, root2) {
		t.Error("commit err, changed but get same root")
		return
	}

	dupState.Load(root2)
	stateCheckKvStored(t, dupState, storeMap, changedMap)
}
