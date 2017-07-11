package angine

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"sync"
	"time"

	"go.uber.org/zap"

	"gitlab.zhonganonline.com/ann/angine/blockchain"
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

type (
	// Engine is a high level abstraction of all the state, consensus, mempool blah blah...
	Engine struct {
		mtx     sync.Mutex
		tune    *EngineTunes
		hooked  bool
		started bool

		driver       IKey
		nodeInfo     *p2p.NodeInfo
		blockstore   *blockchain.BlockStore
		mempool      *mempool.Mempool
		consensus    *consensus.ConsensusState
		stateMachine *state.State
		p2pSwitch    *p2p.Switch
		eventSwitch  *types.EventSwitch
		refuseList   *refuse_list.RefuseList

		logger *zap.Logger
	}

	EngineTunes struct {
		Conf        cfg.Config
		Genesis     *types.GenesisDoc
		Listener    p2p.Listener
		LogPath     string
		Environment string
	}

	IKey interface {
		GetAddress() []byte
		SignVote(chainID string, vote *types.Vote) error
		SignProposal(chainID string, proposal *types.Proposal) error
		GetPrivateKey() crypto.PrivKey
	}
)

// NewEngine makes and returns a new engine, which can be used directly after being imported
func NewEngine(driver IKey, tune *EngineTunes) *Engine {
	apphash := []byte{}
	dbBackend := tune.Conf.GetString("db_backend")
	dbDir := tune.Conf.GetString("db_dir")
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	stateM := state.GetState(tune.Conf, stateDB)
	if stateM == nil {
		if stateM = state.MakeGenesisState(stateDB, tune.Genesis); stateM == nil {
			cmn.Exit(cmn.Fmt("Fail to get genesis state"))
		}
	}
	tune.Conf.Set("chain_id", stateM.ChainID)

	logPath := path.Join(tune.LogPath, "engine-"+stateM.ChainID)
	cmn.EnsureDir(logPath, 0700)
	logger := InitializeLog(tune.Environment, logPath)

	refuseList := refuse_list.NewRefuseList(dbBackend, dbDir)
	eventSwitch := types.NewEventSwitch(logger)
	fastSync := fastSyncable(tune.Conf, driver.GetAddress(), stateM.Validators)
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
	bcReactor := blockchain.NewBlockchainReactor(logger, tune.Conf, stateLastHeight, blockStore, fastSync)
	mem := mempool.NewMempool(logger, tune.Conf)
	for _, p := range stateM.Plugins {
		mem.RegisterFilter(NewMempoolFilter(p.CheckTx))
	}
	memReactor := mempool.NewMempoolReactor(logger, tune.Conf, mem)

	consensusState := consensus.NewConsensusState(logger, tune.Conf, stateM, blockStore, mem)
	consensusState.SetPrivValidator(driver)
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
		fmt.Println(stateM.LastBlockHeight)
		return nil
	})

	privKey := driver.GetPrivateKey()
	p2psw := p2p.NewSwitch(logger, tune.Conf.GetConfig("p2p"))
	p2psw.AddReactor("MEMPOOL", memReactor)
	p2psw.AddReactor("BLOCKCHAIN", bcReactor)
	p2psw.AddReactor("CONSENSUS", consensusReactor)

	if tune.Conf.GetBool("pex_reactor") {
		addrBook := p2p.NewAddrBook(logger, tune.Conf.GetString("addrbook_file"), tune.Conf.GetBool("addrbook_strict"))
		addrBook.Start()
		pexReactor := p2p.NewPEXReactor(logger, addrBook)
		p2psw.AddReactor("PEX", pexReactor)
	}

	p2psw.SetNodePrivKey(privKey.(crypto.PrivKeyEd25519))
	p2psw.SetAuthByCA(authByCA(stateM.ChainID, &stateM.Validators, logger))
	p2psw.SetAddToRefuselist(addToRefuselist(refuseList))
	p2psw.SetRefuseListFilter(refuseListFilter(refuseList))

	if tune.Listener != nil {
		p2psw.AddListener(tune.Listener)
	}

	setEventSwitch(eventSwitch, bcReactor, memReactor, consensusReactor)
	initCorePlugins(stateM, privKey.(crypto.PrivKeyEd25519), p2psw, &stateM.Validators, refuseList)

	return &Engine{
		tune:         tune,
		stateMachine: stateM,
		p2pSwitch:    p2psw,
		eventSwitch:  &eventSwitch,
		refuseList:   refuseList,
		driver:       driver,
		blockstore:   blockStore,
		mempool:      mem,
		consensus:    consensusState,

		logger: logger,
	}
}

