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

package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/dappledger/AnnChain/gemmill/go-wire"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	dbm "github.com/dappledger/AnnChain/gemmill/modules/go-db"
	"github.com/dappledger/AnnChain/gemmill/types"
)

/*
Simple low level store for blocks.

There are three types of information stored:
 - BlockMeta:   Meta information about each block
 - Block part:  Parts of each block, aggregated w/ PartSet
 - Commit:      The commit part of each block, for gossiping precommit votes

Currently the precommit signatures are duplicated in the Block parts as
well as the Commit.  In the future this may change, perhaps by moving
the Commit data outside the Block.

Panics indicate probable corruption in the data
*/
type BlockStore struct {
	db        dbm.DB
	archiveDB dbm.DB

	mtx          sync.RWMutex
	height       int64
	originHeight int64
}

func NewBlockStore(db, archiveDB dbm.DB) *BlockStore {
	bsjson := LoadBlockStoreStateJSON(db)
	return &BlockStore{
		height:       bsjson.Height,
		originHeight: bsjson.OriginHeight,
		db:           db,
		archiveDB:    archiveDB,
	}
}

// Height() returns the last known contiguous block height.
func (bs *BlockStore) Height() int64 {
	bs.mtx.RLock()
	defer bs.mtx.RUnlock()
	return bs.height
}

func (bs *BlockStore) OriginHeight() int64 {
	return bs.originHeight
}

func (bs *BlockStore) SetOriginHeight(height int64) {
	bs.originHeight = height
}

func (bs *BlockStore) GetReader(key []byte) io.Reader {
	bytez := bs.db.Get(key)
	if bytez == nil {
		return nil
	}
	return bytes.NewReader(bytez)
}

func (bs *BlockStore) GetReaderFromArchive(key []byte) io.Reader {
	bytez := bs.archiveDB.Get(key)
	if bytez == nil {
		return nil
	}
	return bytes.NewReader(bytez)
}

func (bs *BlockStore) LoadBlock(height int64) *types.Block {
	var n int
	var err error
	r := bs.GetReader(calcBlockMetaKey(height))
	if r == nil {
		return nil
	}
	meta := wire.ReadBinary(&types.BlockMeta{}, r, 0, &n, &err).(*types.BlockMeta)
	if err != nil {
		gcmn.PanicCrisis(gcmn.Fmt("Error reading block meta: %v", err))
	}
	bytez := []byte{}
	for i := 0; i < meta.PartsHeader.Total; i++ {
		part := bs.LoadBlockPart(height, i)
		bytez = append(bytez, part.Bytes...)
	}
	block := wire.ReadBinary(&types.Block{}, bytes.NewReader(bytez), 0, &n, &err).(*types.Block)
	if err != nil {
		gcmn.PanicCrisis(gcmn.Fmt("Error reading block: %v", err))
	}
	return block
}

func (bs *BlockStore) LoadBlockPart(height int64, index int) *types.Part {
	var n int
	var err error
	r := bs.GetReader(calcBlockPartKey(height, index))
	if r == nil {
		return nil
	}
	part := wire.ReadBinary(&types.Part{}, r, 0, &n, &err).(*types.Part)
	if err != nil {
		gcmn.PanicCrisis(gcmn.Fmt("Error reading block part: %v", err))
	}
	return part
}

func (bs *BlockStore) LoadBlockMeta(height int64) *types.BlockMeta {
	var n int
	var err error
	r := bs.GetReader(calcBlockMetaKey(height))
	if r == nil {
		return nil
	}
	meta := wire.ReadBinary(&types.BlockMeta{}, r, 0, &n, &err).(*types.BlockMeta)
	if err != nil {
		gcmn.PanicCrisis(gcmn.Fmt("Error reading block meta: %v", err))
	}
	return meta
}

