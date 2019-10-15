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
	"fmt"

	"go.uber.org/zap"

	cfg "github.com/dappledger/AnnChain/gemmill/config"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/types"
)

var (
	tpsc = NewTPSCalculator(10)
)

//--------------------------------------------------
// Execute the block

// Execute the block to mutate State.
// Validates block and then executes Data.Txs in the block.
func (s *State) ExecBlock(eventSwitch types.EventSwitch, block *types.Block, blockPartsHeader types.PartSetHeader, round int64) error {
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
	changedValidators := make([]*types.ValidatorAttr, 0)

	err := s.blockExecutable.BeginBlock(block, eventSwitch, &blockPartsHeader)
	if err != nil {
		return err
	}

	if _, err := s.execBlockOnApp(eventSwitch, block, round); err != nil {
		return err
	}

	err = s.blockExecutable.EndBlock(block, eventSwitch, &blockPartsHeader, changedValidators, nextValSet)
	if err != nil {
		return err
	}
	// plugins modify changedValidators inplace
	// All good!
	// Update validator accums and set state variables
	nextValSet.IncrementAccum(1)
	s.SetBlockAndValidators(block.Header, blockPartsHeader, valSet, nextValSet)

	// save state with updated height/blockhash/validators
	// but stale apphash, in case we fail between Commit and Save
	s.SaveIntermediate()

	return nil
}

func (s *State) execBlockOnApp(eventSwitch types.EventSwitch, block *types.Block, round int64) ([]*types.ValidatorAttr, error) {
	// Run ExTxs of block
	ed := types.NewEventDataHookExecute(block.Height, round, block)
	types.FireEventHookExecute(eventSwitch, ed) // Run Txs of block
	res := <-ed.ResCh
	eventCache := types.NewEventCache(eventSwitch)
	err := s.blockExecutable.ExecBlock(block, eventCache, &res)
	if err != nil {
		return nil, err
	}

	go func() {
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
	}()

	if res.Error != nil {
		return nil, res.Error
	}

	tpsc.AddRecord(uint32(len(res.ValidTxs) + len(res.InvalidTxs)))
	tps := tpsc.TPS()

	log.Info("Executed block",
		zap.Int64("height", block.Height),
		zap.Int64("txs", block.NumTxs),
		zap.Int("valid", len(res.ValidTxs)),
		zap.Int("invalid", len(res.InvalidTxs)),
		zap.Int("extended", len(block.Data.ExTxs)),
		zap.Int("tps", tps),
		zap.String("receiptHash", fmt.Sprintf("%X", block.Header.ReceiptsHash)))

	return nil, nil
}

// return a bit array of validators that signed the last commit
// NOTE: assumes commits have already been authenticated
func commitBitArrayFromBlock(block *types.Block) *gcmn.BitArray {
	signed := gcmn.NewBitArray(len(block.LastCommit.Precommits))
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
	return s.blockVerifier.ValidateBlock(block)
}

// ApplyBlock executes the block, then commits and updates the mempool atomically
func (s *State) ApplyBlock(eventSwitch types.EventSwitch, block *types.Block, partsHeader types.PartSetHeader, mempool types.IMempool, round int64) error {
	// Run the block on the State:
	// + update validator sets
	// + run txs on the proxyAppConn
	err := s.ExecBlock(eventSwitch, block, partsHeader, round)
	if err != nil {
		return errors.New(gcmn.Fmt("Exec failed for application: %v", err))
	}
	// lock mempool, commit state, update mempoool
	err = s.CommitStateUpdateMempool(eventSwitch, block, mempool, round)
	if err != nil {
		return errors.New(gcmn.Fmt("Commit failed for application: %v", err))
	}

	return nil
}

// mempool must be locked during commit and update
// because state is typically reset on Commit and old txs must be replayed
// against committed state before new txs are run in the mempool, lest they be invalid
func (s *State) CommitStateUpdateMempool(eventSwitch types.EventSwitch, block *types.Block, mempool types.IMempool, round int64) error {
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
	// 	return errors.New(gcmn.Fmt("Error on replay: %v", err))
	// }

	// Save the state
	// h.state.Save()

	// TODO: (on restart) replay mempool

	return nil
}
