/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package node

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/db"
	"go.uber.org/zap"

	ac "github.com/dappledger/AnnChain/angine/config"
	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	cmn "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/lib/go-merkle"
	"github.com/dappledger/AnnChain/module/xlib"
	"github.com/dappledger/AnnChain/module/xlib/def"
	acfg "github.com/dappledger/AnnChain/src/chain/config"
	"github.com/dappledger/AnnChain/src/chain/log"
)

const (
	BASE_APP_NAME = "metropolis"
)

type (
	// AppName just wraps around string, maybe we can use the new feat in Go1.9
	AppName string

	// ChainID just wraps around string
	ChainID string

	// MetropolisState abstracts the structure of the information
	// that we need to make the application consistency verifiable
	MetropolisState struct {
		OrgStateHash   []byte
		OrgHeights     []def.INT
		EventStateHash []byte
	}

	// Metropolis defines the application
	Metropolis struct {
		agtypes.BaseApplication

		mtx  sync.Mutex
		node *Node

		// core abstracts the orgnode within application
		core Core

		config      *viper.Viper
		angineHooks agtypes.Hooks
		logger      *zap.Logger

		// top level app state
		state *MetropolisState

		// Organization
		OrgStateDB db.DB
		// OrgState is a DL that keeps tracks of subchains' state
		OrgState *OrgState
		OrgApps  map[ChainID]AppName
		// Orgs contains all the subchains this node is effectively in
		Orgs map[ChainID]*OrgNode
		// TODO: persist onto disk
		PendingOrgTxs map[string]*OrgTx

		// Events
		EventStateDB   db.DB
		EventState     *EventState
		EventWarehouse *EventWarehouse
		EventCodeBase  db.DB
		// TODO: persist onto disk
		PendingEventRequestTxs map[string]*EventRequestTx

		// TODO: add economy stimulus
		// Economy
		// ecoStateDB db.DB
		// ecoState   *EcoState

	}

	// LastBlockInfo is just a must for every angine-based application
	LastBlockInfo struct {
		Height def.INT `msgpack:"height"`
		// hash from the top level state
		Hash []byte `msgpack:"hash"`
	}
)

var (
	// ErrUnknownTx is exported because I am not sure if it will be needed outside of this pkg
	ErrUnknownTx = fmt.Errorf("please give me something that I actually know about")
)

// NewLastBlockInfo just a convience to generate an empty LastBlockInfo
func NewLastBlockInfo() *LastBlockInfo {
	return &LastBlockInfo{
		Height: 0,
		Hash:   make([]byte, 0),
	}
}

// NewMetropolis initialize all the necessary parts of the application:
// 1. state of metropolis
// 2. init BaseApplication
// 3. open databases and generate orgstate, eventstate and so on
// 4. set up the angine hooks
func NewMetropolis(logger *zap.Logger, conf *viper.Viper) *Metropolis {
	datadir := conf.GetString("db_dir")
	met := Metropolis{
		config: conf,
		logger: logger,

		state: &MetropolisState{
			OrgHeights:     make([]def.INT, 0),
			OrgStateHash:   make([]byte, 0),
			EventStateHash: make([]byte, 0),
		},

		OrgApps: make(map[ChainID]AppName),
		Orgs:    make(map[ChainID]*OrgNode),

		PendingOrgTxs:          make(map[string]*OrgTx),
		PendingEventRequestTxs: make(map[string]*EventRequestTx),
	}

	var err error
	if err = met.BaseApplication.InitBaseApplication(BASE_APP_NAME, datadir); err != nil {
		cmn.PanicCrisis(err)
	}
	if met.OrgStateDB, err = db.NewGoLevelDB("orgStatus", datadir); err != nil {
		cmn.PanicCrisis(err)
	}
	met.OrgState = NewOrgState(met.OrgStateDB)
	if met.EventStateDB, err = db.NewGoLevelDB("eventstate", datadir); err != nil {
		cmn.PanicCrisis(err)
	}
	met.EventState = NewEventState(met.EventStateDB)
	ewhDB, err := db.NewGoLevelDB("eventwarehouse", datadir)
	if err != nil {
		cmn.PanicCrisis(err)
	}
	met.EventWarehouse = NewEventWarehouse(ewhDB)

	met.EventCodeBase, err = db.NewGoLevelDB("eventcodebase", datadir)
	if err != nil {
		cmn.PanicCrisis(err)
	}

	met.angineHooks = agtypes.Hooks{
		OnExecute: agtypes.NewHook(met.OnExecute),
		OnCommit:  agtypes.NewHook(met.OnCommit),
	}

	return &met
}

