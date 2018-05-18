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
	"os"
	"path"
	"strings"

	"gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-common"
	"gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-config"
)

const (
	RUNTIME_ENV     = "ANN_RUNTIME"
	DEFAULT_RUNTIME = ".ann_runtime"
	DATADIR         = "data"
	CONFIGFILE      = "config.toml"
)

func parseConfigTpl(moniker string, root string) (conf string) {
	conf = strings.Replace(CONFIGTPL, "__MONIKER__", moniker, -1)
	conf = strings.Replace(conf, "__CONFROOT__", root, -1)
	return
}

func RuntimeDir(root string) string {
	if root != "" {
		return root
	}
	if runtimePath, exists := os.LookupEnv(RUNTIME_ENV); exists {
		return runtimePath
	}
	// return path.Join(os.Getenv("HOME"), DEFAULT_RUNTIME)
	return path.Join("/alidata1/", DEFAULT_RUNTIME)
}

func InitRuntime(root string) {
	common.EnsureDir(root, 0700)
	common.EnsureDir(path.Join(root, DATADIR), 0700)
	configFilePath := path.Join(root, CONFIGFILE)
	if !common.FileExists(configFilePath) {
		common.MustWriteFile(configFilePath, []byte(parseConfigTpl("anonymous", root)), 0644)
	}
}

func GetConfig(root string) (conf *config.MapConfig) {
	var err error

	runtime := RuntimeDir(root)
	configAbs := path.Join(runtime, CONFIGFILE)
	InitRuntime(runtime)

	if conf, err = config.ReadMapConfigFromFile(configAbs); err != nil {
		common.Exit(common.Fmt("Could not read config: %v", err))
	}

	// Set defaults or panic
	if conf.IsSet("chain_id") {
		common.Exit("Cannot set 'chain_id' via config.toml")
	}
	if conf.IsSet("revision_file") {
		common.Exit("Cannot set 'revision_file' via config.toml. It must match what's in the Makefile")
	}

	FillInDefaults(runtime, conf)

	return
}

func FillInDefaults(root string, conf *config.MapConfig) *config.MapConfig {
	conf.SetRequired("chain_id") // blows up if you try to use it before setting.
	conf.SetRequired("environment")

	conf.SetDefault("environment", "development")
	conf.SetDefault("log_mode", "")
	conf.SetDefault("datadir", root)
	conf.SetDefault("genesis_file", path.Join(root, "genesis.json"))
	conf.SetDefault("moniker", "anonymous")
	conf.SetDefault("node_laddr", "tcp://0.0.0.0:46656")
	conf.SetDefault("seeds", "")
	conf.SetDefault("fast_sync", true)
	conf.SetDefault("skip_upnp", false)
	conf.SetDefault("addrbook_file", path.Join(root, "addrbook.json"))
	conf.SetDefault("addrbook_strict", false) // disable to allow connections locally
	conf.SetDefault("pex_reactor", false)     // enable for peer exchange
	conf.SetDefault("priv_validator_file", path.Join(root, "priv_validator.json"))
	conf.SetDefault("db_backend", "leveldb")
	conf.SetDefault("db_dir", path.Join(root, DATADIR))
	conf.SetDefault("rpc_laddr", "tcp://0.0.0.0:46657")
	conf.SetDefault("grpc_laddr", "")
	conf.SetDefault("api_laddr", "")
	conf.SetDefault("revision_file", path.Join(root, "revision"))
	conf.SetDefault("cs_wal_dir", path.Join(root, DATADIR, "cs.wal"))
	conf.SetDefault("cs_wal_light", false)
	conf.SetDefault("filter_peers", false)

	conf.SetDefault("block_size", 3000)       // max number of txs
	conf.SetDefault("block_part_size", 65536) // part size 64K
	conf.SetDefault("disable_data_hash", false)
	conf.SetDefault("timeout_propose", 3000)
	conf.SetDefault("timeout_propose_delta", 500)
	conf.SetDefault("timeout_prevote", 3000)
	conf.SetDefault("timeout_prevote_delta", 500)
	conf.SetDefault("timeout_precommit", 3000)
	conf.SetDefault("timeout_precommit_delta", 500)
	conf.SetDefault("timeout_commit", 3000)
	conf.SetDefault("skip_timeout_commit", false)

	conf.SetDefault("mempool_recheck", true)
	conf.SetDefault("mempool_recheck_empty", true)
	conf.SetDefault("mempool_broadcast", true)
	conf.SetDefault("mempool_wal_dir", "") // path.Join(root, DATADIR, "mempool.wal")
	conf.SetDefault("mempool_enable_txs_limits", false)

	conf.SetDefault("signbyCA", "")

	conf.SetDefault("p2p", map[string]interface{}{"connection_reset_wait": 300})

	conf.SetDefault("test", false)
	conf.SetDefault("pprof", false)

	conf.SetDefault("log_path", "")

	conf.Set("db_type", "sqlite3")
	conf.Set("db_conn_str", "") // some types of database will need this

	return conf
}
