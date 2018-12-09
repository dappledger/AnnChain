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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine/blockchain"
	ac "github.com/dappledger/AnnChain/angine/config"
	"github.com/dappledger/AnnChain/angine/consensus"
	"github.com/dappledger/AnnChain/angine/mempool"
	"github.com/dappledger/AnnChain/angine/plugin"
	"github.com/dappledger/AnnChain/angine/refuse_list"
	"github.com/dappledger/AnnChain/angine/state"
	"github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/ann-module/lib/ed25519"
	cmn "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	cfg "github.com/dappledger/AnnChain/ann-module/lib/go-config"
	crypto "github.com/dappledger/AnnChain/ann-module/lib/go-crypto"
	dbm "github.com/dappledger/AnnChain/ann-module/lib/go-db"
	"github.com/dappledger/AnnChain/ann-module/lib/go-events"
	p2p "github.com/dappledger/AnnChain/ann-module/lib/go-p2p"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
)

const version = "0.6.0"

type (
	// Angine is a high level abstraction of all the state, consensus, mempool blah blah...
	Angine struct {
		Tune *AngineTunes

		mtx     sync.Mutex
		tune    *AngineTunes
		hooked  bool
		started bool

		statedb       dbm.DB
		blockdb       dbm.DB
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

	AngineTunes struct {
		Runtime string
		Conf    *cfg.MapConfig
	}
)

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
		InitAccounts: []types.InitInfo{
			types.InitInfo{
				StartingBalance: "100000000000000",
				Address:         "0x65188459a1dc65984a0c7d4a397ed3986ed0c853", // privatekey:7cb4880c2d4863f88134fd01a250ef6633cc5e01aeba4c862bedbf883a148ba8
			},
		},
	}
	genDoc.Validators = gVals
	return genDoc, genDoc.SaveAs(path)
}

func integrityCheck(conf cfg.Config) error {
	privFile := conf.GetString("priv_validator_file")
	genFile := conf.GetString("genesis_file")
	if !cmn.FileExists(privFile) {
		return fmt.Errorf("PrivValidator file needed: %s", privFile)
	}
	if !cmn.FileExists(genFile) {
		return fmt.Errorf("Genesis file needed: %s", genFile)
	}
	return nil
}

