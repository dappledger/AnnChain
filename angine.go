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

package angine

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"gitlab.zhonganonline.com/ann/angine/blockchain"
	ac "gitlab.zhonganonline.com/ann/angine/config"
	"gitlab.zhonganonline.com/ann/angine/consensus"
	"gitlab.zhonganonline.com/ann/angine/mempool"
	"gitlab.zhonganonline.com/ann/angine/plugin"
	"gitlab.zhonganonline.com/ann/angine/refuse_list"
	"gitlab.zhonganonline.com/ann/angine/state"
	"gitlab.zhonganonline.com/ann/angine/types"
	"gitlab.zhonganonline.com/ann/ann-module/lib/ed25519"
	cmn "gitlab.zhonganonline.com/ann/ann-module/lib/go-common"
	cfg "gitlab.zhonganonline.com/ann/ann-module/lib/go-config"
	crypto "gitlab.zhonganonline.com/ann/ann-module/lib/go-crypto"
	dbm "gitlab.zhonganonline.com/ann/ann-module/lib/go-db"
	"gitlab.zhonganonline.com/ann/ann-module/lib/go-events"
	p2p "gitlab.zhonganonline.com/ann/ann-module/lib/go-p2p"
	"gitlab.zhonganonline.com/ann/ann-module/lib/go-wire"
)

const version = "0.6.0"

type (
	// Angine is a high level abstraction of all the state, consensus, mempool blah blah...
	Angine struct {
		Tune *Tunes

		mtx     sync.Mutex
		tune    *Tunes
		hooked  bool
		started bool

		statedb       dbm.DB
		blockdb       dbm.DB
		querydb       dbm.DB
		privValidator *types.PrivValidator
		blockstore    *blockchain.BlockStore
		mempool       *mempool.Mempool
		consensus     *consensus.ConsensusState
		stateMachine  *state.State
		p2pSwitch     *p2p.Switch
		eventSwitch   *types.EventSwitch
		refuseList    *refuse_list.RefuseList
		p2pHost       string
		p2pPort       uint16
		genesis       *types.GenesisDoc

		logger *zap.Logger

		getSpecialVote func([]byte, *types.Validator) ([]byte, error)
	}

	// Tunes wraps two different kinds of configurations for angine
	Tunes struct {
		Runtime string
		Conf    *cfg.MapConfig
	}
)

// Defaults to tcp
func ProtocolAndAddress(listenAddr string) (string, string) {
	protocol, address := "tcp", listenAddr
	parts := strings.SplitN(address, "://", 2)
	if len(parts) == 2 {
		protocol, address = parts[0], parts[1]
	}
	return protocol, address
}

// Initialize generates genesis.json and priv_validator.json automatically.
// It is usually used with commands like "init" before user put the node into running.
func Initialize(tune *Tunes) {
	var conf *cfg.MapConfig
	if tune.Conf == nil {
		conf = ac.GetConfig(tune.Runtime)
	} else {
		conf = tune.Conf
	}
	priv := genPrivFile(conf.GetString("priv_validator_file"))
	gvs := []types.GenesisValidator{types.GenesisValidator{
		PubKey:     priv.PubKey,
		Amount:     100,
		IsCA:       true,
		RPCAddress: conf.GetString("rpc_laddr"),
	}}
	genDoc, err := genGenesiFile(conf.GetString("genesis_file"), gvs)
	if err != nil {
		cmn.PanicSanity(err)
	}

	fmt.Println("Initialized ", genDoc.ChainID, "genesis", conf.GetString("genesis_file"), "priv_validator", conf.GetString("priv_validator_file"))
	fmt.Println("Check the files generated, make sure everything is OK.")
}

