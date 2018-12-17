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

package session

import (
	"time"
)

type Session struct {
	checkTime  int64
	expireTime int64
	exitChan   chan bool
	mMap       *Map
}

type SessionValue struct {
	CreateTime int64
	Value      interface{}
}

func NewSession(expireTime, checkTime int64) *Session {

	s := &Session{
		mMap:       new(Map),
		expireTime: expireTime,
		checkTime:  checkTime,
		exitChan:   make(chan bool, 0),
	}

	go func() {
		chanCheck := time.NewTicker(time.Duration(s.checkTime))
		defer chanCheck.Stop()
		for {
			select {
			case <-chanCheck.C:
				s.mMap.RLockRange(func(key interface{}, value interface{}) {
					if time.Now().Unix()-value.(SessionValue).CreateTime >= s.expireTime {
						s.mMap.UnsafeDel(key)
					}
				})
			case <-s.exitChan:
				return
			}

		}
	}()

	return s
}

func (s *Session) SetSession(key, value interface{}) {
	s.mMap.Set(key, SessionValue{
		Value:      value,
		CreateTime: time.Now().Unix(),
	})
}

func (s *Session) GetSessionAndDelete(key interface{}) interface{} {
	sv := s.mMap.GetAndDel(key)
	if sv == nil {
		return nil
	}
	return sv.(SessionValue).Value
}

func (s *Session) GetSession(key interface{}) interface{} {
	sv := s.mMap.Get(key)
	if sv == nil {
		return nil
	}
	return sv.(SessionValue).Value
}

func (s *Session) Close() {
	close(s.exitChan)
}
