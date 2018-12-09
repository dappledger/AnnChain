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
