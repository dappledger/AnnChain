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

func TestOnlyNumLetterUnderline(t *testing.T) {
	str1 := "abc1abc2_"
	if !OnlyNumLetterUnderline(str1) {
		t.Error(str1, OnlyNumLetterUnderline(str1))
		return
	}
	str2 := "_12bc1_aAAAAVBBbc2"
	if !OnlyNumLetterUnderline(str2) {
		t.Error(str2, OnlyNumLetterUnderline(str2))
		return
	}
	str3 := "$" + str1
	if OnlyNumLetterUnderline(str3) {
		t.Error(str3, OnlyNumLetterUnderline(str3))
		return
	}
	str4 := str1 + "_/\"" + str1
	if OnlyNumLetterUnderline(str4) {
		t.Error(str4, OnlyNumLetterUnderline(str4))
		return
	}
	str5 := str1 + "_/\""
	if OnlyNumLetterUnderline(str5) {
		t.Error(str5, OnlyNumLetterUnderline(str5))
		return
	}
	str6 := "%"
	if OnlyNumLetterUnderline(str6) {
		t.Error(str6, OnlyNumLetterUnderline(str6))
		return
	}
}
