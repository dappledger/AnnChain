// Copyright © 2017 ZhongAn Technology
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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	rtypes "github.com/dappledger/AnnChain/chain/types"
	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/common/math"
	"github.com/dappledger/AnnChain/eth/core"
	"github.com/dappledger/AnnChain/eth/core/rawdb"
	estate "github.com/dappledger/AnnChain/eth/core/state"
	etypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/core/vm"
	"github.com/dappledger/AnnChain/eth/ethdb"
	"github.com/dappledger/AnnChain/eth/params"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/modules/go-merkle"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
)

const (
	AppName         = "evm"
	DatabaseCache   = 128
	DatabaseHandles = 1024

	// With 2.2 GHz Intel Core i7, 16 GB 2400 MHz DDR4, 256GB SSD, we tested following contract, it takes about 24157 gas and 171.193µs.
	// function setVal(uint256 _val) public {
	//	val = _val;
	//	emit SetVal(_val,_val);
	//  emit SetValByWho("a name which length is bigger than 32 bytes",msg.sender, _val);
	// }
	// So we estimate that running out of 100000000 gas may be taken at least 1s to 10s
	EVMGasLimit uint64 = 100000000
)

//reference ethereum BlockChain
type BlockChainEvm struct {
	db ethdb.Database
}

func NewBlockChain(db ethdb.Database) *BlockChainEvm {
	return &BlockChainEvm{db}
}

func (bc *BlockChainEvm) GetHeader(hash common.Hash, number uint64) *etypes.Header {
	//todo cache,reference core/headerchain.go
	header := rawdb.ReadHeader(bc.db, hash, number)
	if header == nil {
		return nil
	}
	return header
}

var (
	ReceiptsPrefix = []byte("receipts-")
	KvPrefix       = []byte("kvstore-")

	EmptyTrieRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	IsHomestead = true
	evmConfig   = vm.Config{EVMGasLimit: EVMGasLimit}

	errQuitExecute = fmt.Errorf("quit executing block")
)

type EVMApp struct {
	pool *ethTxPool
	gtypes.BaseApplication
	AngineHooks gtypes.Hooks

	core gtypes.Core

	datadir string
	Config  *viper.Viper

	currentHeader *etypes.Header
	chainConfig   *params.ChainConfig

	stateDb      ethdb.Database
	stateMtx     sync.Mutex
	state        *estate.StateDB
	currentState *estate.StateDB

	receipts etypes.Receipts
	kvs      rtypes.KVS
	Signer   etypes.Signer
}

type LastBlockInfo struct {
	Height  int64
	AppHash []byte
}

