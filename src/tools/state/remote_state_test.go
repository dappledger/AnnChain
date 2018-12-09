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
	"math/big"
	"path"
	"testing"

	"github.com/tendermint/tmlibs/db"

	"github.com/dappledger/AnnChain/module/xlib/mlist"
	"github.com/dappledger/AnnChain/src/chain/log"
)

const (
	LOG_PATH     = "./testlog"
	RMT_TEST_NUM = 100
	//	secSlc = []string{
	//		"f826a86b031e646c884aa64d9e2309ef053cb2a578f29cca84d31a37d8d38c7a",
	//		"a8971729fbc199fb3459529cebcd8704791fc699d88ac89284f23ff8e7fca7d6",
	//		"cb08d54e8559e9febd752c9be74a919156a2fc5bf32a8d91e2ee1129aaff7712",
	//	}
	PER_NUM int64 = 666
)

func GetInitSt() (remote *RemoteState, storeMp *mlist.MapList, accSlc []string) {
	d := db.NewMemDB()
	logger := log.Initialize("development", path.Join(LOG_PATH, "node.output.log"), path.Join(LOG_PATH, "node.err.log"))

	storeMp = mlist.NewMapList()
	for i := 0; i < RMT_TEST_NUM; i++ {
		recordi := i + 1
		storeMp.Set(fmt.Sprintf(STORE_PRE, recordi), []byte(fmt.Sprintf(STORE_TAIL, recordi)))
	}
	remote = NewRemoteState(d, logger)
	accSlc = []string{
		"0x77ece0684aa9c007bb1ffbe230fe4712e5bfa156",
		"0x7752b42608a0f1943c19fc5802cb027e60b4c911",
		"0x955546f92461d6e470b8a78932f3568cf50c394a",
	}
	return
}

func remoteTransferToCreateAcc(t *testing.T, remote *RemoteState, accSlc []string) ([]byte, error) {
	if remote.CreateRemoteAcc(accSlc[0], big.NewInt(PER_NUM*int64(len(accSlc)))) == nil {
		t.Error("create account failed,acc:", accSlc[0])
		return nil, nil
	}

	// check Transfer
	for i := range accSlc {
		if i == 0 {
			continue
		}
		if !remote.Transfer(accSlc[0], accSlc[i], big.NewInt(PER_NUM)) {
			t.Error("transfer failed,to:", accSlc[i])
		}
	}
	return remote.Commit()
}

func remoteCheckTransferCreate(t *testing.T, remote *RemoteState, accSlc []string) {
	for i := range accSlc {
		acc := remote.RemoteAcc(accSlc[i])
		if acc == nil {
			t.Error("account not found:", accSlc[i])
			return
		}
		if acc.GetBalance().Cmp(big.NewInt(PER_NUM)) != 0 {
			t.Error(accSlc[i], "acc get balance err:", acc.GetBalance(), ",should be:", big.NewInt(PER_NUM))
			return
		}
	}
}

func _remoteSetAccKv(remote *RemoteState, storeMp *mlist.MapList, accSlc []string) {
	for i := range accSlc {
		var kvCount int
		storeMp.Exec(func(k string, e interface{}) {
			vbytes := e.([]byte)
			kvCount++
			remote.AccSetKv(accSlc[i], k, vbytes)
		})
		kvCount = 0
	}
}

func remoteSetAccKv(remote *RemoteState, storeMp *mlist.MapList, accSlc []string) ([]byte, error) {
	_remoteSetAccKv(remote, storeMp, accSlc)
	return remote.Commit()
}

