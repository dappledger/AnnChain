package types

import (
	"sync"
)

type (
	Hooker interface {
		Sync(int, int, *Block)
		Async(int, int, *Block)
		Result() interface{}
	}

	Hook struct {
		wg       sync.WaitGroup
		done     chan struct{}
		res      interface{}
		callback func(height, round int, block *Block) error
	}

	Hooks struct {
		OnNewRound  *Hook
		OnPropose   *Hook
		OnPrevote   *Hook
		OnPrecommit *Hook
		OnCommit    *Hook
	}

	CommitResult struct {
		AppHash      []byte
		ReceiptsHash []byte
	}
)

func NewHook(cb func(int, int, *Block) error) *Hook {
	return &Hook{
		callback: cb,
		done:     make(chan struct{}, 1),
	}
}

func (h *Hook) Result() interface{} {
	<-h.done
	return h.res
}

func (h *Hook) Sync(height, round int, block *Block) {
	h.wg.Add(1)
	go func() {
		h.res = h.callback(height, round, block)
		h.wg.Done()
		h.done <- struct{}{}
	}()
	h.wg.Wait()
}

func (h *Hook) Async(height, round int, block *Block) {
	h.drain()
	go func() {
		h.res = h.callback(height, round, block)
		h.done <- struct{}{}
	}()
}

func (h *Hook) drain() {
	select {
	case <-h.done:
	default:
	}
}
