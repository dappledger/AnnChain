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

//func TestTrustline(t *testing.T) {
//	dm := getDM()

//	td := &types.TrustData{
//		Account: ethcmn.StringToAddress("AccountID"),
//		Asset: types.Asset{
//			Type:   1,
//			Code:   "USD",
//			Issuer: ethcmn.StringToAddress("Issuer"),
//		},
//		Balance:      ethcmn.Big("0"),
//		Limit:        ethcmn.Big("200000000000000000"),
//		Flags:        types.TYPE_TRUST_FLAG(2),
//		LastModified: time.Now().Unix(),
//	}

//	// Insert
//	_, err := dm.AddTrustData(td)
//	if err != nil {
//		t.Fatal(err)
//	}

//	// Query
//	qtd, err := dm.LoadTrustData(&td.Account, &td.Asset.Issuer, td.Asset.Code)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if qtd == nil {
//		t.Fatal("insert: not exist")
//	}
//	if !reflect.DeepEqual(qtd.Account, td.Account) ||
//		!reflect.DeepEqual(qtd.Asset, td.Asset) ||
//		!reflect.DeepEqual(qtd.Balance, td.Balance) ||
//		!reflect.DeepEqual(qtd.Limit, td.Limit) ||
//		!reflect.DeepEqual(qtd.Flags, td.Flags) {

//		fmt.Println("before", td.Account.Hex())
//		fmt.Println("after", qtd.Account.Hex())
//		t.Fatal("insert: data incorrect")
//	}

//	// Update
//	qtd.Balance.SetInt64(3000000004000500009)
//	err = dm.UpdateTrustData(qtd)
//	if err != nil {
//		t.Fatal(err)
//	}
//	qtd2, err := dm.LoadTrustData(&td.Account, &td.Asset.Issuer, td.Asset.Code)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if !reflect.DeepEqual(qtd2.Balance, new(big.Int).SetInt64(3000000004000500009)) {
//		t.Fatal("update: failed")
//	}

//	// Delete
//	err = dm.DropTrustData(&td.Account, &td.Asset.Issuer, td.Asset.Code)
//	if err != nil {
//		t.Fatal(err)
//	}
//	qtd, err = dm.LoadTrustData(&td.Account, &td.Asset.Issuer, td.Asset.Code)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if qtd != nil {
//		t.Fatal("delete: already exist")
//	}
//}

//func TestLedger(t *testing.T) {
//	dm := getDM()

//	var rbak *types.LedgerHeaderData

//	// Insert 10 recrods
//	seq := time.Now().UnixNano()
//	for i := int64(0); i < 10; i++ {
//		hd := &types.LedgerHeaderData{
//			Sequence:         big.NewInt(seq + i),
//			Hash:             ethcmn.BigToHash(big.NewInt(seq + i)),
//			PrevHash:         ethcmn.BigToHash(big.NewInt(seq)),
//			TransactionCount: uint64(rand.Int() % 1000),
//			OperactionCount:  uint64(rand.Int() % 10000),
//			ClosedAt:         time.Now(),
//			TotalCoins:       big.NewInt(rand.Int63() % 10000000000),
//			FeePool:          big.NewInt(rand.Int63() % 10000000000),
//			BaseFee:          big.NewInt(200),
//			BaseReserve:      big.NewInt(20000000000),
//			MaxTxSetSize:     10000,
//		}

//		_, err := dm.AddLedgerHeaderData(hd)
//		if err != nil {
//			t.Fatal(err)
//		}

//		rbak = hd
//	}

//	// Query Single
//	d, err := dm.QueryLedgerHeaderData(rbak.Sequence)
//	if err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println("ledgerid:", d.LedgerID)
//	fmt.Println("t_prev:", rbak.ClosedAt.UnixNano())
//	fmt.Println("t_curr:", d.ClosedAt.UnixNano())
//	rbak.LedgerID = d.LedgerID // auto_increament
//	rbak.ClosedAt = d.ClosedAt // time struct is complicated
//	if !reflect.DeepEqual(d, rbak) {
//		fmt.Println("prev:", rbak)
//		fmt.Println("curr:", d)
//		t.Fatal("query single: diffenent from the one inserted")
//	}

//	// Query asc
//	ds, err := dm.QueryAllLedgerHeaderData(2, 7, "asc")
//	if err != nil {
//		t.Fatal(err)
//	}
//	lastID := ds[0].LedgerID
//	for _, v := range ds {
//		fmt.Println("Query asc:", v.LedgerID)
//		if v.LedgerID < lastID {
//			t.Fatal("Query asc: failed")
//		}
//	}

//	// Query desc
//	ds, err = dm.QueryAllLedgerHeaderData(9, 5, "desc")
//	if err != nil {
//		t.Fatal(err)
//	}
//	lastID = ds[0].LedgerID
//	for _, v := range ds {
//		fmt.Println("Query desc:", v.LedgerID)
//		if v.LedgerID > lastID {
//			t.Fatal("Query desc: failed")
//		}
//	}
//}

//func TestNothing(t *testing.T) {

//}
