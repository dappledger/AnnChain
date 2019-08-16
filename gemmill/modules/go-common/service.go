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

/*

Classical-inheritance-style service declarations.
Services can be started, then stopped, then optionally restarted.
Users can override the OnStart/OnStop methods.
By default, these methods are guaranteed to be called at most once.
A call to Reset will panic, unless OnReset is overwritten, allowing OnStart/OnStop to be called again.
Caller must ensure that Start() and Stop() are not called concurrently.
It is ok to call Stop() without calling Start() first.
Services cannot be re-started unless OnReset is overwritten to allow it.

Typical usage:

type FooService struct {
	BaseService
	// private fields
}

func NewFooService() *FooService {
	fs := &FooService{
		// init
	}
	fs.BaseService = *NewBaseService(log, "FooService", fs)
	return fs
}

func (fs *FooService) OnStart() error {
	fs.BaseService.OnStart() // Always call the overridden method.
	// initialize private fields
	// start subroutines, etc.
}

func (fs *FooService) OnStop() error {
	fs.BaseService.OnStop() // Always call the overridden method.
	// close/destroy private fields
	// stop subroutines, etc.
}

*/
package common

import (
	"sync/atomic"

	log "github.com/dappledger/AnnChain/gemmill/modules/go-log"
)

type Service interface {
	Start() (bool, error)
	OnStart() error

	Stop() bool
	OnStop()

	Reset() (bool, error)
	OnReset() error

	IsRunning() bool

	String() string
}

type BaseService struct {
	name    string
	started uint32 // atomic
	stopped uint32 // atomic
	Quit    chan struct{}

	// The "subclass" of BaseService
	impl Service
}

func NewBaseService(name string, impl Service) *BaseService {
	return &BaseService{
		name: name,
		Quit: make(chan struct{}),
		impl: impl,
	}
}

// Implements Servce
func (bs *BaseService) Start() (bool, error) {
	if atomic.CompareAndSwapUint32(&bs.started, 0, 1) {
		if atomic.LoadUint32(&bs.stopped) == 1 {
			log.Warn(Fmt("Not starting %v -- already stopped", bs.name))
			return false, nil
		} else {
			log.Debug(Fmt("Starting %v", bs.name))
		}
		err := bs.impl.OnStart()
		return true, err
	} else {
		log.Debug(Fmt("Not starting %v -- already started", bs.name))
		return false, nil
	}
}

// Implements Service
// NOTE: Do not put anything in here,
// that way users don't need to call BaseService.OnStart()
func (bs *BaseService) OnStart() error { return nil }

// Implements Service
func (bs *BaseService) Stop() bool {
	if atomic.CompareAndSwapUint32(&bs.stopped, 0, 1) {
		log.Info(Fmt("Stopping %v", bs.name))
		bs.impl.OnStop()
		close(bs.Quit)
		return true
	} else {
		log.Debug(Fmt("Stopping %v (ignoring: already stopped)", bs.name))
		return false
	}
}

// Implements Service
// NOTE: Do not put anything in here,
// that way users don't need to call BaseService.OnStop()
func (bs *BaseService) OnStop() {}

// Implements Service
func (bs *BaseService) Reset() (bool, error) {
	if atomic.CompareAndSwapUint32(&bs.stopped, 1, 0) {
		// whether or not we've started, we can reset
		atomic.CompareAndSwapUint32(&bs.started, 1, 0)

		return true, bs.impl.OnReset()
	} else {
		log.Debug(Fmt("Can't reset %v. Not stopped", bs.name))
		return false, nil
	}
	// never happens
	return false, nil
}

// Implements Service
func (bs *BaseService) OnReset() error {
	PanicSanity("The service cannot be reset")
	return nil
}

// Implements Service
func (bs *BaseService) IsRunning() bool {
	return atomic.LoadUint32(&bs.started) == 1 && atomic.LoadUint32(&bs.stopped) == 0
}

func (bs *BaseService) Wait() {
	<-bs.Quit
}

// Implements Servce
func (bs *BaseService) String() string {
	return bs.name
}

//----------------------------------------

type QuitService struct {
	BaseService
}

func NewQuitService(name string, impl Service) *QuitService {
	log.Warn("QuitService is deprecated, use BaseService instead")
	return &QuitService{
		BaseService: *NewBaseService(name, impl),
	}
}
