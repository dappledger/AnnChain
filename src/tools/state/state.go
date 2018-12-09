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
	"fmt"
	"sync"

	"github.com/tendermint/merkleeyes/iavl"
	"github.com/tendermint/tmlibs/db"

	"github.com/dappledger/AnnChain/module/xlib"
	"github.com/dappledger/AnnChain/module/xlib/mlist"
	cvtypes "github.com/dappledger/AnnChain/src/types"
)

type State struct {
	mtx      sync.Mutex
	database db.DB
	rootHash []byte
	trie     *iavl.IAVLTree

	// dirty just holds everything that will be write down to disk on commit
	dirty *mlist.MapList
	// dataCache is just an effecient improvement
	dataCache *mlist.MapList
}

var (
	ErrKeyExisted          = fmt.Errorf("key already existed")
	ErrKeyNotExist         = fmt.Errorf("key not existed")
	ErrInvalidNonce        = fmt.Errorf("invalid nonce")
	ErrInsufficientBalance = fmt.Errorf("balance is lower than the amount to be executed")
)

func NewState(database db.DB) *State {
	state := &State{
		database:  database,
		trie:      iavl.NewIAVLTree(1024, database),
		dirty:     mlist.NewMapList(),
		dataCache: mlist.NewMapList(),
	}
	return state
}

func (os *State) Lock() {
	os.mtx.Lock()
}

func (os *State) Unlock() {
	os.mtx.Unlock()
}

func (os *State) CreateData(acc cvtypes.StateDataItfc) (err error) {
	os.mtx.Lock()
	accID := acc.Key()
	if os.dataCache.Has(accID) || os.trie.Has([]byte(accID)) {
		os.mtx.Unlock()
		return ErrKeyExisted
	}
	os.dataCache.Set(accID, acc)
	os.dirty.Set(accID, nil)
	os.mtx.Unlock()
	return
}

func (os *State) GetData(accID string, fromBytes cvtypes.FromBytesFunc) (acc cvtypes.StateDataItfc, err error) {
	os.mtx.Lock()
	if acdata, exist := os.dataCache.Get(accID); exist {
		os.mtx.Unlock()
		if acdata != nil {
			acc, _ = acdata.(cvtypes.StateDataItfc)
		} else {
			err = ErrKeyNotExist
		}
		return
	}
	_, bytes, exist := os.trie.Get([]byte(accID))
	if !exist {
		err = ErrKeyNotExist
		os.mtx.Unlock()
		return
	}
	if acc, err = fromBytes(accID, bytes); err != nil {
		os.mtx.Unlock()
		return nil, err
	}
	os.dataCache.Set(accID, acc)
	os.mtx.Unlock()
	return
}

func (os *State) ExistData(accID string) bool {
	os.mtx.Lock()
	defer os.mtx.Unlock()

	if os.dataCache.Has(accID) {
		return true
	}
	return os.trie.Has([]byte(accID))
}

// RemoveData acts in a sync-block way
// remove related bufferred data and remove the data from db immediately
func (os *State) RemoveData(accID string) bool {
	os.mtx.Lock()
	os.dirty.Del(accID)
	os.dataCache.Del(accID)
	_, removed := os.trie.Remove([]byte(accID))
	os.mtx.Unlock()
	return removed
}

// Commit returns the new root bytes
func (os *State) Commit() ([]byte, error) {
	os.mtx.Lock()
	var err error
	os.dirty.Exec(func(key string, e interface{}) {
		accData, exist := os.dataCache.Get(key)
		if !exist {
			fmt.Println("not hit")
			return
		}
		if xlib.CheckItfcNil(accData) {
			os.trie.Remove([]byte(key))
			return
		}
		if acc, ok := accData.(cvtypes.StateDataItfc); ok {
			if err = acc.OnCommit(); err != nil {
				// TODO do revert
				fmt.Println("commit err:", err)
				return
			}
			var serial []byte
			if serial, err = acc.Bytes(); err != nil {
				// TODO do revert
				fmt.Println("serial err:", err)
				return
			}
			if len(serial) == 0 {
				os.trie.Remove([]byte(key))
			} else {
				os.trie.Set([]byte(key), serial)
			}
		}
	})
	os.dirty.Reset()
	os.rootHash = os.trie.Save()
	os.clearJournal()
	os.mtx.Unlock()
	return os.rootHash, nil
}

func (os *State) clearJournal() {
}

// Load dumps all the buffer, start every thing from a clean slate
func (os *State) Load(root []byte) {
	os.mtx.Lock()
	os.dataCache.Reset()
	os.dirty.Reset()
	os.trie.Load(root)
	os.mtx.Unlock()
}

// Reload works the same as Load, just for semantic purpose
func (os *State) Reload(root []byte) {
	os.Lock()
	//os.dataCache.Reset()
	os.dirty.Reset()
	os.trie.Load(root)
	os.Unlock()
}

func (os *State) ModifyData(key string, data cvtypes.StateDataItfc) {
	os.Lock()
	os.dataCache.Set(key, data)
	os.dirty.Set(key, nil)
	os.Unlock()
}

// MarkModified puts accounts into dirty cache and they will be persisted during commit
func (os *State) MarkModified(key string) {
	os.Lock()
	os.dirty.Set(key, nil)
	os.Unlock()
}

func (os *State) Copy() (cp *State) {
	os.mtx.Lock()
	cp = &State{
		database:  os.database,
		rootHash:  os.rootHash,
		trie:      os.trie.Copy().(*iavl.IAVLTree),
		dataCache: mlist.NewMapList(),
		dirty:     mlist.NewMapList(),
	}

	os.dataCache.Exec(func(key string, ac interface{}) {
		if xlib.CheckItfcNil(ac) {
			cp.dirty.Set(key, nil)
			return
		}
		if vac, ok := ac.(cvtypes.StateDataItfc); ok {
			cp.dataCache.Set(key, vac.Copy())
			// don't double copy the data
			if os.dirty.Has(key) {
				cp.dirty.Set(key, nil)
			}
		}
	})

	os.mtx.Unlock()
	return
}
