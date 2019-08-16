// Copyright © 2017 ZhongAn Technology
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

package global

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/dappledger/AnnChain/chain/core"
	"github.com/dappledger/AnnChain/gemmill/config"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-utils"
)

const (
	DefaultRuntimeDir = "~/.genesis"
	DefaultApiAddr    = "127.0.0.1:9011"
	DefaultCrypto     = crypto.CryptoTypeZhongAn

	DefaultApp = "evm"
)

type GlobalFlags struct {
	Debug      bool
	RuntimeDir string
	LogDir     string
	ApiAddr    string
}

var (
	globalFlags = GlobalFlags{}

	globalConf *viper.Viper
)

func GFlags() *GlobalFlags {
	return &globalFlags
}

func GConf() *viper.Viper {
	return globalConf
}

type NoRuntimeError struct {
	Path string
}

var _ error = NoRuntimeError{}

func (err NoRuntimeError) Error() string {
	return fmt.Sprintf("no runtime dir is found in %s.\nplease run: 'genesis init'", err.Path)
}

func ReadRuntime() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return filepath.Join(os.Getenv("HOME") + DefaultRuntimeDir)
}

func CheckInitialized(path string) error {
	if !ConfigIsInitialized(path) {
		return NoRuntimeError{Path: path}
	}
	return nil
}

func ConfigIsInitialized(cfgPath string) bool {
	if !utils.FileExists(cfgPath) {
		return false
	}
	return true
}

func GenConf() *viper.Viper {
	globalConf = core.DefaultConf()
	//globalConf.Set("runtime", GFlags().RuntimeDir)
	return globalConf
}

func CheckAndReadRuntimeConfig(runtimeDir string) (err error) {
	if runtimeDir == "" {
		runtimeDir = ReadRuntime()
	} else {
		if runtimeDir, err = homedir.Expand(runtimeDir); err != nil {
			return err
		}
	}

	// globalFlags.RuntimeDir = runtimeDir
	if err := CheckInitialized(runtimeDir); err != nil {
		return err
	}

	if globalConf, err = config.ReadConfig(runtimeDir); err != nil {
		return err
	}
	cmdAllKv := viper.AllSettings()
	for k, v := range cmdAllKv {
		globalConf.Set(k, v)
	}
	return nil
}

func CheckAppName(appName string) bool {
	switch appName {
	case "evm":
		return true
	}
	return false
}

func CheckCryptoType(ctype string) bool {
	switch ctype {
	case "ed25519", "secp256k1":
		return true
	}
	return false
}
