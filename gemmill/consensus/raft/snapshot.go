package raft

import (
	"fmt"
	"github.com/hashicorp/raft"
)

type BlockChainSnapshot struct {
	Height int64
	Hash   []byte
}

func (s *BlockChainSnapshot) Persist(sink raft.SnapshotSink) error {
	_, err := sink.Write([]byte(fmt.Sprintf("%d-%x", s.Height, s.Hash)))
	return err
}

func (s *BlockChainSnapshot) Release() {
	// nothing to do
}