// NewAngine makes and returns a new angine, which can be used directly after being imported
func NewAngine(lgr *zap.Logger, tune *Tunes) *Angine {
	var conf *cfg.MapConfig
	if tune.Conf == nil {
		conf = ac.GetConfig(tune.Runtime)
	} else {
		conf = tune.Conf
	}

	dbBackend := conf.GetString("db_backend")
	dbDir := conf.GetString("db_dir")
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	blockStoreDB := dbm.NewDB("blockstore", dbBackend, dbDir)
	querydb, err := ensureQueryDB(dbDir)
	if err != nil {
		lgr.Error("angine error", zap.Error(err))
		fmt.Println(err)
	}
	gotGenesis := true
	genesis, err := getGenesisFile(conf) // ignore any error
	if err != nil {
		gotGenesis = false
	}
	stateM, err := getOrMakeState(conf, stateDB, genesis)
	if err != nil {
		lgr.Error("angine error", zap.Error(err))
		return nil
	}

	if gotGenesis {
		conf.Set("chain_id", genesis.ChainID)
	}
	chainID := conf.GetString("chain_id")

	logger, err := getLogger(conf, chainID)
	if err != nil {
		lgr.Error("fail to get logger", zap.Error(err))
		return nil
	}

	privValidator := types.LoadOrGenPrivValidator(logger, conf.GetString("priv_validator_file"))
	refuseList := refuse_list.NewRefuseList(dbBackend, dbDir)
	eventSwitch := types.NewEventSwitch(logger)
	if _, err := eventSwitch.Start(); err != nil {
		logger.Error("fail to start event switch", zap.Error(err))
		return nil
	}

	gb := make([]byte, 0)
	if gotGenesis {
		gb = wire.JSONBytes(genesis)
	}
	p2psw, err := prepareP2P(logger, conf, gb, privValidator, refuseList)
	p2pListener := p2psw.Listeners()[0]

	if tune.Conf == nil {
		tune.Conf = conf
	} else if tune.Runtime == "" {
		tune.Runtime = conf.GetString("datadir")
	}
	angine := &Angine{
		Tune: tune,

		statedb: stateDB,
		blockdb: blockStoreDB,
		querydb: querydb,
		tune:    tune,

		p2pSwitch:     p2psw,
		eventSwitch:   &eventSwitch,
		refuseList:    refuseList,
		privValidator: privValidator,
		p2pHost:       p2pListener.ExternalAddress().IP.String(),
		p2pPort:       p2pListener.ExternalAddress().Port,
		genesis:       genesis,

		logger: logger,
	}

	if gotGenesis {
		assembleStateMachine(angine, stateM)
	} else {
		p2psw.SetGenesisUnmarshal(func(b []byte) error {
			g := types.GenesisDocFromJSON(b)
			if g.ChainID != chainID {
				return fmt.Errorf("wrong chain id from genesis, expect %v, got %v", chainID, g.ChainID)
			}
			if err := g.SaveAs(conf.GetString("genesis_file")); err != nil {
				return err
			}
			angine.genesis = g
			assembleStateMachine(angine, state.MakeGenesisState(stateDB, g))

			return nil
		})
	}

	return angine
}

func (ang *Angine) SetSpecialVoteRPC(f func([]byte, *types.Validator) ([]byte, error)) {
	ang.getSpecialVote = f
}

