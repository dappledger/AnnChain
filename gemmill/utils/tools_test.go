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

package utils

import (
	"testing"
)

func TestSortInt64Slc(t *testing.T) {
	var slc = []int64{1, 129, 20, 4, 45, 66}
	Int64Slice(slc).Sort()
	pre := slc[0]
	for i := range slc {
		if pre > slc[i] {
			t.Error("sort err")
			return
		}
		pre = slc[i]
	}
}
