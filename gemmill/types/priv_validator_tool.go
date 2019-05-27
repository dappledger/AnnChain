package types

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dappledger/AnnChain/gemmill/utils"
)

var (
	ErrFileNotFound       = errors.New("priv_validator.json not found")
	ErrBranchIsUsed       = errors.New("priv_validator:branch name is used")
	ErrPVRevertFromBackup = errors.New("priv_validator:revert from backup, not find data")
)

const (
	PRIV_FILE_NAME = "priv_validator.json"
)

type PrivValidatorTool struct {
	dir string
	pv  *PrivValidator
}

func (pt *PrivValidatorTool) Init(dir string) error {
	var err error
	pt.pv, err = LoadPrivValidator(dir)
	if err != nil {
		return err
	}
	if pt.pv == nil {
		return ErrFileNotFound
	}
	return nil
}

func (pt *PrivValidatorTool) backupName(branchName string) string {
	return fmt.Sprintf("%v/%v-%v.json", filepath.Dir(pt.pv.filePath), PRIV_FILE_NAME, branchName)
}

func (pt *PrivValidatorTool) BackupData(branchName string) error {
	bkName := pt.backupName(branchName)
	find, err := utils.PathExists(bkName)
	if err != nil {
		return err
	}
	if find {
		return ErrBranchIsUsed
	}
	preDir := pt.pv.filePath
	pt.pv.SetFile(bkName)
	pt.pv.Save()
	pt.pv.SetFile(preDir)
	return nil
}

func (pt *PrivValidatorTool) RevertFromBackup(branchName string) error {
	bkName := pt.backupName(branchName)
	find, err := utils.PathExists(bkName)
	if err != nil {
		return err
	}
	if !find {
		return ErrPVRevertFromBackup
	}
	utils.CopyFile(pt.pv.filePath, bkName)
	return nil
}

func (pt *PrivValidatorTool) DelBackup(branchName string) {
	os.Remove(pt.backupName(branchName))
}

func (pt *PrivValidatorTool) SaveNewPrivV(toHeight int64) error {
	pt.pv.LastHeight = toHeight
	pt.pv.LastRound = 0
	pt.pv.LastStep = 0
	pt.pv.LastSignature = nil
	pt.pv.LastSignBytes = make([]byte, 0)
	pt.pv.Save()
	return nil
}