func (ang *Angine) ConnectApp(app types.Application) error {
	ang.hooked = true
	hooks := app.GetAngineHooks()
	if hooks.OnExecute == nil || hooks.OnCommit == nil {
		ang.logger.Error("At least implement OnExecute & OnCommit, otherwise what your application is for?")
		return fmt.Errorf("no hooks implemented")
	}

	types.AddListenerForEvent(*ang.eventSwitch, "angine", types.EventStringHookNewRound(), func(ed types.TMEventData) {
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
		types.AddListenerForEvent(*ang.eventSwitch, "angine", types.EventStringHookPropose(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPropose)
			hooks.OnPropose.Async(data.Height, data.Round, nil, nil, nil)
		})
	}
	if hooks.OnPrevote != nil {
		types.AddListenerForEvent(*ang.eventSwitch, "angine", types.EventStringHookPrevote(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPrevote)
			hooks.OnPrevote.Async(data.Height, data.Round, data.Block, nil, nil)
		})
	}
	if hooks.OnPrecommit != nil {
		types.AddListenerForEvent(*ang.eventSwitch, "angine", types.EventStringHookPrecommit(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPrecommit)
			hooks.OnPrecommit.Async(data.Height, data.Round, data.Block, nil, nil)
		})
	}
	types.AddListenerForEvent(*ang.eventSwitch, "angine", types.EventStringHookExecute(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookExecute)
		hooks.OnExecute.Sync(data.Height, data.Round, data.Block)
		result := hooks.OnExecute.Result()
		if r, ok := result.(types.ExecuteResult); ok {
			data.ResCh <- r
		} else {
			data.ResCh <- types.ExecuteResult{}
		}

	})
	types.AddListenerForEvent(*ang.eventSwitch, "angine", types.EventStringHookCommit(), func(ed types.TMEventData) {
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

	if ang.genesis == nil {
		return nil
	}

	info := app.Info()
	if err := ang.RecoverFromCrash(info.LastBlockAppHash, int(info.LastBlockHeight)); err != nil {
		return err
	}

	return nil
}

func (ang *Angine) PrivValidator() *types.PrivValidator {
	return ang.privValidator
}

func (ang *Angine) Genesis() *types.GenesisDoc {
	return ang.genesis
}

func (ang *Angine) P2PHost() string {
	return ang.p2pHost
}

func (ang *Angine) P2PPort() uint16 {
	return ang.p2pPort
}

func (ang *Angine) DialSeeds(seeds []string) {
	ang.p2pSwitch.DialSeeds(seeds)
}

func (ang *Angine) Start() error {
	ang.mtx.Lock()
	defer ang.mtx.Unlock()
	if ang.started {
		return fmt.Errorf("can't start angine twice")
	}
	if !ang.hooked {
		ang.hookDefaults()
	}
	if _, err := ang.p2pSwitch.Start(); err == nil {
		ang.started = true
	} else {
		return err
	}

	seeds := ang.tune.Conf.GetString("seeds")
	if seeds != "" {
		ang.DialSeeds(strings.Split(seeds, ","))
	}

	return nil
}

// Stop just wrap around swtich.Stop, which will stop reactors, listeners,etc
func (ang *Angine) Stop() bool {
	ang.refuseList.Stop()
	ang.statedb.Close()
	ang.blockdb.Close()
	ang.querydb.Close()
	return ang.p2pSwitch.Stop()
}

func (ang *Angine) RegisterNodeInfo(ni *p2p.NodeInfo) {
	ang.p2pSwitch.SetNodeInfo(ni)
}

func (ang *Angine) GetNodeInfo() *p2p.NodeInfo {
	return ang.p2pSwitch.NodeInfo()
}

func (ang *Angine) Height() int {
	return ang.blockstore.Height()
}

func (ang *Angine) GetBlock(height int) (*types.Block, *types.BlockMeta) {
	if height == 0 {
		return nil, nil
	}
	return ang.blockstore.LoadBlock(height), ang.blockstore.LoadBlockMeta(height)
}

func (ang *Angine) BroadcastTx(tx []byte) error {
	return ang.mempool.CheckTx(tx)
}

func (ang *Angine) BroadcastTxCommit(tx []byte) error {
	if err := ang.mempool.CheckTx(tx); err != nil {
		return err
	}
	committed := make(chan types.EventDataTx, 1)
	eventString := types.EventStringTx(tx)
	timer := time.NewTimer(60 * 2 * time.Second)
	types.AddListenerForEvent(*ang.eventSwitch, "angine", eventString, func(data types.TMEventData) {
		committed <- data.(types.EventDataTx)
	})
	defer func() {
		(*ang.eventSwitch).(events.EventSwitch).RemoveListenerForEvent(eventString, "angine")
	}()
	select {
	case res := <-committed:
		if res.Code == types.CodeType_OK {
			return nil
		}
		return fmt.Errorf(res.Error)
	case <-timer.C:
		return fmt.Errorf("Timed out waiting for transaction to be included in a block")
	}
}

func (ang *Angine) FlushMempool() {
	ang.mempool.Flush()
}

func (ang *Angine) GetValidators() (int, *types.ValidatorSet) {
	return ang.stateMachine.LastBlockHeight, ang.stateMachine.Validators
}

func (ang *Angine) GetP2PNetInfo() (bool, []string, []*types.Peer) {
	listening := ang.p2pSwitch.IsListening()
	listeners := []string{}
	for _, l := range ang.p2pSwitch.Listeners() {
		listeners = append(listeners, l.String())
	}
	peers := make([]*types.Peer, 0, ang.p2pSwitch.Peers().Size())
	for _, p := range ang.p2pSwitch.Peers().List() {
		peers = append(peers, &types.Peer{
			NodeInfo:         *p.NodeInfo,
			IsOutbound:       p.IsOutbound(),
			ConnectionStatus: p.Connection().Status(),
		})
	}
	return listening, listeners, peers
}

func (ang *Angine) GetNumPeers() int {
	o, i, d := ang.p2pSwitch.NumPeers()
	return o + i + d
}

func (ang *Angine) GetConsensusStateInfo() (string, []string) {
	roundState := ang.consensus.GetRoundState()
	peerRoundStates := make([]string, 0, ang.p2pSwitch.Peers().Size())
	for _, p := range ang.p2pSwitch.Peers().List() {
		peerState := p.Data.Get(types.PeerStateKey).(*consensus.PeerState)
		peerRoundState := peerState.GetRoundState()
		peerRoundStateStr := p.Key + ":" + string(wire.JSONBytes(peerRoundState))
		peerRoundStates = append(peerRoundStates, peerRoundStateStr)
	}
	return roundState.String(), peerRoundStates
}

func (ang *Angine) GetNumUnconfirmedTxs() int {
	return ang.mempool.Size()
}

func (ang *Angine) GetUnconfirmedTxs() []types.Tx {
	return ang.mempool.Reap(-1)
}

func (ang *Angine) IsNodeValidator(pub crypto.PubKey) bool {
	edPub := pub.(crypto.PubKeyEd25519)
	_, vals := ang.consensus.GetValidators()
	for _, v := range vals {
		if edPub.KeyString() == v.PubKey.KeyString() {
			return true
		}
	}
	return false
}

func (ang *Angine) GetBlacklist() []string {
	return ang.refuseList.ListAllKey()
}

func (ang *Angine) Query(queryType byte, load []byte) (interface{}, error) {
	return ang.QueryExecutionResult(load)
}

func (ang *Angine) QueryExecutionResult(txHash []byte) (*types.TxExecutionResult, error) {
	item := ang.querydb.Get(txHash)
	if len(item) == 0 {
		return nil, fmt.Errorf("no execution result for %v", txHash)
	}
	info := &types.TxExecutionResult{}
	if err := info.FromBytes(item); err != nil {
		return nil, err
	}
	return info, nil
}

// Recover world status
// Replay all blocks after blockHeight and ensure the result matches the current state.
func (ang *Angine) RecoverFromCrash(appHash []byte, appBlockHeight int) error {
	storeBlockHeight := ang.blockstore.Height()
	stateBlockHeight := ang.stateMachine.LastBlockHeight

	if storeBlockHeight == 0 {
		return nil // no blocks to replay
	}

	ang.logger.Info("Replay Blocks", zap.Int("appHeight", appBlockHeight), zap.Int("storeHeight", storeBlockHeight), zap.Int("stateHeight", stateBlockHeight))

	if storeBlockHeight < appBlockHeight {
		// if the app is ahead, there's nothing we can do
		return state.ErrAppBlockHeightTooHigh{CoreHeight: storeBlockHeight, AppHeight: appBlockHeight}
	} else if storeBlockHeight == appBlockHeight {
		// We ran Commit, but if we crashed before state.Save(),
		// load the intermediate state and update the state.AppHash.
		// NOTE: If ABCI allowed rollbacks, we could just replay the
		// block even though it's been committed
		stateAppHash := ang.stateMachine.AppHash
		lastBlockAppHash := ang.blockstore.LoadBlock(storeBlockHeight).AppHash

		if bytes.Equal(stateAppHash, appHash) {
			// we're all synced up
			ang.logger.Debug("RelpayBlocks: Already synced")
		} else if bytes.Equal(stateAppHash, lastBlockAppHash) {
			// we crashed after commit and before saving state,
			// so load the intermediate state and update the hash
			ang.stateMachine.LoadIntermediate()
			ang.stateMachine.AppHash = appHash
			ang.logger.Debug("RelpayBlocks: Loaded intermediate state and updated state.AppHash")
		} else {
			cmn.PanicSanity(cmn.Fmt("Unexpected state.AppHash: state.AppHash %X; app.AppHash %X, lastBlock.AppHash %X", stateAppHash, appHash, lastBlockAppHash))
		}
		return nil
	} else if storeBlockHeight == appBlockHeight+1 &&
		storeBlockHeight == stateBlockHeight+1 {
		// We crashed after saving the block
		// but before Commit (both the state and app are behind),
		// so just replay the block

		// check that the lastBlock.AppHash matches the state apphash
		block := ang.blockstore.LoadBlock(storeBlockHeight)
		if !bytes.Equal(block.Header.AppHash, appHash) {
			return state.ErrLastStateMismatch{Height: storeBlockHeight, Core: block.Header.AppHash, App: appHash}
		}

		blockMeta := ang.blockstore.LoadBlockMeta(storeBlockHeight)
		// h.nBlocks++
		// replay the latest block
		return ang.stateMachine.ApplyBlock(*ang.eventSwitch, block, blockMeta.PartsHeader, MockMempool{}, 0)
	} else if storeBlockHeight != stateBlockHeight {
		// unless we failed before committing or saving state (previous 2 case),
		// the store and state should be at the same height!
		if storeBlockHeight == stateBlockHeight+1 {
			ang.stateMachine.AppHash = appHash
			ang.stateMachine.LastBlockHeight = storeBlockHeight
			ang.stateMachine.LastBlockID = ang.blockstore.LoadBlockMeta(storeBlockHeight).Header.LastBlockID
			ang.stateMachine.LastBlockTime = ang.blockstore.LoadBlockMeta(storeBlockHeight).Header.Time
		} else {
			cmn.PanicSanity(cmn.Fmt("Expected storeHeight (%d) and stateHeight (%d) to match.", storeBlockHeight, stateBlockHeight))
		}
	} else {
		// store is more than one ahead,
		// so app wants to replay many blocks
		// replay all blocks starting with appBlockHeight+1
		// var eventCache types.Fireable // nil
		// TODO: use stateBlockHeight instead and let the consensus state do the replay
		for h := appBlockHeight + 1; h <= storeBlockHeight; h++ {
			// h.nBlocks++
			block := ang.blockstore.LoadBlock(h)
			blockMeta := ang.blockstore.LoadBlockMeta(h)
			ang.stateMachine.ApplyBlock(*ang.eventSwitch, block, blockMeta.PartsHeader, MockMempool{}, 0)
		}
		if !bytes.Equal(ang.stateMachine.AppHash, appHash) {
			return fmt.Errorf("Ann state.AppHash does not match AppHash after replay. Got %X, expected %X", appHash, ang.stateMachine.AppHash)
		}
		return nil
	}
	return nil
}

func (ang *Angine) hookDefaults() {
	types.AddListenerForEvent(*ang.eventSwitch, "angine", types.EventStringHookNewRound(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookNewRound)
		data.ResCh <- types.NewRoundResult{}
	})
	types.AddListenerForEvent(*ang.eventSwitch, "angine", types.EventStringHookExecute(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookExecute)
		data.ResCh <- types.ExecuteResult{}
	})
	types.AddListenerForEvent(*ang.eventSwitch, "angine", types.EventStringHookCommit(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookCommit)
		data.ResCh <- types.CommitResult{}
	})
}

