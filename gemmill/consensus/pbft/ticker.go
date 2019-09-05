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

package pbft

import (
	"time"

	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	log "github.com/dappledger/AnnChain/gemmill/modules/go-log"
)

var (
	tickTockBufferSize = 10
)

// TimeoutTicker is a timer that schedules timeouts
// conditional on the height/round/step in the timeoutInfo.
// The timeoutInfo.Duration may be non-positive.
type TimeoutTicker interface {
	Start() (bool, error)
	Stop() bool
	Chan() <-chan timeoutInfo       // on which to receive a timeout
	ScheduleTimeout(ti timeoutInfo) // reset the timer
}

// timeoutTicker wraps time.Timer,
// scheduling timeouts only for greater height/round/step
// than what it's already seen.
// Timeouts are scheduled along the tickChan,
// and fired on the tockChan.
type timeoutTicker struct {
	gcmn.BaseService

	timer    *time.Timer
	tickChan chan timeoutInfo
	tockChan chan timeoutInfo
}

func NewTimeoutTicker() TimeoutTicker {
	tt := &timeoutTicker{
		timer:    time.NewTimer(0),
		tickChan: make(chan timeoutInfo, tickTockBufferSize),
		tockChan: make(chan timeoutInfo, tickTockBufferSize),
	}
	tt.stopTimer() // don't want to fire until the first scheduled timeout
	tt.BaseService = *gcmn.NewBaseService("TimeoutTicker", tt)
	return tt
}

func (t *timeoutTicker) OnStart() error {
	t.BaseService.OnStart()

	go t.timeoutRoutine()

	return nil
}

func (t *timeoutTicker) OnStop() {
	t.BaseService.OnStop()
	t.stopTimer()
}

func (t *timeoutTicker) Chan() <-chan timeoutInfo {
	return t.tockChan
}

// The timeoutRoutine is alwaya available to read from tickChan (it won't block).
// The scheduling may fail if the timeoutRoutine has already scheduled a timeout for a later height/round/step.
func (t *timeoutTicker) ScheduleTimeout(ti timeoutInfo) {
	t.tickChan <- ti
}

//-------------------------------------------------------------

// stop the timer and drain if necessary
func (t *timeoutTicker) stopTimer() {
	// Stop() returns false if it was already fired or was stopped
	if !t.timer.Stop() {
		select {
		case <-t.timer.C:
		default:
			log.Debug("Timer already stopped")
		}
	}
}

// send on tickChan to start a new timer.
// timers are interupted and replaced by new ticks from later steps
// timeouts of 0 on the tickChan will be immediately relayed to the tockChan
func (t *timeoutTicker) timeoutRoutine() {
	log.Debug("Starting timeout routine")
	var ti timeoutInfo
	for {
		select {
		case newti := <-t.tickChan:
			// ignore tickers for old height/round/step
			if newti.Height < ti.Height {
				continue
			} else if newti.Height == ti.Height {
				if newti.Round < ti.Round {
					continue
				} else if newti.Round == ti.Round {
					if ti.Step > 0 && newti.Step <= ti.Step {
						continue
					}
				}
			}

			// stop the last timer
			t.stopTimer()

			// update timeoutInfo and reset timer
			// NOTE time.Timer allows duration to be non-positive
			ti = newti
			t.timer.Reset(ti.Duration)
			log.Debugw("Scheduled timeout", "dur", ti.Duration, "height", ti.Height, "round", ti.Round, "step", ti.Step)
		case <-t.timer.C:
			log.Infow("Timed out", "dur", ti.Duration, "height", ti.Height, "round", ti.Round, "step", ti.Step)
			// go routine here gaurantees timeoutRoutine doesn't block.
			// Determinism comes from playback in the receiveRoutine.
			// We can eliminate it by merging the timeoutRoutine into receiveRoutine
			//  and managing the timeouts ourselves with a millisecond ticker
			go func(toi timeoutInfo) { t.tockChan <- toi }(ti)
		case <-t.Quit:
			return
		}
	}
}
