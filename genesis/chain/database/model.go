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

package database

import (
	"database/sql"
	"time"
)

// AccData object for db
type AccData struct {
	DataID    uint64 `db:"dataid"`
	AccountID string `db:"accountid"`
	DataKey   string `db:"datakey"`
	DataValue string `db:"datavalue"`
	Category  string `db:"category"`
}

// LedgerHeader object for db
type LedgerHeader struct {
	LedgerID         uint64    `db:"ledgerid"`
	Height           string    `db:"height"`
	Hash             string    `db:"hash"`
	PrevHash         string    `db:"prevhash"`
	TransactionCount uint64    `db:"transactioncount"`
	ClosedAt         time.Time `db:"closedat"`
	TotalCoins       string    `db:"totalcoins"`
	BaseFee          string    `db:"basefee"`
	MaxTxSetSize     uint64    `db:"maxtxsetsize"`
}

// TxData object for db
type TxData struct {
	TxID            uint64 `db:"txid"`
	TxHash          string `db:"txhash"`
	LedgerHash      string `db:"ledgerhash"`
	Height          string `db:"height"`
	CreateDate      int64  `db:"createdat"`
	Account         string `db:"account"`
	Target          string `db:"target"`
	OpType          string `db:"optype"`
	AccountSequence string `db:"accountsequence"`
	FeePaid         string `db:"feepaid"`
	ResultCode      uint   `db:"resultcode"`
	ResultCodes     string `db:"resultcodes"`
	Memo            string `db:"memo"`
}

// Action object for db
type Action struct {
	ActionID    uint64         `db:"actionid"`
	Typei       int            `db:"typei"`
	Type        string         `db:"type"`
	Height      string         `db:"height"`
	TxHash      string         `db:"txhash"`
	FromAccount sql.NullString `db:"fromaccount"`
	ToAccount   sql.NullString `db:"toaccount"`
	CreateAt    uint64         `db:"createat"`
	JData       string         `db:"jdata"`
}

// Effect object for db
type Effect struct {
	EffectID uint64 `db:"effectid"`
	Typei    int    `db:"typei"`
	Type     string `db:"type"`
	Height   string `db:"height"`
	TxHash   string `db:"txhash"`
	ActionID uint64 `db:"actionid"`
	Account  string `db:"account"`
	CreateAt uint64 `db:"createat"`
	JData    string `db:"jdata"`
}
