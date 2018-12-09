package state

import (
	"errors"
	"fmt"

	cfg "github.com/spf13/viper"
	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	dbm "github.com/dappledger/AnnChain/module/lib/go-db"
	"github.com/dappledger/AnnChain/module/xlib/def"
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
	st.lastState = LoadState(nil, st.db)
	if st.lastState == nil || st.lastState.LastBlockHeight <= 0 {
		return ErrStateIsNil
	}
	return nil
}

func (st *StateTool) ChangeToIntermidiate() {
}

func (st *StateTool) LastHeight() def.INT {
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
func (st *StateTool) SaveNewState(lastBlock *pbtypes.Block, lastBlockMeta *pbtypes.BlockMeta, lastBlockID *pbtypes.BlockID) error {
	newState := st.lastState.Copy()
	newState.AppHash = lastBlock.Header.AppHash
	newState.LastBlockHeight = lastBlock.Header.Height
	newState.LastBlockID = *lastBlockID
	newState.LastBlockTime = lastBlockMeta.Header.Time
	newState.Save()
	return nil
}
