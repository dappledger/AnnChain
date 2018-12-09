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

package common

import (
	"sync"
	"time"
)

/*
ThrottleTimer fires an event at most "dur" after each .Set() call.
If a short burst of .Set() calls happens, ThrottleTimer fires once.
If a long continuous burst of .Set() calls happens, ThrottleTimer fires
at most once every "dur".
*/
type ThrottleTimer struct {
	Name string
	Ch   chan struct{}
	quit chan struct{}
	dur  time.Duration

	mtx   sync.Mutex
	timer *time.Timer
	isSet bool
}

func NewThrottleTimer(name string, dur time.Duration) *ThrottleTimer {
	var ch = make(chan struct{})
	var quit = make(chan struct{})
	var t = &ThrottleTimer{Name: name, Ch: ch, dur: dur, quit: quit}
	t.mtx.Lock()
	t.timer = time.AfterFunc(dur, t.fireRoutine)
	t.mtx.Unlock()
	t.timer.Stop()
	return t
}

func (t *ThrottleTimer) fireRoutine() {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	select {
	case t.Ch <- struct{}{}:
		t.isSet = false
	case <-t.quit:
		// do nothing
	default:
		t.timer.Reset(t.dur)
	}
}

func (t *ThrottleTimer) Set() {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	if !t.isSet {
		t.isSet = true
		t.timer.Reset(t.dur)
	}
}

func (t *ThrottleTimer) Unset() {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.isSet = false
	t.timer.Stop()
}

// For ease of .Stop()'ing services before .Start()'ing them,
// we ignore .Stop()'s on nil ThrottleTimers
func (t *ThrottleTimer) Stop() bool {
	if t == nil {
		return false
	}
	close(t.quit)
	t.mtx.Lock()
	defer t.mtx.Unlock()
	return t.timer.Stop()
}
