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
	"errors"
	"fmt"

	dbm "github.com/dappledger/AnnChain/gemmill/modules/go-db"
	"github.com/dappledger/AnnChain/gemmill/types"
	cfg "github.com/spf13/viper"
)

var (
	ErrBranchNameUsed   = errors.New("state db:branch name has been used")
	ErrStateIsNil       = errors.New("the state has no data")
	ErrRevertFromBackup = errors.New("state revert from backup, not find data")
)

type StateTool struct {
	db        dbm.DB
	lastState *State
}

func (st *StateTool) Init(config *cfg.Viper) error {
	st.db = StateDB(config)
	st.lastState = LoadState(st.db)
	if st.lastState == nil || st.lastState.LastBlockHeight <= 0 {
		return ErrStateIsNil
	}
	return nil
}

func (st *StateTool) ChangeToIntermidiate() {
}

func (st *StateTool) NewValSet(lastHeight int64) *types.ValidatorSet {
	newVal := st.lastState.Validators.Copy()
	for i := range newVal.Validators {
		newVal.Validators[i].Accum = 0
	}
	newVal.IncrementAccum(lastHeight)
	return newVal
}

func (st *StateTool) LastHeight() int64 {
	return st.lastState.LastBlockHeight
}

func (st *StateTool) backupName(branchName string) []byte {
	return []byte(fmt.Sprintf("%s-%s", stateKey, branchName))
}

func (st *StateTool) BackupLastState(branchName string) error {
	saveKey := st.backupName(branchName)
	preBs := st.db.Get(saveKey)
	if len(preBs) > 0 {
		return ErrBranchNameUsed
	}
	st.lastState.SaveToKey(saveKey)
	// SaveIntermediate()
	return nil
}

func (st *StateTool) RevertFromBackup(branchName string) error {
	preKeyName := st.backupName(branchName)
	bs := st.db.Get(preKeyName)
	if len(preKeyName) == 0 {
		return ErrRevertFromBackup
	}
	st.db.Set(stateKey, bs)
	return nil
}

func (st *StateTool) DelBackup(branchName string) {
	st.db.Delete(st.backupName(branchName))
}

// back to height of lastBlock
// TODO whatif the validator_set has been changed
func (st *StateTool) SaveNewState(valSet *types.ValidatorSet, nextBlock, lastBlock *types.Block, lastBlockMeta *types.BlockMeta, lastBlockID *types.BlockID) error {
	newState := st.lastState.Copy()
	newState.AppHash = nextBlock.Header.AppHash
	newState.LastBlockHeight = lastBlock.Header.Height
	newState.LastBlockID = *lastBlockID
	newState.LastBlockTime = lastBlockMeta.Header.Time
	newState.ReceiptsHash = nextBlock.Header.ReceiptsHash
	newState.Validators = valSet
	newState.Save()
	return nil
}
