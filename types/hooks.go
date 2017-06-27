package types

import (
	"sync"
)

type (
	Hooker interface {
		Sync(int, int, *Block) (interface{}, error)
		Async(int, int, *Block)
		Result() interface{}
	}

	Hook struct {
		wg       sync.WaitGroup
		done     chan struct{}
		res      interface{}
		err      error
		callback func(height, round int, block *Block) (interface{}, error)
	}

	Hooks struct {
		OnNewRound  *Hook
		OnPropose   *Hook
		OnPrevote   *Hook
		OnPrecommit *Hook
		OnCommit    *Hook
		OnExecute   *Hook
	}
)

func NewHook(cb func(int, int, *Block) (interface{}, error)) *Hook {
	return &Hook{
		callback: cb,
		done:     make(chan struct{}, 1),
	}
}

func (h *Hook) Result() interface{} {
	<-h.done
	return h.res
}

func (h *Hook) Error() error {
	return h.err
}

func (h *Hook) Sync(height, round int, block *Block) (interface{}, error) {
	h.res = nil
	h.err = nil
	h.drain()
	h.wg.Add(1)
	go func() {
		h.res, h.err = h.callback(height, round, block)
		h.wg.Done()
		h.done <- struct{}{}
	}()
	h.wg.Wait()

	return h.res, h.err
}

func (h *Hook) Async(height, round int, block *Block, cb func(interface{}), onError func(error)) {
	var (
		res interface{}
		err error
	)
	h.drain()
	go func() {
		res, err = h.callback(height, round, block)
		if err != nil && onError != nil {
			onError(err)
		} else {
			if cb != nil {
				cb(res)
			}
		}
		h.done <- struct{}{}
	}()
}

func (h *Hook) drain() {
	select {
	case <-h.done:
	default:
	}
}
