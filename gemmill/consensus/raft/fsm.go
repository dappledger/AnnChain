package raft

import (
	"bytes"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/hashicorp/raft"

	"github.com/dappledger/AnnChain/gemmill/blockchain"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	"github.com/dappledger/AnnChain/gemmill/state"
	"github.com/dappledger/AnnChain/gemmill/types"

	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
)

type BlockChainFSM struct {
	conf            *config
	evsw            types.EventSwitch
	blockExecutable state.IBlockExecutable
	mempool         types.TxPool
	state           *state.State
	appliedCh       chan *types.Block
	blockStore      *blockchain.BlockStore
	mut             sync.Mutex
}

func newBlockChainFSM(conf *config, mempool types.TxPool, blockStore *blockchain.BlockStore, state *state.State) *BlockChainFSM {

	return &BlockChainFSM{
		conf:       conf,
		mempool:    mempool,
		state:      state,
		appliedCh:  make(chan *types.Block),
		blockStore: blockStore,
	}
}

func (b *BlockChainFSM) SetEventSwitch(evsw types.EventSwitch) {
	b.evsw = evsw
}

func (b *BlockChainFSM) Apply(l *raft.Log) interface{} {

	start := time.Now()
	r := bytes.NewBuffer(l.Data)
	var err error
	n := 0
	block := wire.ReadBinary(&types.Block{}, r, types.MaxBlockSize, &n, &err).(*types.Block)
	partSet := types.NewPartSetFromData(l.Data, b.conf.blockPartSize)

	log.Debug("FSM.Apply start apply", zap.Int64("height", block.Height))
	if b.blockStore.Height()+1 != block.Height {
		log.Warn("FSM.Apply, found dup block", zap.Int64("height", block.Height))
		return nil
	}
	b.blockStore.SaveBlock(block, partSet, &types.Commit{})

	stateCopy := b.state.Copy()
	if err := stateCopy.ApplyBlock(b.evsw, block, partSet.Header(), b.mempool, 0); err != nil {
		log.Error("FSM.Apply state ApplyBlock", zap.Int64("height", block.Height), zap.Error(err))
		return err
	}
	stateCopy.Save()

	b.state = stateCopy
	end := time.Now()
	log.Info("BlockChainFSM Applied Block", zap.Int64("height", block.Height), zap.Int("txs num", len(block.Data.Txs)), zap.Duration("taken", end.Sub(start)))

	b.appliedCh <- block
	return nil
}

func (b *BlockChainFSM) createProposalBlock(proposerAddr []byte) *types.Block {

	txs := b.mempool.Reap(b.conf.blockSize)

	if len(txs) < b.conf.blockSize {
		t1 := time.NewTimer(b.conf.emptyBlockInterval)
	L1:
		for {
			select {
			case <-t1.C:
				break L1
			case <-time.After(time.Millisecond * 100):
				txs = b.mempool.Reap(b.conf.blockSize)
				if len(txs) > 0 {
					t1.Stop()
					break L1
				}
			}
		}
	}

	h := b.state.LastBlockHeight + 1

	block, _ := types.MakeBlock(h, b.state.ChainID, txs, nil, &types.Commit{}, proposerAddr,
		b.state.LastBlockID, b.state.Validators.Hash(), b.state.AppHash, b.state.ReceiptsHash, b.conf.blockPartSize)
	return block
}

func (b *BlockChainFSM) Snapshot() (raft.FSMSnapshot, error) {

	return &BlockChainSnapshot{
		Height: b.state.LastBlockHeight,
		Hash:   b.state.LastBlockID.Hash,
	}, nil
}

func (b *BlockChainFSM) Restore(r io.ReadCloser) error {

	io.Copy(ioutil.Discard, r)
	return nil
}

func (b *BlockChainFSM) AppliedCh() <-chan *types.Block {
	return b.appliedCh
}
