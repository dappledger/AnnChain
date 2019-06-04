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

package evm

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	rtypes "github.com/dappledger/AnnChain/chain/types"
	ethcmn "github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/common/math"
	ethcore "github.com/dappledger/AnnChain/eth/core"
	"github.com/dappledger/AnnChain/eth/core/rawdb"
	ethstate "github.com/dappledger/AnnChain/eth/core/state"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	ethvm "github.com/dappledger/AnnChain/eth/core/vm"
	"github.com/dappledger/AnnChain/eth/ethdb"
	ethparams "github.com/dappledger/AnnChain/eth/params"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/modules/go-merkle"
	atypes "github.com/dappledger/AnnChain/gemmill/types"
)

const (
	OfficialAddress     = "0x7752b42608a0f1943c19fc5802cb027e60b4c911"
	StateRemoveEmptyObj = true
	DatabaseCache       = 128
	DatabaseHandles     = 1024
	APP_NAME            = "evm"
)

//reference ethereum BlockChain
type BlockChainEvm struct {
	db ethdb.Database
}

type Hashs []ethcmn.Hash
type BeginExecFunc func() (ExecFunc, EndExecFunc)
type ExecFunc func(index int, raw []byte, tx *ethtypes.Transaction) error
type EndExecFunc func(bs []byte, err error) bool

func NewBlockChain(db ethdb.Database) *BlockChainEvm {
	return &BlockChainEvm{db}
}
func (bc *BlockChainEvm) GetHeader(hash ethcmn.Hash, number uint64) *ethtypes.Header {

	//todo cache,reference core/headerchain.go
	header := rawdb.ReadHeader(bc.db, hash, number)
	if header == nil {
		return nil
	}
	return header
}

var (
	ReceiptsPrefix  = []byte("receipts-")
	BlockHashPrefix = []byte("blockhash-")
)

type LastBlockInfo struct {
	Height  int64
	AppHash []byte
}

type EVMApp struct {
	atypes.BaseApplication
	AngineHooks atypes.Hooks

	core atypes.Core

	datadir string
	Config  *viper.Viper

	currentHeader *ethtypes.Header
	chainConfig   *ethparams.ChainConfig

	stateDb      ethdb.Database
	stateMtx     sync.Mutex
	state        *ethstate.StateDB
	currentState *ethstate.StateDB

	receipts    ethtypes.Receipts
	valid_hashs Hashs

	Signer ethtypes.Signer
}

const (
	// With 2.2 GHz Intel Core i7, 16 GB 2400 MHz DDR4, 256GB SSD, we tested following contract, it takes about 24157 gas and 171.193µs.
	// function setVal(uint256 _val) public {
	//	val = _val;
	//	emit SetVal(_val,_val);
	//  emit SetValByWho("a name which length is bigger than 32 bytes",msg.sender, _val);
	// }
	// So we estimate that running out of 100000000 gas may be taken at least 1s to 10s
	EVMGasLimit uint64 = 100000000
)

var (
	EmptyTrieRoot = ethcmn.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	IsHomestead   = true

	lastBlockKey = []byte("lastblock")
	evmConfig    = ethvm.Config{EVMGasLimit: EVMGasLimit}

	errQuitExecute = fmt.Errorf("quit executing block")
)

func makeCurrentHeader(block *atypes.Block, header *atypes.Header) *ethtypes.Header {
	return &ethtypes.Header{
		ParentHash: ethcmn.BytesToHash(block.Header.LastBlockID.Hash),
		Difficulty: big.NewInt(0),
		GasLimit:   math.MaxBig256.Uint64(),
		Time:       big.NewInt(block.Header.Time.Unix()),
		Number:     big.NewInt(header.Height),
	}
}

func stateKey(block *atypes.Block, height, round int64) string {
	return ethcmn.Bytes2Hex(block.Hash())
}

func OpenDatabase(datadir string, name string, cache int, handles int) (ethdb.Database, error) {
	return ethdb.NewLDBDatabase(filepath.Join(datadir, name), cache, handles)
}

