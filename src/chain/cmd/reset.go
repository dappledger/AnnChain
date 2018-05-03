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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	acfg "github.com/dappledger/AnnChain/angine/config"
	"github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/src/chain/config"
	"go.uber.org/zap"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset PrivValidator, clean the data and shards",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		angineconf := acfg.GetConfig(viper.GetString("runtime"))
		conf := config.GetConfig(angineconf)
		os.RemoveAll(conf.GetString("db_dir"))
		os.RemoveAll(conf.GetString("shards"))
		resetPrivValidator(angineconf.GetString("priv_validator_file"))
	},
}

func init() {
	RootCmd.AddCommand(resetCmd)
}

func resetPrivValidator(privValidatorFile string) {
	var (
		privValidator *types.PrivValidator
		logger        *zap.Logger
	)

	if _, err := os.Stat(privValidatorFile); err == nil {
		privValidator = types.LoadPrivValidator(logger, privValidatorFile)
		privValidator.Reset()
		fmt.Println("Reset PrivValidator", "file", privValidatorFile)
	} else {
		privValidator = types.GenPrivValidator(logger, nil)
		privValidator.SetFile(privValidatorFile)
		privValidator.Save()
		fmt.Println("Generated PrivValidator", "file", privValidatorFile)
	}
}
