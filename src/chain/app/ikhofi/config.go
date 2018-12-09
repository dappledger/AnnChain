/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */

package ikhofi

import (
	"os"
	"path"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/dappledger/AnnChain/module/lib/go-common"
	"path/filepath"
)

const (
	RUNTIME_ENV     = "ANN_RUNTIME"
	IKHOFI_PATH     = "IKHOFI_PATH"
	DEFAULT_RUNTIME = ".ann_runtime"
	DATADIR         = "contract_data"
	DBDATADIR       = "chaindata"
	CONFIGFILE      = "ikhofi.yaml"
)

type IkhofiConfig struct {
	Db struct {
		DbType            string `yaml:"type"`
		DbPath            string `yaml:"path"`
		CacheSize         int    `yaml:"cacheSize"`
		DestroyAfterClose bool   `yaml:"destroyAfterClose"`
	}
}

func runtimeDir(root string) string {
	if root != "" {
		if root[0:1] != "/" {
			pwd, _ := os.Getwd()
			root = filepath.Join(pwd, root)
		}
		return root
	}
	if runtimePath, exists := os.LookupEnv(RUNTIME_ENV); exists {
		return runtimePath
	}
	return path.Join(os.Getenv("HOME"), DEFAULT_RUNTIME)
}

func getYamlBytes(dbPath string) (yamlBytes []byte, err error) {
	cfg := IkhofiConfig{}
	err = yaml.Unmarshal([]byte(CONFIGTPL), &cfg)
	if err != nil {
		return
	}
	cfg.Db.DbType = "leveldb"
	cfg.Db.DbPath = dbPath
	cfg.Db.CacheSize = 67108864
	cfg.Db.DestroyAfterClose = false
	yamlBytes, err = yaml.Marshal(&cfg)
	return
}

func initRuntime(root string) string {
	common.EnsureDir(root, 0700)
	common.EnsureDir(path.Join(root, DATADIR), 0700)
	configFilePath := path.Join(root, CONFIGFILE)
	if !common.FileExists(configFilePath) {
		yamlBytes, err := getYamlBytes(path.Join(root, DATADIR, DBDATADIR, "/"))
		if err != nil {
			common.Exit("can not generate ikhofi.yaml file.")
		}
		common.MustWriteFile(configFilePath, yamlBytes, 0644)
	}
	return configFilePath
}

func GetIkhofiPath(path string) string {
	if path != "" {
		return path
	}
	if ikhofiPath, exists := os.LookupEnv(IKHOFI_PATH); exists {
		return ikhofiPath
	}
	common.Exit("IKHOFI_PATH is nil.")
	return ""
}

func InitIkhofiConfig(root string, conf *viper.Viper) *viper.Viper {
	runtime := runtimeDir(root)
	configFilePath := initRuntime(runtime)

	conf.SetDefault("ikhofi_config", configFilePath)

	return conf
}