func NewEVMApp(config *viper.Viper) (*EVMApp, error) {
	app := &EVMApp{
		datadir:     config.GetString("db_dir"),
		chainConfig: ethparams.MainnetChainConfig,
		Config:      config,
	}

	app.AngineHooks = atypes.Hooks{
		OnNewRound: atypes.NewHook(app.OnNewRound),
		OnCommit:   atypes.NewHook(app.OnCommit),
		OnPrevote:  atypes.NewHook(app.OnPrevote),
		OnExecute:  atypes.NewHook(app.OnExecute),
	}

	app.Signer = new(ethtypes.HomesteadSigner)
	var err error
	if err = app.BaseApplication.InitBaseApplication(APP_NAME, app.datadir); err != nil {
		log.Error("InitBaseApplication error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}

	if app.stateDb, err = OpenDatabase(app.datadir, "chaindata", DatabaseCache, DatabaseHandles); err != nil {
		log.Error("OpenDatabase error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}

	return app, nil
}

func (app *EVMApp) writeGenesis() error {

	switch ethcore.GetEvmLimitType() {
	case ethcore.EvmLimitTypeTx, ethcore.EvmLimitTypeBalance:
	default:
		return nil
	}

	if app.getLastAppHash() != EmptyTrieRoot {
		return nil
	}

	g := ethcore.DefaultGenesis()
	b := g.ToBlock(app.stateDb)
	app.SaveLastBlock(LastBlockInfo{Height: 0, AppHash: b.Root().Bytes()})
	return nil
}

func (app *EVMApp) Start() (err error) {

	if err := app.writeGenesis(); err != nil {
		app.Stop()
		log.Error("write genesis err:", zap.Error(err))
		return err
	}

	lastBlock := &LastBlockInfo{
		Height:  0,
		AppHash: make([]byte, 0),
	}
	if res, err := app.LoadLastBlock(lastBlock); err == nil && res != nil {
		lastBlock = res.(*LastBlockInfo)
	}
	if err != nil {
		log.Error("fail to load last block", zap.Error(err))
		return
	}

	// Load evm state when starting
	trieRoot := EmptyTrieRoot
	if len(lastBlock.AppHash) > 0 {
		trieRoot = ethcmn.BytesToHash(lastBlock.AppHash)
	}

	if app.state, err = ethstate.New(trieRoot, ethstate.NewDatabase(app.stateDb)); err != nil {
		app.Stop()
		log.Error("fail to new ethstate", zap.Error(err))
		return
	}

	return nil
}

func (app *EVMApp) getLastAppHash() ethcmn.Hash {
	lastBlock := &LastBlockInfo{
		Height:  0,
		AppHash: make([]byte, 0),
	}
	if res, err := app.LoadLastBlock(lastBlock); err == nil && res != nil {
		lastBlock = res.(*LastBlockInfo)
	}
	if len(lastBlock.AppHash) > 0 {
		return ethcmn.BytesToHash(lastBlock.AppHash)
	}
	return EmptyTrieRoot
}

func (app *EVMApp) Stop() {
	app.BaseApplication.Stop()
	app.stateDb.Close()
}

func (app *EVMApp) GetAngineHooks() atypes.Hooks {
	return app.AngineHooks
}

func (app *EVMApp) CompatibleWithAngine() {}

func (app *EVMApp) BeginExecute() {
}

func (app *EVMApp) OnNewRound(height, round int64, block *atypes.Block) (interface{}, error) {
	return atypes.NewRoundResult{}, nil
}

func (app *EVMApp) OnPrevote(height, round int64, block *atypes.Block) (interface{}, error) {
	return nil, nil
}

func (app *EVMApp) genExecFun(block *atypes.Block, res *atypes.ExecuteResult) BeginExecFunc {

	blockHash := ethcmn.BytesToHash(block.Hash())
	app.currentHeader = makeCurrentHeader(block, block.Header)
	return func() (ExecFunc, EndExecFunc) {
		state := app.currentState
		stateSnapshot := state.Snapshot()
		temReceipt := make([]*ethtypes.Receipt, 0)
		temTxHash := make([]ethcmn.Hash, 0)

		execFunc := func(txIndex int, raw []byte, tx *ethtypes.Transaction) error {
			gp := new(ethcore.GasPool).AddGas(math.MaxBig256.Uint64())

			txBytes, err := rlp.EncodeToBytes(tx)
			if err != nil {
				return err
			}
			txhash := ethcmn.BytesToHash(atypes.Tx(txBytes).Hash())
			state.Prepare(txhash, blockHash, txIndex)

			bc := NewBlockChain(app.stateDb)
			receipt, _, err := ethcore.ApplyTransaction(
				app.chainConfig,
				bc,
				nil, // coinbase ,maybe use local account
				gp,
				state,
				app.currentHeader,
				tx,
				new(uint64),
				evmConfig)

			if err != nil {
				return err
			}

			if tx.To() != nil {
				receipt.To = *(tx.To())
			}
			receipt.Height = app.currentHeader.Number.Uint64()
			receipt.Timestamp = new(big.Int).SetInt64(time.Now().Unix())
			temReceipt = append(temReceipt, receipt)
			temTxHash = append(temTxHash, txhash)

			return nil
		}

		endFunc := func(raw []byte, err error) bool {
			if err != nil {
				log.Warn("[evm execute],apply transaction", zap.Error(err))
				state.RevertToSnapshot(stateSnapshot)
				temReceipt = nil
				temTxHash = nil
				res.InvalidTxs = append(res.InvalidTxs, atypes.ExecuteInvalidTx{Bytes: raw, Error: err})
				return true
			}
			app.receipts = append(app.receipts, temReceipt...)
			app.valid_hashs = append(app.valid_hashs, temTxHash...)
			res.ValidTxs = append(res.ValidTxs, raw)

			return true
		}
		return execFunc, endFunc
	}
}

func (app *EVMApp) OnExecute(height, round int64, block *atypes.Block) (interface{}, error) {
	var (
		res atypes.ExecuteResult
		err error
	)

	if app.currentState, err = ethstate.New(app.getLastAppHash(), ethstate.NewDatabase(app.stateDb)); err != nil {
		return nil, errors.Wrap(err, "create StateDB failed")
	}
	exeWithCPUSerialVeirfy(nil, block.Data.Txs, app.genExecFun(block, &res))

	return res, err
}

// OnCommit run in a sync way, we don't need to lock stateDupMtx, but stateMtx is still needed
func (app *EVMApp) OnCommit(height, round int64, block *atypes.Block) (interface{}, error) {
	appHash, err := app.currentState.Commit(StateRemoveEmptyObj)
	if err != nil {
		return nil, err
	}

	if err := app.currentState.Database().TrieDB().Commit(appHash, false); err != nil {
		return nil, err
	}

	app.stateMtx.Lock()
	if app.state, err = ethstate.New(appHash, ethstate.NewDatabase(app.stateDb)); err != nil {
		app.stateMtx.Unlock()
		return nil, errors.Wrap(err, "create StateDB failed")
	}
	app.stateMtx.Unlock()

	app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: appHash.Bytes()})
	rHash := app.SaveReceipts()
	bHash := app.SaveBlocks(block.Hash())
	app.receipts = nil
	app.valid_hashs = nil

	log.Info("application save to db", zap.String("appHash", fmt.Sprintf("%X", appHash.Bytes())), zap.String("receiptHash", fmt.Sprintf("%X", rHash)))

	return atypes.CommitResult{
		AppHash:      appHash.Bytes(),
		ReceiptsHash: rHash,
		BlockHash:    bHash,
	}, nil
}

func (app *EVMApp) CheckTx(bs []byte) error {
	return atypes.Tx(bs).Deal(func(txbs atypes.Tx) error {
		tx := &ethtypes.Transaction{}
		err := rlp.DecodeBytes(txbs, tx)
		if err != nil {
			return err
		}
		from, err := ethtypes.Sender(app.Signer, tx)
		if err != nil {
			return err
		}

		app.stateMtx.Lock()
		defer app.stateMtx.Unlock()
		// Last but not least check for nonce errors
		nonce := tx.Nonce()
		getNonce := app.state.GetNonce(from)
		if getNonce > nonce {
			txhash := atypes.Tx(bs).Hash()
			return fmt.Errorf("nonce(%d) different with getNonce(%d), transaction already exists %v", nonce, getNonce, hex.EncodeToString(txhash))
		}
		// Transactor should have enough funds to cover the costs
		// cost == V + GP * GL
		if app.state.GetBalance(from).Cmp(tx.Cost()) < 0 {
			return fmt.Errorf("not enough funds")
		}
		return nil
	})
}

func (app *EVMApp) SaveBlocks(blockHash []byte) []byte {

	blockBatch := app.stateDb.NewBatch()

	storageBlockBytes, err := rlp.EncodeToBytes(app.valid_hashs)
	if err != nil {
		fmt.Println("wrong rlp encode:" + err.Error())
		return nil
	}

	key := append(BlockHashPrefix, blockHash...)

	if err := blockBatch.Put(key, storageBlockBytes); err != nil {
		fmt.Println("batch block failed:" + err.Error())
		return nil
	}

	if err := blockBatch.Write(); err != nil {
		fmt.Println("persist block failed:" + err.Error())
		return nil
	}

	bHash := merkle.SimpleHashFromBinaries([]interface{}{app.valid_hashs})

	return bHash
}

func (app *EVMApp) SaveReceipts() []byte {
	savedReceipts := make([][]byte, 0, len(app.receipts))
	receiptBatch := app.stateDb.NewBatch()

	for _, receipt := range app.receipts {
		storageReceipt := (*ethtypes.ReceiptForStorage)(receipt)
		storageReceiptBytes, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			fmt.Println("wrong rlp encode:" + err.Error())
			return nil
		}

		key := append(ReceiptsPrefix, receipt.TxHash.Bytes()...)

		if err := receiptBatch.Put(key, storageReceiptBytes); err != nil {
			fmt.Println("batch receipt failed:" + err.Error())
			return nil
		}
		savedReceipts = append(savedReceipts, storageReceiptBytes)
	}
	if err := receiptBatch.Write(); err != nil {
		fmt.Println("persist receipts failed:" + err.Error())
		return nil
	}

	rHash := merkle.SimpleHashFromHashes(savedReceipts)

	return rHash
}

func (app *EVMApp) Info() (resInfo atypes.ResultInfo) {
	lb := &LastBlockInfo{
		AppHash: make([]byte, 0),
		Height:  0,
	}
	if res, err := app.LoadLastBlock(lb); err == nil {
		lb = res.(*LastBlockInfo)
	}

	resInfo.LastBlockAppHash = lb.AppHash
	resInfo.LastBlockHeight = lb.Height
	resInfo.Version = "0.7.0"
	resInfo.Data = "default app with evm-1.8.21"
	return
}

func (app *EVMApp) Query(query []byte) atypes.Result {
	var res atypes.Result
	action := query[0]
	load := query[1:]
	switch action {
	case rtypes.QueryType_Contract:
		res = app.queryContract(load)
	case rtypes.QueryType_Nonce:
		res = app.queryNonce(load)
	case rtypes.QueryType_Balance:
		res = app.queryBalance(load)
	case rtypes.QueryType_Receipt:
		res = app.queryReceipt(load)
	case rtypes.QueryType_Existence:
		res = app.queryContractExistence(load)
	case rtypes.QueryType_PayLoad:
		res = app.queryPayLoad(load)
	case rtypes.QueryType_BlockHash:
		res = app.queryBlockHash(load)
	default:
		res = atypes.NewError(atypes.CodeType_BaseInvalidInput, "unimplemented query")
	}

	return res
}

func (app *EVMApp) queryContractExistence(load []byte) atypes.Result {
	tx := new(ethtypes.Transaction)
	err := rlp.DecodeBytes(load, tx)
	if err != nil {
		return atypes.NewError(atypes.CodeType_BaseInvalidInput, err.Error())
	}
	contractAddr := tx.To()

	app.stateMtx.Lock()
	hashBytes := app.state.GetCodeHash(*contractAddr).Bytes()
	app.stateMtx.Unlock()

	if bytes.Equal(tx.Data(), hashBytes) {
		return atypes.NewResultOK(append([]byte{}, byte(0x01)), "contract exists")
	}
	return atypes.NewResultOK(append([]byte{}, byte(0x00)), "constract doesn't exist")
}

func (app *EVMApp) queryContract(load []byte) atypes.Result {
	tx := new(ethtypes.Transaction)
	err := rlp.DecodeBytes(load, tx)
	if err != nil {
		return atypes.NewError(atypes.CodeType_BaseInvalidInput, err.Error())
	}

	txMsg, _ := tx.AsMessage(app.Signer)

	bc := NewBlockChain(app.stateDb)
	envCxt := ethcore.NewEVMContext(txMsg, app.currentHeader, bc, nil)

	app.stateMtx.Lock()
	vmEnv := ethvm.NewEVM(envCxt, app.state.Copy(), app.chainConfig, evmConfig)
	gpl := new(ethcore.GasPool).AddGas(math.MaxBig256.Uint64())
	res, _, _, err := ethcore.ApplyMessage(vmEnv, txMsg, gpl) // we don't care about gasUsed
	if err != nil {
		log.Warn("query apply msg err", zap.Error(err))
	}
	app.stateMtx.Unlock()

	return atypes.NewResultOK(res, "")
}

func (app *EVMApp) queryNonce(addrBytes []byte) atypes.Result {
	if len(addrBytes) != 20 {
		return atypes.NewError(atypes.CodeType_BaseInvalidInput, "Invalid address")
	}
	addr := ethcmn.BytesToAddress(addrBytes)

	app.stateMtx.Lock()
	nonce := app.state.GetNonce(addr)
	app.stateMtx.Unlock()

	data, err := rlp.EncodeToBytes(nonce)
	if err != nil {
		log.Warn("query error", zap.Error(err))
	}
	return atypes.NewResultOK(data, "")
}

func (app *EVMApp) queryBalance(addrBytes []byte) atypes.Result {
	if len(addrBytes) != 20 {
		return atypes.NewError(atypes.CodeType_BaseInvalidInput, "Invalid address")
	}
	addr := ethcmn.BytesToAddress(addrBytes)

	app.stateMtx.Lock()
	balance := app.state.GetBalance(addr)
	app.stateMtx.Unlock()

	data, err := rlp.EncodeToBytes(balance)
	if err != nil {
		log.Warn("query error", zap.Error(err))
	}
	return atypes.NewResultOK(data, "")
}

func (app *EVMApp) queryReceipt(txHashBytes []byte) atypes.Result {
	key := append(ReceiptsPrefix, txHashBytes...)
	data, err := app.stateDb.Get(key)
	if err != nil {
		return atypes.NewError(atypes.CodeType_InternalError, "fail to get receipt for tx:"+string(key))
	}
	return atypes.NewResultOK(data, "")
}

func (app *EVMApp) queryBlockHash(blockHashBytes []byte) atypes.Result {
	key := append(BlockHashPrefix, blockHashBytes...)
	data, err := app.stateDb.Get(key)
	if err != nil {
		return atypes.NewError(atypes.CodeType_InternalError, "fail to get txs for blockhash:"+string(key))
	}
	return atypes.NewResultOK(data, "")
}

func (app *EVMApp) queryPayLoad(txHashBytes []byte) atypes.Result {
	if len(txHashBytes) == 0 {
		return atypes.NewError(atypes.CodeType_BaseInvalidInput, "Empty query")
	}

	var res atypes.Result
	data, err := app.core.Query(txHashBytes[0], txHashBytes[1:])
	if err != nil {
		return atypes.NewError(atypes.CodeType_InternalError, err.Error())
	}

	if value, ok := data.([]byte); ok {
		res.Data = value
	}

	res.Code = atypes.CodeType_OK
	return res
}

func (app *EVMApp) SetCore(core atypes.Core) {
	app.core = core
}
