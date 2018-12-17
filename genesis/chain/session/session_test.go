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
	"testing"
	"time"
)

func TestBench(t *testing.T) {

	s := NewSession(10, 2)

	s.SetSession("name", "fhy")

	for {
		v := s.GetSession("name")
		if v == nil {
			s.Close()
			return
		} else {
			t.Log(v)
		}
		time.Sleep(time.Second)
	}
}
