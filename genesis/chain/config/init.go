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
	"io/ioutil"
	"path"

	ac "github.com/dappledger/AnnChain/angine/config"
	at "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/ann-module/lib/go-config"
	c "github.com/dappledger/AnnChain/genesis/tools/config"
)

// LoadDefaultConfig set default config
func LoadDefaultConfig(conf *config.MapConfig, confPath string) {

	conf.SetDefault("db_type", "sqlite3")
	conf.SetDefault("db_conn_str", "sqlite3") // some types of database will need this

	var root string
	if confPath != "" {
		root = confPath
	} else {
		root = ac.RuntimeDir("")
	}
	cfg := c.LoadConfigFile(path.Join(root, ac.MYCONFIGFILE))

	conf.SetDefault("base_fee", cfg.GetInt("base_fee"))
	conf.SetDefault("base_reserve", cfg.GetInt("base_reserve"))
	conf.SetDefault("max_txset_size", cfg.GetInt("max_txset_size"))
	conf.SetDefault("init_official", true)
}

func GetInitialIssueAccount(conf config.Config) ([]at.InitInfo, error) {
	genesispath := conf.GetString("genesis_file")
	jsonbody, err := ioutil.ReadFile(genesispath)
	if err != nil {
		return nil, err
	}
	gendoc := at.GenesisDocFromJSON(jsonbody)
	return gendoc.InitAccounts, nil
}
