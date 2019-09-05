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

package gemmill

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	rpcserver "github.com/dappledger/AnnChain/gemmill/rpc/server"

	"github.com/dappledger/AnnChain/gemmill/consensus/pbft"
	"github.com/dappledger/AnnChain/gemmill/consensus/raft"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/dappledger/AnnChain/gemmill/archive"
	"github.com/dappledger/AnnChain/gemmill/blockchain"
	config "github.com/dappledger/AnnChain/gemmill/config"
	"github.com/dappledger/AnnChain/gemmill/consensus"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	"github.com/dappledger/AnnChain/gemmill/mempool"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	dbm "github.com/dappledger/AnnChain/gemmill/modules/go-db"
	"github.com/dappledger/AnnChain/gemmill/modules/go-events"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/dappledger/AnnChain/gemmill/plugin"
	"github.com/dappledger/AnnChain/gemmill/refuse_list"
	"github.com/dappledger/AnnChain/gemmill/state"
	"github.com/dappledger/AnnChain/gemmill/trace"
	"github.com/dappledger/AnnChain/gemmill/types"
	"github.com/dappledger/AnnChain/gemmill/utils/zip"
)

const version = "0.9.0"

// Angine is a high level abstraction of all the state, consensus, mempool blah blah...
type Angine struct {
	Tune *Tunes

	mtx     sync.Mutex
	tune    *Tunes
	hooked  bool
	started bool

	app types.Application

	dbs           map[string]dbm.DB
	privValidator *types.PrivValidator
	blockstore    *blockchain.BlockStore
	dataArchive   *archive.Archive
	conf          *viper.Viper
	txPool        types.TxPool
	consensus     consensus.Engine
	traceRouter   *trace.Router
	stateMachine  *state.State
	p2pSwitch     *p2p.Switch
	eventSwitch   *types.EventSwitch
	refuseList    *refuse_list.RefuseList
	p2pHost       string
	p2pPort       uint16
	genesis       *types.GenesisDoc
	addrBook      *p2p.AddrBook
	plugins       []plugin.IPlugin

	getAdminVote func([]byte, *types.Validator) ([]byte, error)

	queryPayLoadTxParser func([]byte) ([]byte, error)

	apis []map[string]*rpcserver.RPCFunc
}

type Tunes struct {
	Runtime string
	Conf    *viper.Viper
}

// Initialize generates genesis.json and priv_validator.json automatically.
// It is usually used with commands like "init" before user put the node into running.
func Initialize(tune *Tunes, chainId string) {
	crypto.NodeInit(crypto.CryptoType)
	if err := config.InitRuntime(tune.Runtime, chainId, tune.Conf); err != nil {
		gcmn.Exit(gcmn.Fmt("Init Runtime error: %v", err))
	}
}

// NewAngine makes and returns a new angine, which can be used directly after being imported
func NewAngine(app types.Application, tune *Tunes) (angine *Angine, err error) {
	var conf *viper.Viper
	if tune.Conf == nil {
		conf, err = config.ReadConfig(tune.Runtime)
		if err != nil {
			return nil, err
		}
	} else {
		conf = tune.Conf
	}

	conf.AutomaticEnv()

	var refuseList *refuse_list.RefuseList
	dbs := openDBs(conf)
	defer func() {
		if angine == nil {
			if refuseList != nil {
				refuseList.Stop()
			}

			for _, db := range dbs {
				db.Close()
			}
		}
	}()

	dbBackend := conf.GetString("db_backend")
	dbDir := conf.GetString("db_dir")

	genesis, err := getGenesisFile(conf)
	if err != nil {
		if err != GENESIS_NOT_FOUND {
			return nil, err
		}
		genesis = nil
	}

	logger, err := getLogger(conf)
	if err != nil {
		fmt.Println("failed to get logger: ", err)
		return nil, err
	}
	log.SetLog(logger)
	crypto.NodeInit(crypto.CryptoType)
	privValidator, err := types.LoadPrivValidator(conf.GetString("priv_validator_file"))
	if err != nil {
		fmt.Println("LoadPrivValidator error: ", err)
		return nil, err
	}
	refuseList = refuse_list.NewRefuseList(dbBackend, dbDir)
	p2psw, err := prepareP2P(conf, genesis, privValidator, refuseList)
	if err != nil {
		fmt.Println("prepare p2p error: ", err)
		return nil, err
	}

	p2pListener := p2psw.Listeners()[0]
	dataArchive := archive.NewArchive(dbBackend, dbDir, conf.GetInt64("threshold_blocks"))
	eventSwitch := types.NewEventSwitch()
	angine = &Angine{
		Tune: tune,

		dbs:         dbs,
		tune:        tune,
		dataArchive: dataArchive,
		conf:        conf,

		p2pSwitch:     p2psw,
		eventSwitch:   &eventSwitch,
		refuseList:    refuseList,
		privValidator: privValidator,
		p2pHost:       p2pListener.ExternalAddress().IP.String(),
		p2pPort:       p2pListener.ExternalAddress().Port,
		genesis:       genesis,
	}

	angine.app = app
	err = angine.buildState(genesis)

	if tune.Conf == nil {
		tune.Conf = conf
	} else if tune.Runtime == "" {
		tune.Runtime = conf.GetString("runtime")
	}

	return
}