// Hash is merely a wrapper of MetropolisState.Hash
func (met *Metropolis) Hash() []byte {
	return met.state.Hash()
}

// Lock locks Metropolis universally
func (met *Metropolis) Lock() {
	met.mtx.Lock()
}

// Unlock undos what Lock does
func (met *Metropolis) Unlock() {
	met.mtx.Unlock()
}

// SetNode
func (met *Metropolis) SetNode(n *Node) {
	met.node = n
}

// SetCore gives Metropolis a way to make some calls to the node running underneath
// it is more abstractive than SetNode, which I am planning to deprecate.
func (met *Metropolis) SetCore(c Core) {
	met.core = c
}

// BroadcastTx just passes the transaction into core
func (met *Metropolis) BroadcastTx(tx []byte) error {
	return met.core.GetEngine().BroadcastTx(tx)
}

// GetNodePubKey gets our universal public key
func (met *Metropolis) GetNodePubKey() crypto.PubKey {
	return met.core.GetEngine().PrivValidator().GetPubKey()
}

// Stop stops all still running organization
func (met *Metropolis) Stop() {
	for i := range met.Orgs {
		met.Orgs[i].Stop()
	}
}

// Load gets all those ledgers have been persisted back alive by a metropolisstate hash
func (met *Metropolis) Load(lb *LastBlockInfo) error {
	met.Lock()
	defer met.Unlock()

	state := &MetropolisState{}
	if lb == nil {
		return nil
	}

	bs := met.Database.Get(lb.Hash)
	if err := state.FromBytes(bs); err != nil {
		return errors.Wrap(err, "fail to restore metropolis state")
	}
	met.state = state
	met.OrgState.Load(met.state.OrgStateHash)
	met.EventState.Load(met.state.EventStateHash)
	return nil
}

// Start will restore organization according to orgtx history
func (met *Metropolis) Start() error {
	if err := met.spawnOffchainEventListener(); err != nil {
		met.logger.Error("fail to start event server", zap.Error(err))
		return err
	}

	lastBlock := NewLastBlockInfo()
	if res, err := met.LoadLastBlock(lastBlock); err == nil {
		lastBlock = res.(*LastBlockInfo)
	}
	if lastBlock.Hash == nil || len(lastBlock.Hash) == 0 {
		return nil
	}
	if err := met.Load(lastBlock); err != nil {
		met.logger.Error("fail to load metropolis state", zap.Error(err))
		return err
	}

	for _, height := range met.state.OrgHeights {
		block, _, err := met.core.GetEngine().GetBlock(def.INT(height))
		if err != nil {
			return err
		}
		for _, tx := range block.Data.Txs {
			txBytes := agtypes.UnwrapTx(tx)
			switch {
			case IsOrgCancelTx(tx):
				met.restoreOrgCancel(txBytes)
			case IsOrgConfirmTx(tx):
				met.restoreOrgConfirm(txBytes)
			case IsOrgTx(tx):
				met.restoreOrg(txBytes)
			}
		}
	}

	return nil
}

// GetAngineHooks returns the hooks we defined for angine to grasp
func (met *Metropolis) GetAngineHooks() agtypes.Hooks {
	return met.angineHooks
}

func (met *Metropolis) GetAttributes() map[string]string {
	return nil
}

// CompatibleWithAngine just exists to satisfy types.Application
func (met *Metropolis) CompatibleWithAngine() {}

