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

package app

import (
	"bytes"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"encoding/binary"
	"encoding/hex"
	"strconv"

	at "github.com/dappledger/AnnChain/angine/types"
	cmn "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	cfg "github.com/dappledger/AnnChain/ann-module/lib/go-config"
	"github.com/dappledger/AnnChain/ann-module/lib/go-merkle"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
	"github.com/dappledger/AnnChain/ann-module/xlib"
	dcfg "github.com/dappledger/AnnChain/genesis/chain/config"
	"github.com/dappledger/AnnChain/genesis/chain/database"
	"github.com/dappledger/AnnChain/genesis/chain/database/basesql"
	"github.com/dappledger/AnnChain/genesis/chain/datamanager"
	"github.com/dappledger/AnnChain/genesis/chain/version"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/eth/core/state"
	ethtypes "github.com/dappledger/AnnChain/genesis/eth/core/types"
	"github.com/dappledger/AnnChain/genesis/eth/ethdb"
	"github.com/dappledger/AnnChain/genesis/eth/rlp"
	"github.com/dappledger/AnnChain/genesis/types"
	"go.uber.org/zap"
)

const (
	OfficialAddress     = "0xed1de12230e28f561c67e63e5b765a671af2afb2"
	StateRemoveEmptyObj = false

	LDatabaseCache   = 128
	LDatabaseHandles = 1024
)

type LastBlockInfo struct {
	Height       uint64
	StateRoot    []byte
	AppHash      []byte
	TotalCoin    string
	Feepool      string
	InflationSeq uint64
}

type blockExeInfo struct {
	txDatas        []*types.TransactionData
	effectG        []*types.EffectGroup
	inflationOccur bool
}

type stateDup struct {
	height   int
	round    int
	key      string
	state    *state.StateDB
	stateMtx *sync.Mutex
	receipts []*types.Receipt
}

type GenesisApp struct {
	config cfg.Config

	stateApp    *state.StateDB
	stateAppMtx sync.Mutex // protected concurrent changes of app.state

	currentHeader *types.AppHeader
	tempHeader    *types.AppHeader // for executing tx

	blockExeInfo *blockExeInfo

	chainDb ethdb.Database // Block chain database

	tmpStateDup *stateDup

	AngineHooks at.Hooks
	opM         OperationManager

	dataM *datamanager.DataManager

	txCache *cmn.CMap

	EvmCurrentHeader *ethtypes.Header

	Init_Accounts []at.InitInfo
}

var (
	EmptyTrieRoot  = ethcmn.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	ReceiptsPrefix = []byte("receipts-")
	lastBlockKey   = []byte("lastblock")
	logger         *zap.Logger
)

func init() {}

func newStateDup(state *state.StateDB, block *at.Block, height, round int) *stateDup {

	stateCopy := state.DeepCopy()

	if stateCopy == nil {
		cmn.PanicCrisis("state deep copy failed")
	}

	return &stateDup{
		height:   height,
		round:    round,
		key:      stateKey(block),
		state:    stateCopy,
		stateMtx: &sync.Mutex{},
	}
}

func stateKey(block *at.Block) string {
	return ethcmn.Bytes2Hex(block.Hash())
}

func OpenDatabase(datadir string, name string, cache int, handles int) (ethdb.Database, error) {
	return ethdb.NewLDBDatabase(filepath.Join(datadir, name), cache, handles)
}

