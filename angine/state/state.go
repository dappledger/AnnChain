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

package state

import (
	"bytes"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine/plugin"
	"github.com/dappledger/AnnChain/angine/types"
	. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	cfg "github.com/dappledger/AnnChain/ann-module/lib/go-config"
	dbm "github.com/dappledger/AnnChain/ann-module/lib/go-db"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
)

var (
	stateKey             = []byte("stateKey")
	stateIntermediateKey = []byte("stateIntermediateKey")
)

//-----------------------------------------------------------------------------

// NOTE: not goroutine-safe.
type State struct {
	started bool
	logger  *zap.Logger

	// mtx for writing to db
	mtx sync.Mutex
	db  dbm.DB

	// should not change
	GenesisDoc *types.GenesisDoc
	ChainID    string

	// updated at end of ExecBlock
	LastBlockHeight int // Genesis state has this set to 0.  So, Block(H=0) does not exist.
	LastBlockID     types.BlockID
	LastBlockTime   time.Time
	Validators      *types.ValidatorSet
	LastValidators  *types.ValidatorSet // block.LastCommit validated against this

	// AppHash is updated after Commit
	AppHash []byte
	// ReceiptsHash is updated only after eval the txs
	ReceiptsHash []byte

	Plugins []IPlugin

	Tpsc *TPSCalculator
}

func LoadState(db dbm.DB) *State {
	return loadState(db, stateKey)
}

func loadState(db dbm.DB, key []byte) *State {
	s := &State{db: db}
	buf := db.Get(key)
	if len(buf) == 0 {
		return nil
	}

	r, n, err := bytes.NewReader(buf), new(int), new(error)
	wire.ReadBinaryPtr(&s, r, 0, n, err)
	if *err != nil {
		Exit(Fmt("Data has been corrupted or its spec has changed: %v\n", *err))
	}
	// TODO: ensure that buf is completely read.

	return s
}

// logger will be copied
func (s *State) Copy() *State {
	return &State{
		db:              s.db,
		logger:          s.logger,
		GenesisDoc:      s.GenesisDoc,
		ChainID:         s.ChainID,
		LastBlockHeight: s.LastBlockHeight,
		LastBlockID:     s.LastBlockID,
		LastBlockTime:   s.LastBlockTime,
		Validators:      s.Validators.Copy(),
		LastValidators:  s.LastValidators.Copy(),
		AppHash:         s.AppHash,
		ReceiptsHash:    s.ReceiptsHash,
		Plugins:         s.Plugins,
		Tpsc:            s.Tpsc,
	}
}

func (s *State) Save() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.db.SetSync(stateKey, s.Bytes())
}

func (s *State) SaveIntermediate() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.db.SetSync(stateIntermediateKey, s.Bytes())
}

// Load the intermediate state into the current state
// and do some sanity checks
func (s *State) LoadIntermediate() {
	s2 := loadState(s.db, stateIntermediateKey)
	if s.ChainID != s2.ChainID {
		PanicSanity(Fmt("State mismatch for ChainID. Got %v, Expected %v", s2.ChainID, s.ChainID))
	}

	if s.LastBlockHeight+1 != s2.LastBlockHeight {
		PanicSanity(Fmt("State mismatch for LastBlockHeight. Got %v, Expected %v", s2.LastBlockHeight, s.LastBlockHeight+1))
	}

	if !bytes.Equal(s.Validators.Hash(), s2.LastValidators.Hash()) {
		PanicSanity(Fmt("State mismatch for LastValidators. Got %X, Expected %X", s2.LastValidators.Hash(), s.Validators.Hash()))
	}

	if !bytes.Equal(s.AppHash, s2.AppHash) {
		PanicSanity(Fmt("State mismatch for AppHash. Got %X, Expected %X", s2.AppHash, s.AppHash))
	}

	if !bytes.Equal(s.ReceiptsHash, s2.ReceiptsHash) {
		PanicSanity(Fmt("State mismatch for ReceiptsHash. Got %X, Expected %X", s2.ReceiptsHash, s.ReceiptsHash))
	}

	s.setBlockAndValidators(s2.LastBlockHeight, s2.LastBlockID, s2.LastBlockTime, s2.Validators.Copy(), s2.LastValidators.Copy())
}

