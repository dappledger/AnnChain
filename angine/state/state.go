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
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	sttpb "github.com/dappledger/AnnChain/angine/protos/state"
	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	. "github.com/dappledger/AnnChain/module/lib/go-common"
	dbm "github.com/dappledger/AnnChain/module/lib/go-db"
	"github.com/dappledger/AnnChain/module/lib/go-events"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

var (
	stateKey             = []byte("stateKey")
	stateIntermediateKey = []byte("stateIntermediateKey")
)

//-----------------------------------------------------------------------------

type IBlockExecutable interface {
	BeginBlock(*agtypes.BlockCache, events.Fireable, *pbtypes.PartSetHeader)
	ExecBlock(*agtypes.BlockCache, events.Fireable, *agtypes.ExecuteResult)
	EndBlock(*agtypes.BlockCache, events.Fireable, *pbtypes.PartSetHeader, []*agtypes.ValidatorAttr, *agtypes.ValidatorSet)
}

// NOTE: not goroutine-safe.
type State struct {
	// mtx for writing to db
	mtx sync.Mutex

	logger          *zap.Logger
	db              dbm.DB
	blockExecutable IBlockExecutable

	// should not change
	GenesisDoc *agtypes.GenesisDoc
	ChainID    string

	// updated at end of ExecBlock
	LastBlockHeight def.INT
	// Genesis state has this set to 0.  So, Block(H=0) does not exist.
	LastBlockID        pbtypes.BlockID
	LastBlockTime      def.INT
	Validators         *agtypes.ValidatorSet
	LastValidators     *agtypes.ValidatorSet // block.LastCommit validated against this
	LastNonEmptyHeight def.INT
	// some where in the past, maybe jump through lots of blocks

	// AppHash is updated after Commit
	AppHash []byte
	// ReceiptsHash is updated only after eval the txs
	ReceiptsHash []byte
}

func (s *State) FillDataFromPbBytes(pbbys []byte) error {
	var pbState sttpb.State
	err := agtypes.UnmarshalData(pbbys, &pbState)
	if err != nil {
		return err
	}
	s.GenesisDoc = agtypes.GenesisDocFromJSON(pbState.GenesisDoc.JSONData)
	s.ChainID = pbState.ChainID
	s.LastBlockHeight = pbState.LastBlockHeight
	s.LastBlockID = (*pbState.LastBlockID)
	s.LastBlockTime = pbState.LastBlockTime
	s.Validators = agtypes.ValSetFromJsonBytes(pbState.Validators.JSONData)
	s.LastValidators = agtypes.ValSetFromJsonBytes(pbState.LastValidators.JSONData)
	s.LastNonEmptyHeight = pbState.LastNonEmptyHeight
	s.AppHash = pbState.AppHash
	s.ReceiptsHash = pbState.ReceiptsHash
	//s.Plugins, err = PluginsFromPbData(pbState.Plugins, s.logger, s.db)
	return err
}

func (s *State) PbBytes() ([]byte, error) {
	var pbState sttpb.State
	genesBys, err := s.GenesisDoc.JSONBytes()
	if err != nil {
		return nil, err
	}
	vsetBys, err := s.Validators.JSONBytes()
	if err != nil {
		return nil, err
	}
	lvsetBys, err := s.LastValidators.JSONBytes()
	if err != nil {
		return nil, err
	}
	//if pbState.Plugins, err = ToPbPlugins(s.Plugins); err != nil {
	//	return nil, err
	//}
	pbState.GenesisDoc = &sttpb.GenesisDoc{
		JSONData: genesBys,
	}
	pbState.ChainID = s.ChainID
	pbState.LastBlockHeight = s.LastBlockHeight
	pbState.LastBlockID = &s.LastBlockID
	pbState.LastBlockTime = s.LastBlockTime
	pbState.Validators = &sttpb.ValidatorSet{
		JSONData: vsetBys,
	}
	pbState.LastValidators = &sttpb.ValidatorSet{
		JSONData: lvsetBys,
	}
	pbState.LastNonEmptyHeight = s.LastNonEmptyHeight
	pbState.AppHash = s.AppHash
	pbState.ReceiptsHash = s.ReceiptsHash
	return agtypes.MarshalData(&pbState)
}

