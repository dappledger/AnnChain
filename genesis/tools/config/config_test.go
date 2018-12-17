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

package config

import (
	"testing"
	"fmt"
)

var testCfgFile = "./cfg.json"

func TestParse(t *testing.T) {
	cfg := LoadConfigFile(testCfgFile)
	if cfg.GetFloat("port") != 8010 || cfg.GetString("role") != "cc" || cfg.GetBool("idgen") == false || cfg.GetFloat("sid") != 0 {
		t.Error("Fatal")
	}

	// fmt.Println(cfg.GetFloat("port"))
	fmt.Println(cfg.GetInt("port"))
}