func NewEVMApp(config *viper.Viper) (*EVMApp, error) {
	app := &EVMApp{
		datadir:     config.GetString("db_dir"),
		Config:      config,
		chainConfig: params.MainnetChainConfig,
		Signer:      new(etypes.HomesteadSigner),
	}

	app.AngineHooks = gtypes.Hooks{
		OnNewRound: gtypes.NewHook(app.OnNewRound),
		OnCommit:   gtypes.NewHook(app.OnCommit),
		OnPrevote:  gtypes.NewHook(app.OnPrevote),
		OnExecute:  gtypes.NewHook(app.OnExecute),
	}

	var err error
	if err = app.BaseApplication.InitBaseApplication(AppName, app.datadir); err != nil {
		log.Error("InitBaseApplication error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}

	if app.stateDb, err = OpenDatabase(app.datadir, "chaindata", DatabaseCache, DatabaseHandles); err != nil {
		log.Error("OpenDatabase error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}

	app.pool = NewEthTxPool(app, config)

	return app, nil
}

func OpenDatabase(datadir string, name string, cache int, handles int) (ethdb.Database, error) {
	return ethdb.NewLDBDatabase(filepath.Join(datadir, name), cache, handles)
}

func (app *EVMApp) writeGenesis() error {
	if app.getLastAppHash() != EmptyTrieRoot {
		return nil
	}

	g := core.DefaultGenesis()
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
		trieRoot = common.BytesToHash(lastBlock.AppHash)
	}
	app.pool.Start(lastBlock.Height)
	if app.state, err = estate.New(trieRoot, estate.NewDatabase(app.stateDb)); err != nil {
		app.Stop()
		log.Error("fail to new state", zap.Error(err))
		return
	}

	return nil
}

func (app *EVMApp) getLastAppHash() common.Hash {
	lastBlock := &LastBlockInfo{
		Height:  0,
		AppHash: make([]byte, 0),
	}
	if res, err := app.LoadLastBlock(lastBlock); err == nil && res != nil {
		lastBlock = res.(*LastBlockInfo)
	}
	if len(lastBlock.AppHash) > 0 {
		return common.BytesToHash(lastBlock.AppHash)
	}
	return EmptyTrieRoot
}

func (app *EVMApp) GetTxPool() gtypes.TxPool {
	return app.pool
}

func (app *EVMApp) Stop() {
	app.BaseApplication.Stop()
	app.stateDb.Close()
}

func (app *EVMApp) GetAngineHooks() gtypes.Hooks {
	return app.AngineHooks
}

func (app *EVMApp) CompatibleWithAngine() {}

func (app *EVMApp) BeginExecute() {}

func (app *EVMApp) OnNewRound(height, round int64, block *gtypes.Block) (interface{}, error) {
	return gtypes.NewRoundResult{}, nil
}

func (app *EVMApp) OnPrevote(height, round int64, block *gtypes.Block) (interface{}, error) {
	return nil, nil
}

func (app *EVMApp) excuteTx(blockHash common.Hash, state *estate.StateDB, txIndex int, raw []byte, tx *etypes.Transaction) (*etypes.Receipt, error) {
	gp := new(core.GasPool).AddGas(math.MaxBig256.Uint64())
	txBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	txhash := gtypes.Tx(txBytes).Hash()
	state.Prepare(common.BytesToHash(txhash), blockHash, txIndex)

	bc := NewBlockChain(app.stateDb)
	receipt, _, err := core.ApplyTransaction(
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
		return nil, err
	}
	return receipt, nil
}

func (app *EVMApp) excuteKV(state *estate.StateDB, tx *etypes.Transaction) (*rtypes.KV, error) {
	kvData := &rtypes.KV{}
	if err := rlp.DecodeBytes(tx.Data(), kvData); err != nil {
		return nil, err
	}
	from, _ := etypes.Sender(app.Signer, tx)
	state.SetNonce(from, state.GetNonce(from)+1)
	return kvData, nil
}

func (app *EVMApp) genExecFun(block *gtypes.Block, res *gtypes.ExecuteResult) BeginExecFunc {
	blockHash := common.BytesToHash(block.Hash())
	app.currentHeader = makeCurrentHeader(block, block.Header)

	return func() (ExecFunc, EndExecFunc) {
		state := app.currentState
		stateSnapshot := state.Snapshot()
		temReceipt := make([]*etypes.Receipt, 0)
		temKv := make([]*rtypes.KV, 0)

		execFunc := func(txIndex int, raw []byte, tx *etypes.Transaction) error {
			if tx.OpCode() == rtypes.Op_KV {
				kv, err := app.excuteKV(state, tx)
				if err != nil {
					return err
				}
				temKv = append(temKv, kv)
			} else {
				receipt, err := app.excuteTx(blockHash, state, txIndex, raw, tx)
				if err != nil {
					return err
				}
				temReceipt = append(temReceipt, receipt)
			}

			return nil
		}

		endFunc := func(raw []byte, err error) bool {
			if err != nil {
				log.Warn("[evm execute],apply transaction", zap.Error(err))
				state.RevertToSnapshot(stateSnapshot)
				temReceipt = nil
				temKv = nil
				res.InvalidTxs = append(res.InvalidTxs, gtypes.ExecuteInvalidTx{Bytes: raw, Error: err})
				return true
			}
			app.receipts = append(app.receipts, temReceipt...)
			app.kvs = append(app.kvs, temKv...)
			res.ValidTxs = append(res.ValidTxs, raw)
			return true
		}
		return execFunc, endFunc
	}
}

func makeCurrentHeader(block *gtypes.Block, header *gtypes.Header) *etypes.Header {
	return &etypes.Header{
		ParentHash: common.BytesToHash(block.Header.LastBlockID.Hash),
		Difficulty: big.NewInt(0),
		GasLimit:   math.MaxBig256.Uint64(),
		Time:       big.NewInt(block.Header.Time.Unix()),
		Number:     big.NewInt(header.Height),
	}
}

func (app *EVMApp) OnExecute(height, round int64, block *gtypes.Block) (interface{}, error) {
	var (
		res gtypes.ExecuteResult
		err error
	)

	if app.currentState, err = estate.New(app.getLastAppHash(), estate.NewDatabase(app.stateDb)); err != nil {
		return nil, errors.Wrap(err, "create StateDB failed")
	}
	exeWithCPUParallelVeirfy(app.Signer, block.Data.Txs, nil, app.genExecFun(block, &res))

	m := make(map[string]int)
	for _, tx := range block.Data.Txs {
		m[string(tx)]++
	}
	dups := 0
	for _, v := range m {
		if v > 1 {
			dups++
		}
	}
	return res, err
}

// OnCommit run in a sync way, we don't need to lock stateDupMtx, but stateMtx is still needed
func (app *EVMApp) OnCommit(height, round int64, block *gtypes.Block) (interface{}, error) {
	appHash, err := app.currentState.Commit(true)
	if err != nil {
		return nil, err
	}

	if err := app.currentState.Database().TrieDB().Commit(appHash, false); err != nil {
		return nil, err
	}

	app.stateMtx.Lock()
	if app.state, err = estate.New(appHash, estate.NewDatabase(app.stateDb)); err != nil {
		app.stateMtx.Unlock()
		return nil, errors.Wrap(err, "create StateDB failed")
	}
	app.stateMtx.Unlock()

	app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: appHash.Bytes()})

	rHash, err := app.SaveReceipts()
	if err != nil {
		log.Error("application save receipts", zap.Error(err), zap.Int64("height", block.Height))
	}

	app.receipts = nil
	app.pool.updateToState()
	log.Info("application save to db", zap.String("appHash", fmt.Sprintf("%X", appHash.Bytes())), zap.String("receiptHash", fmt.Sprintf("%X", rHash)))

	return gtypes.CommitResult{
		AppHash:      appHash.Bytes(),
		ReceiptsHash: rHash,
	}, nil
}

func (app *EVMApp) CheckTx(bs []byte) error {
	tx := &etypes.Transaction{}
	err := rlp.DecodeBytes(bs, tx)
	if err != nil {
		return err
	}
	from, _ := etypes.Sender(app.Signer, tx)

	app.stateMtx.Lock()
	defer app.stateMtx.Unlock()
	// Last but not least check for nonce errors
	nonce := tx.Nonce()
	getNonce := app.state.GetNonce(from)
	if getNonce > nonce {
		txhash := gtypes.Tx(bs).Hash()
		return fmt.Errorf("nonce(%d) different with getNonce(%d), transaction already exists %v", nonce, getNonce, hex.EncodeToString(txhash))
	}
	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	if tx.OpCode() == rtypes.Op_KV {
		kvData := &rtypes.KV{}
		if err := rlp.DecodeBytes(tx.Data(), kvData); err != nil {
			return fmt.Errorf("rlp decode to kv error %s", err.Error())
		}
		if len(kvData.Key) > 256 || len(kvData.Value) > 512 {
			return fmt.Errorf("key or value too big,MaxKey:256,MaxValue:512")
		}
		if ok, _ := app.stateDb.Has(append(KvPrefix, kvData.Key...)); ok {
			return fmt.Errorf("duplicate key :%v", kvData.Key)
		}
	} else {
		if app.state.GetBalance(from).Cmp(tx.Cost()) < 0 {
			return fmt.Errorf("not enough funds")
		}
	}

	return nil
}

func (app *EVMApp) SaveReceipts() ([]byte, error) {
	savedReceipts := make([][]byte, 0, len(app.receipts)+len(app.kvs))
	receiptBatch := app.stateDb.NewBatch()

	for _, receipt := range app.receipts {
		storageReceipt := (*etypes.ReceiptForStorage)(receipt)
		storageReceiptBytes, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			return nil, fmt.Errorf("wrong rlp encode:%v", err.Error())
		}

		key := append(ReceiptsPrefix, receipt.TxHash.Bytes()...)
		if err := receiptBatch.Put(key, storageReceiptBytes); err != nil {
			return nil, fmt.Errorf("batch receipt failed:%v", err.Error())
		}
		savedReceipts = append(savedReceipts, storageReceiptBytes)
	}

	for _, kv := range app.kvs {
		kvBytes, err := rlp.EncodeToBytes(kv)
		if err != nil {
			return nil, fmt.Errorf("wrong rlp encode:%v", err.Error())
		}
		key := append(KvPrefix, kv.Key...)
		if err := receiptBatch.Put(key, kv.Value); err != nil {
			return nil, fmt.Errorf("batch receipt failed:%v", err.Error())
		}
		savedReceipts = append(savedReceipts, kvBytes)
	}

	if err := receiptBatch.Write(); err != nil {
		return nil, fmt.Errorf("persist receipts failed:%v", err.Error())
	}
	rHash := merkle.SimpleHashFromHashes(savedReceipts)
	return rHash, nil
}

