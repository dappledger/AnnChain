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

package state

import (
	"time"
)

type blockExeInfo struct {
	txExecuted uint32
	timeCost   time.Duration
}

type TPSCalculator struct {
	count  uint32
	offset uint32
	data   []blockExeInfo
	lastR  time.Time
}

func NewTPSCalculator(count uint32) *TPSCalculator {
	return &TPSCalculator{
		count: count,
		lastR: time.Now(),
		data:  make([]blockExeInfo, count),
	}
}

func (c *TPSCalculator) AddRecord(txExcuted uint32) {
	now := time.Now()
	if c.offset == c.count {
		c.offset = 0
	}

	c.data[c.offset] = blockExeInfo{
		txExecuted: txExcuted,
		timeCost:   now.Sub(c.lastR),
	}
	c.offset++
	c.lastR = now
}

func (c *TPSCalculator) TPS() int {
	var totalTime time.Duration
	var totalExecuted uint32
	for _, v := range c.data {
		if v.timeCost.Nanoseconds() != 0 {
			totalTime += v.timeCost
			totalExecuted += v.txExecuted
		}
	}
	return int(float64(totalExecuted) / totalTime.Seconds())
}
