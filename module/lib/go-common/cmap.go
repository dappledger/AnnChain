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

import "sync"

// CMap is a goroutine-safe map
type CMap struct {
	m map[string]interface{}
	l sync.Mutex
}

func NewCMap() *CMap {
	return &CMap{
		m: make(map[string]interface{}, 0),
	}
}

func (cm *CMap) Set(key string, value interface{}) {
	cm.l.Lock()
	defer cm.l.Unlock()
	cm.m[key] = value
}

func (cm *CMap) Get(key string) interface{} {
	cm.l.Lock()
	defer cm.l.Unlock()
	return cm.m[key]
}

func (cm *CMap) Has(key string) bool {
	cm.l.Lock()
	defer cm.l.Unlock()
	_, ok := cm.m[key]
	return ok
}

func (cm *CMap) Delete(key string) {
	cm.l.Lock()
	defer cm.l.Unlock()
	delete(cm.m, key)
}

func (cm *CMap) Size() int {
	cm.l.Lock()
	defer cm.l.Unlock()
	return len(cm.m)
}

func (cm *CMap) Clear() {
	cm.l.Lock()
	defer cm.l.Unlock()
	cm.m = make(map[string]interface{}, 0)
}

func (cm *CMap) Values() []interface{} {
	cm.l.Lock()
	defer cm.l.Unlock()
	items := []interface{}{}
	for _, v := range cm.m {
		items = append(items, v)
	}
	return items
}
