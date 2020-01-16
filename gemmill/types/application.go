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

	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/modules/go-db"
)

var lastBlockKey = []byte("lastblock")

type TxPoolApplication interface {
	Application
	GetTxPool() TxPool
}

type Application interface {
	GetAngineHooks() Hooks
	CompatibleWithAngine()
	CheckTx(bs []byte) (from common.Address,nonce uint64, err error)
	Query([]byte) Result
	Info() ResultInfo
	Start() error
	Stop()
	SetCore(Core)
}

type Core interface {
	Query(byte, []byte) (interface{}, error)
	GetBlockMeta(height int64) (*BlockMeta, error)
}

// type AppMaker func(config.Config) Application

type BaseApplication struct {
	Database         db.DB
	InitializedState bool
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
	ba.SaveLastBlockByKey(lastBlockKey, lastBlock)
}

func (ba *BaseApplication) SaveLastBlockByKey(key []byte, lastBlock interface{}) {
	buf, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(lastBlock, buf, n, err)
	if *err != nil {
		gcmn.PanicCrisis(*err)
	}
	ba.Database.SetSync(key, buf.Bytes())
}

// Initialized returns if a BaseApplication based app has been fully initialized
func (ba *BaseApplication) Initialized() bool {
	return ba.InitializedState
}

// Stop handles freeing resources taken by BaseApplication
func (ba *BaseApplication) Stop() {
	ba.Database.Close()
}
