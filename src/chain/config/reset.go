package config

import (
	"path"

	acfg "github.com/dappledger/AnnChain/angine/config"
	cmn "github.com/dappledger/AnnChain/module/lib/go-common"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	DATADIR   = "data"
	SHARDSDIR = "shards"
)

func GetConfig(agconf *viper.Viper) (conf *viper.Viper) {
	var err error
	runtime := agconf.GetString("runtime")
	acfg.InitRuntime(runtime, agconf.GetString("chain_id"), nil)

	conf = viper.New()
	conf.SetEnvPrefix("ANGINE")
	conf.SetConfigFile(path.Join(runtime, acfg.CONFIGFILE))
	if err = conf.ReadInConfig(); err != nil {
		cmn.Exit(errors.Wrap(err, "angine configuration").Error())
	}
	if conf.IsSet("chain_id") {
		cmn.Exit("Cannot set 'chain_id' via config.toml")
	}
	if conf.IsSet("revision_file") {
		cmn.Exit("Cannot set 'revision_file' via config.toml. It must match what's in the Makefile")
	}

	SetDefaults(runtime, conf)

	return
}

func SetDefaults(runtime string, conf *viper.Viper) *viper.Viper {
	conf.SetDefault("db_dir", path.Join(runtime, DATADIR))
	conf.SetDefault("shards", path.Join(runtime, SHARDSDIR))

	return conf
}
