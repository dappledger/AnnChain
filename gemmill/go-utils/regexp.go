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
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	regexp_NotNumLetterUnderline *regexp.Regexp
)

func init() {
	//regexp_NotNumLetterUnderline = regexp.MustCompile(`^[0-9a-zA-Z_]+`)
	regexp_NotNumLetterUnderline = regexp.MustCompile(`[\W]+`)
}

func OnlyNumLetterUnderline(str string) bool {
	return regexp_NotNumLetterUnderline.FindStringIndex(str) == nil
}

func CheckNumber(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

func CheckIPAddrSlc(str string) bool {
	addrSlc := strings.Split(str, ",")
	if len(addrSlc) == 0 {
		return false
	}
	for i := range addrSlc {
		if _, err := net.ResolveTCPAddr("tcp", addrSlc[i]); err != nil {
			return false
		}
	}
	return true
}