func NewGenesisApp(config cfg.Config, _logger *zap.Logger) *GenesisApp {

	var err error

	datadir := config.GetString("db_dir")

	app := GenesisApp{
		config: config,
	}

	if app.chainDb, err = OpenDatabase(datadir, "chaindata", LDatabaseCache, LDatabaseHandles); err != nil {
		cmn.PanicCrisis(err)
	}
	lastBlock := app.LoadLastBlock()

	trieRoot := EmptyTrieRoot

	if len(lastBlock.StateRoot) > 0 {
		trieRoot = ethcmn.BytesToHash(lastBlock.StateRoot)
	}

	if app.stateApp, err = state.New(trieRoot, app.chainDb); err != nil {
		cmn.PanicCrisis(err)
	}

	app.blockExeInfo = &blockExeInfo{}

	lastBlockTotalCoin, _ := big.NewInt(0).SetString(lastBlock.TotalCoin, 10)

	lastBlockFeePool, _ := big.NewInt(0).SetString(lastBlock.Feepool, 10)

	app.currentHeader = &types.AppHeader{
		PrevHash:  ethcmn.BytesToLedgerHash(lastBlock.AppHash),
		TotalCoin: lastBlockTotalCoin,
		Feepool:   lastBlockFeePool,
		Height:    new(big.Int),
		BaseFee:   new(big.Int),
	}

	app.tempHeader = app.currentHeader

	if app.Init_Accounts, err = dcfg.GetInitialIssueAccount(config); err != nil {
		cmn.PanicCrisis(fmt.Errorf("fail to setup initial accounts, error: %s", err.Error()))
	}

	if config.GetBool("init_official") && trieRoot == EmptyTrieRoot {
		//initial issue lumens to accounts get from initialFile
		totalcoin := new(big.Int).SetUint64(0)

		for idx := range app.Init_Accounts {
			addr := ethcmn.HexToAddress(app.Init_Accounts[idx].Address)
			app.stateApp.CreateAccount(addr)
			amount, succ := new(big.Int).SetString(app.Init_Accounts[idx].StartingBalance, 10)
			if !succ {
				cmn.PanicCrisis("fail to convert startingbalance")
			}
			app.stateApp.AddBalance(addr, amount, "init account")
			totalcoin.Add(totalcoin, amount)
		}

		app.currentHeader.TotalCoin = totalcoin

		if apphash, err := app.stateApp.Commit(StateRemoveEmptyObj); err != nil {
			cmn.PanicCrisis(fmt.Errorf("fail to setup initial funds, error: %s", err.Error()))
		} else {
			app.stateApp, _ = app.stateApp.New(apphash)
		}

	}
	// initialize data manager
	app.dataM, err = datamanager.NewDataManager(config, _logger, func(dbname string) database.Database {
		dbi := &basesql.Basesql{}
		err := dbi.Init(dbname, config, _logger)
		if err != nil {
			cmn.PanicCrisis(err)
		}
		return dbi
	})
	if err != nil {
		cmn.PanicCrisis(err)
	}

	app.AngineHooks = at.Hooks{
		OnNewRound: at.NewHook(app.OnNewRound),
		OnCommit:   at.NewHook(app.OnCommit),
		OnExecute:  at.NewHook(app.OnExecute),
	}

	app.opM.Init(app.dataM, &app)

	app.txCache = cmn.NewCMap()

	logger = _logger

	return &app
}

func (app *GenesisApp) Start() {
	version.InitNodeInfo("genesis")
}

func (app *GenesisApp) Stop() {
	app.chainDb.Close()
	app.dataM.Close()
}

func (app *GenesisApp) makeTempHeader(block *at.Block) {
	app.tempHeader = &types.AppHeader{
		// do not fill here
		StateRoot: app.currentHeader.StateRoot,

		// use block info
		Height:   new(big.Int).SetInt64(int64(block.Height)),
		ClosedAt: block.Header.Time,

		// dynamic get
		BaseFee: app.ParseBaseFee(block),

		MaxTxSetSize: app.ParseMaxTxSetSize(block),

		// global save
		PrevHash:  app.currentHeader.PrevHash,
		TotalCoin: app.currentHeader.TotalCoin,
		Feepool:   app.currentHeader.Feepool,
	}
}

func (app *GenesisApp) CompatibleWithAngine() {
}

func (app *GenesisApp) GetAngineHooks() at.Hooks {
	return app.AngineHooks
}

func (app *GenesisApp) checkBeforeExecute(stateDup *stateDup, bs []byte) (*types.Transaction, error) {

	var tx *types.Transaction

	// retrive if in cache
	if txbs := app.txCache.Get(string(bs)); txbs != nil {
		tx = txbs.(*types.Transaction)
	} else {
		tx = new(types.Transaction)
		err := rlp.DecodeBytes(bs, &tx.Data)
		if err != nil {
			logger.Warn("Decode Bytes  failed:" + err.Error())
			return nil, err
		}
	}

	if tx.Nonce() != stateDup.state.GetNonce(tx.GetFrom()) {
		return nil, fmt.Errorf("bad nonce")
	}

	return tx, nil
}

