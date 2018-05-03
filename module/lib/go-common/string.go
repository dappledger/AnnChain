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
	"fmt"
	"strings"
)

var Fmt = fmt.Sprintf

func RightPadString(s string, totalLength int) string {
	remaining := totalLength - len(s)
	if remaining > 0 {
		s = s + strings.Repeat(" ", remaining)
	}
	return s
}

func LeftPadString(s string, totalLength int) string {
	remaining := totalLength - len(s)
	if remaining > 0 {
		s = strings.Repeat(" ", remaining) + s
	}
	return s
}

//SanitizeHex trim the prefix '0x'|'0X' if present
func SanitizeHex(hex string) string {
	return strings.TrimPrefix(strings.ToLower(hex), "0x")
}
