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

package datamanager

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"sync"

	"github.com/dappledger/AnnChain/ann-module/lib/go-config"
	"github.com/dappledger/AnnChain/genesis/chain/database"
	"github.com/dappledger/AnnChain/genesis/chain/database/basesql"
	"github.com/dappledger/AnnChain/genesis/chain/log"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
	_ "github.com/mattn/go-sqlite3"
)

func getDM() *DataManager {
	cfg := config.NewMapConfig(nil)
	cfg.Set("db_type", "sqlite3")
	dbDir := fmt.Sprintf("D:/sqlitedb/sqliteDB")
	if err := os.Mkdir(dbDir, 0777); err != nil {
		panic(err)
	}

	cfg.Set("db_dir", dbDir)

	logger := log.Initialize("", "", "output.log", "err.log")

	if _, err := splitlog.NewLog("D:/sqlitedb/", "genesis", 2); err != nil {
		panic(err)
	}

	dm, err := NewDataManager(cfg, logger, func(dbname string) database.Database {
		dbi := &basesql.Basesql{}
		err := dbi.Init(dbname, cfg, logger)
		if err != nil {
			panic(err)
		}
		return dbi
	})
	if err != nil {
		panic(err)
	}

	return dm
}

func TestAsync(t *testing.T) {

	var wg sync.WaitGroup

	dm := getDM()

	wg.Add(1)
	go func() {

		dm.QTxBegin()
		for i := 0; i < 1000000; i++ {
			data := &types.TransactionData{
				TxID:            uint64(i),
				TxHash:          ethcmn.BytesToHash([]byte("")),
				LedgerHash:      ethcmn.BytesToHash([]byte("")),
				Height:          big.NewInt(int64(i)),
				CreateDate:      time.Now().Unix(),
				Account:         ethcmn.BytesToAddress([]byte("123")),
				AccountSequence: big.NewInt(int64(i)),
				FeePaid:         big.NewInt(int64(100)),
				ResultCode:      1,
				ResultCodes:     "21",
				Memo:            "sdfasdfaf",
			}
			if _, err := dm.AddTransaction(data); err != nil {
				t.Log(err)
			}
		}
		dm.QTxCommit()
		wg.Done()
	}()

	wg.Wait()

}
