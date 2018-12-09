package types

import (
	"errors"
	"fmt"
)

const (
	databaseCache   = 128
	databaseHandles = 1024
)

var (
	ErrRevertFromBackup = errors.New("revert from backup,not find data")
	ErrBranchNameUsed   = errors.New("app:branch name has been used")
)

type BaseAppTool struct {
	BaseApplication
}

func (t *BaseAppTool) backupName(branchName string) []byte {
	return []byte(fmt.Sprintf("%s-%s", lastBlockKey, branchName))
}

func (t *BaseAppTool) RevertFromBackup(branchName string) error {
	preKeyName := t.backupName(branchName)
	bs := t.Database.Get(preKeyName)
	if len(bs) == 0 {
		return ErrRevertFromBackup
	}
	t.Database.Set(lastBlockKey, bs)
	return nil
}

func (t *BaseAppTool) DelBackup(branchName string) {
	t.Database.Delete(t.backupName(branchName))
}

func (t *BaseAppTool) BackupLastBlockData(branchName string, lastBlock interface{}) error {
	preKeyName := t.backupName(branchName)
	dataBs := t.Database.Get(preKeyName)
	if len(dataBs) > 0 {
		return ErrBranchNameUsed
	}
	t.SaveLastBlockByKey(preKeyName, lastBlock)
	return nil
}
