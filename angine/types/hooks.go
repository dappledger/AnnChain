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
	"sync"

	"github.com/dappledger/AnnChain/module/xlib/def"
)

type (
	Hooker interface {
		Sync(def.INT, def.INT, *BlockCache) (interface{}, error)
		Async(def.INT, def.INT, *BlockCache)
		Result() interface{}
	}

	HookCallbackFunc func(height, round def.INT, block *BlockCache) (interface{}, error)

	Hook struct {
		wg       sync.WaitGroup
		done     chan struct{}
		res      interface{}
		err      error
		callback HookCallbackFunc
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

func NewHook(cb HookCallbackFunc) *Hook {
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

func (h *Hook) Sync(height, round def.INT, block *BlockCache) (interface{}, error) {
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

func (h *Hook) Async(height, round def.INT, block *BlockCache, cb func(interface{}), onError func(error)) {
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
