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

	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine/plugin"
	"github.com/dappledger/AnnChain/angine/types"
	cmn "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	cfg "github.com/dappledger/AnnChain/ann-module/lib/go-config"
)

//--------------------------------------------------
// Execute the block

// Execute the block to mutate State.
// Validates block and then executes Data.Txs in the block.
func (s *State) ExecBlock(eventSwitch types.EventSwitch, block *types.Block, blockPartsHeader types.PartSetHeader, round int) error {
	// Validate the block.
	if err := s.validateBlock(block); err != nil {
		return ErrInvalidBlock(err)
	}

	// compute bitarray of validators that signed
	signed := commitBitArrayFromBlock(block)
	_ = signed // TODO send on begin block

	// copy the valset
	valSet := s.Validators.Copy()
	nextValSet := valSet.Copy()

	s.execBeginBlockOnPlugins(block)

	// Execute the block txs
	// changedValidators, err := execBlockOnProxyApp(eventCache, proxyAppConn, block, s)
	// if err != nil {
	// There was some error in proxyApp
	// TODO Report error and wait for proxyApp to be available.
	// return ErrProxyAppConn(err)
	// }

	changedValidators := make([]*types.ValidatorAttr, 0)
	s.execBlockOnApp(eventSwitch, block, round)
	// plugins modify changedValidators inplace
	s.execEndBlockOnPlugins(block, changedValidators, nextValSet)

	// All good!
	// Update validator accums and set state variables
	nextValSet.IncrementAccum(1)
	s.SetBlockAndValidators(block.Header, blockPartsHeader, valSet, nextValSet)

	// save state with updated height/blockhash/validators
	// but stale apphash, in case we fail between Commit and Save
	s.SaveIntermediate()

	return nil
}

func (s *State) execBlockOnApp(eventSwitch types.EventSwitch, block *types.Block, round int) ([]*types.ValidatorAttr, error) {
	// Run ExTxs of block
	for i, tx := range block.Data.ExTxs {
		s.deliverTxOnPlugins(tx, i)
	}
	ed := types.NewEventDataHookExecute(block.Height, round, block)
	types.FireEventHookExecute(eventSwitch, ed) // Run Txs of block
	res := <-ed.ResCh
	eventCache := types.NewEventCache(eventSwitch)
	for _, tx := range res.ValidTxs {
		txev := types.EventDataTx{
			Tx:   tx,
			Code: types.CodeType_OK,
		}
		types.FireEventTx(eventCache, txev)
	}
	for _, invalid := range res.InvalidTxs {
		txev := types.EventDataTx{
			Tx:    invalid.Bytes,
			Code:  types.CodeType_InvalidTx,
			Error: invalid.Error.Error(),
		}
		types.FireEventTx(eventCache, txev)
	}
	eventCache.Flush()

	if res.Error != nil {
		return nil, res.Error
	}

	s.Tpsc.AddRecord(uint32(len(res.ValidTxs)))

	tps := s.Tpsc.TPS()

	if s.logger != nil {
		s.logger.Info("Executed block",
			zap.Int("height", block.Height),
			zap.Int("txs", block.NumTxs),
			zap.Int("valid", len(res.ValidTxs)),
			zap.Int("invalid", len(res.InvalidTxs)),
			zap.Int("extended", len(block.Data.ExTxs)),
			zap.Int("tps", tps))
	}

	return nil, nil
}

// return a bit array of validators that signed the last commit
// NOTE: assumes commits have already been authenticated
func commitBitArrayFromBlock(block *types.Block) *cmn.BitArray {
	signed := cmn.NewBitArray(len(block.LastCommit.Precommits))
	for i, precommit := range block.LastCommit.Precommits {
		if precommit != nil {
			signed.SetIndex(i, true) // val_.LastCommitHeight = block.Height - 1
		}
	}
	return signed
}

//-----------------------------------------------------
// Validate block

func (s *State) ValidateBlock(block *types.Block) error {
	return s.validateBlock(block)
}

