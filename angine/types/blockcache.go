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

package types

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-merkle"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

const MaxBlockSize = 22020096 // 21MB TODO make it configurable

type IBlockTick interface {
	Tick(*BlockCache)
}

type BlockCache struct {
	*pbtypes.Block

	datac   *DataCache
	commitc *CommitCache
}

// TODO: version
func MakeBlock(maker []byte, height def.INT, chainID string, alltxs []Tx, commit *CommitCache,
	prevBlockID pbtypes.BlockID, valHash, appHash, receiptsHash []byte, partSize, nonEmptyHeight def.INT) (*BlockCache, *PartSet) {
	extxs := []Tx{}
	txs := []Tx{}
	for _, tx := range alltxs {
		if IsExtendedTx(tx) {
			extxs = append(extxs, tx)
			continue
		}
		txs = append(txs, tx)
	}
	block := &pbtypes.Block{
		Header: &pbtypes.Header{
			ChainID:            chainID,
			Height:             height,
			Time:               time.Now().UnixNano(),
			NumTxs:             def.INT(len(alltxs)),
			Maker:              maker,
			LastBlockID:        &prevBlockID,
			ValidatorsHash:     valHash,
			AppHash:            appHash,      // state merkle root of txs from the previous block.
			ReceiptsHash:       receiptsHash, // receipts hash from the previous block
			LastNonEmptyHeight: nonEmptyHeight,
		},
		LastCommit: commit.Commit,
		Data: &pbtypes.Data{
			Txs:   Txs(txs).ToBytes(),
			ExTxs: Txs(extxs).ToBytes(),
		},
	}

	blockCache := &BlockCache{
		Block: block,
		datac: &DataCache{
			Data: block.GetData(),
		},
		commitc: commit,
	}

	blockCache.FillHeader()
	return blockCache, blockCache.MakePartSet(partSize)
}

func MakeBlockCache(block *pbtypes.Block) *BlockCache {
	if block.GetData() == nil || block.GetLastCommit() == nil {
		return nil
	}
	return &BlockCache{
		Block: block,
		datac: &DataCache{
			Data: block.GetData(),
		},
		commitc: &CommitCache{
			Commit: block.GetLastCommit(),
		},
	}
}

func (b *BlockCache) CommitCache() *CommitCache {
	return b.commitc
}

func (b *BlockCache) DataCache() *DataCache {
	return b.datac
}

// Basic validation that doesn't involve state data.
func (b *BlockCache) ValidateBasic(chainID string, lastBlockHeight def.INT, lastBlockID pbtypes.BlockID,
	lastBlockTime def.INT, appHash, receiptsHash []byte, nonEmptyHeight def.INT) error {
	bheader := b.GetHeader()
	bdata := b.datac
	bcommit := b.commitc
	if bheader.ChainID != chainID {
		return errors.New(Fmt("Wrong Block.Header.ChainID. Expected %v, got %v", chainID, bheader.ChainID))
	}
	if bheader.Height != lastBlockHeight+1 {
		return errors.New(Fmt("(%s) Wrong Block.Header.Height. Expected %v, got %v", chainID, lastBlockHeight+1, bheader.Height))
	}
	if bheader.LastNonEmptyHeight != nonEmptyHeight {
		return errors.New(Fmt("(%s) Wrong Block.Header.LastNonEmptyHeight. Expected %v, got %v", chainID, nonEmptyHeight, bheader.LastNonEmptyHeight))
	}
	/*	TODO: Determine bounds for Time
		See blockchain/reactor "stopSyncingDurationMinutes"

		if !b.Time.After(lastBlockTime) {
			return errors.New("Invalid Block.Header.Time")
		}
	*/
	if bheader.NumTxs != def.INT(len(bdata.Txs)+len(bdata.ExTxs)) {
		return errors.New(Fmt("(%s) Wrong Block.Header.NumTxs. Expected %v, got %v", chainID, len(bdata.Txs)+len(bdata.ExTxs), bheader.NumTxs))
	}
	if !bheader.LastBlockID.Equals(&lastBlockID) {
		return errors.New(Fmt("(%s) Wrong Block.Header.LastBlockID.  Expected %v, got %v", chainID, lastBlockID, (*bheader.LastBlockID)))
	}
	if !bytes.Equal(bheader.LastCommitHash, bcommit.Hash()) {
		return errors.New(Fmt("(%s) Wrong Block.Header.LastCommitHash.  Expected %X, got %X", chainID, bheader.LastCommitHash, bcommit.Hash()))
	}
	if bheader.Height != 1 {
		if err := bcommit.ValidateBasic(); err != nil {
			return err
		}
	}
	if !bytes.Equal(bheader.DataHash, bdata.Hash()) {
		return errors.New(Fmt("(%s) Wrong Block.Header.DataHash.  Expected %X, got %X", chainID, bheader.DataHash, bdata.Hash()))
	}
	if !bytes.Equal(bheader.AppHash, appHash) {
		return errors.New(Fmt("(%s) Wrong Block.Header.AppHash.  Expected %X, got %X", chainID, appHash, bheader.AppHash))
	}
	if !bytes.Equal(bheader.ReceiptsHash, receiptsHash) {
		return errors.New(Fmt("(%s) Wrong Block.Header.ReceiptsHash.  Expected %X, got %X", chainID, receiptsHash, bheader.ReceiptsHash))
	}
	// NOTE: the AppHash and ValidatorsHash are validated later.
	return nil
}

