package config

import (
	"path"

	"github.com/spf13/viper"

	anginecfg "github.com/dappledger/AnnChain/angine/config"
	"github.com/dappledger/AnnChain/module/lib/go-common"
)

const (
	ShardsDir  = "shards"
	ShardsData = "data"
)

// LoadDefaultShardConfig should only be called after initAnnChainRuntime
func LoadDefaultAngineConfig(datadir, chainID string, conf map[string]interface{}) (*viper.Viper, error) {
	shardPath := path.Join(datadir, ShardsDir, chainID)
	if err := common.EnsureDir(shardPath, 0700); err != nil {
		return nil, err
	}
	if err := common.EnsureDir(path.Join(shardPath, ShardsData), 0700); err != nil {
		return nil, err
	}

	c := anginecfg.SetDefaults(shardPath, viper.New())
	for k, v := range conf {
		c.Set(k, v)
	}
	c.Set("chain_id", chainID)

	loadDefaultSqlConfig(c)
	return c, nil
}

func loadDefaultSqlConfig(c *viper.Viper) {
	c.Set("db_type", "sqlite3")
	c.Set("db_conn_str", "")
}
