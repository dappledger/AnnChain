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
	"bytes"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/tendermint/tmlibs/db"
	"go.uber.org/zap"

	cmn "gitlab.zhonganonline.com/ann/ann-module/lib/go-common"
	"gitlab.zhonganonline.com/ann/ann-module/lib/go-config"
	"gitlab.zhonganonline.com/ann/ann-module/lib/go-wire"
)

var lastBlockKey = []byte("lastblock")

type Application interface {
	GetAngineHooks() Hooks
	CompatibleWithAngine()
	CheckTx([]byte) error
	Query([]byte) Result
	Info() ResultInfo
	Start()
	Stop()
	SetCore(Core)
	Initialized() bool
}

type AppMaker func(*zap.Logger, config.Config) Application

// -------------- BaseApplication ---------------

type BaseApplication struct {
	Database         db.DB
	InitializedState bool
}

func (ba *BaseApplication) InitBaseApplication(name string, datadir string) (err error) {
	if ba.Database, err = db.NewGoLevelDB(name, datadir); err != nil {
		return err
	}
	return nil
}

func (ba *BaseApplication) LoadLastBlock(t interface{}) (res interface{}, err error) {
	buf := ba.Database.Get(lastBlockKey)
	if len(buf) != 0 {
		r, n, err := bytes.NewReader(buf), new(int), new(error)
		res = wire.ReadBinaryPtr(t, r, 0, n, err)
		if *err != nil {
			return nil, *err
		}
	} else {
		return nil, errors.New("empty")
	}
	return
}

func (ba *BaseApplication) SaveLastBlock(lastBlock interface{}) {
	buf, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(lastBlock, buf, n, err)
	if *err != nil {
		cmn.PanicCrisis(*err)
	}
	ba.Database.SetSync(lastBlockKey, buf.Bytes())
}

func (ba *BaseApplication) Initialized() bool {
	return ba.InitializedState
}

// ------------ CommApplication --------------

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