func (s *State) validateBlock(block *types.Block) error {
	// Basic block validation.
	err := block.ValidateBasic(s.ChainID, s.LastBlockHeight, s.LastBlockID, s.LastBlockTime, s.AppHash, s.ReceiptsHash)
	if err != nil {
		return err
	}

	// Validate block LastCommit.
	if block.Height == 1 {
		if len(block.LastCommit.Precommits) != 0 {
			return errors.New("Block at height 1 (first block) should have no LastCommit precommits")
		}
	} else {
		if len(block.LastCommit.Precommits) != s.LastValidators.Size() {
			return errors.New(cmn.Fmt("Invalid block commit size. Expected %v, got %v",
				s.LastValidators.Size(), len(block.LastCommit.Precommits)))
		}
		err := s.LastValidators.VerifyCommit(
			s.ChainID, s.LastBlockID, block.Height-1, block.LastCommit)
		if err != nil {
			return err
		}
	}

	return nil
}

// ApplyBlock executes the block, then commits and updates the mempool atomically
func (s *State) ApplyBlock(eventSwitch types.EventSwitch, block *types.Block, partsHeader types.PartSetHeader, mempool types.IMempool, round int) error {
	// Run the block on the State:
	// + update validator sets
	// + run txs on the proxyAppConn
	err := s.ExecBlock(eventSwitch, block, partsHeader, round)
	if err != nil {
		return errors.New(cmn.Fmt("Exec failed for application: %v", err))
	}
	// lock mempool, commit state, update mempoool
	err = s.CommitStateUpdateMempool(eventSwitch, block, mempool, round)
	if err != nil {
		return errors.New(cmn.Fmt("Commit failed for application: %v", err))
	}

	return nil
}

// mempool must be locked during commit and update
// because state is typically reset on Commit and old txs must be replayed
// against committed state before new txs are run in the mempool, lest they be invalid
func (s *State) CommitStateUpdateMempool(eventSwitch types.EventSwitch, block *types.Block, mempool types.IMempool, round int) error {
	mempool.Update(int64(block.Height), append(block.Txs, block.ExTxs...))
	ed := types.NewEventDataHookCommit(block.Height, round, block)
	types.FireEventHookCommit(eventSwitch, ed)
	res := <-ed.ResCh
	s.AppHash = res.AppHash
	s.ReceiptsHash = res.ReceiptsHash
	return nil
}

//----------------------------------------------------------------
// Handshake with app to sync to latest state of core by replaying blocks

// TODO: Should we move blockchain/store.go to its own package?
type BlockStore interface {
	Height() int
	LoadBlock(height int) *types.Block
	LoadBlockMeta(height int) *types.BlockMeta
}

type Handshaker struct {
	config cfg.Config
	state  *State
	store  BlockStore

	nBlocks int // number of blocks applied to the state
}

func NewHandshaker(config cfg.Config, state *State, store BlockStore) *Handshaker {
	return &Handshaker{config, state, store, 0}
}

// TODO: retry the handshake/replay if it fails ?
func (h *Handshaker) Handshake() error {
	// blockHeight := int(res.LastBlockHeight)
	// appHash := res.LastBlockAppHash

	// TODO: check version

	// replay blocks up to the latest in the blockstore
	// err = h.ReplayBlocks(appHash, blockHeight)
	// if err != nil {
	// 	return errors.New(Fmt("Error on replay: %v", err))
	// }

	// Save the state
	// h.state.Save()

	// TODO: (on restart) replay mempool

	return nil
}

func (s *State) execBeginBlockOnPlugins(block *types.Block) {
	params := &plugin.BeginBlockParams{Block: block}
	for _, p := range s.Plugins {
		p.BeginBlock(params)
	}
}

// return bool false will not go into proxyapp
func (s *State) deliverTxOnPlugins(tx []byte, i int) bool {
	ret := true
	for _, p := range s.Plugins {
		passon, _ := p.DeliverTx(tx, i)
		ret = ret && passon
	}
	return ret
}

func (s *State) deliverBlockOnPlugins(block *types.Block) {
	for i, tx := range block.Txs {
		for _, p := range s.Plugins {
			p.DeliverTx(tx, i)
		}
	}
}

func (s *State) execEndBlockOnPlugins(block *types.Block, valattrs []*types.ValidatorAttr, n *types.ValidatorSet) {
	params := &plugin.EndBlockParams{Block: block, ChangedValidators: valattrs, NextValidatorSet: n}
	for _, p := range s.Plugins {
		p.EndBlock(params)
	}
}