func (e *Engine) ConnectApp(app types.Application) {
	e.hooked = true
	hooks := app.GetEngineHooks()
	if hooks.OnExecute == nil || hooks.OnCommit == nil {
		cmn.PanicSanity("At least implement OnExecute & OnCommit, otherwise what your application is for")
	}

	types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookNewRound(), func(ed types.TMEventData) {
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
		types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookPropose(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPropose)
			hooks.OnPropose.Async(data.Height, data.Round, nil, nil, nil)
		})
	}
	if hooks.OnPrevote != nil {
		types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookPrevote(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPrevote)
			hooks.OnPrevote.Async(data.Height, data.Round, data.Block, nil, nil)
		})
	}
	if hooks.OnPrecommit != nil {
		types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookPrecommit(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPrecommit)
			hooks.OnPrecommit.Async(data.Height, data.Round, data.Block, nil, nil)
		})
	}
	types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookExecute(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookExecute)
		hooks.OnExecute.Sync(data.Height, data.Round, data.Block)
		result := hooks.OnExecute.Result()
		if r, ok := result.(types.ExecuteResult); ok {
			data.ResCh <- r
		} else {
			data.ResCh <- types.ExecuteResult{}
		}

	})
	types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookCommit(), func(ed types.TMEventData) {
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
		cmn.PanicSanity("replay blocks on engine start failed")
	}
}

func (e *Engine) DialSeeds(seeds []string) {
	e.p2pSwitch.DialSeeds(seeds)
}

// AddListener abstract a role for a listener
// TODO:: implement role
func (e *Engine) AddListener(role string, l p2p.Listener) {
	e.p2pSwitch.AddListener(l)
}

func (e *Engine) Start() error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	if e.started {
		return errors.New("can't start engine twice")
	}
	if !e.hooked {
		e.hookDefaults()
	}
	if _, err := e.p2pSwitch.Start(); err == nil {
		e.started = true
	} else {
		return err
	}
	return nil
}

// Stop just wrap around swtich.Stop, which will stop reactors, listeners,etc
func (e *Engine) Stop() bool {
	return e.p2pSwitch.Stop()
}

func (e *Engine) RegisterNodeInfo(ni *p2p.NodeInfo) {
	e.nodeInfo = ni
	e.p2pSwitch.SetNodeInfo(ni)
}

func (e *Engine) Height() int {
	return e.blockstore.Height()
}

func (e *Engine) GetBlock(height int) (*types.Block, *types.BlockMeta) {
	if height == 0 {
		return nil, nil
	}
	return e.blockstore.LoadBlock(height), e.blockstore.LoadBlockMeta(height)
}

func (e *Engine) BroadcastTx(tx []byte) error {
	return e.mempool.CheckTx(tx)
}

func (e *Engine) BroadcastTxCommit(tx []byte) error {
	committed := make(chan types.EventDataTx, 1)
	eventString := types.EventStringTx(tx)
	types.AddListenerForEvent(*e.eventSwitch, "engine", eventString, func(data types.TMEventData) {
		committed <- data.(types.EventDataTx)
	})
	defer func() {
		(*e.eventSwitch).(events.EventSwitch).RemoveListenerForEvent(eventString, "engine")
	}()

	if err := e.mempool.CheckTx(tx); err != nil {
		return err
	}
	timer := time.NewTimer(60 * 2 * time.Second)
	select {
	case <-committed:
		return nil
	case <-timer.C:
		return fmt.Errorf("Timed out waiting for transaction to be included in a block")
	}
}

func (e *Engine) FlushMempool() {
	e.mempool.Flush()
}

func (e *Engine) GetValidators() (int, []*types.Validator) {
	return e.stateMachine.LastBlockHeight, e.stateMachine.Validators.Validators
}

func (e *Engine) GetP2PNetInfo() (bool, []string, []*types.Peer) {
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

func (e *Engine) GetNumPeers() int {
	o, i, d := e.p2pSwitch.NumPeers()
	return o + i + d
}

func (e *Engine) GetConsensusStateInfo() (string, []string) {
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

func (e *Engine) GetNumUnconfirmedTxs() int {
	return e.mempool.Size()
}

func (e *Engine) GetUnconfirmedTxs() []types.Tx {
	return e.mempool.Reap(-1)
}

func (e *Engine) IsNodeValidator(pub crypto.PubKey) bool {
	edPub := pub.(crypto.PubKeyEd25519)
	_, vals := e.consensus.GetValidators()
	for _, v := range vals {
		if edPub.KeyString() == v.PubKey.KeyString() {
			return true
		}
	}
	return false
}

func (e *Engine) GetBlacklist() []string {
	return e.refuseList.ListAllKey()
}

// Recover world status
// Replay all blocks after blockHeight and ensure the result matches the current state.
func (e *Engine) RecoverFromCrash(appHash []byte, appBlockHeight int) error {
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

func (e *Engine) hookDefaults() {
	types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookNewRound(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookNewRound)
		data.ResCh <- types.NewRoundResult{}
	})
	types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookExecute(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookExecute)
		data.ResCh <- types.ExecuteResult{}
	})
	types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookCommit(), func(ed types.TMEventData) {
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