func (a *Angine) APIs() []map[string]*rpcserver.RPCFunc {
	return a.apis
}

func (a *Angine) SetQueryPayLoadTxParser(fn func([]byte) ([]byte, error)) {
	a.queryPayLoadTxParser = fn
}

func (a *Angine) OnRecvExchangeData(data *p2p.ExchangeData) error {
	if data == nil {
		return nil
	}
	if a.genesis == nil {
		if len(data.GenesisJSON) == 0 {
			// TODO wait ...
			return errors.New("no genesis file found in other node")
		}
		othGenesis, err := types.GenesisDocFromJSONRet(data.GenesisJSON)
		if err != nil {
			// TODO log err
			fmt.Println("oth genesis err:", err)
			return err
		}
		a.p2pSwitch.GetExchangeData().GenesisJSON = data.GenesisJSON
		if err = a.buildState(othGenesis); err != nil {
			// TODO log err
			fmt.Println("build state err:", err)
			return err
		}
		if a.stateMachine == nil {
			return errors.New("state generaterr")
		}
		a.p2pSwitch.Start()
	}
	return nil
}

func (a *Angine) buildState(genesis *types.GenesisDoc) error {
	stateM, err := getOrMakeState(a.conf, a.dbs["state"], genesis)
	if err != nil {
		fmt.Println("getOrMakeState error: ", err)
		return err
	}

	if stateM == nil {
		// delay assemble
		a.p2pSwitch.SetDealExchangeDataFunc(a.OnRecvExchangeData)
		return nil
	}
	a.assembleStateMachine(stateM)
	return nil
}

func openDBs(conf *viper.Viper) map[string]dbm.DB {
	dbBackend := conf.GetString("db_backend")
	dbDir := conf.GetString("db_dir")
	dbArchiveDir := conf.GetString("db_archive_dir")
	dbs := make(map[string]dbm.DB)
	dbs["state"] = dbm.NewDB("state", dbBackend, dbDir)
	dbs["blockstore"] = dbm.NewDB("blockstore", dbBackend, dbDir)
	dbs["archive"] = dbm.NewDB("blockstore", dbBackend, dbArchiveDir)
	dbs["votechannel"] = dbm.NewDB("votechannel", dbBackend, dbDir)

	return dbs
}

func closeDBs(a *Angine) {
	for _, db := range a.dbs {
		db.Close()
	}
}

func (ang *Angine) assembleStateMachine(stateM *state.State) {
	conf := ang.tune.Conf
	conf.Set("chain_id", stateM.ChainID)

	fastSync := fastSyncable(conf, ang.privValidator.GetAddress(), stateM.Validators)

	blockStore := blockchain.NewBlockStore(ang.dbs["blockstore"], ang.dbs["archive"])
	_, stateLastHeight, _ := stateM.GetLastBlockInfo()
	bcReactor := blockchain.NewBlockchainReactor(conf, stateLastHeight, blockStore, fastSync, ang.dataArchive)
	var txPool types.TxPool
	if txPoolApp, isType := ang.app.(types.TxPoolApplication); isType {
		log.Info("app implemented tx pool")
		txPool = txPoolApp.GetTxPool()
	} else {
		log.Info("app does not implement tx pool, use default mempool")
		txPool = mempool.NewMempool(conf)
	}
	memReactor := mempool.NewTxReactor(conf, txPool)

	var consensusEngine consensus.Engine

	if conf.Get("consensus") == "raft" {

		consensusState, err := raft.NewConsensusState(conf, *ang.eventSwitch, blockStore, stateM, txPool, ang.privValidator)
		if err != nil {
			log.Fatal("assembleStateMachine with raft err", zap.Error(err))
		}
		consensusEngine = consensusState

		bcReactor.SetBlockVerifier(func(bID types.BlockID, h int64, lc *types.Commit) error { return nil })

		ang.apis = append(ang.apis, consensusState.NewPublicAPI().API())

	} else {

		consensusState := pbft.NewConsensusState(conf, stateM, blockStore, txPool)
		consensusState.SetPrivValidator(ang.privValidator)

		consensusReactor := pbft.NewConsensusReactor(consensusState, fastSync)
		consensusState.BindReactor(consensusReactor)
		ang.p2pSwitch.AddReactor("CONSENSUS", consensusReactor)
		setEventSwitch(*ang.eventSwitch, bcReactor, memReactor, consensusReactor)

		consensusEngine = consensusState

		bcReactor.SetBlockVerifier(func(bID types.BlockID, h int64, lc *types.Commit) error {
			return stateM.Validators.VerifyCommit(stateM.ChainID, bID, h, lc)
		})
	}

	bcReactor.SetBlockExecuter(func(blk *types.Block, pst *types.PartSet, c *types.Commit) error {
		blockStore.SaveBlock(blk, pst, c)
		if err := stateM.ApplyBlock(*ang.eventSwitch, blk, pst.Header(), txPool, -1); err != nil {
			log.Error("bc,ApplyBlock err", zap.Int64("height", blk.Height), zap.Error(err))
			return err
		}
		stateM.Save()
		log.Debug("save to db", zap.Int64("height", blk.Height), zap.String("state receiptHash", fmt.Sprintf("%X", stateM.ReceiptsHash)), zap.String("block receiptHash", fmt.Sprintf("%X", blk.ReceiptsHash)))
		return nil
	})

	ang.p2pSwitch.AddReactor("MEMPOOL", memReactor)
	ang.p2pSwitch.AddReactor("BLOCKCHAIN", bcReactor)

	var addrBook *p2p.AddrBook
	if conf.GetBool("pex_reactor") {
		addrBook = p2p.NewAddrBook(conf.GetString("addrbook_file"), conf.GetBool("addrbook_strict"))
		pexReactor := p2p.NewPEXReactor(addrBook)
		ang.p2pSwitch.AddReactor("PEX", pexReactor)
	}

	if conf.GetBool("auth_by_ca") {
		ang.p2pSwitch.SetAuthByCA(authByCA(conf, &stateM.Validators))
	}

	setEventSwitch(*ang.eventSwitch, bcReactor, memReactor, consensusEngine)

	ang.blockstore = blockStore
	ang.consensus = consensusEngine
	ang.txPool = txPool

	ang.stateMachine = stateM
	ang.genesis = stateM.GenesisDoc
	ang.addrBook = addrBook
	ang.stateMachine.SetBlockExecutable(ang)
	ang.stateMachine.SetBlockVerifier(consensusEngine)

	ang.InitPlugins()
	for _, p := range ang.plugins {
		txPool.RegisterFilter(types.NewTxpoolFilter(p.CheckTx))
	}
}

