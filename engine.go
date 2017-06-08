package angine

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

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
	p2p "gitlab.zhonganonline.com/ann/ann-module/lib/go-p2p"
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
	}

	EngineTunes struct {
		Conf cfg.Config
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
	refuseList := refuse_list.NewRefuseList(dbBackend, dbDir)
	eventSwitch := types.NewEventSwitch()
	fastSync := fastSyncable(tune.Conf, driver.GetAddress(), stateM.Validators)
	if _, err := eventSwitch.Start(); err != nil {
		cmn.PanicSanity(cmn.Fmt("Failed to start switch: %v", err))
	}

	tune.Conf.Set("chain_id", stateM.ChainID)

	blockStoreDB := dbm.NewDB("blockstore", dbBackend, dbDir)
	blockStore := blockchain.NewBlockStore(blockStoreDB)
	if block := blockStore.LoadBlock(blockStore.Height()); block != nil {
		apphash = block.AppHash
	}
	_ = apphash // just bypass golint
	_, stateLastHeight, _ := stateM.GetLastBlockInfo()
	bcReactor := blockchain.NewBlockchainReactor(tune.Conf, stateLastHeight, blockStore, fastSync)
	bcReactor.SetBlockVerifier(func(bID types.BlockID, h int, lc *types.Commit) error {
		return stateM.Validators.VerifyCommit(stateM.ChainID, bID, h, lc)
	})
	bcReactor.SetBlockExecuter(func(blk *types.Block, pst *types.PartSet, c *types.Commit) error {
		blockStore.SaveBlock(blk, pst, c)
		// TODO: should we be firing events? need to fire NewBlock events manually ...
		// NOTE: we could improve performance if we
		// didn't make the app commit to disk every block
		// ... but we would need a way to get the hash without it persisting
		if err := stateM.ApplyBlock(eventSwitch, blk, pst.Header(), MockMempool{}); err != nil {
			return err
		}
		stateM.Save()
		return nil
	})

	mem := mempool.NewMempool(tune.Conf)
	for _, p := range stateM.Plugins {
		mem.RegisterFilter(NewMempoolFilter(p.CheckTx))
	}
	memReactor := mempool.NewMempoolReactor(tune.Conf, mem)

	consensusState := consensus.NewConsensusState(tune.Conf, stateM, blockStore, mem)
	consensusState.SetPrivValidator(driver)
	consensusReactor := consensus.NewConsensusReactor(consensusState, fastSync)

	// Make p2p network switch
	sw := p2p.NewSwitch(tune.Conf.GetConfig("p2p"))
	sw.AddReactor("MEMPOOL", memReactor)
	sw.AddReactor("BLOCKCHAIN", bcReactor)
	sw.AddReactor("CONSENSUS", consensusReactor)

	// TODO: just ignore the pex for a short while
	// Optionally, start the pex reactor
	// TODO: this is a dev feature, it needs some love
	// if config.GetBool("pex_reactor") {
	//	addrBook := p2p.NewAddrBook(config.GetString("addrbook_file"), config.GetBool("addrbook_strict"))
	//	addrBook.Start()
	//	pexReactor := p2p.NewPEXReactor(addrBook)
	//	sw.AddReactor("PEX", pexReactor)
	// }
	privKey := driver.GetPrivateKey()
	sw.SetNodePrivKey(privKey.(crypto.PrivKeyEd25519))
	sw.SetAuthByCA(authByCA(&stateM.Validators))
	sw.SetAddToRefuselist(addToRefuselist(refuseList))
	sw.SetRefuseListFilter(refuseListFilter(refuseList))
	setEventSwitch(eventSwitch, bcReactor, memReactor, consensusReactor)

	initCorePlugins(stateM, privKey.(crypto.PrivKeyEd25519), sw, &stateM.Validators, refuseList)

	return &Engine{
		tune:         tune,
		stateMachine: stateM,
		p2pSwitch:    sw,
		eventSwitch:  &eventSwitch,
		refuseList:   refuseList,
		driver:       driver,
		blockstore:   blockStore,
		mempool:      mem,
		consensus:    consensusState,
	}
}