func (app *EVMApp) Info() (resInfo gtypes.ResultInfo) {
	lb := &LastBlockInfo{
		AppHash: make([]byte, 0),
		Height:  0,
	}
	if res, err := app.LoadLastBlock(lb); err == nil {
		lb = res.(*LastBlockInfo)
	}

	resInfo.LastBlockAppHash = lb.AppHash
	resInfo.LastBlockHeight = lb.Height
	resInfo.Version = "alpha 0.2"
	resInfo.Data = "default app with evm-1.5.9"
	return
}

func (app *EVMApp) Query(query []byte) (res gtypes.Result) {
	action := query[0]
	load := query[1:]
	switch action {
	case rtypes.QueryType_Contract:
		res = app.queryContract(load, 0)
	case rtypes.QueryTypeContractByHeight:
		if len(load) < 8 {
			return gtypes.NewError(gtypes.CodeType_BaseInvalidInput, "wrong height")
		}
		h := binary.BigEndian.Uint64(load[len(load)-8:])
		res = app.queryContract(load[:len(load)-8], h)
	case rtypes.QueryType_Nonce:
		res = app.queryNonce(load)
	case rtypes.QueryType_Receipt:
		res = app.queryReceipt(load)
	case rtypes.QueryType_Existence:
		res = app.queryContractExistence(load)
	case rtypes.QueryType_PayLoad:
		res = app.queryPayLoad(load)
	case rtypes.QueryType_TxRaw:
		res = app.queryTransaction(load)
	case rtypes.QueryType_Key:
		res = app.queryKey(load)
	case rtypes.QueryType_Key_Prefix:
		res = app.queryKeyWithPrefix(load)
	default:
		res = gtypes.NewError(gtypes.CodeType_BaseInvalidInput, "unimplemented query")
	}

	// check if contract exists
	return res
}