func (app *GenesisApp) CheckSignTx(tx *types.Transaction) at.Result {
	if err := tx.CheckSig(); err != nil {
		return at.NewError(at.CodeType_BaseInvalidSignature, err.Error())
	}
	return app.opM.PreCheck(tx)
}

// ExecuteTx execute tx one by one in the loop, without lock, so should always be called between Lock() and Unlock() on the *stateDup
func (app *GenesisApp) ExecuteTx(stateDup *stateDup, bs []byte) (err error) {

	var (
		tx *types.Transaction
	)

	if tx, err = app.checkBeforeExecute(stateDup, bs); err != nil {
		app.txCache.Delete(string(bs))
		return
	}

	state := stateDup.state

	// begin db tx
	if err = app.dataM.OpTxBegin(); err != nil {
		logger.Warn("Begin database tx failed:" + err.Error())
		app.txCache.Delete(string(bs))
		return
	}

	// begin statedb tx
	stateSnapshot := state.Snapshot()

	// take fee first
	state.SubBalance(tx.GetFrom(), tx.BaseFee(), "tx cost")

	// do execute tx
	err = app.opM.ExecTx(stateDup, tx)

	// log execute result
	txData := tx.GetDBTxData(err)

	app.blockExeInfo.txDatas = append(app.blockExeInfo.txDatas, txData)

	// check executing result
	if err != nil {
		state.RevertToSnapshot(stateSnapshot)
		app.dataM.OpTxRollback() // error is not important here
		app.txCache.Delete(string(bs))
		return
	}

	// commit db tx
	if err = app.dataM.OpTxCommit(); err != nil {
		logger.Error("Commit database tx failed:" + err.Error())
		app.txCache.Delete(string(bs))
		return
	}

	// Increment the nonce for the next transaction
	state.SetNonce(tx.GetFrom(), state.GetNonce(tx.GetFrom())+1)

	app.txCache.Delete(string(bs))

	// Collect operations and effects
	action, effects := tx.GetOperatorItfc().GetOperationEffects()

	action.GetActionBase().CreateAt = tx.GetCreateTime()

	action.GetActionBase().TxHash = tx.Hash()

	for idx := range effects {
		effects[idx].GetEffectBase().CreateAt = tx.GetCreateTime()
		effects[idx].GetEffectBase().TxHash = tx.Hash()
	}

	app.blockExeInfo.effectG = append(app.blockExeInfo.effectG, &types.EffectGroup{
		Action:  action,
		Effects: effects,
	})

	app.tempHeader.Feepool = app.tempHeader.Feepool.Add(app.tempHeader.Feepool, txData.FeePaid)
	app.tempHeader.TotalCoin = app.tempHeader.TotalCoin.Sub(app.tempHeader.TotalCoin, txData.FeePaid)

	return
}

func (app *GenesisApp) OnNewRound(height, round int, block *at.Block) (interface{}, error) {
	return at.NewRoundResult{}, nil
}

func (app *GenesisApp) OnExecute(height, round int, block *at.Block) (interface{}, error) {

	var (
		res at.ExecuteResult
		err error
	)

	app.EvmCurrentHeader = app.makeCurrentHeader(block)

	app.stateAppMtx.Lock()
	app.tmpStateDup = newStateDup(app.stateApp, block, height, round)
	app.stateAppMtx.Unlock()

	app.makeTempHeader(block)

	app.tmpStateDup.stateMtx.Lock()

	for _, tx := range block.Data.Txs {
		if err := app.ExecuteTx(app.tmpStateDup, tx); err != nil {
			res.InvalidTxs = append(res.InvalidTxs, at.ExecuteInvalidTx{Bytes: tx, Error: err})
		} else {
			res.ValidTxs = append(res.ValidTxs, tx)
			app.tempHeader.TxCount++
		}
	}
	app.tmpStateDup.stateMtx.Unlock()

	return res, err
}

