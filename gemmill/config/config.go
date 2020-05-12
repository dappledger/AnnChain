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

	crypto "github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-utils"
	cmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/types"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	RUNTIME_ENV     = "ANGINE_RUNTIME"
	DEFAULT_RUNTIME = ".genesis"
	DATADIR         = "data"
	ARCHIVEDIR      = "data/archive"
	CONFIGFILE      = "config.toml"
)

func runtimeDir(root string) string {
	if root != "" {
		root, _ = homedir.Expand(root)
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

	root = runtimeDir(root)
	var err error
	if err = cmn.EnsureDir(root, 0700); err != nil {
		return err
	}

	if err = cmn.EnsureDir(path.Join(root, DATADIR), 0700); err != nil {
		return err
	}

	configFilePath := path.Join(root, CONFIGFILE)
	if cmn.FileExists(configFilePath) {
		return fmt.Errorf("%v already exists!", configFilePath)
	}
	wconf := newWriteConfig(setConf)
	err = wconf.WriteConfigAs(configFilePath)
	if err != nil {
		return err
	}

	SetDefaults(root, setConf)
	setConf.AutomaticEnv()

	priv, err := genPrivFile(crypto.CryptoType, setConf.GetString("priv_validator_file"), getPrivkeyFromConf(setConf))
	if err != nil {
		return fmt.Errorf("%v generate priv_validator error", err)
	}

	setConf.AutomaticEnv()

	var genDoc *types.GenesisDoc
	if gvsJson := setConf.GetString("genesis_json_file"); len(gvsJson) > 0 {
		oriData, err := utils.ReadFileDataFromCmd(gvsJson)
		if err != nil {
			return err
		}
		genDoc, err = types.GenesisDocFromJSONRet(oriData)
		if err != nil {
			return err
		}
		genDoc.SaveAs(setConf.GetString("genesis_file"))
	}
	if genDoc == nil {
		genDoc, err = genGenesiFile(setConf.GetString("genesis_file"), chainId, priv)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Initialized chain_id: %v genesis_file: %v priv_validator: %v\n", genDoc.ChainID, setConf.GetString("genesis_file"), setConf.GetString("priv_validator_file"))
	fmt.Println("Check the files generated, make sure everything is OK.")

	return nil
}

func mergeDefaultConf(passedInConf *viper.Viper) {
	conf := DefaultConfig()
	for _, k := range conf.AllKeys() {
		if !passedInConf.IsSet(k) {
			passedInConf.Set(k, conf.Get(k))
		}
	}
}

func InitConfig(root, chainID string, setConf *viper.Viper) (*viper.Viper, error) {
	runtime := runtimeDir(root)
	InitRuntime(runtime, chainID, setConf)
	return ReadConfig(root)
}

// ReadConfig returns a ready-to-go config instance with all defaults filled in
func ReadConfig(root string) (*viper.Viper, error) {
	runtimeDir := runtimeDir(root)
	conf := viper.New()
	conf.SetEnvPrefix("ANGINE")
	conf.SetConfigFile(path.Join(runtimeDir, CONFIGFILE))
	SetDefaults(runtimeDir, conf)

	if err := conf.ReadInConfig(); err != nil {
		return conf, err
	}

	if conf.IsSet("chain_id") {
		return conf, errors.New("Cannot set 'chain_id' via config.toml")
	}
	return conf, nil
}

func genPrivFile(cryptoType, path string, privkey crypto.PrivKey) (*types.PrivValidator, error) {
	privValidator, err := types.GenPrivValidator(cryptoType, privkey)
	if err != nil {
		return nil, err
	}
	privValidator.SetFile(path)
	return privValidator, privValidator.Save()
}

func GenChainID() string {
	return cmn.Fmt("genesis-%v", cmn.RandStr(6))
}

func genGenesiFile(path string, chainId string, priv *types.PrivValidator) (*types.GenesisDoc, error) {
	if len(chainId) == 0 {
		chainId = GenChainID()
	}
	genDoc := &types.GenesisDoc{
		ChainID: chainId,
		Plugins: "adminOp,querycache",
		Validators: []types.GenesisValidator{types.GenesisValidator{
			PubKey: priv.PubKey,
			Amount: 100,
			IsCA:   true,
		}},
	}
	return genDoc, genDoc.SaveAs(path)
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

// SetDefaults sets all the default configs for angine
func SetDefaults(runtime string, conf *viper.Viper) *viper.Viper {
	conf.SetDefault("runtime", runtime)
	conf.SetDefault("genesis_file", path.Join(runtime, "genesis.json"))
	conf.SetDefault("priv_validator_file", path.Join(runtime, "priv_validator.json"))
	conf.SetDefault("addrbook_file", path.Join(runtime, "addrbook.json"))
	conf.SetDefault("addrbook_strict", false) // disable to allow connections locally
	conf.SetDefault("pex_reactor", true)      // enable for peer exchange
	conf.SetDefault("db_dir", path.Join(runtime, DATADIR))
	conf.SetDefault("db_archive_dir", path.Join(runtime, ARCHIVEDIR))
	conf.SetDefault("revision_file", path.Join(runtime, "revision"))
	conf.SetDefault("filter_peers", false)
	conf.SetDefault("p2p_proxy_addr", "")

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
	if pkData := conf.Get("gen_privkey"); !utils.CheckItfcNil(pkData) {
		privkey = pkData.(crypto.PrivKey)
		conf.Set("gen_privkey", nil)
	}
	return
}

func DefaultConfig() (conf *viper.Viper) {
	conf = viper.New()
	conf.Set("app_name", "evm")
	conf.Set("environment", "production")
	conf.Set("moniker", "__MONIKER__")
	conf.Set("db_backend", "leveldb")
	conf.Set("moniker", "anonymous")
	conf.Set("p2p_laddr", "tcp://0.0.0.0:46656")
	conf.Set("rpc_laddr", "tcp://0.0.0.0:46657")
	conf.Set("seeds", "")
	conf.Set("auth_by_ca", true)
	conf.Set("signbyCA", "")
	conf.Set("fast_sync", true)
	conf.Set("skip_upnp", true)
	conf.Set("log_path", "")
	conf.Set("audit_log_path", "audit.log")
	conf.Set("broadcast_tx_time_out", 30)
	conf.Set("non_validator_node_auth", true)
	conf.SetDefault("tracerouter_msg_ttl", 5)
	conf.Set("threshold_blocks", 0)
	conf.SetDefault("block_size", 5000)

	return conf
}