func setEventSwitch(evsw types.EventSwitch, eventables ...types.Eventable) {
	for _, e := range eventables {
		e.SetEventSwitch(evsw)
	}
}

func addToRefuselist(refuseList *refuse_list.RefuseList) func([32]byte) error {
	return func(pk [32]byte) error {
		refuseList.AddRefuseKey(pk)
		return nil
	}
}

func refuseListFilter(refuseList *refuse_list.RefuseList) func(crypto.PubKeyEd25519) error {
	return func(pubkey crypto.PubKeyEd25519) error {
		if refuseList.QueryRefuseKey(pubkey) {
			return fmt.Errorf("%s in refuselist", pubkey.KeyString())
		}
		return nil
	}
}

func authByCA(chainID string, ppValidators **types.ValidatorSet, log *zap.Logger) func(*p2p.NodeInfo) error {
	valset := *ppValidators
	chainIDBytes := []byte(chainID)
	return func(peerNodeInfo *p2p.NodeInfo) error {
		msg := append(peerNodeInfo.PubKey[:], chainIDBytes...)
		for _, val := range valset.Validators {
			if !val.IsCA {
				continue // CA must be validator
			}
			valPk := [32]byte(val.PubKey.(crypto.PubKeyEd25519))
			signedPkByte64, err := types.StringTo64byte(peerNodeInfo.SigndPubKey)
			if err != nil {
				return err
			}
			if ed25519.Verify(&valPk, msg, &signedPkByte64) {
				log.Sugar().Infow("Peer handshake", "peerNodeInfo", peerNodeInfo)
				return nil
			}
		}
		err := fmt.Errorf("Reject Peer, has no CA sig")
		log.Warn(err.Error())
		return err
	}
}