// OnCommit run in a sync way, we don't need to lock stateDupMtx, but stateAppMtx is still needed
func (app *GenesisApp) OnCommit(height, round int, block *at.Block) (interface{}, error) {

	var (
		stateRoot ethcmn.Hash
		err       error
	)

	// commit levelDB
	app.tmpStateDup.stateMtx.Lock()
	stateRoot, err = app.tmpStateDup.state.Commit(StateRemoveEmptyObj)
	app.tmpStateDup.stateMtx.Unlock()
	if err != nil {
		app.SaveLastBlock(app.currentHeader.Hash(), app.currentHeader)
		return nil, err
	}

	receiptHash := app.SaveReceipts(app.tmpStateDup)

	app.currentHeader = app.tempHeader
	app.currentHeader.StateRoot = stateRoot

	appHash := app.currentHeader.Hash()
	app.SaveLastBlock(appHash, app.currentHeader)

	err = app.SaveDBData()
	if err != nil {
		logger.Error("Save db data failed:" + err.Error())
		panic(err)
	}

	app.blockExeInfo = &blockExeInfo{}

	app.stateAppMtx.Lock()
	app.stateApp, err = app.tmpStateDup.state.New(stateRoot)
	app.stateAppMtx.Unlock()

	app.currentHeader.PrevHash = ethcmn.BytesToLedgerHash(appHash)

	return at.CommitResult{
		AppHash:      appHash,
		ReceiptsHash: receiptHash,
	}, nil
}

func (app *GenesisApp) SaveReceipts(stdup *stateDup) []byte {

	savedReceipts := make([][]byte, 0, len(stdup.receipts))

	receiptBatch := app.chainDb.NewBatch()

	for _, receipt := range stdup.receipts {

		storageReceipt := (*types.Receipt)(receipt)

		storageReceiptBytes, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			logger.Error("wrong rlp encode" + err.Error())
			continue
		}

		key := append(ReceiptsPrefix, receipt.TxHash.Bytes()...)

		if err := receiptBatch.Put(key, storageReceiptBytes); err != nil {
			logger.Error("batch receipt failed" + err.Error())
			continue
		}
		savedReceipts = append(savedReceipts, storageReceiptBytes)
	}
	if err := receiptBatch.Write(); err != nil {
		logger.Error("persist receipts failed" + err.Error())
	}
	return merkle.SimpleHashFromHashes(savedReceipts)
}

// SaveDBData save data into sql-db
func (app *GenesisApp) SaveDBData() error {
	// begin dbtx
	err := app.dataM.QTxBegin()
	if err != nil {
		return err
	}

	// Save ledgerheader
	ledgerHeader := app.currentHeader.GetLedgerHeaderData()
	_, err = app.dataM.AddLedgerHeaderData(ledgerHeader)
	if err != nil {
		app.dataM.QTxRollback()
		return err
	}
	stmt, err := app.dataM.PrepareTransaction()
	if err != nil {
		app.dataM.QTxRollback()
		return err
	}
	for _, v := range app.blockExeInfo.txDatas {
		v.LedgerHash = ethcmn.BytesToLedgerHash(app.currentHeader.Hash())
		v.Height = app.currentHeader.Height
		err = app.dataM.AddTransactionStmt(stmt, v)
		if err != nil {
			app.dataM.QTxRollback()
			return err
		}
	}
	stmt.Close()

	//save action
	stmt, err = app.dataM.PrepareAction()
	if err != nil {
		app.dataM.QTxRollback()
		return err
	}
	for _, a := range app.blockExeInfo.effectG {
		a.Action.GetActionBase().Height = app.currentHeader.Height
		err = app.dataM.AddActionDataStmt(stmt, a.Action)
		if err != nil {
			app.dataM.QTxRollback()
			return err
		}
	}
	stmt.Close()

	//save effect
	stmt, err = app.dataM.PrepareEffect()
	if err != nil {
		app.dataM.QTxRollback()
		return err
	}
	for _, a := range app.blockExeInfo.effectG {
		for _, e := range a.Effects {
			e.GetEffectBase().Height = app.currentHeader.Height
			e.GetEffectBase().ActionID = a.ActionID
			err = app.dataM.AddEffectDataStmt(stmt, e)
			if err != nil {
				app.dataM.QTxRollback()
				return err
			}
		}
	}
	stmt.Close()
	// commit dbtx
	err = app.dataM.QTxCommit()
	if err != nil {
		return err
	}

	return nil
}