func (e *Angine) ConnectApp(app types.Application) {
	e.hooked = true
	hooks := app.GetAngineHooks()
	if hooks.OnExecute == nil || hooks.OnCommit == nil {
		gcmn.PanicSanity("At least implement OnExecute & OnCommit, otherwise what your application is for")
	}

	types.AddListenerForEvent(*e.eventSwitch, "angine", types.EventStringHookNewRound(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookNewRound)
		if hooks.OnNewRound == nil {
			data.ResCh <- types.NewRoundResult{}
			return
		}
		hooks.OnNewRound.Sync(data.Height, data.Round, nil)
		result := hooks.OnNewRound.Result()
		if r, ok := result.(types.NewRoundResult); ok {
			data.ResCh <- r
		} else {
			data.ResCh <- types.NewRoundResult{}
		}
	})
	if hooks.OnPropose != nil {
		types.AddListenerForEvent(*e.eventSwitch, "angine", types.EventStringHookPropose(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPropose)
			hooks.OnPropose.Async(data.Height, data.Round, nil, nil, nil)
		})
	}
	if hooks.OnPrevote != nil {
		types.AddListenerForEvent(*e.eventSwitch, "angine", types.EventStringHookPrevote(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPrevote)
			hooks.OnPrevote.Async(data.Height, data.Round, data.Block, nil, nil)
		})
	}
	if hooks.OnPrecommit != nil {
		types.AddListenerForEvent(*e.eventSwitch, "angine", types.EventStringHookPrecommit(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPrecommit)
			hooks.OnPrecommit.Async(data.Height, data.Round, data.Block, nil, nil)
		})
	}
	types.AddListenerForEvent(*e.eventSwitch, "angine", types.EventStringHookExecute(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookExecute)
		hooks.OnExecute.Sync(data.Height, data.Round, data.Block)
		result := hooks.OnExecute.Result()
		if r, ok := result.(types.ExecuteResult); ok {
			data.ResCh <- r
		} else {
			data.ResCh <- types.ExecuteResult{}
		}

	})
	types.AddListenerForEvent(*e.eventSwitch, "angine", types.EventStringHookCommit(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookCommit)
		if hooks.OnCommit == nil {
			data.ResCh <- types.CommitResult{}
			return
		}
		hooks.OnCommit.Sync(data.Height, data.Round, data.Block)
		result := hooks.OnCommit.Result()
		if cs, ok := result.(types.CommitResult); ok {
			data.ResCh <- cs
		} else {
			data.ResCh <- types.CommitResult{}
		}
	})

	e.app = app

	if e.genesis == nil {
		return
	}

	info := app.Info()
	if err := e.RecoverFromCrash(info.LastBlockAppHash, int64(info.LastBlockHeight)); err != nil {
		gcmn.PanicSanity(fmt.Sprintf("replay blocks on angine start failed,err:%v", err))
	}
}

func (e *Angine) PrivValidator() *types.PrivValidator {
	return e.privValidator
}

func (e *Angine) Genesis() *types.GenesisDoc {
	return e.genesis
}

func (e *Angine) P2PHost() string {
	return e.p2pHost
}

func (e *Angine) P2PPort() uint16 {
	return e.p2pPort
}

func (e *Angine) DialSeeds(seeds []string) {
	e.p2pSwitch.DialSeeds(seeds)
}

func (e *Angine) RegisterNodeInfo(ni *p2p.NodeInfo) {
	e.p2pSwitch.SetNodeInfo(ni)
}

