package evm

import (
	"errors"

	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

const (
	databaseCache   = 128
	databaseHandles = 1024
)

var (
	ErrRevertFromBackup = errors.New("revert from backup,not find data")
	ErrDataTransfer     = errors.New("data transfer err")
)

type AppTool struct {
	agtypes.BaseAppTool
	lastBlock LastBlockInfo
}

func (t *AppTool) Init(datadir string) error {
	if err := t.InitBaseApplication(APP_NAME, datadir); err != nil {
		return err
	}
	ret, err := t.LoadLastBlock(&t.lastBlock)
	if err != nil {
		return err
	}
	tmp, ok := ret.(*LastBlockInfo)
	if !ok {
		return ErrDataTransfer
	}
	t.lastBlock = *tmp
	return nil
}

func (t *AppTool) LastHeightHash() (def.INT, []byte) {
	return def.INT(t.lastBlock.Height), t.lastBlock.AppHash
}

func (t *AppTool) BackupLastBlock(branchName string) error {
	return t.BackupLastBlockData(branchName, &t.lastBlock)
}

func (t *AppTool) SaveNewLastBlock(fromHeight def.INT, fromAppHash []byte) error {
	newBranchBlock := LastBlockInfo{
		Height:  fromHeight,
		AppHash: fromAppHash,
	}
	t.SaveLastBlock(newBranchBlock)
	// TODO
	return nil
}
