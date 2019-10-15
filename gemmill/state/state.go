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
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"
	"time"

	"github.com/dappledger/AnnChain/gemmill/go-wire"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	dbm "github.com/dappledger/AnnChain/gemmill/modules/go-db"
	"github.com/dappledger/AnnChain/gemmill/modules/go-events"
	"github.com/dappledger/AnnChain/gemmill/types"

	"github.com/spf13/viper"
)

var (
	stateKey             = []byte("stateKey")
	stateIntermediateKey = []byte("stateIntermediateKey")
)

type IBlockExecutable interface {
	BeginBlock(*types.Block, events.Fireable, *types.PartSetHeader) error
	ExecBlock(*types.Block, events.Fireable, *types.ExecuteResult) error
	EndBlock(*types.Block, events.Fireable, *types.PartSetHeader, []*types.ValidatorAttr, *types.ValidatorSet) error
}

type BlockVerifier interface {
	ValidateBlock(*types.Block) error
}

// NOTE: not goroutine-safe.
type State struct {
	started bool

	// mtx for writing to db
	mtx             sync.Mutex
	db              dbm.DB
	blockExecutable IBlockExecutable
	blockVerifier   BlockVerifier

	// should not change
	GenesisDoc *types.GenesisDoc
	ChainID    string

	// updated at end of ExecBlock
	LastBlockHeight    int64 // Genesis state has this set to 0.  So, Block(H=0) does not exist.
	LastBlockID        types.BlockID
	LastBlockTime      time.Time
	Validators         *types.ValidatorSet
	LastValidators     *types.ValidatorSet // block.LastCommit validated against this
	LastNonEmptyHeight int64

	// AppHash is updated after Commit
	AppHash []byte
	// ReceiptsHash is updated only after eval the txs
	ReceiptsHash []byte
}

func (s *State) CheckPubkeyPtr() {
	for _, v := range s.GenesisDoc.Validators {
		fmt.Println("genes:", reflect.TypeOf(v.PubKey))
	}
	for _, v := range s.Validators.Validators {
		fmt.Println("vldts:", reflect.TypeOf(v.PubKey))
	}
	for _, v := range s.LastValidators.Validators {
		fmt.Println("last vldts:", reflect.TypeOf(v.PubKey))
	}
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
		gcmn.Exit(gcmn.Fmt("Data has been corrupted or its spec has changed: %v\n", *err))
	}
	// TODO: ensure that buf is completely read.

	return s
}

func (s *State) Copy() *State {
	return &State{
		db:              s.db,
		blockExecutable: s.blockExecutable,
		blockVerifier:   s.blockVerifier,

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

func StateDB(config *viper.Viper) dbm.DB {
	var (
		db_backend = config.GetString("db_backend")
		db_dir     = config.GetString("db_dir")
	)
	return dbm.NewDB("state", db_backend, db_dir)
}

func (s *State) Save() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.db.SetSync(stateKey, s.Bytes())
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
func (s *State) LoadIntermediate() {
	s2 := loadState(s.db, stateIntermediateKey)
	if s.ChainID != s2.ChainID {
		gcmn.PanicSanity(gcmn.Fmt("State mismatch for ChainID. Got %v, Expected %v", s2.ChainID, s.ChainID))
	}

	if s.LastBlockHeight+1 != s2.LastBlockHeight {
		gcmn.PanicSanity(gcmn.Fmt("State mismatch for LastBlockHeight. Got %v, Expected %v", s2.LastBlockHeight, s.LastBlockHeight+1))
	}

	if !bytes.Equal(s.Validators.Hash(), s2.LastValidators.Hash()) {
		gcmn.PanicSanity(gcmn.Fmt("State mismatch for LastValidators. Got %X, Expected %X", s2.LastValidators.Hash(), s.Validators.Hash()))
	}

	if !bytes.Equal(s.AppHash, s2.AppHash) {
		gcmn.PanicSanity(gcmn.Fmt("State mismatch for AppHash. Got %X, Expected %X", s2.AppHash, s.AppHash))
	}

	if !bytes.Equal(s.ReceiptsHash, s2.ReceiptsHash) {
		gcmn.PanicSanity(gcmn.Fmt("State mismatch for ReceiptsHash. Got %X, Expected %X", s2.ReceiptsHash, s.ReceiptsHash))
	}

	s.setBlockAndValidators(s2.LastBlockHeight, s2.LastNonEmptyHeight, s2.LastBlockID, s2.LastBlockTime, s2.Validators.Copy(), s2.LastValidators.Copy())
}

func (s *State) SetBlockExecutable(ex IBlockExecutable) {
	s.blockExecutable = ex
}

func (s *State) Equals(s2 *State) bool {
	return bytes.Equal(s.Bytes(), s2.Bytes())
}

// Bytes return go-wired []byte
func (s *State) Bytes() []byte {
	buf, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(s, buf, n, err)
	if *err != nil {
		gcmn.PanicCrisis(*err)
	}
	return buf.Bytes()
}

func (s *State) SetBlockVerifier(verifier BlockVerifier) {
	s.blockVerifier = verifier
}

// Mutate state variables to match block and validators
// after running EndBlock
func (s *State) SetBlockAndValidators(header *types.Header, blockPartsHeader types.PartSetHeader, prevValSet, nextValSet *types.ValidatorSet) {
	nonEmptyHeight := s.LastNonEmptyHeight
	if header.NumTxs > 0 {
		nonEmptyHeight = header.Height
	}
	s.setBlockAndValidators(header.Height, nonEmptyHeight, types.BlockID{Hash: header.Hash(), PartsHeader: blockPartsHeader},
		header.Time, prevValSet, nextValSet)
}

func (s *State) setBlockAndValidators(height int64, nonEmptyHeight int64, blockID types.BlockID, blockTime time.Time,
	prevValSet, nextValSet *types.ValidatorSet) {

	s.LastBlockHeight = height
	s.LastBlockID = blockID
	s.LastBlockTime = blockTime
	s.Validators = nextValSet
	s.LastValidators = prevValSet
	s.LastNonEmptyHeight = nonEmptyHeight
}

func (s *State) GetValidators() (*types.ValidatorSet, *types.ValidatorSet) {
	return s.LastValidators, s.Validators
}

func (s *State) GetLastBlockInfo() (types.BlockID, int64, time.Time) {
	return s.LastBlockID, s.LastBlockHeight, s.LastBlockTime
}

func (s *State) GetChainID() string {
	return s.ChainID
}

// Load the most recent state from "state" db,
// or create a new one (and save) from genesis.
func GetState(config *viper.Viper, stateDB dbm.DB) *State {
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
		gcmn.Exit(gcmn.Fmt("The genesis file has no validators"))
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
		}
	}
	validatorSet := types.NewValidatorSet(validators)
	lastValidatorSet := types.NewValidatorSet(nil)

	// TODO: genDoc doesn't need to provide receiptsHash
	return &State{
		db:                 db,
		GenesisDoc:         genDoc,
		ChainID:            genDoc.ChainID,
		LastBlockHeight:    0,
		LastBlockID:        types.BlockID{},
		LastBlockTime:      genDoc.GenesisTime,
		Validators:         validatorSet,
		LastValidators:     lastValidatorSet,
		AppHash:            genDoc.AppHash,
		LastNonEmptyHeight: 0,
	}
}
