package state

import (
	"errors"

	. "gitlab.zhonganonline.com/ann/ann-module/lib/go-common"
	cfg "gitlab.zhonganonline.com/ann/ann-module/lib/go-config"
	"gitlab.zhonganonline.com/ann/angine/plugin"
	"gitlab.zhonganonline.com/ann/angine/types"
)

//--------------------------------------------------
// Execute the block

// Execute the block to mutate State.
// Validates block and then executes Data.Txs in the block.
func (s *State) ExecBlock(eventCache types.Fireable, block *types.Block, blockPartsHeader types.PartSetHeader) error {
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

// Executes block's transactions on proxyAppConn.
// Returns a list of updates to the validator set
// TODO: Generate a bitmap or otherwise store tx validity in state.
// func execBlockOnProxyApp(eventCache types.Fireable, proxyAppConn proxy.AppConnConsensus, block *types.Block, state *State) ([]*abci.Validator, error) {

// 	var validTxs, invalidTxs = 0, 0

// 	// Execute transactions and get hash
// 	proxyCb := func(req *abci.Request, res *abci.Response) {
// 		switch r := res.Value.(type) {
// 		case *abci.Response_DeliverTx:
// 			// TODO: make use of res.Log
// 			// TODO: make use of this info
// 			// Blocks may include invalid txs.
// 			txError := ""
// 			apTx := r.DeliverTx
// 			if apTx.Code == abci.CodeType_OK {
// 				validTxs++
// 			} else {
// 				logger.Debug("Invalid tx", "code", r.DeliverTx.Code, "log", r.DeliverTx.Log, "data", r.DeliverTx.Data)
// 				invalidTxs++
// 				txError = apTx.Code.String()
// 			}
// 			// NOTE: if we count we can access the tx from the block instead of
// 			// pulling it from the req
// 			event := types.EventDataTx{
// 				Tx:    req.GetDeliverTx().Tx,
// 				Data:  apTx.Data,
// 				Code:  apTx.Code,
// 				Log:   apTx.Log,
// 				Error: txError,
// 			}
// 			types.FireEventTx(eventCache, event)
// 		}
// 	}
// 	proxyAppConn.SetResponseCallback(proxyCb)

// 	// Begin block
// 	err := proxyAppConn.BeginBlockSync(block.Hash(), types.TM2PB.Header(block.Header))
// 	if err != nil {
// 		log.Warn("Error in proxyAppConn.BeginBlock", "error", err)
// 		return nil, err
// 	}
// 	// Run txs of block
// 	for i, tx := range block.Txs {
// 		if !state.deliverTxOnPlugins(tx, i) {
// 			validTxs++
// 			continue // special op will bypass app
// 		}

// 		proxyAppConn.DeliverTxAsync(tx, i)
// 		if err := proxyAppConn.Error(); err != nil {
// 			return nil, err
// 		}
// 	}

// 	// End block
// 	respEndBlock, err := proxyAppConn.EndBlockSync(uint64(block.Height))
// 	if err != nil {
// 		log.Warn("Error in proxyAppConn.EndBlock", "error", err)
// 		return nil, err
// 	}

// 	logger.Info("Executed block", "height", block.Height, "valid txs", validTxs, "invalid txs", invalidTxs)

// 	return respEndBlock.Diffs, nil
// }

// return a bit array of validators that signed the last commit
// NOTE: assumes commits have already been authenticated
func commitBitArrayFromBlock(block *types.Block) *BitArray {
	signed := NewBitArray(len(block.LastCommit.Precommits))
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
			return errors.New(Fmt("Invalid block commit size. Expected %v, got %v",
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
func (s *State) ApplyBlock(eventCache types.Fireable, block *types.Block, partsHeader types.PartSetHeader, mempool types.IMempool) error {
	// Run the block on the State:
	// + update validator sets
	// + run txs on the proxyAppConn
	err := s.ExecBlock(eventCache, block, partsHeader)
	if err != nil {
		return errors.New(Fmt("Exec failed for application: %v", err))
	}
	// lock mempool, commit state, update mempoool
	err = s.CommitStateUpdateMempool(block, mempool)
	if err != nil {
		return errors.New(Fmt("Commit failed for application: %v", err))
	}
	return nil
}

// mempool must be locked during commit and update
// because state is typically reset on Commit and old txs must be replayed
// against committed state before new txs are run in the mempool, lest they be invalid
func (s *State) CommitStateUpdateMempool(block *types.Block, mempool types.IMempool) error {
	mempool.Update(block.Height, block.Txs)
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