func LoadState(logger *zap.Logger, db dbm.DB) *State {
	return loadState(logger, db, stateKey)
}

func loadState(logger *zap.Logger, db dbm.DB, key []byte) *State {
	s := &State{db: db}
	bys := db.Get(key)
	if len(bys) == 0 {
		return nil
	}
	if err := s.FillDataFromPbBytes(bys); err != nil {
		Exit(Fmt("Data has been corrupted or its spec has changed: %v\n", err))
	}
	// TODO: ensure that buf is completely read.
	return s
}

func StateDB(config *viper.Viper) dbm.DB {
	var (
		db_backend = config.GetString("db_backend")
		db_dir     = config.GetString("db_dir")
	)
	return dbm.NewDB("state", db_backend, db_dir)
}

// logger will be copied
func (s *State) Copy() *State {
	return &State{
		db:              s.db,
		logger:          s.logger,
		blockExecutable: s.blockExecutable,

		GenesisDoc:         s.GenesisDoc,
		ChainID:            s.ChainID,
		LastBlockHeight:    s.LastBlockHeight,
		LastBlockID:        s.LastBlockID,
		LastBlockTime:      s.LastBlockTime,
		Validators:         s.Validators.Copy(),
		LastValidators:     s.LastValidators.Copy(),
		AppHash:            s.AppHash,
		ReceiptsHash:       s.ReceiptsHash,
		LastNonEmptyHeight: s.LastNonEmptyHeight,
	}
}

func (s *State) Save() {
	s.SaveToKey(stateKey)
}

func (s *State) SaveToKey(key []byte) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.db.SetSync(key, s.Bytes())
}

func (s *State) SaveIntermediate() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.db.SetSync(stateIntermediateKey, s.Bytes())
}

// Load the intermediate state into the current state
// and do some sanity checks
func (s *State) LoadIntermediate() error {
	s2 := loadState(s.logger, s.db, stateIntermediateKey)
	s2.blockExecutable = s.blockExecutable

	if s.ChainID != s2.ChainID {
		return errors.Errorf("State mismatch for ChainID. Got %v, Expected %v", s2.ChainID, s.ChainID)
	}

	if s.LastBlockHeight+1 != s2.LastBlockHeight {
		return errors.Errorf("State mismatch for LastBlockHeight. Got %v, Expected %v", s2.LastBlockHeight, s.LastBlockHeight+1)
	}

	if !bytes.Equal(s.Validators.Hash(), s2.LastValidators.Hash()) {
		return errors.Errorf("State mismatch for LastValidators. Got %X, Expected %X", s2.LastValidators.Hash(), s.Validators.Hash())
	}

	if !bytes.Equal(s.AppHash, s2.AppHash) {
		return errors.Errorf("State mismatch for AppHash. Got %X, Expected %X", s2.AppHash, s.AppHash)
	}

	if !bytes.Equal(s.ReceiptsHash, s2.ReceiptsHash) {
		return errors.Errorf("State mismatch for ReceiptsHash. Got %X, Expected %X", s2.ReceiptsHash, s.ReceiptsHash)
	}

	s.setBlockAndValidators(s2.LastBlockHeight, s2.LastNonEmptyHeight, s2.LastBlockID, s2.LastBlockTime, s2.Validators.Copy(), s2.LastValidators.Copy())

	return nil
}

func (s *State) SetBlockExecutable(ex IBlockExecutable) {
	s.blockExecutable = ex
}

func (s *State) Equals(s2 *State) bool {
	return bytes.Equal(s.Bytes(), s2.Bytes())
}

// Bytes return go-wired []byte
func (s *State) Bytes() []byte {
	bys, err := s.PbBytes()
	if err != nil {
		s.logger.DPanic("[State Bytes]", zap.Error(err))
		return []byte{}
	}
	return bys
}

// Mutate state variables to match block and validators
// after running EndBlock
func (s *State) SetBlockAndValidators(header *pbtypes.Header, blockPartsHeader *pbtypes.PartSetHeader, prevValSet, nextValSet *agtypes.ValidatorSet) {
	nonEmptyHeight := s.LastNonEmptyHeight
	if header.NumTxs > 0 {
		nonEmptyHeight = header.Height
	}

	s.setBlockAndValidators(header.Height, nonEmptyHeight,
		pbtypes.BlockID{Hash: header.Hash(), PartsHeader: blockPartsHeader},
		header.Time,
		prevValSet, nextValSet)
}