func Initialize(tune *AngineTunes) {
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
func NewAngine(tune *AngineTunes) *Angine {
	var conf *cfg.MapConfig
	if tune.Conf == nil {
		conf = ac.GetConfig(tune.Runtime)
	} else {
		conf = tune.Conf
	}

	if err := integrityCheck(conf); err != nil {
		priv := genPrivFile(conf.GetString("priv_validator_file"))
		gvs := []types.GenesisValidator{types.GenesisValidator{
			PubKey:     priv.PubKey,
			Amount:     100,
			IsCA:       true,
			RPCAddress: conf.GetString("rpc_laddr"),
		}}
		_, err := genGenesiFile(conf.GetString("genesis_file"), gvs)
		if err != nil {
			cmn.PanicSanity(err)
		}
	}

	apphash := []byte{}
	dbBackend := conf.GetString("db_backend")
	dbDir := conf.GetString("db_dir")
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	stateM := state.GetState(conf, stateDB)
	genesis := getGenesisFileMust(conf)
	if stateM == nil {
		if stateM = state.MakeGenesisState(stateDB, genesis); stateM == nil {
			cmn.Exit(cmn.Fmt("Fail to get genesis state"))
		}
	}
	conf.Set("chain_id", stateM.ChainID)

	logpath := conf.GetString("log_path")
	if logpath == "" {
		logpath, _ = os.Getwd()
	}
	logpath = path.Join(logpath, "angine-"+stateM.ChainID)
	cmn.EnsureDir(logpath, 0700)
	logger := InitializeLog(conf.GetString("log_mode"), conf.GetString("environment"), logpath)
	stateM.SetLogger(logger)
	privValidator := types.LoadOrGenPrivValidator(logger, conf.GetString("priv_validator_file"))
	refuseList := refuse_list.NewRefuseList(dbBackend, dbDir)
	eventSwitch := types.NewEventSwitch(logger)
	fastSync := fastSyncable(conf, privValidator.GetAddress(), stateM.Validators)
	if _, err := eventSwitch.Start(); err != nil {
		cmn.PanicSanity(cmn.Fmt("Fail to start event switch: %v", err))
	}

	blockStoreDB := dbm.NewDB("blockstore", dbBackend, dbDir)
	blockStore := blockchain.NewBlockStore(blockStoreDB)
	if block := blockStore.LoadBlock(blockStore.Height()); block != nil {
		apphash = block.AppHash
	}
	_ = apphash // just bypass golint
	_, stateLastHeight, _ := stateM.GetLastBlockInfo()
	bcReactor := blockchain.NewBlockchainReactor(logger, conf, stateLastHeight, blockStore, fastSync)
	mem := mempool.NewMempool(logger, conf)
	for _, p := range stateM.Plugins {
		mem.RegisterFilter(NewMempoolFilter(p.CheckTx))
	}
	memReactor := mempool.NewMempoolReactor(logger, conf, mem)

	consensusState := consensus.NewConsensusState(logger, conf, stateM, blockStore, mem)
	consensusState.SetPrivValidator(privValidator)
	consensusReactor := consensus.NewConsensusReactor(logger, consensusState, fastSync)

	bcReactor.SetBlockVerifier(func(bID types.BlockID, h int, lc *types.Commit) error {
		return stateM.Validators.VerifyCommit(stateM.ChainID, bID, h, lc)
	})
	bcReactor.SetBlockExecuter(func(blk *types.Block, pst *types.PartSet, c *types.Commit) error {
		blockStore.SaveBlock(blk, pst, c)
		if err := stateM.ApplyBlock(eventSwitch, blk, pst.Header(), MockMempool{}, -1); err != nil {
			return err
		}
		stateM.Save()
		return nil
	})

	privKey := privValidator.GetPrivateKey()
	p2psw := p2p.NewSwitch(logger, conf.GetConfig("p2p"))
	p2psw.AddReactor("MEMPOOL", memReactor)
	p2psw.AddReactor("BLOCKCHAIN", bcReactor)
	p2psw.AddReactor("CONSENSUS", consensusReactor)

	if conf.GetBool("pex_reactor") {
		addrBook := p2p.NewAddrBook(logger, conf.GetString("addrbook_file"), conf.GetBool("addrbook_strict"))
		addrBook.Start()
		pexReactor := p2p.NewPEXReactor(logger, addrBook)
		p2psw.AddReactor("PEX", pexReactor)
	}

	p2psw.SetNodePrivKey(privKey.(crypto.PrivKeyEd25519))
	//p2psw.SetAuthByCA(authByCA(stateM.ChainID, &stateM.Validators, logger))
	if conf.GetBool("auth_by_ca") {
		p2psw.SetAuthByCA_Local(authByCA_Local(conf, stateM.ChainID, &stateM.Validators, logger))
	}
	p2psw.SetAddToRefuselist(addToRefuselist(refuseList))
	p2psw.SetRefuseListFilter(refuseListFilter(refuseList))

	protocol, address := ProtocolAndAddress(conf.GetString("node_laddr"))
	defaultListener := p2p.NewDefaultListener(logger, protocol, address, conf.GetBool("skip_upnp"))
	p2psw.AddListener(defaultListener)
	angineNodeInfo := &p2p.NodeInfo{
		PubKey:      privKey.PubKey().(crypto.PubKeyEd25519),
		SigndPubKey: conf.GetString("signbyCA"),
		Moniker:     conf.GetString("moniker"),
		ListenAddr:  defaultListener.ExternalAddress().String(),
		Version:     version,
	}
	p2psw.SetNodeInfo(angineNodeInfo)

	setEventSwitch(eventSwitch, bcReactor, memReactor, consensusReactor)
	initCorePlugins(stateM, privKey.(crypto.PrivKeyEd25519), p2psw, &stateM.Validators, refuseList)

	if tune.Conf == nil {
		tune.Conf = conf
	} else if tune.Runtime == "" {
		tune.Runtime = conf.GetString("datadir")
	}

	return &Angine{
		Tune: tune,

		statedb:       stateDB,
		blockdb:       blockStoreDB,
		tune:          tune,
		stateMachine:  stateM,
		p2pSwitch:     p2psw,
		eventSwitch:   &eventSwitch,
		refuseList:    refuseList,
		privValidator: privValidator,
		blockstore:    blockStore,
		mempool:       mem,
		consensus:     consensusState,
		p2pHost:       defaultListener.ExternalAddress().IP.String(),
		p2pPort:       defaultListener.ExternalAddress().Port,
		genesis:       genesis,

		logger: logger,
	}
}

func (e *Angine) SetSpecialVoteRPC(f func([]byte, *types.Validator) ([]byte, error)) {
	e.getSpecialVote = f
}

func (e *Angine) ConnectApp(app types.Application) {
	e.hooked = true
	hooks := app.GetAngineHooks()
	if hooks.OnExecute == nil || hooks.OnCommit == nil {
		cmn.PanicSanity("At least implement OnExecute & OnCommit, otherwise what your application is for")
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

	info := app.Info()
	if err := e.RecoverFromCrash(info.LastBlockAppHash, int(info.LastBlockHeight)); err != nil {
		cmn.PanicSanity("replay blocks on angine start failed")
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

func (e *Angine) Start() error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	if e.started {
		return errors.New("can't start angine twice")
	}
	if !e.hooked {
		e.hookDefaults()
	}
	if _, err := e.p2pSwitch.Start(); err == nil {
		e.started = true
	} else {
		return err
	}

	seeds := e.tune.Conf.GetString("seeds")
	if seeds != "" {
		e.DialSeeds(strings.Split(seeds, ","))
	}

	return nil
}

// Stop just wrap around swtich.Stop, which will stop reactors, listeners,etc
func (e *Angine) Stop() bool {
	e.refuseList.Stop()
	e.statedb.Close()
	e.blockdb.Close()
	return e.p2pSwitch.Stop()
}

func (e *Angine) RegisterNodeInfo(ni *p2p.NodeInfo) {
	e.p2pSwitch.SetNodeInfo(ni)
}

func (e *Angine) GetNodeInfo() *p2p.NodeInfo {
	return e.p2pSwitch.NodeInfo()
}

func (e *Angine) Height() int {
	return e.blockstore.Height()
}

func (e *Angine) GetBlock(height int) (*types.Block, *types.BlockMeta) {
	if height == 0 {
		return nil, nil
	}
	return e.blockstore.LoadBlock(height), e.blockstore.LoadBlockMeta(height)
}

func (e *Angine) BroadcastTx(tx []byte) error {
	return e.mempool.CheckTx(tx)
}

func (e *Angine) BroadcastTxCommit(tx []byte) (*types.ResultBroadcastTxCommit, error) {
	if err := e.mempool.CheckTx(tx); err != nil {
		return nil, err
	}
	committed := make(chan types.EventDataTx, 1)
	eventString := types.EventStringTx(tx)
	timer := time.NewTimer(60 * 2 * time.Second)
	types.AddListenerForEvent(*e.eventSwitch, "angine", eventString, func(data types.TMEventData) {
		committed <- data.(types.EventDataTx)
	})
	defer func() {
		(*e.eventSwitch).(events.EventSwitch).RemoveListenerForEvent(eventString, "angine")
	}()
	select {
	case c := <-committed:
		if c.Code == types.CodeType_OK {
			return &types.ResultBroadcastTxCommit{
				Code: c.Code,
				Data: c.Data,
				Log:  c.Log,
			}, nil
		}
		res := new(types.Result).FromJSON(c.Error)
		return &types.ResultBroadcastTxCommit{
			Code: res.Code,
			Data: res.Data,
			Log:  res.Log,
		}, nil
	case <-timer.C:
		// 如果交易超时了，需要将交易哈希和交易对象返回给上层
		return &types.ResultBroadcastTxCommit{
			Code: types.CodeType_Timeout,
			Log:  "Timed out waiting for transaction to be included in a block",
		}, nil
		// return nil, fmt.Errorf("Timed out waiting for transaction to be included in a block")
	}
}

func (e *Angine) FlushMempool() {
	e.mempool.Flush()
}

func (e *Angine) GetValidators() (int, []*types.Validator) {
	return e.stateMachine.LastBlockHeight, e.stateMachine.Validators.Validators
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
	roundState := e.consensus.GetRoundState()
	peerRoundStates := make([]string, 0, e.p2pSwitch.Peers().Size())
	for _, p := range e.p2pSwitch.Peers().List() {
		peerState := p.Data.Get(types.PeerStateKey).(*consensus.PeerState)
		peerRoundState := peerState.GetRoundState()
		peerRoundStateStr := p.Key + ":" + string(wire.JSONBytes(peerRoundState))
		peerRoundStates = append(peerRoundStates, peerRoundStateStr)
	}
	return roundState.String(), peerRoundStates
}

func (e *Angine) GetNumUnconfirmedTxs() int {
	return e.mempool.Size()
}

func (e *Angine) GetUnconfirmedTxs() []types.Tx {
	return e.mempool.Reap(-1)
}

func (e *Angine) IsNodeValidator(pub crypto.PubKey) bool {
	edPub := pub.(crypto.PubKeyEd25519)
	_, vals := e.consensus.GetValidators()
	for _, v := range vals {
		if edPub.KeyString() == v.PubKey.KeyString() {
			return true
		}
	}
	return false
}

func (e *Angine) GetBlacklist() []string {
	return e.refuseList.ListAllKey()
}

// Recover world status
// Replay all blocks after blockHeight and ensure the result matches the current state.
func (e *Angine) RecoverFromCrash(appHash []byte, appBlockHeight int) error {
	storeBlockHeight := e.blockstore.Height()
	stateBlockHeight := e.stateMachine.LastBlockHeight

	if storeBlockHeight == 0 {
		return nil // no blocks to replay
	}

	e.logger.Info("Replay Blocks", zap.Int("appHeight", appBlockHeight), zap.Int("storeHeight", storeBlockHeight), zap.Int("stateHeight", stateBlockHeight))

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
			e.logger.Debug("RelpayBlocks: Already synced")
		} else if bytes.Equal(stateAppHash, lastBlockAppHash) {
			// we crashed after commit and before saving state,
			// so load the intermediate state and update the hash
			e.stateMachine.LoadIntermediate()
			e.stateMachine.AppHash = appHash
			e.logger.Debug("RelpayBlocks: Loaded intermediate state and updated state.AppHash")
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

func authByCA_Local(conf *cfg.MapConfig, chainID string, ppValidators **types.ValidatorSet, log *zap.Logger) func(*p2p.NodeInfo) error {
	valset := *ppValidators
	return func(peerNodeInfo *p2p.NodeInfo) error {
		if !valset.HasAddress(peerNodeInfo.PubKey.Address()) && !conf.GetBool("non_validator_node_auth") {
			return nil
		}
		msg := peerNodeInfo.PubKey[:]
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

//func authByCA(conf *cfg.MapConfig, chainID string, ppValidators **types.ValidatorSet, log *zap.Logger) func(*p2p.NodeInfo) error {
//	valset := *ppValidators
//	chainIDBytes := []byte(chainID)
//	return func(peerNodeInfo *p2p.NodeInfo) error {
//		//validator node must be signed by CA but normal node  can bypass auth check if config says to
//		if !valset.HasAddress(peerNodeInfo.PubKey.Address()) && !conf.GetBool("non_validator_node_auth") {
//			return nil
//		}
//		msg := append(peerNodeInfo.PubKey[:], chainIDBytes...)
//		for _, val := range valset.Validators {
//			//only CA
//			if !val.IsCA {
//				continue
//			}
//			return nil
//			valPk := [32]byte(val.PubKey.(crypto.PubKeyEd25519))
//			fmt.Println("pyb:", valPk)
//			signedPkByte64, err := types.StringTo64byte(peerNodeInfo.SigndPubKey)
//			if err != nil {
//				return err
//			}
//			if ed25519.Verify(&valPk, msg, &signedPkByte64) {
//				log.Sugar().Infof("Peer handshake", "peerNodeInfo", peerNodeInfo)
//				return nil
//			}
//		}
//		err := fmt.Errorf("Reject Peer,has no CA sig")
//		log.Warn(err.Error())
//		return err
//	}
//}

//func authByCA(chainID string, ppValidators **types.ValidatorSet, log *zap.Logger) func(*p2p.NodeInfo) error {
//	//valset := *ppValidators
//	return func(peerNodeInfo *p2p.NodeInfo) error {
//		{
//			log.Sugar().Infow("Peer handshake without CA check", "peerNodeInfo", peerNodeInfo)
//			return nil
//		}

//		//		msg := peerNodeInfo.PubKey[:]
//		//		for _, val := range valset.Validators {
//		//			if !val.IsCA {
//		//				continue // CA must be validator
//		//			}
//		//			valPk := [32]byte(val.PubKey.(crypto.PubKeyEd25519))
//		//			signedPkByte64, err := types.StringTo64byte(peerNodeInfo.SigndPubKey)
//		//			if err != nil {
//		//				return err
//		//			}
//		//			if ed25519.Verify(&valPk, msg, &signedPkByte64) {
//		//				log.Sugar().Infow("Peer handshake", "peerNodeInfo", peerNodeInfo)
//		//				return nil
//		//			}
//		//		}
//		//		err := fmt.Errorf("Reject Peer, has no CA sig")
//		//		log.Warn(err.Error())
//		//		return err
//	}
//}

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

func getGenesisFileMust(conf cfg.Config) *types.GenesisDoc {
	genDocFile := conf.GetString("genesis_file")
	if !cmn.FileExists(genDocFile) {
		cmn.PanicSanity("missing genesis_file")
	}
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		cmn.Exit(cmn.Fmt("Couldn't read GenesisDoc file: %v", err))
	}
	genDoc := types.GenesisDocFromJSON(jsonBlob)
	if genDoc.ChainID == "" {
		cmn.PanicSanity(cmn.Fmt("Genesis doc %v must include non-empty chain_id", genDocFile))
	}
	conf.Set("chain_id", genDoc.ChainID)

	return genDoc
}

// Defaults to tcp
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
type MempoolFilter struct {
	cb func([]byte) (bool, error)
}

func (m MempoolFilter) CheckTx(tx types.Tx) (bool, error) {
	return m.cb(tx)
}
func NewMempoolFilter(f func([]byte) (bool, error)) MempoolFilter {
	return MempoolFilter{cb: f}
}
