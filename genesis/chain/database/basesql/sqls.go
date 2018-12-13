package basesql

// GetInitSQLs get database initialize sqls
//	opt sqls to create operation tables
//	opi sqls to create operation table-indexs
//	qt  sqls to create query tables
//	qi  sqls to create query table-indexs
func (bs *Basesql) GetInitSQLs() (opt, opi, qt, qi []string) {
	opt = []string{
		createAccDataSQL,
	}
	opi = createOpIndexs

	qt = []string{
		createEffectSQL,
		createLedgerSQL,
		creatActionSQL,
		createTransactionSQL,
	}
	qi = createQIndex

	return
}

var (
	createOpIndexs = []string{
		// index for table accdata
		"CREATE INDEX IF NOT EXISTS accdateaccid ON accdata (accountid)",
	}

	createQIndex = []string{
		// indexs for table ledger
		"CREATE INDEX IF NOT EXISTS height ON ledgerheaders (height)",

		// indexs for table transactions
		"CREATE INDEX IF NOT EXISTS txtxhash ON transactions (txhash)",
		"CREATE INDEX IF NOT EXISTS txaccount ON transactions (account)",

		// indexs for table actions
		"CREATE INDEX IF NOT EXISTS actionstxhash ON actions (txhash)",
		"CREATE INDEX IF NOT EXISTS actionsfromaccount ON actions (fromaccount)",
		"CREATE INDEX IF NOT EXISTS actionstoaccount ON actions (toaccount)",
		"CREATE INDEX IF NOT EXISTS actionscreateat ON actions (createat)",

		// indexs for table effects
		"CREATE INDEX IF NOT EXISTS effectstxhash ON effects (txhash)",
		"CREATE INDEX IF NOT EXISTS effectsaccount ON effects (account)",
		"CREATE INDEX IF NOT EXISTS effectscreateat ON effects (createat)",
	}
)

const (
	createAccDataSQL = `CREATE TABLE IF NOT EXISTS accdata
    (
		dataid			INTEGER 		PRIMARY KEY	AUTOINCREMENT,
		accountid       VARCHAR(66)		NOT NULL,
		datakey			VARCHAR(256)	NOT NULL,
		datavalue		VARCHAR(256)	NOT NULL,
		category		VARCHAR(256)	NOT NULL
	);`

	//========================================================================//

	creatActionSQL = `CREATE TABLE IF NOT EXISTS actions
	(
		actionid			INTEGER	PRIMARY KEY	AUTOINCREMENT,
		typei				INT			NOT NULL,
		type				VARCHAR(32)	NOT NULL,
		height			INT			NOT NULL,
		txhash				VARCHAR(64)	NOT NULL,
		fromaccount			VARCHAR(66),			-- only used in payment
		toaccount			VARCHAR(66),			-- only used in payment
		createat			INT			NOT NULL,
		jdata				TEXT		NOT NULL
	);`

	createEffectSQL = `CREATE TABLE IF NOT EXISTS effects
	(
		effectid          	INTEGER	PRIMARY KEY	AUTOINCREMENT,
		typei				INT,
		type				VARCHAR(32)	NOT NULL,
		height			INT			NOT NULL,
		txhash				VARCHAR(64)	NOT NULL,
		actionid			INT			NOT NULL,
		account				VARCHAR(66)	NOT NULL,
		createat			INT			NOT NULL,
		jdata				TEXT		NOT NULL
	);`

	createLedgerSQL = `CREATE TABLE IF NOT EXISTS ledgerheaders
    (
		ledgerid			INTEGER	PRIMARY KEY	AUTOINCREMENT,
		height			TEXT 		UNIQUE,
		hash				VARCHAR(64) NOT NULL,
		prevhash			VARCHAR(64) NOT NULL,
		transactioncount	INT			NOT NULL,
		closedat			TIMESTAMP 	NOT NULL,
		totalcoins			TEXT	 	NOT NULL,
		basefee				TEXT 		NOT NULL,
		maxtxsetsize		INT 		NOT NULL
	);`

	createTransactionSQL = `CREATE TABLE IF NOT EXISTS transactions
	(
		txid				INTEGER	PRIMARY KEY	AUTOINCREMENT,
		txhash				VARCHAR(64)	NOT NULL,
		ledgerhash			VARCHAR(64)	NOT NULL,
		height				TEXT		NOT NULL,
		createdat			INTEGER	NOT NULL,
		account				VARCHAR(66)	NOT NULL,
		target				VARCHAR(66)	NOT NULL,
		optype				TEXT		NOT NULL,
		accountsequence		TEXT		NOT NULL,
		feepaid				TEXT		NOT NULL,
		resultcode			INT			NOT NULL,
		resultcodes			TEXT		NOT NULL,
		memo				TEXT
	);`
)