func (e *Engine) ConnectHooks(hooks types.Hooks) {
	e.hooked = true

	if hooks.OnNewRound != nil {
		types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookNewRound(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookNewRound)
			hooks.OnNewRound.Async(data.Height, data.Round, nil)
		})
	}
	if hooks.OnPropose != nil {
		types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookPropose(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPropose)
			hooks.OnPropose.Async(data.Height, data.Round, nil)
		})
	}
	if hooks.OnPrevote != nil {
		types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookPrevote(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPrevote)
			hooks.OnPrevote.Async(data.Height, data.Round, data.Block)
		})
	}
	if hooks.OnPrecommit != nil {
		types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookPrecommit(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookPrecommit)
			hooks.OnPrecommit.Async(data.Height, data.Round, data.Block)
		})
	}

	types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookCommit(), func(ed types.TMEventData) {
		data := ed.(types.EventDataHookCommit)
		cs := types.CommitResult{}
		if hooks.OnCommit != nil {
			cs = hooks.OnCommit.Sync(data.Height, data.Round, data.Block).(CommitResult)
		}
		data.ResCh <- cs
	})
}

func (e *Engine) DialSeeds(seeds []string) {
	e.p2pSwitch.DialSeeds(seeds)
}

// AddListener abstract a role for a listener
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
		types.AddListenerForEvent(*e.eventSwitch, "engine", types.EventStringHookCommit(), func(ed types.TMEventData) {
			data := ed.(types.EventDataHookCommit)
			data.ResCh <- types.CommitResult{}
		})
	}
	if _, err := e.p2pSwitch.Start(); err == nil {
		e.started = true
	} else {
		return err
	}
	return nil
}

// Stop just wrap around swtich.Stop, which will stop reactors, listeners,etc
func (e *Engine) Stop() {
	e.p2pSwitch.Stop()
}

func (e *Engine) RegisterNodeInfo(ni *p2p.NodeInfo) {
	e.nodeInfo = ni
	e.p2pSwitch.SetNodeInfo(ni)
}

// Replay for world status
// Replay all blocks after blockHeight and ensure the result matches the current state.
func (e *Engine) ReplayBlocks(appHash []byte, appBlockHeight int) error {
	storeBlockHeight := e.blockstore.Height()
	stateBlockHeight := e.stateMachine.LastBlockHeight
	log.Notice("Replay Blocks", "appHeight", appBlockHeight, "storeHeight", storeBlockHeight, "stateHeight", stateBlockHeight)

	if storeBlockHeight == 0 {
		return nil
	} else if storeBlockHeight < appBlockHeight {
		// if the app is ahead, there's nothing we can do
		return state.ErrAppBlockHeightTooHigh{storeBlockHeight, appBlockHeight}

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
			return state.ErrLastStateMismatch{storeBlockHeight, block.Header.AppHash, appHash}
		}

		blockMeta := e.blockstore.LoadBlockMeta(storeBlockHeight)
		// h.nBlocks++
		var eventCache types.Fireable

		// replay the latest block
		return e.stateMachine.ApplyBlock(eventCache, block, blockMeta.PartsHeader, MockMempool{})
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

		// TODO: use stateBlockHeight instead and let the consensus state
		// do the replay
		// var appHash []byte
		// for i := appBlockHeight + 1; i <= storeBlockHeight; i++ {
		// h.nBlocks++
		// block := e.blockstore.LoadBlock(i)
		// Commit block, get hash back
		// appHash = res.Data
		// }
		// if !bytes.Equal(e.stateMachine.AppHash, appHash) {
		// 	return errors.New(fmt.Sprintf("Ann state.AppHash does not match AppHash after replay. Got %X, expected %X", appHash, e.stateMachine.AppHash))
		// }
		return nil
	}
	return nil
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

func authByCA(ppValidators **types.ValidatorSet) func(*p2p.NodeInfo) error {
	valset := *ppValidators
	return func(peerNodeInfo *p2p.NodeInfo) error {
		for _, val := range valset.Validators {
			if !val.IsCA {
				continue // CA must be validator
			}
			valPk := [32]byte(val.PubKey.(crypto.PubKeyEd25519))
			signedPkByte64, err := types.StringTo64byte(peerNodeInfo.SigndPubKey)
			if err != nil {
				return err
			}
			if ed25519.Verify(&valPk, peerNodeInfo.PubKey[:], &signedPkByte64) {
				log.Info("Peer handshake", "peerNodeInfo", peerNodeInfo)
				return nil
			}
		}
		err := fmt.Errorf("Reject Peer , has no CA sig")
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

func (m MockMempool) Lock()                             {}
func (m MockMempool) Unlock()                           {}
func (m MockMempool) Update(height int, txs []types.Tx) {}

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