func (app *GenesisApp) LoadLastBlock() (lastBlock LastBlockInfo) {
	buf, _ := app.chainDb.Get(lastBlockKey)
	if len(buf) != 0 {
		r, n, err := bytes.NewReader(buf), new(int), new(error)
		wire.ReadBinaryPtr(&lastBlock, r, 0, n, err)
		if *err != nil {
			logger.Warn("lastblockinfo has been corrupted")
		}
	} else {
		lastBlock.TotalCoin = "0"
		lastBlock.Feepool = "0"
	}

	return lastBlock
}

func (app *GenesisApp) SaveLastBlock(appHash []byte, header *types.AppHeader) {
	lastBlock := LastBlockInfo{
		Height:    uint64(header.Height.Int64()),
		StateRoot: header.StateRoot.Bytes(),
		AppHash:   appHash,
		TotalCoin: header.TotalCoin.String(),
		Feepool:   header.Feepool.String(),
	}

	buf, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(lastBlock, buf, n, err)
	if *err != nil {
		cmn.PanicCrisis(*err)
	}
	app.chainDb.Put(lastBlockKey, buf.Bytes())
}

func (app *GenesisApp) CheckTx(bs []byte) at.Result {

	var err error

	tx := &types.Transaction{}

	err = rlp.DecodeBytes(bs, &tx.Data)

	if err != nil {
		return at.NewError(at.CodeType_WrongRLP, err.Error())
	}

	tx.SetCreateTime(uint64(time.Now().UnixNano()))

	app.stateAppMtx.Lock()

	if !app.stateApp.Exist(tx.GetFrom()) {
		app.stateAppMtx.Unlock()
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}
	// Cost checking
	if !app.checkEnoughFee(tx.GetFrom(), tx) {
		app.stateAppMtx.Unlock()
		return at.NewError(at.CodeType_BaseInsufficientFunds, at.CodeType_BaseInsufficientFunds.String())
	}

	app.stateAppMtx.Unlock()

	// check base fee
	if tx.BaseFee() == nil || tx.BaseFee().Cmp(app.currentHeader.BaseFee) < 0 {
		return at.NewError(at.CodeType_BaseInsufficientFunds, at.CodeType_BaseInsufficientFunds.String())
	}

	if ret := app.CheckSignTx(tx); ret.IsErr() {
		return ret
	}

	app.txCache.Set(string(bs), tx)

	return at.NewResultOK(nil, "")
}

func (app *GenesisApp) checkEnoughFee(from ethcmn.Address, tx *types.Transaction) bool {
	rest := new(big.Int).Sub(app.stateApp.GetBalance(from), tx.BaseFee())
	if rest.Cmp(big.NewInt(0)) < 0 {
		return false
	}
	return true
}

// query Info
func (app *GenesisApp) Info() (resInfo at.ResultInfo) {
	lb := app.LoadLastBlock()
	resInfo.LastBlockAppHash = lb.AppHash
	resInfo.LastBlockHeight = lb.Height
	resInfo.Version = "alpha 0.1"
	resInfo.Data = "default app with evm-1.5.9"
	return
}

// query account's nonce
func (app *GenesisApp) QueryNonce(address string) at.NewRPCResult {

	account := ethcmn.HexToAddress(address)

	app.stateAppMtx.Lock()
	nonce := app.stateApp.GetNonce(account)
	app.stateAppMtx.Unlock()

	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, nonce)

	return at.NewRpcResultOK(b, "")
}

// query accout info
func (app *GenesisApp) QueryAccount(address string) at.NewRPCResult {

	if !ethcmn.IsHexAddress(address) {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}
	if strings.Index(address, "0x") == 0 {
		address = address[2:]
	}

	account := ethcmn.HexToAddress(address)

	app.stateAppMtx.Lock()

	accountSO := app.stateApp.GetStateObject(account)

	app.stateAppMtx.Unlock()

	if xlib.CheckItfcNil(accountSO) {
		return at.NewRpcError(at.CodeType_BaseUnknownAddress, "Unknown address")
	}
	var show types.ShowAccount
	accountSO.FillShow(&show)
	// Default paging query 200, order = desc
	datas, err := app.dataM.QueryAccData(account, "desc")
	if err != nil {
		logger.Warn("[query account],load accdata err:", zap.String("err", err.Error()))
		return at.NewRpcError(at.CodeType_InternalError, fmt.Sprintf("get accdata fail:%v", err))
	}
	show.Data = datas

	return at.NewRpcResultOK(show, "")
}

