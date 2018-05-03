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
	if err := t.InitBaseApplication(BASE_APP_NAME, datadir); err != nil {
		return err
	}
	lb := NewLastBlockInfo()
	ret, err := t.LoadLastBlock(lb)
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
	return def.INT(t.lastBlock.Height), t.lastBlock.Hash
}

func (t *AppTool) BackupLastBlock(branchName string) error {
	return t.BackupLastBlockData(branchName, &t.lastBlock)
}

func (t *AppTool) SaveNewLastBlock(fromHeight def.INT, fromAppHash []byte) error {
	newBranchBlock := LastBlockInfo{
		Height: fromHeight,
		Hash:   fromAppHash,
	}
	t.SaveLastBlock(newBranchBlock)
	// TODO
	return nil
}
