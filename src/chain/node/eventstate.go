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


package node

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/tendermint/merkleeyes/iavl"
	"github.com/tendermint/tmlibs/db"
)

type CodeHash = []byte

// EventState manages event subscription relationships between subchains
type EventState struct {
	mtx      sync.Mutex
	database db.DB
	rootHash []byte
	trie     *iavl.IAVLTree

	dirty        map[string]struct{}
	order        []string
	accountCache map[string]*EventAccount
}

// EventAccount abstracts the model used in the EventState, mainly focusing on who am I listening and who is listening me
type EventAccount struct {
	master *EventState

	ChainID  string
	MySubers map[string]CodeHash
	MyPubers map[string]CodeHash
}

func NewEventState(database db.DB) *EventState {
	return &EventState{
		database:     database,
		trie:         iavl.NewIAVLTree(1024, database),
		accountCache: make(map[string]*EventAccount),
		dirty:        make(map[string]struct{}),
		order:        make([]string, 0),
	}
}

func (es *EventState) Lock() {
	es.mtx.Lock()
}

func (es *EventState) Unlock() {
	es.mtx.Unlock()
}

func (es *EventState) ModifyAccount(a *EventAccount) {
	es.accountCache[a.ChainID] = a
	if _, ok := es.dirty[a.ChainID]; !ok {
		es.dirty[a.ChainID] = struct{}{}
		es.order = append(es.order, a.ChainID)
	}
}

func (es *EventState) CreateAccount(chainID string) (accnt *EventAccount, err error) {
	es.mtx.Lock()
	_, ok := es.accountCache[chainID]
	if ok || es.trie.Has([]byte(chainID)) {
		es.mtx.Unlock()
		return nil, ErrAccountExisted
	}
	accnt = NewEventAccount(es, chainID)
	es.accountCache[chainID] = accnt
	es.dirty[chainID] = struct{}{}
	es.order = append(es.order, chainID)
	es.mtx.Unlock()
	return
}

func (es *EventState) GetAccount(chainID string) (accnt *EventAccount, err error) {
	es.mtx.Lock()
	defer es.mtx.Unlock()
	accnt, ok := es.accountCache[chainID]
	if ok {
		return
	}
	_, bytes, exist := es.trie.Get([]byte(chainID))
	if !exist {
		err = ErrAccountNotExist
		return
	}
	accnt = &EventAccount{}
	if err := accnt.FromBytes(bytes, es); err != nil {
		return nil, err
	}
	es.accountCache[chainID] = accnt
	return
}

func (es *EventState) ExistAccount(chainID string) bool {
	es.mtx.Lock()
	if _, ok := es.accountCache[chainID]; ok {
		es.mtx.Unlock()
		return ok
	}
	es.mtx.Unlock()
	return es.trie.Has([]byte(chainID))
}

func (es *EventState) Commit() ([]byte, error) {
	es.mtx.Lock()
	for _, id := range es.order {
		es.trie.Set([]byte(es.accountCache[id].ChainID), es.accountCache[id].ToBytes())
	}
	es.rootHash = es.trie.Save()
	es.mtx.Unlock()
	return es.rootHash, nil
}

func (es *EventState) Load(root []byte) {
	es.mtx.Lock()
	es.accountCache = make(map[string]*EventAccount)
	es.dirty = make(map[string]struct{})
	es.order = make([]string, 0)
	es.trie.Load(root)
	es.mtx.Unlock()
}

func (es *EventState) Reload(root []byte) {
	es.mtx.Lock()
	es.accountCache = make(map[string]*EventAccount)
	es.dirty = make(map[string]struct{})
	es.order = make([]string, 0)
	es.trie.Load(root)
	es.mtx.Unlock()
}

func (es *EventState) Copy() *EventState {
	cp := &EventState{
		database:     es.database,
		rootHash:     es.rootHash,
		trie:         es.trie,
		accountCache: make(map[string]*EventAccount),
		dirty:        make(map[string]struct{}),
		order:        make([]string, len(es.order), cap(es.order)),
	}
	for k := range es.accountCache {
		cp.accountCache[k] = es.accountCache[k].Copy()
	}
	for k := range es.dirty {
		cp.dirty[k] = struct{}{}
	}
	for i, id := range es.order {
		cp.order[i] = id
	}
	return cp
}

func NewEventAccount(es *EventState, chainID string) *EventAccount {
	return &EventAccount{
		master:   es,
		ChainID:  chainID,
		MyPubers: make(map[string]CodeHash),
		MySubers: make(map[string]CodeHash),
	}
}

func (ea *EventAccount) GetSubscribers() map[string]CodeHash {
	return ea.MySubers
}

func (ea *EventAccount) GetPublishers() map[string]CodeHash {
	return ea.MyPubers
}

func (ea *EventAccount) AddSubscriber(sub string, hash CodeHash) {
	ea.MySubers[sub] = hash
	ea.master.ModifyAccount(ea)
}

func (ea *EventAccount) AddPublisher(pub string, hash CodeHash) {
	ea.MyPubers[pub] = hash
	ea.master.ModifyAccount(ea)
}

func (ea *EventAccount) RemoveSubscriber(sub string) {
	delete(ea.MySubers, sub)
	ea.master.ModifyAccount(ea)
}

func (ea *EventAccount) RemovePublisher(pub string) {
	delete(ea.MyPubers, pub)
	ea.master.ModifyAccount(ea)
}

func (ea *EventAccount) FromBytes(bs []byte, es *EventState) error {
	gdc := gob.NewDecoder(bytes.NewBuffer(bs))
	if err := gdc.Decode(ea); err != nil {
		return err
	}
	ea.master = es
	return nil
}

func (ea *EventAccount) ToBytes() []byte {
	buf := &bytes.Buffer{}
	gec := gob.NewEncoder(buf)
	if err := gec.Encode(ea); err != nil {
		return nil
	}
	return buf.Bytes()
}

func (ea *EventAccount) Copy() *EventAccount {
	cp := &EventAccount{
		master:   ea.master,
		ChainID:  ea.ChainID,
		MySubers: make(map[string]CodeHash),
		MyPubers: make(map[string]CodeHash),
	}

	for k, v := range ea.MySubers {
		cp.MySubers[k] = v
	}
	for k, v := range ea.MyPubers {
		cp.MyPubers[k] = v
	}

	return cp
}
