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

	cmn "gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-common"
	"gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-config"
	"gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-db"
	"gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-wire"
)

var lastBlockKey = []byte("lastblock")

type Application interface {
	GetAngineHooks() Hooks
	CompatibleWithAngine()
	CheckTx([]byte) Result
	QueryNonce(string) Result
	QueryAccount(string) Result
	QueryLedgers(string, uint64, uint64) Result
	QueryLedger(uint64) Result
	QueryPayments(string, uint64, uint64) Result
	QueryAccountPayments(string, string, uint64, uint64) Result
	QueryPayment(string) Result
	QueryTransactions(string, uint64, uint64) Result
	QueryTransaction(string) Result
	QueryAccountTransactions(string, string, uint64, uint64) Result
	QueryLedgerTransactions(uint64, string, uint64, uint64) Result
	QueryDoContract([]byte) Result
	QueryContractExist(string) Result
	QueryReceipt(string) Result
	QueryAccountManagedatas(string, string, uint64, uint64) Result
	QueryAccountManagedata(string, string) Result
	Info() ResultInfo
	Start()
	Stop()
}

type AppMaker func(config.Config) Application

type BaseApplication struct {
	Database db.DB
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
