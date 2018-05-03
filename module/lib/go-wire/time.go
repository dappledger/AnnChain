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

package wire

import (
	"io"
	"time"

	. "github.com/dappledger/AnnChain/module/lib/go-common"
)

/*
Writes nanoseconds since epoch but with millisecond precision.
This is to ease compatibility with Javascript etc.
*/

func WriteTime(t time.Time, w io.Writer, n *int, err *error) {
	nanosecs := t.UnixNano()
	millisecs := nanosecs / 1000000
	WriteInt64(millisecs*1000000, w, n, err)
}

func ReadTime(r io.Reader, n *int, err *error) time.Time {
	t := ReadInt64(r, n, err)
	if t%1000000 != 0 {
		PanicSanity("Time cannot have sub-millisecond precision")
	}
	return time.Unix(0, t)
}