// The +2/3 and other Precommit-votes for block at `height`.
// This Commit comes from block.LastCommit for `height+1`.
func (bs *BlockStore) LoadBlockCommit(height int64) *types.Commit {
	var n int
	var err error
	r := bs.GetReader(calcBlockCommitKey(height))
	if r == nil {
		return nil
	}
	commit := wire.ReadBinary(&types.Commit{}, r, 0, &n, &err).(*types.Commit)
	if err != nil {
		gcmn.PanicCrisis(gcmn.Fmt("Error reading commit: %v", err))
	}
	return commit
}

// NOTE: the Precommit-vote heights are for the block at `height`
func (bs *BlockStore) LoadSeenCommit(height int64) *types.Commit {
	var n int
	var err error
	r := bs.GetReader(calcSeenCommitKey(height))
	if r == nil {
		return nil
	}
	commit := wire.ReadBinary(&types.Commit{}, r, 0, &n, &err).(*types.Commit)
	if err != nil {
		gcmn.PanicCrisis(gcmn.Fmt("Error reading commit: %v", err))
	}
	return commit
}

// blockParts: Must be parts of the block
// seenCommit: The +2/3 precommits that were seen which committed at height.
//             If all the nodes restart after committing a block,
//             we need this to reload the precommits to catch-up nodes to the
//             most recent height.  Otherwise they'd stall at H-1.
func (bs *BlockStore) SaveBlock(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit) {
	height := block.Height
	if height != bs.Height()+1 {
		gcmn.PanicSanity(gcmn.Fmt("BlockStore can only save contiguous blocks. Wanted %v, got %v", bs.Height()+1, height))
	}
	if !blockParts.IsComplete() {
		gcmn.PanicSanity(gcmn.Fmt("BlockStore can only save complete block part sets"))
	}

	// Save block meta
	meta := types.NewBlockMeta(block, blockParts)
	metaBytes := wire.BinaryBytes(meta)
	bs.db.Set(calcBlockMetaKey(height), metaBytes)

	// Save block parts
	for i := 0; i < blockParts.Total(); i++ {
		bs.saveBlockPart(height, i, blockParts.GetPart(i))
	}

	// Save block commit (duplicate and separate from the Block)
	blockCommitBytes := wire.BinaryBytes(block.LastCommit)
	bs.db.Set(calcBlockCommitKey(height-1), blockCommitBytes)

	// Save seen commit (seen +2/3 precommits for block)
	// NOTE: we can delete this at a later height
	seenCommitBytes := wire.BinaryBytes(seenCommit)
	bs.db.Set(calcSeenCommitKey(height), seenCommitBytes)

	// Save new BlockStoreStateJSON descriptor
	BlockStoreStateJSON{Height: height, OriginHeight: bs.originHeight}.Save(bs.db)

	// Done!
	bs.mtx.Lock()
	bs.height = height
	bs.mtx.Unlock()

	// Flush
	bs.db.SetSync(nil, nil)
}

func (bs *BlockStore) SaveBlockToArchive(height int64, block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit) {
	// Save block meta
	meta := types.NewBlockMeta(block, blockParts)
	metaBytes := wire.BinaryBytes(meta)
	bs.archiveDB.Set(calcBlockMetaKey(height), metaBytes)

	// Save block parts
	for i := 0; i < blockParts.Total(); i++ {
		bs.savePartToArchive(height, i, blockParts.GetPart(i))
	}

	// Save block commit (duplicate and separate from the Block)
	blockCommitBytes := wire.BinaryBytes(block.LastCommit)
	bs.archiveDB.Set(calcBlockCommitKey(height-1), blockCommitBytes)

	// Save seen commit (seen +2/3 precommits for block)
	// NOTE: we can delete this at a later height
	seenCommitBytes := wire.BinaryBytes(seenCommit)
	bs.archiveDB.Set(calcSeenCommitKey(height), seenCommitBytes)
}