func (app *EVMApp) queryKey(load []byte) gtypes.Result {
	value, err := app.stateDb.Get(append(KvPrefix, load...))
	if err != nil {
		return gtypes.NewError(gtypes.CodeType_InternalError, "fail to get value for key:"+string(load))
	}
	return gtypes.NewResultOK(value, "")
}

func (app *EVMApp) queryKeyWithPrefix(load []byte) gtypes.Result {
	st := &struct {
		Prefix []byte
		SeeKey []byte
		Limit  uint32
	}{}
	if err := rlp.DecodeBytes(load, st); err != nil {
		return gtypes.NewError(gtypes.CodeType_WrongRLP, "rlp decode error:"+string(load))
	}
	if st.Limit > 200 {
		st.Limit = 200
	}
	kvs, err := app.stateDb.GetWithPrefix(append(KvPrefix, st.Prefix...), append(KvPrefix, st.SeeKey...), st.Limit, len(KvPrefix))
	if err != nil {
		return gtypes.NewError(gtypes.CodeType_InternalError, "fail to get value for key:"+string(st.SeeKey))
	}
	bytKvs, err := rlp.EncodeToBytes(kvs)
	if err != nil {
		return gtypes.NewError(gtypes.CodeType_WrongRLP, "rlp encode error:"+err.Error())
	}
	return gtypes.NewResultOK(bytKvs, "")
}

func (app *EVMApp) queryContractExistence(load []byte) gtypes.Result {
	tx := new(etypes.Transaction)
	err := rlp.DecodeBytes(load, tx)
	if err != nil {
		return gtypes.NewError(gtypes.CodeType_BaseInvalidInput, err.Error())
	}
	contractAddr := tx.To()

	app.stateMtx.Lock()
	hashBytes := app.state.GetCodeHash(*contractAddr).Bytes()
	app.stateMtx.Unlock()

	if bytes.Equal(tx.Data(), hashBytes) {
		return gtypes.NewResultOK(append([]byte{}, byte(0x01)), "contract exists")
	}
	return gtypes.NewResultOK(append([]byte{}, byte(0x00)), "constract doesn't exist")
}