func (e *Angine) GetNodeInfo() *p2p.NodeInfo {
	return e.p2pSwitch.NodeInfo()
}

func (e *Angine) Height() int64 {
	return e.blockstore.Height()
}

func (e *Angine) OriginHeight() int64 {
	return e.blockstore.OriginHeight()
}

func (e *Angine) GetBlock(height int64) (block *types.Block, meta *types.BlockMeta, err error) {

	meta, err = e.GetBlockMeta(height)
	if err != nil {
		return
	}
	if height > e.blockstore.OriginHeight() {
		block = e.blockstore.LoadBlock(height)
	} else {
		archiveDB, errN := e.newArchiveDB(height)
		if errN != nil {
			err = errN
			return
		}
		defer archiveDB.Close()
		newStore := blockchain.NewBlockStore(archiveDB, nil)
		block = newStore.LoadBlock(height)
	}
	return
}

func (e *Angine) GetBlockMeta(height int64) (meta *types.BlockMeta, err error) {

	if height == 0 {
		err = errors.New("height must be greater than 0")
		return
	}
	if height > e.Height() {
		err = fmt.Errorf("height(%d) must be less than the current blockchain height(%d)", height, e.Height())
		return
	}
	if height > e.blockstore.OriginHeight() {
		meta = e.blockstore.LoadBlockMeta(height)
	} else {
		archiveDB, errN := e.newArchiveDB(height)
		if errN != nil {
			err = errN
			return
		}
		defer archiveDB.Close()
		newStore := blockchain.NewBlockStore(archiveDB, nil)
		meta = newStore.LoadBlockMeta(height)
	}
	return
}

func (e *Angine) newArchiveDB(height int64) (archiveDB dbm.DB, err error) {

	fileHash := string(e.dataArchive.QueryFileHash(height))
	archiveDir := e.conf.GetString("db_archive_dir")
	_, err = os.Stat(filepath.Join(archiveDir, fileHash+".zip"))
	if err != nil {
		// download archive data zip
		err = e.dataArchive.Client.DownloadFile(fileHash, filepath.Join(archiveDir, fileHash+".zip"))
		if err != nil {
			return
		}
		err = zip.Decompress(filepath.Join(archiveDir, fileHash+".zip"), filepath.Join(archiveDir, fileHash+".db"))
		if err != nil {
			return
		}
	}

	archiveDB = dbm.NewDB(fileHash, e.conf.GetString("db_backend"), archiveDir)
	return
}

func (e *Angine) NoneGenesis() bool {
	return e.genesis == nil
}

func (e *Angine) Start() error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	if e.started {
		return errors.New("can't start angine twice")
	}
	if !e.hooked {
		e.hookDefaults()
	}

	if e.stateMachine != nil {
		if _, err := e.p2pSwitch.Start(); err != nil {
			return err
		}
	}

	if _, err := (*e.eventSwitch).Start(); err != nil {
		fmt.Println("fail to start event switch, error: ", err)
		return err
	}

	e.started = true
	seeds := e.tune.Conf.GetString("seeds")
	if seeds != "" {
		e.DialSeeds(strings.Split(seeds, ","))
	}

	return nil
}

// Stop just wrap around swtich.Stop, which will stop reactors, listeners,etc
func (ang *Angine) Stop() bool {
	ret := ang.p2pSwitch.Stop()
	ang.Destroy()
	return ret
}

// Destroy is called after something go south while before angine.Start has been called
func (ang *Angine) Destroy() {
	for _, p := range ang.plugins {
		p.Stop()
	}

	if ang.refuseList != nil {
		ang.refuseList.Stop()
	}
	ang.dataArchive.Close()
	closeDBs(ang)
}

func (e *Angine) BroadcastTx(tx []byte) error {
	return e.txPool.ReceiveTx(tx)
}

func (e *Angine) BroadcastTxCommit(tx []byte) (err error) {
	if err = e.txPool.ReceiveTx(tx); err != nil {
		return
	}
	committed := make(chan types.EventDataTx, 1)
	eventString := types.EventStringTx(tx)
	types.AddListenerForEvent(*e.eventSwitch, "angine", eventString, func(data types.TMEventData) {
		select {
		case committed <- data.(types.EventDataTx):
			return
		case <-time.NewTimer(2 * time.Second).C:
			return
		}
	})
	defer func() {
		(*e.eventSwitch).(events.EventSwitch).RemoveListenerForEvent(eventString, "angine")
	}()
	select {
	case c := <-committed: // in EventDataTx, Only Code and Error is used
		if c.Code == types.CodeType_OK {
			return
		}
		err = errors.New(c.Error)
		return
	case <-time.NewTimer(60 * 2 * time.Second).C:
		err = errors.New("Timed out waiting for transaction to be included in a block")
		return
	}
}

func (e *Angine) FlushMempool() {
	e.txPool.Flush()
}

func (e *Angine) GetValidators() (int64, *types.ValidatorSet) {
	return e.stateMachine.LastBlockHeight, e.stateMachine.Validators
}