// OnCommit persists state that we define to be consistent in a cross-block way
func (met *Metropolis) OnCommit(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	var err error

	if met.state.OrgStateHash, err = met.OrgState.Commit(); err != nil {
		met.logger.Error("fail to commit orgState", zap.Error(err))
		return nil, err
	}
	if met.state.EventStateHash, err = met.EventState.Commit(); err != nil {
		met.logger.Error("fail to commit eventState", zap.Error(err))
		return nil, err
	}
	met.state.OrgHeights = sterilizeOrgHeights(met.state.OrgHeights)

	lastBlock := LastBlockInfo{Height: height, Hash: met.Hash()}
	met.Database.SetSync(lastBlock.Hash, met.state.ToBytes())
	met.SaveLastBlock(lastBlock)

	defer func() {
		met.OrgState.Reload(met.state.OrgStateHash)
		met.EventState.Reload(met.state.EventStateHash)
	}()

	return agtypes.CommitResult{AppHash: lastBlock.Hash}, nil
}

// IsTxKnown is a fast way to identify txs unknown
func (met *Metropolis) IsTxKnown(bs []byte) bool {
	return IsOrgRelatedTx(bs)
}

// CheckTx returns nil if sees an unknown tx

// Info gives information about the application in general
func (met *Metropolis) Info() (resInfo agtypes.ResultInfo) {
	lb := NewLastBlockInfo()
	if res, err := met.LoadLastBlock(lb); err == nil {
		lb = res.(*LastBlockInfo)
	}

	resInfo.LastBlockAppHash = lb.Hash
	resInfo.LastBlockHeight = lb.Height
	resInfo.Version = "alpha 0.2"
	resInfo.Data = "metropolis with organizations and inter-organization communication"
	return
}

// SetOption can dynamicly change some options of the application
func (met *Metropolis) SetOption() {}

// Query now gives the ability to query transaction execution result with the transaction hash
func (met *Metropolis) Query(query []byte) agtypes.Result {
	if len(query) == 0 {
		return agtypes.NewResultOK([]byte{}, "Empty query")
	}
	var res agtypes.Result

	switch query[0] {
	case agtypes.QueryTxExecution:
		qryRes, err := met.core.GetEngine().Query(query[0], query[1:])
		if err != nil {
			return agtypes.NewError(pbtypes.CodeType_InternalError, err.Error())
		}
		info, ok := qryRes.(*agtypes.TxExecutionResult)
		if !ok {
			return agtypes.NewError(pbtypes.CodeType_InternalError, err.Error())
		}
		res.Code = pbtypes.CodeType_OK
		res.Data, _ = info.ToBytes()
		return res
	case QueryEvents:
		keys := make([]string, 0)
		for k := range met.EventWarehouse.state.state {
			keys = append(keys, k)
		}

		buffers := &bytes.Buffer{}
		encoder := gob.NewEncoder(buffers)
		if err := encoder.Encode(keys); err != nil {
			res.Code = pbtypes.CodeType_InternalError
			res.Log = err.Error()
		} else {
			res.Code = pbtypes.CodeType_OK
			res.Data = buffers.Bytes()
		}

		return res
	}

	return res
}

// SetOrg wraps some atomic operations that need to be done when this node create/join an organization
func (met *Metropolis) SetOrg(chainID, appname string, o *OrgNode) {
	met.mtx.Lock()
	met.Orgs[ChainID(chainID)] = o
	met.OrgApps[ChainID(chainID)] = AppName(appname)
	met.mtx.Unlock()
}

// GetOrg gets the orgnode associated to the chainid, or gives an error if we are not connected to the chain
func (met *Metropolis) GetOrg(id string) (*OrgNode, error) {
	met.mtx.Lock()
	defer met.mtx.Unlock()
	if o, ok := met.Orgs[ChainID(id)]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("no such org: %s", id)
}