func (b *BlockCache) FillHeader() {
	bheader := b.Header
	if bheader.LastCommitHash == nil {
		bheader.LastCommitHash = b.commitc.Hash()
	}
	if bheader.DataHash == nil {
		bheader.DataHash = b.datac.Hash()
	}
}

// Computes and returns the block hash.
// If the block is incomplete, block hash is nil for safety.
func (b *BlockCache) Hash() []byte {
	// fmt.Println(">>", b.Data)
	if b.Header == nil || b.Data == nil || b.LastCommit == nil {
		return nil
	}
	b.FillHeader()
	return b.Header.Hash()
}

func (b *BlockCache) MakePartSet(partSize def.INT) *PartSet {
	bys, _ := MarshalData(b.Block)
	return NewPartSetFromData(bys, partSize)
}

// Convenience.
// A nil block never hashes to anything.
// Nothing hashes to a nil hash.
func (b *BlockCache) HashesTo(hash []byte) bool {
	if len(hash) == 0 {
		return false
	}
	if b == nil {
		return false
	}
	return bytes.Equal(b.Hash(), hash)
}

func (b *BlockCache) String() string {
	return b.StringIndented("")
}

func (b *BlockCache) StringIndented(indent string) string {
	if b == nil {
		return "nil-Block"
	}
	return fmt.Sprintf(`Block{
%s  %v
%s  %v
%s  %v
%s}#%X`,
		indent, b.Header.StringIndented(indent+"  "),
		indent, b.datac.StringIndented(indent+"  "),
		indent, b.commitc.StringIndented(indent+"  "),
		indent, b.Hash())
}

func (b *BlockCache) StringShort() string {
	if b == nil {
		return "nil-Block"
	}
	return fmt.Sprintf("Block#%X", b.Hash())
}

//-------------------------------------

// NOTE: Commit is empty for height 1, but never nil.
type CommitCache struct {
	// NOTE: The Precommits are in order of address to preserve the bonded ValidatorSet order.
	// Any peer with a block can gossip precommits by index with a peer without recalculating the
	// active ValidatorSet.

	*pbtypes.Commit
	// Volatile
	firstPrecommit *pbtypes.Vote
	hash           []byte
	bitArray       *BitArray
}

func NewCommitCache(c *pbtypes.Commit) *CommitCache {
	return &CommitCache{
		Commit: c,
	}
}

func (commit *CommitCache) FirstPrecommit() *pbtypes.Vote {
	if len(commit.Precommits) == 0 {
		return nil
	}
	if commit.firstPrecommit != nil {
		return commit.firstPrecommit
	}
	for _, precommit := range commit.Precommits {
		if precommit.Exist() {
			commit.firstPrecommit = precommit
			return precommit
		}
	}
	return nil
}

func (commit *CommitCache) Height() def.INT {
	if len(commit.Precommits) == 0 {
		return 0
	}
	return commit.FirstPrecommit().GetData().GetHeight()
}

