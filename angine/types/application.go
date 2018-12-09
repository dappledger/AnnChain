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

package types

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/tendermint/tmlibs/db"
	"github.com/vmihailenco/msgpack"

	cmn "github.com/dappledger/AnnChain/module/lib/go-common"
)

var lastBlockKey = []byte("lastblock")

type Application interface {
	GetAngineHooks() Hooks
	CompatibleWithAngine()
	CheckTx([]byte) error
	Query([]byte) Result
	Info() ResultInfo
	Start() error
	Stop()
	Initialized() bool
}

// BaseApplication defines the default save/load last block implementations
// You can write all on your own, but embed this will save u some breath
type BaseApplication struct {
	Database         db.DB
	InitializedState bool
}

// ValSetLoader default funtion that return nil
func (ba *BaseApplication) ValSetLoader() func(height, round, size int) *ValidatorSet {
	return nil
}

// SuspectValidator suggest punishment when needed
func (ba *BaseApplication) SuspectValidator(pubkey []byte, reason string) {
	// Nothing to do here
	fmt.Printf("punish: %X\n", pubkey)
}

// InitBaseApplication must be the first thing to be called when an application embeds BaseApplication
func (ba *BaseApplication) InitBaseApplication(name string, datadir string) (err error) {
	if ba.Database, err = db.NewGoLevelDB(name, datadir); err != nil {
		return err
	}
	ba.InitializedState = true
	return nil
}

func (ba *BaseApplication) LoadLastBlock(t interface{}) (res interface{}, err error) {
	bys := ba.Database.Get(lastBlockKey)
	if bys != nil && len(bys) != 0 {
		err = msgpack.Unmarshal(bys, t)
		res = t
	} else {
		err = errors.New("empty")
	}
	return
}

func (ba *BaseApplication) SaveLastBlock(lastBlock interface{}) {
	ba.SaveLastBlockByKey(lastBlockKey, lastBlock)
}

func (ba *BaseApplication) SaveLastBlockByKey(key []byte, lastBlock interface{}) {
	bys, err := msgpack.Marshal(lastBlock)
	if err != nil {
		cmn.PanicCrisis(err)
	}
	ba.Database.SetSync(key, bys)
}

// Initialized returns if a BaseApplication based app has been fully initialized
func (ba *BaseApplication) Initialized() bool {
	return ba.InitializedState
}

// Stop handles freeing resources taken by BaseApplication
func (ba *BaseApplication) Stop() {
	ba.Database.Close()
}

// CommApplication defines an app with basic net io ability
type CommApplication struct {
	mtx      sync.Mutex
	Listener net.Listener
}

func (ca *CommApplication) Lock() {
	ca.mtx.Lock()
}

func (ca *CommApplication) Unlock() {
	ca.mtx.Unlock()
}

func (ca *CommApplication) Listen(addr string) (net.Listener, error) {
	var tryListenSeconds = 10
	var listener net.Listener
	var err error

	for i := 0; i < tryListenSeconds; i++ {
		listener, err = net.Listen("tcp", addr)
		if err == nil {
			break
		}
		if i < tryListenSeconds {
			time.Sleep(1 * time.Second)
		}
	}
	if err != nil {
		return nil, err
	}

	// TODO: UPnP ?
	ca.Listener = listener

	return listener, nil
}

func (ca *CommApplication) Stop() {
	ca.Listener.Close()
}