// RemoveOrg wraps atomic operations needed to be done when node gets removed from an organization
func (met *Metropolis) RemoveOrg(id string) error {
	met.mtx.Lock()
	orgNode := met.Orgs[ChainID(id)]
	orgNode.Stop()
	if orgNode.IsRunning() && !orgNode.Stop() {
		met.mtx.Unlock()
		return fmt.Errorf("node is still running, error cowordly")
	}
	delete(met.Orgs, ChainID(id))
	delete(met.OrgApps, ChainID(id))
	met.mtx.Unlock()
	return nil
}

// createOrgNode start a new org node in goroutine, so the app could use runtime.Goexit when error occurs
func (met *Metropolis) createOrgNode(tx *OrgTx) (*OrgNode, error) {
	if !AppExists(tx.App) {
		return nil, fmt.Errorf("no such app: %s", tx.App)
	}
	runtime := ac.RuntimeDir(met.config.GetString("runtime"))
	conf, err := acfg.LoadDefaultAngineConfig(runtime, tx.ChainID, tx.Config)
	if err != nil {
		return nil, err
	}

	if !cmn.FileExists(conf.GetString("priv_validator_file")) {
		mainPV := met.core.GetEngine().PrivValidator()
		privKey := mainPV.PrivKey
		pubKey := mainPV.PubKey
		// addr := mainPV.Address
		privVal := &agtypes.PrivValidator{
			// Address: addr,
			PrivKey: privKey,
			PubKey:  pubKey,
			Signer:  agtypes.NewDefaultSigner(privKey.PrivKey),
		}
		privVal.SetFile(conf.GetString("priv_validator_file"))
		privVal.Save()
	}

	if tx.Genesis.ChainID != "" && !cmn.FileExists(conf.GetString("genesis_file")) {
		if err := tx.Genesis.SaveAs(conf.GetString("genesis_file")); err != nil {
			met.logger.Error("fail to save org's genesis", zap.String("chainid", tx.ChainID), zap.String("path", conf.GetString("genesis_file")))
			return nil, fmt.Errorf("fail to save org's genesis")
		}
	}

	resChan := make(chan *OrgNode, 1)
	go func(c chan<- *OrgNode) {
		applog := log.Initialize(met.config.GetString("environment"), path.Join(conf.GetString("log_path"), tx.ChainID, "output.log"), path.Join(conf.GetString("log_path"), tx.ChainID, "err.log"))
		n := NewOrgNode(applog, tx.App, conf, met)
		if n == nil {
			met.logger.Error("startOrgNode failed", zap.Error(err))
			c <- nil
			return
		}
		if err := n.Start(); err != nil {
			met.logger.Error("startOrgNode failed", zap.Error(err))
			c <- nil
			return
		}
		c <- n
	}(resChan)

	var node *OrgNode
	select {
	case node = <-resChan:
	case <-time.After(10 * time.Second):
	}

	if node != nil {
		fmt.Printf("organization %s is running\n", node.GetChainID())
		return node, nil
	}
	return nil, fmt.Errorf("fail to start org node, check the log")
}

func (ms *MetropolisState) Hash() []byte {
	return merkle.SimpleHashFromBinary(ms)
}

func (ms *MetropolisState) ToBytes() []byte {
	var bs []byte
	buf := bytes.NewBuffer(bs)
	gec := gob.NewEncoder(buf)
	if err := gec.Encode(ms); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (ms *MetropolisState) FromBytes(bs []byte) error {
	bsReader := bytes.NewReader(bs)
	gdc := gob.NewDecoder(bsReader)
	return gdc.Decode(ms)
}

// remove duplicates and sort
func sterilizeOrgHeights(orgH []def.INT) []def.INT {
	hm := make(map[def.INT]struct{})
	for _, h := range orgH {
		hm[h] = struct{}{}
	}
	uniqueHeights := make([]def.INT, 0)
	for k := range hm {
		uniqueHeights = append(uniqueHeights, def.INT(k))
	}
	xlib.Int64Slice(uniqueHeights).Sort()
	return uniqueHeights
}