func (commit *CommitCache) Round() def.INT {
	if len(commit.Precommits) == 0 {
		return 0
	}
	return commit.FirstPrecommit().GetData().GetRound()
}

func (commit *CommitCache) Type() pbtypes.VoteType {
	return pbtypes.VoteType_Precommit
}

func (commit *CommitCache) CSize() int {
	if commit == nil {
		return 0
	}
	return len(commit.Precommits)
}

func (commit *CommitCache) BitArray() *BitArray {
	if commit.bitArray == nil {
		commit.bitArray = NewBitArray(len(commit.Precommits))
		for i, precommit := range commit.Precommits {
			commit.bitArray.SetIndex(i, precommit.Exist())
		}
	}
	return commit.bitArray
}

func (commit *CommitCache) GetByIndex(index int) *pbtypes.Vote {
	return commit.Precommits[index]
}

func (commit *CommitCache) IsCommit() bool {
	if len(commit.Precommits) == 0 {
		return false
	}
	return true
}

func (commit *CommitCache) ValidateBasic() error {
	if commit.BlockID.IsZero() {
		return errors.New("Commit cannot be for nil block")
	}
	if len(commit.Precommits) == 0 {
		return errors.New("No precommits in commit")
	}
	height, round := commit.Height(), commit.Round()

	// validate the precommits
	for _, precommit := range commit.Precommits {
		// It's OK for precommits to be missing.
		if !precommit.Exist() {
			continue
		}
		pdata := precommit.GetData()
		// Ensure that all votes are precommits
		if pdata.Type != pbtypes.VoteType_Precommit {
			return fmt.Errorf("Invalid commit vote. Expected precommit, got %v",
				pdata.Type)
		}
		// Ensure that all heights are the same
		if pdata.Height != height {
			return fmt.Errorf("Invalid commit precommit height. Expected %v, got %v",
				height, pdata.Height)
		}
		// Ensure that all rounds are the same
		if pdata.Round != round {
			return fmt.Errorf("Invalid commit precommit round. Expected %v, got %v",
				round, pdata.Round)
		}
	}
	return nil
}

func (commit *CommitCache) Hash() []byte {
	if commit.hash == nil {
		bs := make([]interface{}, len(commit.Precommits))
		for i, precommit := range commit.Precommits {
			bs[i] = precommit
		}
		commit.hash = merkle.SimpleHashFromBinaries(bs)
	}
	return commit.hash
}

func (commit *CommitCache) StringIndented(indent string) string {
	if commit == nil {
		return "nil-Commit"
	}
	precommitStrings := make([]string, len(commit.Precommits))
	for i, precommit := range commit.Precommits {
		precommitStrings[i] = precommit.String()
	}
	return fmt.Sprintf(`Commit{
%s  BlockID:    %v
%s  Precommits: %v
%s}#%X`,
		indent, commit.BlockID,
		indent, strings.Join(precommitStrings, "\n"+indent+"  "),
		indent, commit.hash)
}

//-----------------------------------------------------------------------------

type DataCache struct {
	*pbtypes.Data

	hash []byte
}

func (data *DataCache) Hash() []byte {
	if data.hash == nil {
		txs := BytesToTxSlc(data.Txs)
		extxs := BytesToTxSlc(data.ExTxs)
		data.hash = merkle.SimpleHashFromTwoHashes(Txs(txs).Hash(), Txs(extxs).Hash())
	}
	return data.hash
}

func (data *DataCache) StringIndented(indent string) string {
	if data == nil {
		return "nil-Data"
	}
	txStrings := make([]string, MinInt(len(data.Txs)+len(data.ExTxs), 21))
	for i, tx := range append(data.Txs, data.ExTxs...) {
		if i == 20 {
			txStrings[i] = fmt.Sprintf("... (%v total)", len(data.Txs))
			break
		}
		txStrings[i] = fmt.Sprintf("Tx:%v", tx)
	}
	return fmt.Sprintf(`Data{
%s  %v
%s}#%X`,
		indent, strings.Join(txStrings, "\n"+indent+"  "),
		indent, data.hash)
}