func (app *EVMApp) queryContract(load []byte, height uint64) gtypes.Result {
	tx := new(etypes.Transaction)
	err := rlp.DecodeBytes(load, tx)
	if err != nil {
		return gtypes.NewError(gtypes.CodeType_BaseInvalidInput, err.Error())
	}

	from, err := app.Signer.Sender(tx)
	if err != nil {
		return gtypes.NewError(gtypes.CodeType_BaseInvalidInput, err.Error())
	}
	txMsg := etypes.NewMessage(from, tx.To(), 0, tx.Value(), tx.Gas(), tx.GasPrice(), tx.Data(), false)

	bc := NewBlockChain(app.stateDb)

	var vmEnv *vm.EVM

	if height == 0 {

		envCxt := core.NewEVMContext(txMsg, app.currentHeader, bc, nil)

		app.stateMtx.Lock()
		vmEnv = vm.NewEVM(envCxt, app.state.Copy(), app.chainConfig, evmConfig)
		app.stateMtx.Unlock()
	} else {
		//appHash save in next block AppHash
		height++
		blockMeta, err := app.core.GetBlockMeta(int64(height))
		if err != nil {
			return gtypes.NewError(gtypes.CodeType_BaseInvalidInput, err.Error())
		}
		ethHeader := makeETHHeader(blockMeta.Header)
		envCxt := core.NewEVMContext(txMsg, ethHeader, bc, nil)

		trieRoot := EmptyTrieRoot
		if len(blockMeta.Header.AppHash) > 0 {
			trieRoot = common.BytesToHash(blockMeta.Header.AppHash)
		}

		state, err := estate.New(trieRoot, estate.NewDatabase(app.stateDb))
		if err != nil {
			return gtypes.NewError(gtypes.CodeType_BaseInvalidInput, err.Error())
		}
		vmEnv = vm.NewEVM(envCxt, state, app.chainConfig, evmConfig)
	}

	gpl := new(core.GasPool).AddGas(math.MaxBig256.Uint64())
	res, _, _, err := core.ApplyMessage(vmEnv, txMsg, gpl) // we don't care about gasUsed
	if err != nil {
		log.Warn("query apply msg err", zap.Error(err))
	}

	return gtypes.NewResultOK(res, "")
}

func makeETHHeader(header *gtypes.Header) *etypes.Header {
	return &etypes.Header{
		ParentHash: common.BytesToHash(header.LastBlockID.Hash),
		Difficulty: big.NewInt(0),
		GasLimit:   math.MaxBig256.Uint64(),
		Time:       big.NewInt(header.Time.Unix()),
		Number:     big.NewInt(header.Height),
	}
}

func (app *EVMApp) queryNonce(addrBytes []byte) gtypes.Result {
	if len(addrBytes) != 20 {
		return gtypes.NewError(gtypes.CodeType_BaseInvalidInput, "Invalid address")
	}
	addr := common.BytesToAddress(addrBytes)

	app.stateMtx.Lock()
	nonce := app.state.GetNonce(addr)
	app.stateMtx.Unlock()

	data, err := rlp.EncodeToBytes(nonce)
	if err != nil {
		log.Warn("query error", zap.Error(err))
	}
	return gtypes.NewResultOK(data, "")
}

func (app *EVMApp) queryReceipt(txHashBytes []byte) gtypes.Result {
	key := append(ReceiptsPrefix, txHashBytes...)
	data, err := app.stateDb.Get(key)
	if err != nil {
		return gtypes.NewError(gtypes.CodeType_InternalError, "fail to get receipt for tx:"+string(key))
	}
	return gtypes.NewResultOK(data, "")
}

func (app *EVMApp) queryTransaction(txHashBytes []byte) gtypes.Result {
	if len(txHashBytes) == 0 {
		return gtypes.NewError(gtypes.CodeType_BaseInvalidInput, "Empty query")
	}

	var res gtypes.Result
	data, err := app.core.Query(txHashBytes[0], txHashBytes[1:])
	if err != nil {
		return gtypes.NewError(gtypes.CodeType_InternalError, err.Error())
	}

	bs, err := rlp.EncodeToBytes(data)
	if err != nil {
		return gtypes.NewError(gtypes.CodeType_InternalError, err.Error())
	}
	res.Data = bs
	res.Code = gtypes.CodeType_OK
	return res
}

func (app *EVMApp) queryPayLoad(txHashBytes []byte) gtypes.Result {
	if len(txHashBytes) == 0 {
		return gtypes.NewError(gtypes.CodeType_BaseInvalidInput, "Empty query")
	}

	var res gtypes.Result
	data, err := app.core.Query(txHashBytes[0], txHashBytes[1:])
	if err != nil {
		return gtypes.NewError(gtypes.CodeType_InternalError, err.Error())
	}

	if value, ok := data.([]byte); ok {
		res.Data = value
	}

	res.Code = gtypes.CodeType_OK
	return res
}

func (app *EVMApp) SetCore(core gtypes.Core) {
	app.core = core
}
