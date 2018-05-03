package blockchain

import (
	"errors"
	"fmt"

	cfg "github.com/spf13/viper"
	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	dbm "github.com/dappledger/AnnChain/module/lib/go-db"
	"github.com/dappledger/AnnChain/module/xlib/def"
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

func LoadBlockStore(blockStoreDB, blockArchiveDB dbm.DB, height def.INT) (*pbtypes.Block, *pbtypes.BlockMeta, *pbtypes.BlockID) {
	blockStore := NewBlockStore(blockStoreDB, blockArchiveDB)
	nextBlock := blockStore.LoadBlock(height + 1)
	if nextBlock == nil {
		return nil, nil, &pbtypes.BlockID{}
	}
	blockc, blockMeta := blockStore.LoadBlock(height), blockStore.LoadBlockMeta(height)
	return blockc.Block, blockMeta, nextBlock.Header.LastBlockID
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

func (st *StoreTool) LoadBlock(height def.INT) (*pbtypes.Block, *pbtypes.BlockMeta, *pbtypes.BlockID) {
	return LoadBlockStore(st.db, st.archDB, height)
}

func (st *StoreTool) LastHeight() def.INT {
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

func (st *StoreTool) SaveNewLastBlock(toHeight def.INT) error {
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