func (s *State) setBlockAndValidators(
	height, nonEmptyHeight def.INT, blockID pbtypes.BlockID, blockTime def.INT,
	prevValSet, nextValSet *agtypes.ValidatorSet) {

	s.LastBlockHeight = height
	s.LastBlockID = blockID
	s.LastBlockTime = blockTime
	s.Validators = nextValSet
	s.LastValidators = prevValSet
	s.LastNonEmptyHeight = nonEmptyHeight
}

func (s *State) SetLogger(logger *zap.Logger) {
	s.logger = logger
}

func (s *State) GetValidators() (*agtypes.ValidatorSet, *agtypes.ValidatorSet) {
	return s.LastValidators, s.Validators
}

func (s *State) GetLastBlockInfo() (pbtypes.BlockID, def.INT, def.INT) {
	return s.LastBlockID, s.LastBlockHeight, s.LastBlockTime
}

func (s *State) GetChainID() string {
	return s.ChainID
}

// Load the most recent state from "state" db,
// or create a new one (and save) from genesis.
func GetState(logger *zap.Logger, config *viper.Viper, stateDB dbm.DB) *State {
	state := LoadState(logger, stateDB)
	if state == nil {
		if state = MakeGenesisStateFromFile(logger, stateDB, config.GetString("genesis_file")); state != nil {
			state.Save()
		}
	}
	return state
}

//-----------------------------------------------------------------------------
// Genesis

func MakeGenesisStateFromFile(logger *zap.Logger, db dbm.DB, genDocFile string) *State {
	genDocJSON, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil
	}
	genDoc := agtypes.GenesisDocFromJSON(genDocJSON)
	return MakeGenesisState(logger, db, genDoc)
}

func MakeGenesisState(logger *zap.Logger, db dbm.DB, genDoc *agtypes.GenesisDoc) *State {
	if len(genDoc.Validators) == 0 {
		Exit(Fmt("The genesis file has no validators"))
	}

	if genDoc.GenesisTime.IsZero() {
		genDoc.GenesisTime = agtypes.Time{time.Now()}
	}

	// Make validators slice
	validators := make([]*agtypes.Validator, len(genDoc.Validators))
	for i, val := range genDoc.Validators {
		pubKey := val.PubKey
		address := pubKey.Address()

		// Make validator
		validators[i] = &agtypes.Validator{
			Address:     address,
			PubKey:      pubKey,
			VotingPower: val.Amount,
			IsCA:        val.IsCA,
		}
	}
	validatorSet := agtypes.NewValidatorSet(validators)
	lastValidatorSet := agtypes.NewValidatorSet(nil)

	// TODO: genDoc doesn't need to provide receiptsHash
	return &State{
		db:                 db,
		logger:             logger,
		GenesisDoc:         genDoc,
		ChainID:            genDoc.ChainID,
		LastBlockHeight:    0,
		LastBlockID:        pbtypes.BlockID{},
		LastBlockTime:      genDoc.GenesisTime.UnixNano(),
		Validators:         validatorSet,
		LastValidators:     lastValidatorSet,
		AppHash:            genDoc.AppHash,
		LastNonEmptyHeight: 0,
	}
}

func MakeState(logger *zap.Logger, db dbm.DB) *State {
	validatorSet := agtypes.NewValidatorSet(nil)
	lastValidatorSet := agtypes.NewValidatorSet(nil)
	return &State{
		db:                 db,
		logger:             logger,
		GenesisDoc:         nil,
		ChainID:            "",
		LastBlockHeight:    0,
		LastBlockID:        pbtypes.BlockID{},
		LastBlockTime:      time.Now().UnixNano(),
		Validators:         validatorSet,
		LastValidators:     lastValidatorSet,
		AppHash:            nil,
		LastNonEmptyHeight: 0,
	}
}

func fillPrivate(state *State, db dbm.DB, logger *zap.Logger) {
	state.db = db
	state.logger = logger
}
