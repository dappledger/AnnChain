// Copyright © 2017 ZhongAn Technology
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

package blockchain

import (
	"errors"
	"fmt"

	dbm "github.com/dappledger/AnnChain/gemmill/modules/go-db"
	"github.com/dappledger/AnnChain/gemmill/types"
	cfg "github.com/spf13/viper"
)

var (
	ErrBlockIsNil      = errors.New("the chain has no block data")
	ErrBranchNameUsed  = errors.New("blockchain:branch name has been used")
	ErrConvertToFuture = errors.New("can't revert to future height")
	ErrRevertBackup    = errors.New("revert from backup,not find data")
)

func BlockStoreDB(config *cfg.Viper) (dbm.DB, dbm.DB) {
	var (
		db_backend     = config.GetString("db_backend")
		db_dir         = config.GetString("db_dir")
		db_archive_dir = config.GetString("db_archive_dir")
	)
	return dbm.NewDB("blockstore", db_backend, db_dir),
		dbm.NewDB("blockstore", db_backend, db_archive_dir)
}

func LoadBlockStore(blockStoreDB, blockArchiveDB dbm.DB, height int64) (*types.Block, *types.BlockMeta, *types.BlockID) {
	blockStore := NewBlockStore(blockStoreDB, blockArchiveDB)
	nextBlock := blockStore.LoadBlock(height + 1)
	if nextBlock == nil {
		return nil, nil, &types.BlockID{}
	}
	blockc, blockMeta := blockStore.LoadBlock(height), blockStore.LoadBlockMeta(height)
	return blockc, blockMeta, &nextBlock.Header.LastBlockID
}

type StoreTool struct {
	db        dbm.DB
	archDB    dbm.DB
	lastBlock BlockStoreStateJSON
}

func (st *StoreTool) Init(config *cfg.Viper) error {
	st.db, st.archDB = BlockStoreDB(config)
	st.lastBlock = LoadBlockStoreStateJSON(st.db)
	if st.lastBlock.Height <= 0 {
		return ErrBlockIsNil
	}
	return nil
}

func (st *StoreTool) LoadBlock(height int64) (*types.Block, *types.BlockMeta, *types.BlockID) {
	return LoadBlockStore(st.db, st.archDB, height)
}

func (st *StoreTool) LastHeight() int64 {
	return st.lastBlock.Height
}

func (st *StoreTool) backupName(branchName string) []byte {
	return []byte(fmt.Sprintf("%s-%s", blockStoreKey, branchName))
}

func (st *StoreTool) BackupLastBlock(branchName string) error {
	preKeyName := st.backupName(branchName)
	dataBs := st.db.Get(preKeyName)
	if len(dataBs) > 0 {
		return ErrBranchNameUsed
	}
	st.lastBlock.SaveByKey(preKeyName, st.db)
	return nil
}

func (st *StoreTool) DelBackup(branchName string) {
	st.db.Delete(st.backupName(branchName))
}

func (st *StoreTool) RevertFromBackup(branchName string) error {
	bs := st.db.Get(st.backupName(branchName))
	if len(bs) == 0 {
		return ErrRevertBackup
	}
	st.db.Set(blockStoreKey, bs)
	return nil
}

func (st *StoreTool) SaveNewLastBlock(toHeight int64) error {
	if toHeight >= st.lastBlock.Height {
		return ErrConvertToFuture
	}
	originHeight := st.lastBlock.OriginHeight
	if originHeight > toHeight {
		// 从最低高度 归档
		originHeight = toHeight
	}
	newLastBlockStore := BlockStoreStateJSON{
		Height:       toHeight,
		OriginHeight: originHeight,
	}
	newLastBlockStore.Save(st.db)
	return nil
}
