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
	"fmt"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/dappledger/AnnChain/angine/types"
	cmn "github.com/dappledger/AnnChain/module/lib/go-common"
	crypto "github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/xlib"
)

const (
	// RUNTIME_ENV defines the name of the environment variable for runtime path
	RUNTIME_ENV = "ANGINE_RUNTIME"
	// DEFAULT_RUNTIME defines the default path for runtime path relative to $HOME
	DEFAULT_RUNTIME = ".angine"
	// DATADIR is the data dir in the runtime, basically you don't change this never
	DATADIR    = "data"
	ARCHIVEDIR = "data/archive"
	// CONFIGFILE is the name of the configuration file name in the runtime path for angine
	CONFIGFILE = "config.toml"
)

var runtimePath string

func RuntimeDir(root string) string {
	if root != "" {
		return root
	}
	runtimePath := os.Getenv(RUNTIME_ENV)
	if len(runtimePath) == 0 {
		runtimePath, _ = homedir.Dir()
	}
	return path.Join(runtimePath, DEFAULT_RUNTIME)
}

// InitRuntime makes all the necessary directorys for angine's runtime
// and generate the config template for you if it is not there already
func InitRuntime(root string, chainId string, setConf *viper.Viper) error {
	root = RuntimeDir(root)

	// ~/.angine
	err := cmn.EnsureDir(root, 0700)
	if err != nil {
		return err
	}

	// ~/.angine/data
	_ = cmn.EnsureDir(path.Join(root, DATADIR), 0700)

	configFilePath := path.Join(root, CONFIGFILE)
	if cmn.FileExists(configFilePath) {
		return errors.New("config.toml already exists!")
	}
	fmt.Println("Using config file:", configFilePath)

	wconf := newWriteConfig(setConf)
	if setConf == nil {
		setConf = wconf
	}
	err = wconf.WriteConfigAs(configFilePath)
	if err != nil {
		return err
	}

	SetDefaults(root, wconf)
	wconf.AutomaticEnv()

	// priv_validator.json
	priv := genPrivFile(wconf.GetString("priv_validator_file"), getPrivkeyFromConf(setConf))
	gvs := []types.GenesisValidator{types.GenesisValidator{
		PubKey: priv.PubKey,
		Amount: 100,
		IsCA:   true,
	}}
	var genDoc *types.GenesisDoc
	if gvsJson := setConf.GetString("genesis_json_file"); len(gvsJson) > 0 {
		oriData, err := xlib.ReadFileDataFromCmd(gvsJson)
		if err != nil {
			return err
		}
		genDoc, err = types.GenesisDocFromJSONRet(oriData)
		if err != nil {
			return err
		}
		genDoc.SaveAs(wconf.GetString("genesis_file"))
	}

	if genDoc == nil {
		// genesis.json
		genDoc, err = genGenesiFile(wconf.GetString("genesis_file"), chainId, gvs)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Initialized chain_id: %v genesis_file: %v priv_validator: %v\n", genDoc.ChainID, wconf.GetString("genesis_file"), wconf.GetString("priv_validator_file"))
	fmt.Println("Check the files generated, make sure everything is OK.")

	return nil
}

func newWriteConfig(newConf *viper.Viper) (conf *viper.Viper) {
	defConf := DefaultConfig()
	if newConf != nil {
		for _, k := range defConf.AllKeys() {
			if newConf.IsSet(k) {
				defConf.Set(k, newConf.Get(k))
			}
		}
	}
	return defConf
}

func InitConfig(root, chainID string, setConf *viper.Viper) (conf *viper.Viper) {
	runtime := RuntimeDir(root)
	InitRuntime(runtime, chainID, setConf)
	return GetConfig(root)
}

// GetConfig returns a ready-to-go config instance with all defaults filled in
func GetConfig(root string) (conf *viper.Viper) {
	runtimeDir := RuntimeDir(root)

	conf = viper.New()

	conf.SetEnvPrefix("ANGINE")
	conf.SetConfigFile(path.Join(runtimeDir, CONFIGFILE))
	SetDefaults(runtimeDir, conf)

	if err := conf.ReadInConfig(); err != nil {
		cmn.PanicSanity(err)
	}

	if conf.IsSet("chain_id") {
		err := errors.New("Cannot set 'chain_id' via config.toml")
		cmn.PanicSanity(err)
	}
	return
}

func genPrivFile(path string, privkey crypto.PrivKey) *types.PrivValidator {
	privValidator := types.GenPrivValidator(nil, privkey)
	privValidator.SetFile(path)
	privValidator.Save()
	return privValidator
}

func genGenesiFile(path string, chainId string, gVals []types.GenesisValidator) (*types.GenesisDoc, error) {
	if len(chainId) == 0 {
		chainId = cmn.Fmt("annchain-%v", cmn.RandStr(6))
		// chainId = "civilization"
	}
	genDoc := &types.GenesisDoc{
		ChainID: chainId,
		Plugins: "specialop,querycache",
	}
	genDoc.Validators = gVals
	return genDoc, genDoc.SaveAs(path)
}

// SetDefaults sets all the default configs for angine
func SetDefaults(runtime string, conf *viper.Viper) *viper.Viper {
	conf.SetDefault("environment", "development")
	conf.SetDefault("runtime", runtime)
	conf.SetDefault("genesis_file", path.Join(runtime, "genesis.json"))
	conf.SetDefault("moniker", "anonymous")
	// conf.SetDefault("p2p_laddr", "tcp://0.0.0.0:46656")
	conf.SetDefault("seeds", "")
	conf.SetDefault("auth_by_ca", false)               // auth by ca general switch
	conf.SetDefault("non_validator_auth_by_ca", false) // whether non-validator nodes need auth by ca, only effective when auth_by_ca is true
	conf.SetDefault("fast_sync", true)
	conf.SetDefault("skip_upnp", true)
	conf.SetDefault("addrbook_file", path.Join(runtime, "addrbook.json"))
	conf.SetDefault("addrbook_strict", false) // disable to allow connections locally
	conf.SetDefault("pex_reactor", true)      // enable for peer exchange
	conf.SetDefault("priv_validator_file", path.Join(runtime, "priv_validator.json"))
	conf.SetDefault("db_backend", "leveldb")
	conf.SetDefault("db_dir", path.Join(runtime, DATADIR))
	conf.SetDefault("db_archive_dir", path.Join(runtime, ARCHIVEDIR))
	conf.SetDefault("revision_file", path.Join(runtime, "revision"))
	conf.SetDefault("filter_peers", false)

	conf.SetDefault("signbyCA", "") // auth signature from CA
	conf.SetDefault("log_path", "")

	conf.SetDefault("enable_incentive", false)
	setMempoolDefaults(conf)
	setConsensusDefaults(conf)

	return conf
}

func setMempoolDefaults(conf *viper.Viper) {
	conf.SetDefault("mempool_broadcast", true)
	conf.SetDefault("mempool_wal_dir", path.Join(conf.GetString("runtime"), DATADIR, "mempool.wal"))
	conf.SetDefault("mempool_recheck", false)
	conf.SetDefault("mempool_recheck_empty", false)
	conf.SetDefault("mempool_enable_txs_limits", false)
}

func setConsensusDefaults(conf *viper.Viper) {
	conf.SetDefault("cs_wal_dir", path.Join(conf.GetString("runtime"), DATADIR, "cs.wal"))
	conf.SetDefault("cs_wal_light", false)
	conf.SetDefault("block_size", 5000)       // max number of txs
	conf.SetDefault("block_part_size", 65536) // part size 64K
	conf.SetDefault("disable_data_hash", false)
	conf.SetDefault("timeout_propose", 3000)
	conf.SetDefault("timeout_propose_delta", 500)
	conf.SetDefault("timeout_prevote", 1000)
	conf.SetDefault("timeout_prevote_delta", 500)
	conf.SetDefault("timeout_precommit", 1000)
	conf.SetDefault("timeout_precommit_delta", 500)
	conf.SetDefault("timeout_commit", 1000)
	conf.SetDefault("skip_timeout_commit", false)

	conf.SetDefault("tracerouter_msg_ttl", 5) // seconds
}

func getPrivkeyFromConf(conf *viper.Viper) (privkey crypto.PrivKey) {
	if conf == nil {
		return nil
	}
	if pkData := conf.Get("gen_privkey"); !xlib.CheckItfcNil(pkData) {
		privkey = pkData.(crypto.PrivKey)
		conf.Set("gen_privkey", nil)
	}
	return
}