// query all ledger's info
func (app *GenesisApp) QueryLedgers(order string, limit uint64, cursor uint64) at.NewRPCResult {
	return app.queryAllLedgers(cursor, limit, order)
}

// query ledger info
func (app *GenesisApp) QueryLedger(height uint64) at.NewRPCResult {
	sequence := new(big.Int).SetUint64(height)
	return app.queryLedger(sequence)
}

// query all payments
func (app *GenesisApp) QueryPayments(order string, limit uint64, cursor uint64) at.NewRPCResult {
	var query types.ActionsQuery
	query.Order = order
	query.Limit = limit
	query.Cursor = cursor

	query.Typei = uint64(types.OP_S_PAYMENT.OpInt())

	return app.queryPaymentsData(query)
}

// query account's payments
func (app *GenesisApp) QueryAccountPayments(address string, order string, limit uint64, cursor uint64) at.NewRPCResult {
	if !ethcmn.IsHexAddress(address) {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}
	if strings.Index(address, "0x") == 0 {
		address = address[2:]
	}
	account := ethcmn.HexToAddress(address)

	var query types.ActionsQuery
	query.Order = order
	query.Limit = limit
	query.Cursor = cursor
	query.Account = account

	query.Typei = uint64(types.OP_S_PAYMENT.OpInt())

	return app.queryPaymentsData(query)
}

// query payment with txhash
func (app *GenesisApp) QueryPayment(txhash string) at.NewRPCResult {
	var query types.ActionsQuery

	if txhash == "" {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid txhash")
	}

	hash := ethcmn.HexToHash(txhash)

	if len(hash) != ethcmn.HashLength {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid txhash")
	}

	query.TxHash = hash
	query.Typei = uint64(types.OP_S_PAYMENT.OpInt())

	return app.queryPaymentsData(query)
}

// query all transactions
func (app *GenesisApp) QueryTransactions(order string, limit uint64, cursor uint64) at.NewRPCResult {
	//	var query types.ActionsQuery
	return app.queryAllTxs(cursor, limit, order)
}

// query transaction with txhash
func (app *GenesisApp) QueryTransaction(txhash string) at.NewRPCResult {
	if txhash == "" {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid txhash")
	}

	hash := ethcmn.HexToHash(txhash)
	if hash == types.ZERO_HASH || len(hash) != ethcmn.HashLength {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid txhash")
	}
	var query types.ActionsQuery
	query.TxHash = hash
	query.Begin = 0
	query.End = 0
	query.Typei = types.TypeiUndefined

	return app.queryActionsData(query)
}

// query account's transactions
func (app *GenesisApp) QueryAccountTransactions(address string, order string, limit uint64, cursor uint64) at.NewRPCResult {
	if !ethcmn.IsHexAddress(address) {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}
	if strings.Index(address, "0x") == 0 {
		address = address[2:]
	}
	account := ethcmn.HexToAddress(address)

	if account == types.ZERO_ADDRESS {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}

	return app.queryAccountTxs(account, cursor, limit, order)
}

// query specific ledger's transactions
func (app *GenesisApp) QueryLedgerTransactions(height uint64, order string, limit uint64, cursor uint64) at.NewRPCResult {
	heightStr := strconv.FormatUint(height, 10)
	return app.queryHeightTxs(heightStr, cursor, limit, order)
}

// query contract
func (app *GenesisApp) QueryDoContract(query []byte) at.NewRPCResult {
	return app.queryDoContract(query)
}