func (s *State) Equals(s2 *State) bool {
	return bytes.Equal(s.Bytes(), s2.Bytes())
}

// Bytes return go-wired []byte
func (s *State) Bytes() []byte {
	buf, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(s, buf, n, err)
	if *err != nil {
		PanicCrisis(*err)
	}
	return buf.Bytes()
}

// Mutate state variables to match block and validators
// after running EndBlock
func (s *State) SetBlockAndValidators(header *types.Header, blockPartsHeader types.PartSetHeader, prevValSet, nextValSet *types.ValidatorSet) {
	s.setBlockAndValidators(header.Height,
		types.BlockID{Hash: header.Hash(), PartsHeader: blockPartsHeader},
		header.Time,
		prevValSet, nextValSet)
}

func (s *State) setBlockAndValidators(
	height int, blockID types.BlockID, blockTime time.Time,
	prevValSet, nextValSet *types.ValidatorSet) {

	s.LastBlockHeight = height
	s.LastBlockID = blockID
	s.LastBlockTime = blockTime
	s.Validators = nextValSet
	s.LastValidators = prevValSet
}

func (s *State) SetLogger(logger *zap.Logger) {
	s.logger = logger
}

func (s *State) GetValidators() (*types.ValidatorSet, *types.ValidatorSet) {
	return s.LastValidators, s.Validators
}

func (s *State) GetLastBlockInfo() (types.BlockID, int, time.Time) {
	return s.LastBlockID, s.LastBlockHeight, s.LastBlockTime
}

func (s *State) GetChainID() string {
	return s.ChainID
}

// Load the most recent state from "state" db,
// or create a new one (and save) from genesis.
func GetState(config cfg.Config, stateDB dbm.DB) *State {
	state := LoadState(stateDB)
	if state == nil {
		if state = MakeGenesisStateFromFile(stateDB, config.GetString("genesis_file")); state != nil {
			state.Save()
		}
	}
	return state
}

//-----------------------------------------------------------------------------
// Genesis

func MakeGenesisStateFromFile(db dbm.DB, genDocFile string) *State {
	genDocJSON, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil
	}
	genDoc := types.GenesisDocFromJSON(genDocJSON)
	return MakeGenesisState(db, genDoc)
}

func MakeGenesisState(db dbm.DB, genDoc *types.GenesisDoc) *State {
	if len(genDoc.Validators) == 0 {
		Exit(Fmt("The genesis file has no validators"))
	}

	if genDoc.GenesisTime.IsZero() {
		genDoc.GenesisTime = time.Now()
	}

	// Make validators slice
	validators := make([]*types.Validator, len(genDoc.Validators))
	for i, val := range genDoc.Validators {
		pubKey := val.PubKey
		address := pubKey.Address()

		// Make validator
		validators[i] = &types.Validator{
			Address:     address,
			PubKey:      pubKey,
			VotingPower: val.Amount,
			IsCA:        val.IsCA,
			RPCAddress:  val.RPCAddress,
		}
	}
	validatorSet := types.NewValidatorSet(validators)
	lastValidatorSet := types.NewValidatorSet(nil)

	var plugins []IPlugin
	ps := strings.Split(genDoc.Plugins, ",")
	for _, pn := range ps {
		switch pn {
		case "specialop":
			plugins = append(plugins, plugin.NewSpecialop(&db))
		case "":
			// no core_plugins is allowed, so just ignore it
		default:
			PanicSanity(Fmt("Invalid plugin: %v", pn))
		}
	}

	// TODO: genDoc doesn't need to provide receiptsHash
	return &State{
		db:              db,
		GenesisDoc:      genDoc,
		ChainID:         genDoc.ChainID,
		LastBlockHeight: 0,
		LastBlockID:     types.BlockID{},
		LastBlockTime:   genDoc.GenesisTime,
		Validators:      validatorSet,
		LastValidators:  lastValidatorSet,
		AppHash:         genDoc.AppHash,
		Plugins:         plugins,
	}
}