func (e *Angine) GetP2PNetInfo() (bool, []string, []*types.Peer) {
	listening := e.p2pSwitch.IsListening()
	listeners := []string{}
	for _, l := range e.p2pSwitch.Listeners() {
		listeners = append(listeners, l.String())
	}
	peers := make([]*types.Peer, 0, e.p2pSwitch.Peers().Size())
	for _, p := range e.p2pSwitch.Peers().List() {
		peers = append(peers, &types.Peer{
			NodeInfo:         *p.NodeInfo,
			IsOutbound:       p.IsOutbound(),
			ConnectionStatus: p.Connection().Status(),
		})
	}
	return listening, listeners, peers
}

func (e *Angine) GetNumPeers() int {
	o, i, d := e.p2pSwitch.NumPeers()
	return o + i + d
}

func (e *Angine) GetConsensusStateInfo() (string, []string) {
	c, ok := e.consensus.(*pbft.ConsensusState)
	if !ok {
		return "", nil
	}
	roundState := c.GetRoundState()
	peerRoundStates := make([]string, 0, e.p2pSwitch.Peers().Size())
	for _, p := range e.p2pSwitch.Peers().List() {
		peerState := p.Data.Get(types.PeerStateKey).(*pbft.PeerState)
		peerRoundState := peerState.GetRoundState()
		peerRoundStateStr := p.Key + ":" + string(wire.JSONBytes(peerRoundState))
		peerRoundStates = append(peerRoundStates, peerRoundStateStr)
	}
	return roundState.String(), peerRoundStates
}

func (e *Angine) GetNumUnconfirmedTxs() int {
	return e.txPool.Size()
}

func (e *Angine) GetUnconfirmedTxs() []types.Tx {
	return e.txPool.Reap(-1)
}

func (e *Angine) IsNodeValidator(pub crypto.PubKey) bool {
	_, vals := e.consensus.GetValidators()
	for _, v := range vals {
		if pub.KeyString() == v.PubKey.KeyString() {
			return true
		}
	}
	return false
}

func (e *Angine) GetBlacklist() []string {
	return e.refuseList.ListAllKey()
}

func (ang *Angine) Query(queryType byte, load []byte) (interface{}, error) {
	switch queryType {
	case types.QueryTxExecution:
		data, err := ang.QueryPayLoad(load)
		if err != nil {
			return nil, err
		}
		return data, nil
	case types.QueryTx:
		return ang.QueryTransaction(load)
	}

	return nil, errors.Errorf("[Angine Query] no such query type: %v", queryType)
}

func (ang *Angine) QueryTransaction(load []byte) (interface{}, error) {
	for _, p := range ang.plugins {

		if qc, ok := p.(*plugin.QueryCachePlugin); ok {

			info, err := qc.ExecutionResult(load)
			if err != nil {
				return nil, err
			}

			block, _, err := ang.GetBlock(int64(info.Height))
			if err != nil {
				return nil, fmt.Errorf("[Angine Query] fail to get block:%v", err)
			}

			if int(info.Index) >= len(block.Txs) {
				return nil, fmt.Errorf("[Angine Query] fail to get block, invalid tx index")
			}
			tx := block.Data.Txs[info.Index]
			timestamp := block.Header.Time.UnixNano()

			t := types.ResultTransaction{
				BlockHash:        info.BlockHash,
				BlockHeight:      info.Height,
				TransactionIndex: info.Index,
				RawTransaction:   []byte(tx),
				Timestamp:        uint64(timestamp),
			}
			return &t, err
		}
	}
	return nil, errors.New("not found")
}

func (ang *Angine) QueryPayLoad(load []byte) (interface{}, error) {

	tx, err := ang.QueryTransaction(load)
	if err != nil {
		return nil, err
	}
	rawTx := tx.(*types.ResultTransaction)
	return ang.queryPayLoadTxParser(rawTx.RawTransaction)
}

