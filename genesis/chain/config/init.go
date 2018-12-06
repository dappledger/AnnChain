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
func LoadDefaultConfig(conf *config.MapConfig) {

	conf.SetDefault("db_type", "sqlite3")
	conf.SetDefault("db_conn_str", "sqlite3") // some types of database will need this

	root := ac.RuntimeDir("")
	cfg := c.LoadConfigFile(path.Join(root, ac.MYCONFIGFILE))

	conf.SetDefault("base_fee", cfg.GetInt("base_fee"))
	conf.SetDefault("base_reserve", cfg.GetInt("base_reserve"))
	conf.SetDefault("max_txset_size", cfg.GetInt("max_txset_size"))
	// conf.SetDefault("base_fee", 0)
	// conf.SetDefault("base_reserve", 0)
	// conf.SetDefault("max_txset_size", 5000)

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
