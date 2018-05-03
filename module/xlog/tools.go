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

package xlog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	//"log"
	"runtime"
	"time"
)

func DumpStack() {
	if err := recover(); err != nil {
		var buf bytes.Buffer
		bs := make([]byte, 1<<12)
		num := runtime.Stack(bs, false)
		buf.WriteString(fmt.Sprintf("Panic: %s\n", err))
		buf.Write(bs[:num])
		dumpName := "dump_" + time.Now().Format("20060102-150405")
		nerr := ioutil.WriteFile(dumpName, buf.Bytes(), 0644)
		if nerr != nil {
			fmt.Println("write dump file error", nerr)
			fmt.Println(buf.Bytes())
		}
	}
}