func initCorePlugins(sm *state.State, privkey crypto.PrivKeyEd25519, sw *p2p.Switch, ppValset **types.ValidatorSet, rl *refuse_list.RefuseList) {
	params := &plugin.InitPluginParams{
		Switch:     sw,
		PrivKey:    privkey,
		RefuseList: rl,
		Validators: ppValset,
	}
	for _, plug := range sm.Plugins {
		plug.InitPlugin(params)
	}
}

func fastSyncable(conf cfg.Config, selfAddress []byte, validators *types.ValidatorSet) bool {
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

func getGenesisFile(conf cfg.Config) (*types.GenesisDoc, error) {
	genDocFile := conf.GetString("genesis_file")
	if !cmn.FileExists(genDocFile) {
		return nil, fmt.Errorf("missing genesis_file")
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

func getLogger(conf cfg.Config, chainID string) (*zap.Logger, error) {
	logpath := conf.GetString("log_path")
	if logpath == "" {
		logpath, _ = os.Getwd()
	}
	logpath = path.Join(logpath, "angine-"+chainID)
	if err := cmn.EnsureDir(logpath, 0700); err != nil {
		return nil, err
	}
	if logger := InitializeLog(conf.GetString("environment"), logpath); logger != nil {
		return logger, nil
	}
	return nil, fmt.Errorf("fail to build zap logger")
}

func genPrivFile(path string) *types.PrivValidator {
	privValidator := types.GenPrivValidator(nil)
	privValidator.SetFile(path)
	privValidator.Save()
	return privValidator
}

func genGenesiFile(path string, gVals []types.GenesisValidator) (*types.GenesisDoc, error) {
	genDoc := &types.GenesisDoc{
		ChainID: cmn.Fmt("annchain-%v", cmn.RandStr(6)),
		Plugins: "specialop",
	}
	genDoc.Validators = gVals
	return genDoc, genDoc.SaveAs(path)
}

func checkPrivValidatorFile(conf cfg.Config) error {
	if privFile := conf.GetString("priv_validator_file"); !cmn.FileExists(privFile) {
		return fmt.Errorf("PrivValidator file needed: %s", privFile)
	}
	return nil
}

func checkGenesisFile(conf cfg.Config) error {
	if genFile := conf.GetString("genesis_file"); !cmn.FileExists(genFile) {
		return fmt.Errorf("Genesis file needed: %s", genFile)
	}
	return nil
}

func ensureQueryDB(dbDir string) (*dbm.GoLevelDB, error) {
	if err := cmn.EnsureDir(path.Join(dbDir, "query_cache"), 0775); err != nil {
		return nil, fmt.Errorf("fail to ensure tx_execution_result")
	}
	querydb, err := dbm.NewGoLevelDB("tx_execution_result", path.Join(dbDir, "query_cache"))
	if err != nil {
		return nil, fmt.Errorf("fail to open tx_execution_result")
	}
	return querydb, nil
}

func getOrMakeState(conf cfg.Config, stateDB dbm.DB, genesis *types.GenesisDoc) (*state.State, error) {
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

func prepareP2P(logger *zap.Logger, conf cfg.Config, genesisBytes []byte, privValidator *types.PrivValidator, refuseList *refuse_list.RefuseList) (*p2p.Switch, error) {
	p2psw := p2p.NewSwitch(logger, conf.GetConfig("p2p"), genesisBytes)
	protocol, address := ProtocolAndAddress(conf.GetString("node_laddr"))
	defaultListener, err := p2p.NewDefaultListener(logger, protocol, address, conf.GetBool("skip_upnp"))
	if err != nil {
		return nil, errors.Wrap(err, "prepareP2P")
	}
	nodeInfo := &p2p.NodeInfo{
		PubKey:      privValidator.PubKey.(crypto.PubKeyEd25519),
		SigndPubKey: conf.GetString("signbyCA"),
		Moniker:     conf.GetString("moniker"),
		ListenAddr:  defaultListener.ExternalAddress().String(),
		Version:     version,
	}
	privKey := privValidator.PrivKey
	p2psw.AddListener(defaultListener)
	p2psw.SetNodeInfo(nodeInfo)
	p2psw.SetNodePrivKey(privKey.(crypto.PrivKeyEd25519))
	p2psw.SetAddToRefuselist(addToRefuselist(refuseList))
	p2psw.SetRefuseListFilter(refuseListFilter(refuseList))

	return p2psw, nil
}

func assembleStateMachine(angine *Angine, stateM *state.State) {
	conf := angine.tune.Conf

	fastSync := fastSyncable(conf, angine.privValidator.GetAddress(), stateM.Validators)
	stateM.SetLogger(angine.logger)
	stateM.SetQueryDB(angine.querydb)

	blockStore := blockchain.NewBlockStore(angine.blockdb)
	_, stateLastHeight, _ := stateM.GetLastBlockInfo()
	bcReactor := blockchain.NewBlockchainReactor(angine.logger, conf, stateLastHeight, blockStore, fastSync)
	mem := mempool.NewMempool(angine.logger, conf)
	for _, p := range stateM.Plugins {
		mem.RegisterFilter(NewMempoolFilter(p.CheckTx))
	}
	memReactor := mempool.NewMempoolReactor(angine.logger, conf, mem)

	consensusState := consensus.NewConsensusState(angine.logger, conf, stateM, blockStore, mem)
	consensusState.SetPrivValidator(angine.privValidator)
	consensusReactor := consensus.NewConsensusReactor(angine.logger, consensusState, fastSync)
	consensusState.BindReactor(consensusReactor)

	bcReactor.SetBlockVerifier(func(bID types.BlockID, h int, lc *types.Commit) error {
		return stateM.Validators.VerifyCommit(stateM.ChainID, bID, h, lc)
	})
	bcReactor.SetBlockExecuter(func(blk *types.Block, pst *types.PartSet, c *types.Commit) error {
		blockStore.SaveBlock(blk, pst, c)
		if err := stateM.ApplyBlock(*angine.eventSwitch, blk, pst.Header(), MockMempool{}, -1); err != nil {
			return err
		}
		stateM.Save()
		return nil
	})

	privKey := angine.privValidator.GetPrivateKey()

	angine.p2pSwitch.AddReactor("MEMPOOL", memReactor)
	angine.p2pSwitch.AddReactor("BLOCKCHAIN", bcReactor)
	angine.p2pSwitch.AddReactor("CONSENSUS", consensusReactor)

	if conf.GetBool("pex_reactor") {
		addrBook := p2p.NewAddrBook(angine.logger, conf.GetString("addrbook_file"), conf.GetBool("addrbook_strict"))
		addrBook.Start()
		pexReactor := p2p.NewPEXReactor(angine.logger, addrBook)
		angine.p2pSwitch.AddReactor("PEX", pexReactor)
	}

	angine.p2pSwitch.SetAuthByCA(authByCA(stateM.ChainID, &stateM.Validators, angine.logger))

	setEventSwitch(*angine.eventSwitch, bcReactor, memReactor, consensusReactor)
	initCorePlugins(stateM, privKey.(crypto.PrivKeyEd25519), angine.p2pSwitch, &stateM.Validators, angine.refuseList)

	angine.blockstore = blockStore
	angine.consensus = consensusState
	angine.mempool = mem
	angine.stateMachine = stateM
}

// --------------------------------------------------------------------------------

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
type MempoolFilter struct {
	cb func([]byte) (bool, error)
}

func (m MempoolFilter) CheckTx(tx types.Tx) (bool, error) {
	return m.cb(tx)
}

func NewMempoolFilter(f func([]byte) (bool, error)) MempoolFilter {
	return MempoolFilter{cb: f}
}
