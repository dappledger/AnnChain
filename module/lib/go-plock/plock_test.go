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
	"fmt"
	"sync"
	"testing"
	// "time"
)

func TestLockUnlock(t *testing.T) {
	want := 100
	a := 0
	pl := NewPriorityLock()
	wg := sync.WaitGroup{}
	for i := 0; i < want; i++ {
		go func(id int) {
			wg.Add(1)
			pl.Lock()
			a++
			// fmt.Println("I'm sleeping now: ", id)
			// time.Sleep(1 * time.Second)
			// fmt.Println("I've slept for 1s: ", id)
			pl.Unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()

	fmt.Println("a=", a)
	if a != want {
		t.Fail()
	}
}