func remoteCheckAccKvStored(t *testing.T, remote *RemoteState, storeMp *mlist.MapList, accSlc []string, accModified []*mlist.MapList) {
	for i := range accSlc {
		var kvCount int
		storeMp.Exec(func(k string, e interface{}) {
			retV := remote.AccGetKv(accSlc[i], k)
			kvCount++
			var ebytes []byte
			if len(accModified) > 0 {
				canModify, canDel := testStoreModifyOrDel(kvCount)
				if canModify {
					e, ok := accModified[i].Get(k)
					ebytes = e.([]byte)
					if !ok {
						t.Error("not find in accModified:", k)
						return
					}
				}
				if canDel {
					// deled
					if len(retV) != 0 {
						t.Error("acc:", accSlc[i], ",key:", k, ",wrong value:", string(retV), ",should be deleted:", string(retV))
					}
					return
				}
			}
			if len(ebytes) == 0 {
				ebytes, _ = e.([]byte)
			}
			if !bytes.Equal(retV, ebytes) {
				t.Error("acc:", accSlc[i], ",key:", k, ",wrong value:", string(retV), ",should be:", string(ebytes))
				return
			}
		})
		kvCount = 0
	}
}

func remoteModifyDelAccKv(t *testing.T, remote *RemoteState, storeMp *mlist.MapList, accSlc []string) ([]byte, error, []*mlist.MapList) {
	accModified := _remoteModifyDelAccKv(t, remote, storeMp, accSlc)
	bys, err := remote.Commit()
	return bys, err, accModified
}

func _remoteModifyDelAccKv(t *testing.T, remote *RemoteState, storeMp *mlist.MapList, accSlc []string) []*mlist.MapList {
	accModified := make([]*mlist.MapList, len(accSlc))
	for i := range accSlc {
		accModified[i] = mlist.NewMapList()
		var kvCount int
		storeMp.Exec(func(k string, e interface{}) {
			kvCount++
			canModify, canDel := testStoreModifyOrDel(kvCount)
			if canModify {
				modified := fmt.Sprintf(MODIFY_TAIL, kvCount)
				accModified[i].Set(k, []byte(modified))
				remote.AccSetKv(accSlc[i], k, []byte(modified))
			} else if canDel {
				accModified[i].Set(k, nil)
				remote.AccDelKv(accSlc[i], k)
			}
		})
		kvCount = 0
	}
	return accModified
}

func TestRemoteState(t *testing.T) {
	remote, storeMp, accSlc := GetInitSt()
	root, err := remoteTransferToCreateAcc(t, remote, accSlc)
	t.Log("root:", hex.EncodeToString(root), ",err:", err)

	remote.Load(root)
	remoteCheckTransferCreate(t, remote, accSlc)

	root2, err2 := remoteSetAccKv(remote, storeMp, accSlc)
	t.Log("root2:", hex.EncodeToString(root2), ",err2:", err2)
	if bytes.Equal(root, root2) {
		t.Error("commit err, changed but get same root")
		return
	}

	remote.Load(root2)
	remoteCheckAccKvStored(t, remote, storeMp, accSlc, nil)

	root3, err3, accModified := remoteModifyDelAccKv(t, remote, storeMp, accSlc)
	t.Log("root3:", hex.EncodeToString(root3), ",err3:", err3)
	if bytes.Equal(root2, root3) {
		t.Error("commit err, changed but get same root")
		return
	}

	remote.Load(root3)
	remoteCheckAccKvStored(t, remote, storeMp, accSlc, accModified)
}

func TestSnapshot(t *testing.T) {
	remote, storeMp, accSlc := GetInitSt()
	accSlc = accSlc[:1]
	root, err := remoteTransferToCreateAcc(t, remote, accSlc)
	t.Log("root:", hex.EncodeToString(root), ",err:", err)
	root2, err2 := remoteSetAccKv(remote, storeMp, accSlc)
	t.Log("root2:", hex.EncodeToString(root2), ",err2:", err2)
	remoteCheckAccKvStored(t, remote, storeMp, accSlc, nil)

	snapshot := remote.Snapshot()

	modifiedMp := _remoteModifyDelAccKv(t, remote, storeMp, accSlc)
	t.Log("modified length:", modifiedMp[0].Len())
	remoteCheckAccKvStored(t, remote, storeMp, accSlc, modifiedMp)

	remote.RevertToSnapshot(snapshot)
	//remote.Load(root2)
	remoteCheckAccKvStored(t, remote, storeMp, accSlc, nil)
}