// query contract is exist
func (app *GenesisApp) QueryContractExist(address string) at.NewRPCResult {
	var c *types.QueryContractExist

	if !ethcmn.IsHexAddress(address) {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}
	if strings.Index(address, "0x") == 0 {
		address = address[2:]
	}
	contractAccount := ethcmn.HexToAddress(address)

	app.stateAppMtx.Lock()
	hashBytes := app.stateApp.GetCodeHash(contractAccount)
	codeBytes := app.stateApp.GetByteCode(contractAccount)
	app.stateAppMtx.Unlock()

	if len(hashBytes) != ethcmn.HashLength || ethcmn.EmptyHash(hashBytes) {
		c = &types.QueryContractExist{
			IsExist: false,
		}
	} else {
		c = &types.QueryContractExist{
			IsExist:  true,
			CodeHash: hashBytes.Hex(),
			ByteCode: hex.EncodeToString(codeBytes),
		}
	}

	return at.NewRpcResultOK(c, "")
}

// query contract receipt with txhash
func (app *GenesisApp) QueryReceipt(txhash string) at.NewRPCResult {
	hash := ethcmn.HexToHash(txhash)
	if len(hash) != ethcmn.HashLength {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid txhash")
	}
	key := append(ReceiptsPrefix, hash.Bytes()...)

	app.stateAppMtx.Lock()
	queryData, err := app.chainDb.Get(key)
	app.stateAppMtx.Unlock()

	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, "fail to get receipt for tx:"+txhash)
	}

	var receipt types.Receipt
	if err := rlp.DecodeBytes(queryData, &receipt); err != nil {
		return at.NewRpcError(at.CodeType_WrongRLP, "fail to rlp decode")
	}

	return at.NewRpcResultOK(receipt, "")
}

// query account's all managedata
func (app *GenesisApp) QueryAccountManagedatas(address string, order string, limit uint64, cursor uint64) at.NewRPCResult {
	if !ethcmn.IsHexAddress(address) {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}
	if strings.Index(address, "0x") == 0 {
		address = address[2:]
	}
	account := ethcmn.HexToAddress(address)

	if account == types.ZERO_ADDRESS {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}

	return app.queryAccountManagedata(account, "", "", cursor, limit, order)
}

// query account's managedata for key
func (app *GenesisApp) QueryAccountManagedata(address string, key string) at.NewRPCResult {
	if !ethcmn.IsHexAddress(address) {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}
	if strings.Index(address, "0x") == 0 {
		address = address[2:]
	}
	account := ethcmn.HexToAddress(address)

	if account == types.ZERO_ADDRESS {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}
	return app.queryAccountSingleManageData(account, key)
}

func (app *GenesisApp) QueryAccountCategoryManagedata(address string, category string) at.NewRPCResult {
	if !ethcmn.IsHexAddress(address) {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}
	if strings.Index(address, "0x") == 0 {
		address = address[2:]
	}
	account := ethcmn.HexToAddress(address)

	if account == types.ZERO_ADDRESS {
		return at.NewRpcError(at.CodeType_BaseInvalidInput, "Invalid address")
	}
	return app.queryAccountCategoryManageData(account, category)
}

// ParseBaseFee get base fee
func (app *GenesisApp) ParseBaseFee(block *at.Block) *big.Int {
	baseFee := app.config.GetInt("base_fee")

	return new(big.Int).SetInt64(int64(baseFee))
}

// ParseBaseReserve get base reserve
func (app *GenesisApp) ParseBaseReserve(block *at.Block) *big.Int {
	baseReserve := app.config.GetInt("base_reserve")

	return new(big.Int).SetInt64(int64(baseReserve))
}

// ParseMaxTxSetSize get base max tx set size
func (app *GenesisApp) ParseMaxTxSetSize(block *at.Block) uint64 {
	maxTxSetSize := app.config.GetInt("max_txset_size")

	return uint64(maxTxSetSize)
}

func (app *GenesisApp) makeCurrentHeader(block *at.Block) *ethtypes.Header {
	return &ethtypes.Header{
		ParentHash: ethcmn.HexToHash("0x00"),
		Difficulty: big.NewInt(0),
		GasLimit:   ethcmn.MaxBig,
		//		Number:     ethparams.MainNetSpuriousDragon,
		Time:   big.NewInt(block.Header.Time.Unix()),
		Number: big.NewInt(int64(block.Height)),
	}
}
