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

package plock

import (
	"math/rand"
	"sync"
	"time"
)

type (
	lockType int
	priority int

	PriorityLock struct {
		// lockMutex protect the lockList
		lockMutex sync.Mutex

		// lockChann is the real concurrency control channel, make sure only one routine can hold the lock
		lockChann chan *lockItem

		// lockList buffer all the concurrent routines, waiting for their chance to hold the lock
		lockList []*lockItem
	}

	lockItem struct {
		// drainChann is used to block the routine when calling Lock()
		drainChann chan struct{}

		priorityStep    priority
		runningPriority priority
		lockType        lockType
	}
)

const (
	still lockType = iota
	dynamic
)

func (p priority) Less(a interface{}) bool {
	return p < a.(priority)
}

func newLockItem(t lockType) *lockItem {
	item := &lockItem{
		drainChann:      make(chan struct{}),
		priorityStep:    priority(t),
		runningPriority: 0,
		lockType:        t,
	}

	return item
}

func NewPriorityLock() *PriorityLock {
	l := &PriorityLock{
		lockMutex: sync.Mutex{},
		lockChann: make(chan *lockItem, 1),
		lockList:  make([]*lockItem, 0, 1024),
	}

	return l
}

func (pl *PriorityLock) Lock() {
	pl.lock(newLockItem(still))
}

func (pl *PriorityLock) LockDyn() {
	pl.lock(newLockItem(dynamic))
}

func (pl *PriorityLock) lock(item *lockItem) {
	select {
	case pl.lockChann <- item:
	default:
		pl.lockMutex.Lock()
		pl.lockList = append(pl.lockList, item)
		pl.lockMutex.Unlock()
		item.drainChann <- struct{}{}
	}
}

func (pl *PriorityLock) Unlock() {
	var item *lockItem
	select {
	case item = <-pl.lockChann:
	default:
		panic("unmatched unlock")
	}

	select {
	case <-item.drainChann:
	default:
	}
	close(item.drainChann)

	pl.lockMutex.Lock()
	defer pl.lockMutex.Unlock()

	if len(pl.lockList) == 0 {
		return
	}
	maxIdx := 0
	for i := 1; i < len(pl.lockList); i++ {
		if pl.lockList[i].lockType != still {
			pl.lockList[i].runningPriority += pl.lockList[i].priorityStep
		}
		if pl.lockList[maxIdx].runningPriority < pl.lockList[i].runningPriority {
			maxIdx = i
		}
	}

	if pl.lockList[maxIdx].runningPriority <= 5 {
		rand.Seed(time.Now().Unix())
		maxIdx = rand.Intn(len(pl.lockList))
	}
	pl.lockChann <- pl.lockList[maxIdx]
	_ = <-pl.lockList[maxIdx].drainChann
	if maxIdx == len(pl.lockList)-1 {
		pl.lockList = pl.lockList[0:maxIdx]
	} else {
		pl.lockList = append(pl.lockList[0:maxIdx], pl.lockList[maxIdx+1:]...)
	}
}

func (pl *PriorityLock) TryLock() {

}