// Recover world status
// Replay all blocks after blockHeight and ensure the result matches the current state.
func (e *Angine) RecoverFromCrash(appHash []byte, appBlockHeight int64) error {
	storeBlockHeight := e.blockstore.Height()
	stateBlockHeight := e.stateMachine.LastBlockHeight

	if storeBlockHeight == 0 {
		return nil // no blocks to replay
	}

	log.Info("Replay Blocks", zap.Int64("appHeight", appBlockHeight), zap.Int64("storeHeight", storeBlockHeight), zap.Int64("stateHeight", stateBlockHeight))

	if storeBlockHeight < appBlockHeight {
		// if the app is ahead, there's nothing we can do
		return state.ErrAppBlockHeightTooHigh{CoreHeight: storeBlockHeight, AppHeight: appBlockHeight}
	} else if storeBlockHeight == appBlockHeight {
		// We ran Commit, but if we crashed before state.Save(),
		// load the intermediate state and update the state.AppHash.
		// NOTE: If ABCI allowed rollbacks, we could just replay the
		// block even though it's been committed
		stateAppHash := e.stateMachine.AppHash
		lastBlockAppHash := e.blockstore.LoadBlock(storeBlockHeight).AppHash

		if bytes.Equal(stateAppHash, appHash) {
			// we're all synced up
			log.Debug("RelpayBlocks: Already synced")
		} else if bytes.Equal(stateAppHash, lastBlockAppHash) {
			// we crashed after commit and before saving state,
			// so load the intermediate state and update the hash
			e.stateMachine.LoadIntermediate()
			e.stateMachine.AppHash = appHash
			log.Debug("RelpayBlocks: Loaded intermediate state and updated state.AppHash")
		} else {
			gcmn.PanicSanity(gcmn.Fmt("Unexpected state.AppHash: state.AppHash %X; app.AppHash %X, lastBlock.AppHash %X", stateAppHash, appHash, lastBlockAppHash))
		}
		return nil
	} else if storeBlockHeight == appBlockHeight+1 &&
		storeBlockHeight == stateBlockHeight+1 {
		// We crashed after saving the block
		// but before Commit (both the state and app are behind),
		// so just replay the block

		// check that the lastBlock.AppHash matches the state apphash
		block := e.blockstore.LoadBlock(storeBlockHeight)
		if !bytes.Equal(block.Header.AppHash, appHash) {
			return state.ErrLastStateMismatch{Height: storeBlockHeight, Core: block.Header.AppHash, App: appHash}
		}

		blockMeta := e.blockstore.LoadBlockMeta(storeBlockHeight)
		// h.nBlocks++
		// replay the latest block
		return e.stateMachine.ApplyBlock(*e.eventSwitch, block, blockMeta.PartsHeader, MockMempool{}, 0)
	} else if storeBlockHeight != stateBlockHeight {
		// unless we failed before committing or saving state (previous 2 case),
		// the store and state should be at the same height!
		if storeBlockHeight == stateBlockHeight+1 {
			e.stateMachine.AppHash = appHash
			e.stateMachine.LastBlockHeight = storeBlockHeight
			e.stateMachine.LastBlockID = e.blockstore.LoadBlockMeta(storeBlockHeight).Header.LastBlockID
			e.stateMachine.LastBlockTime = e.blockstore.LoadBlockMeta(storeBlockHeight).Header.Time
		} else {
			gcmn.PanicSanity(gcmn.Fmt("Expected storeHeight (%d) and stateHeight (%d) to match.", storeBlockHeight, stateBlockHeight))
		}
	} else {
		// store is more than one ahead,
		// so app wants to replay many blocks
		// replay all blocks starting with appBlockHeight+1
		// var eventCache types.Fireable // nil
		// TODO: use stateBlockHeight instead and let the consensus state do the replay
		for h := appBlockHeight + 1; h <= storeBlockHeight; h++ {
			// h.nBlocks++
			block := e.blockstore.LoadBlock(h)
			blockMeta := e.blockstore.LoadBlockMeta(h)
			e.stateMachine.ApplyBlock(*e.eventSwitch, block, blockMeta.PartsHeader, MockMempool{}, 0)
		}
		if !bytes.Equal(e.stateMachine.AppHash, appHash) {
			return fmt.Errorf("Ann state.AppHash does not match AppHash after replay. Got %X, expected %X", appHash, e.stateMachine.AppHash)
		}
		return nil
	}
	return nil
}

func (e *Angine) hookDefaults() {
	types.AddListenerForEvent(*e.eventSwitch, "angine", types.EventStringHookNewRound(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookNewRound)
		data.ResCh <- types.NewRoundResult{}
	})
	types.AddListenerForEvent(*e.eventSwitch, "angine", types.EventStringHookExecute(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookExecute)
		data.ResCh <- types.ExecuteResult{}
	})
	types.AddListenerForEvent(*e.eventSwitch, "angine", types.EventStringHookCommit(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookCommit)
		data.ResCh <- types.CommitResult{}
	})
}