func (bs *BlockStore) DeleteBlock(height int64) (err error) {

	bytez := bs.db.Get(calcBlockCommitKey(height))
	if bytez == nil {
		err = ErrNotFound
		return
	}
	bs.db.Delete(calcBlockCommitKey(height - 1))
	bs.db.Delete(calcSeenCommitKey(height))
	var n int
	r := bs.GetReader(calcBlockMetaKey(height))
	if r == nil {
		return nil
	}
	meta := wire.ReadBinary(&types.BlockMeta{}, r, 0, &n, &err).(*types.BlockMeta)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error reading block meta: %v", err))
		return
	}
	for i := 0; i < meta.PartsHeader.Total; i++ {
		bs.db.Delete(calcBlockPartKey(height, i))
	}
	bs.db.Delete(calcBlockMetaKey(height))
	bs.db.DeleteSync(nil)

	bs.archiveDB.Delete(calcBlockCommitKey(height - 1))
	bs.archiveDB.Delete(calcSeenCommitKey(height))
	r = bs.GetReaderFromArchive(calcBlockMetaKey(height))
	if r == nil {
		return nil
	}
	meta = wire.ReadBinary(&types.BlockMeta{}, r, 0, &n, &err).(*types.BlockMeta)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error reading archiveDB block meta: %v", err))
		return
	}
	for i := 0; i < meta.PartsHeader.Total; i++ {
		bs.archiveDB.Delete(calcBlockPartKey(height, i))
	}
	bs.archiveDB.Delete(calcBlockMetaKey(height))
	bs.archiveDB.DeleteSync(nil)

	return
}

func (bs *BlockStore) saveBlockPart(height int64, index int, part *types.Part) {
	if height != bs.Height()+1 {
		gcmn.PanicSanity(gcmn.Fmt("BlockStore can only save contiguous blocks. Wanted %v, got %v", bs.Height()+1, height))
	}
	partBytes := wire.BinaryBytes(part)
	bs.db.Set(calcBlockPartKey(height, index), partBytes)
}

func (bs *BlockStore) savePartToArchive(height int64, index int, part *types.Part) {
	partBytes := wire.BinaryBytes(part)
	bs.archiveDB.Set(calcBlockPartKey(height, index), partBytes)
}

func (bs *BlockStore) RevertToHeight(height int64) {
	BlockStoreStateJSON{Height: height, OriginHeight: bs.originHeight}.Save(bs.db)
}

//-----------------------------------------------------------------------------

func calcBlockMetaKey(height int64) []byte {
	return []byte(fmt.Sprintf("H:%v", height))
}

func calcBlockPartKey(height int64, partIndex int) []byte {
	return []byte(fmt.Sprintf("P:%v:%v", height, partIndex))
}

func calcBlockCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("C:%v", height))
}

func calcSeenCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("SC:%v", height))
}

//-----------------------------------------------------------------------------

var blockStoreKey = []byte("blockStore")

type BlockStoreStateJSON struct {
	Height       int64
	OriginHeight int64
}

func (bsj BlockStoreStateJSON) Save(db dbm.DB) {
	bsj.SaveByKey(blockStoreKey, db)
}

func (bsj BlockStoreStateJSON) SaveByKey(key []byte, db dbm.DB) {
	bytes, err := json.Marshal(bsj)
	if err != nil {
		gcmn.PanicSanity(gcmn.Fmt("Could not marshal state bytes: %v", err))
	}
	db.SetSync(key, bytes)
}

func LoadBlockStoreStateJSON(db dbm.DB) BlockStoreStateJSON {
	bytes := db.Get(blockStoreKey)
	if bytes == nil {
		return BlockStoreStateJSON{
			Height:       0,
			OriginHeight: 0,
		}
	}
	bsj := BlockStoreStateJSON{}
	err := json.Unmarshal(bytes, &bsj)
	if err != nil {
		gcmn.PanicCrisis(gcmn.Fmt("Could not unmarshal bytes: %X", bytes))
	}
	return bsj
}
