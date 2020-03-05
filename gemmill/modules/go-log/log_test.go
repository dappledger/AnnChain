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

package log

import "testing"

func TestLog(t *testing.T) {
	logger, err := Initialize("dev", "testdir")
	if err != nil {
		t.Error("initialize err ", err)
		return
	}
	SetLog(logger)
	Error("error log")
	Warn("warn log")
	Info("info log")
}

func TestGetLog(t *testing.T) {
	Error("error log")
	Warn("warn log")
	Info("info log")
	Debug("debug msg")
	GetLog()
	Error("error log")
	Warn("warn log")
	Info("info log")
	Debug("debug msg")
}