func (ang *Angine) BeginBlock(block *types.Block, eventFireable events.Fireable, blockPartsHeader *types.PartSetHeader) error {
	params := &plugin.BeginBlockParams{Block: block}
	for _, p := range ang.plugins {
		_, err := p.BeginBlock(params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ang *Angine) ExecAdminTx(app plugin.AdminApp, tx []byte) error {
	var ip *plugin.AdminOp
	ok := false
	for _, p := range ang.plugins {
		if ip, ok = p.(*plugin.AdminOp); ok {
			break
		}
	}
	if ip != nil {
		return ip.ExecTX(app, tx)
	}
	return fmt.Errorf("there is no plugin.AdminOp")
}

func (ang *Angine) ExecBlock(block *types.Block, eventFireable events.Fireable, executeResult *types.ExecuteResult) error {
	params := &plugin.ExecBlockParams{
		Block:      block,
		ValidTxs:   executeResult.ValidTxs,
		InvalidTxs: executeResult.InvalidTxs,
	}
	for _, p := range ang.plugins {
		if _, ok := p.(*plugin.AdminOp); ok {
			continue
		}
		_, err := p.ExecBlock(params)
		if err != nil {
			return err
		}
	}

	return nil
}

// plugins modify changedValidators inplace
func (ang *Angine) EndBlock(block *types.Block, eventFireable events.Fireable, blockPartsHeader *types.PartSetHeader, changedValAttrs []*types.ValidatorAttr, nextVS *types.ValidatorSet) error {
	params := &plugin.EndBlockParams{
		Block:             block,
		ChangedValidators: changedValAttrs,
		NextValidatorSet:  nextVS,
	}
	for _, p := range ang.plugins {
		_, err := p.EndBlock(params)
		if err != nil {
			return err
		}
	}

	return nil
}

func setEventSwitch(evsw types.EventSwitch, eventables ...types.Eventable) {
	for _, e := range eventables {
		e.SetEventSwitch(evsw)
	}
}

func addToRefuselist(refuseList *refuse_list.RefuseList) func([]byte) error {
	return func(pk []byte) error {
		refuseList.AddRefuseKey(pk)
		return nil
	}
}

func refuseListFilter(refuseList *refuse_list.RefuseList) func(crypto.PubKey) error {
	return func(pubkey crypto.PubKey) error {
		if refuseList.QueryRefuseKey(pubkey.Bytes()) {
			return fmt.Errorf("%s in refuselist", pubkey.KeyString())
		}
		return nil
	}
}

func authByCA(conf *viper.Viper, ppValidators **types.ValidatorSet) func(*p2p.NodeInfo) error {
	valset := *ppValidators
	return func(peerNodeInfo *p2p.NodeInfo) error {
		// validator node must be signed by CA
		// but normal node can bypass auth check if config says so
		if valset.HasAddress(peerNodeInfo.PubKey.Address()) && !conf.GetBool("non_validator_node_auth") {
			return nil
		}
		msg, err := hex.DecodeString(peerNodeInfo.PubKey.KeyString())
		if err != nil {
			return err
		}
		for _, val := range valset.Validators {
			if !val.IsCA {
				continue // CA must be validator
			}

			signedPkByte64, err := types.StringTo64byte(peerNodeInfo.SigndPubKey)
			if err != nil {
				return err
			}
			sig := crypto.SetNodeSignature(signedPkByte64[:])
			if val.PubKey.VerifyBytes(msg, sig) {
				log.Infow("Peer handshake", "peerNodeInfo", peerNodeInfo)
				return nil
			}
		}
		err = fmt.Errorf("REJECT! No peers with valid CA signatures are found!")
		log.Warn(err.Error())
		return err
	}
}

func (ang *Angine) InitPlugins() {
	ps := strings.Split(ang.genesis.Plugins, ",")
	pk := ang.privValidator.GetPrivKey()
	params := &plugin.InitParams{
		StateDB:    ang.dbs["state"],
		Switch:     ang.p2pSwitch,
		PrivKey:    pk,
		RefuseList: ang.refuseList,
		Validators: &ang.stateMachine.Validators,
	}

	for _, pn := range ps {
		switch pn {
		case "adminOp":
			p := &plugin.AdminOp{}
			p.Init(params)
			ang.plugins = append(ang.plugins, p)
		case "querycache":
			p := &plugin.QueryCachePlugin{}
			querydb, err := ensureQueryDB(ang.tune.Conf.GetString("db_dir"))
			if err != nil {
				// querydb failure is something that we can bear with
				log.Error("[QueryCachePlugin Init]", zap.Error(err))
				fmt.Println(err)
			}
			params.StateDB = querydb
			p.Init(params)
			ang.plugins = append(ang.plugins, p)
		case "":
			// no core_plugins is allowed, so just ignore it
		default:

		}
	}
}

func ensureQueryDB(dbDir string) (*dbm.GoLevelDB, error) {
	if err := gcmn.EnsureDir(path.Join(dbDir, "query_cache"), 0775); err != nil {
		return nil, fmt.Errorf("fail to ensure tx_execution_result")
	}
	querydb, err := dbm.NewGoLevelDB("tx_execution_result", path.Join(dbDir, "query_cache"))
	if err != nil {
		return nil, fmt.Errorf("fail to open tx_execution_result")
	}
	return querydb, nil
}

func fastSyncable(conf *viper.Viper, selfAddress []byte, validators *types.ValidatorSet) bool {
	// We don't fast-sync when the only validator is us.
	fastSync := conf.GetBool("fast_sync")
	if validators.Size() == 1 {
		addr, _ := validators.GetByIndex(0)
		if bytes.Equal(selfAddress, addr) {
			fastSync = false
		}
	}
	return fastSync
}

// func getGenesisFileMust(conf cfg.Config) *types.GenesisDoc {
// 	genDocFile := conf.GetString("genesis_file")
// 	if !gcmn.FileExists(genDocFile) {
// 		gcmn.PanicSanity("missing genesis_file")
// 	}
// 	jsonBlob, err := ioutil.ReadFile(genDocFile)
// 	if err != nil {
// 		gcmn.Exit(gcmn.Fmt("Couldn't read GenesisDoc file: %v", err))
// 	}
// 	genDoc := types.GenesisDocFromJSON(jsonBlob)
// 	if genDoc.ChainID == "" {
// 		gcmn.PanicSanity(gcmn.Fmt("Genesis doc %v must include non-empty chain_id", genDocFile))
// 	}
// 	conf.Set("chain_id", genDoc.ChainID)

// 	return genDoc
// }

var (
	GENESIS_NOT_FOUND = errors.New("missing genesis_file")
)

func getGenesisFile(conf *viper.Viper) (*types.GenesisDoc, error) {
	genDocFile := conf.GetString("genesis_file")
	if !gcmn.FileExists(genDocFile) {
		return nil, GENESIS_NOT_FOUND
	}
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil, fmt.Errorf("Couldn't read GenesisDoc file: %v", err)
	}
	genDoc := types.GenesisDocFromJSON(jsonBlob)
	if genDoc.ChainID == "" {
		return nil, fmt.Errorf("Genesis doc %v must include non-empty chain_id", genDocFile)
	}
	conf.Set("chain_id", genDoc.ChainID)

	return genDoc, nil
}

func getLogger(conf *viper.Viper) (*zap.Logger, error) {
	logpath := conf.GetString("log_path")
	if logpath == "" {
		dir, _ := os.Getwd()
		logpath = filepath.Join(dir, "output.log")
	}
	return log.Initialize(conf.GetString("environment"), logpath)
}

func getOrMakeState(conf *viper.Viper, stateDB dbm.DB, genesis *types.GenesisDoc) (*state.State, error) {
	stateM := state.GetState(conf, stateDB)
	if stateM == nil {
		if genesis != nil {
			if stateM = state.MakeGenesisState(stateDB, genesis); stateM == nil {
				return nil, fmt.Errorf("fail to get genesis state")
			}
		}
	}
	return stateM, nil
}

func prepareP2P(conf *viper.Viper, genesis *types.GenesisDoc, privValidator *types.PrivValidator, refuseList *refuse_list.RefuseList) (*p2p.Switch, error) {
	var genesisJSON []byte
	var err error
	if genesis != nil {
		genesisJSON = genesis.JSONBytes()
	}
	p2psw := p2p.NewSwitch(conf)
	protocol, address := ProtocolAndAddress(conf.GetString("p2p_laddr"))
	defaultListener, err := p2p.NewDefaultListener(protocol, address, conf.GetBool("skip_upnp"))
	if err != nil {
		return nil, errors.Wrap(err, "prepareP2P")
	}
	nodeInfo := &p2p.NodeInfo{
		PubKey:      privValidator.GetPubKey(),
		SigndPubKey: conf.GetString("signbyCA"),
		Moniker:     conf.GetString("moniker"),
		ListenAddr:  defaultListener.ExternalAddress().String(),
		Version:     version,
	}
	privKey := privValidator.GetPrivKey()
	p2psw.AddListener(defaultListener)
	p2psw.SetExchangeData(&p2p.ExchangeData{
		GenesisJSON: genesisJSON,
	})
	p2psw.SetNodeInfo(nodeInfo)
	p2psw.SetNodePrivKey(privKey)
	p2psw.SetAddToRefuselist(addToRefuselist(refuseList))
	p2psw.SetRefuseListFilter(refuseListFilter(refuseList))

	return p2psw, nil
}

// ProtocolAndAddress accepts tcp by default
func ProtocolAndAddress(listenAddr string) (string, string) {
	protocol, address := "tcp", listenAddr
	parts := strings.SplitN(address, "://", 2)
	if len(parts) == 2 {
		protocol, address = parts[0], parts[1]
	}
	return protocol, address
}

// Updates to the mempool need to be synchronized with committing a block
// so apps can reset their transient state on Commit
type MockMempool struct {
}

func (m MockMempool) Lock()                               {}
func (m MockMempool) Unlock()                             {}
func (m MockMempool) Update(height int64, txs []types.Tx) {}

type ITxCheck interface {
	CheckTx(types.Tx) (bool, error)
}

// func (ang *Angine) newArchiveDB(height def.INT) (archiveDB dbm.DB, err error) {
//     fileHash := string(ang.dataArchive.QueryFileHash(height))
//     archiveDir := ang.conf.GetString("db_archive_dir")
//     tiClient := ti.NewTiCapsuleClient(
//         ang.conf.GetString("ti_endpoint"),
//         ang.conf.GetString("ti_key"),
//         ang.conf.GetString("ti_secret"),
//     )
//     _, err = os.Stat(filepath.Join(archiveDir, fileHash+".zip"))
//     if err != nil {
//         err = tiClient.DownloadFile(fileHash, filepath.Join(archiveDir, fileHash+".zip"))
//         if err != nil {
//             return
//         }
//         err = zip.Decompress(filepath.Join(archiveDir, fileHash+".zip"), filepath.Join(archiveDir, fileHash+".db"))
//         if err != nil {
//             return
//         }
//     }

//     archiveDB = dbm.NewDB(fileHash, ang.conf.GetString("db_backend"), archiveDir)
//     return
// }

const (
	TIME_OUT_HEALTH = int64(time.Second * 60)
)

func (ag *Angine) HealthStatus() int {
	cur := time.Now().Unix()
	c, ok := ag.consensus.(*pbft.ConsensusState)
	if !ok {
		return http.StatusNoContent
	}

	lcommitTime := c.CommitTime.Unix()

	if cur > lcommitTime+TIME_OUT_HEALTH {
		return int(http.StatusInternalServerError)
	}
	return http.StatusOK
}
