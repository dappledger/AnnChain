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

package commands

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dappledger/AnnChain/chain/commands/global"
	"github.com/dappledger/AnnChain/chain/core"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/utils"
)

func NewRunCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "run",
		Short: "start a deamon blockchain node",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			runtime, _ := cmd.Flags().GetString("runtime")
			if err = global.CheckAndReadRuntimeConfig(runtime); err != nil {
				fmt.Println("load runtime config err:", err)
				return err
			}
			setFlags(cmd, global.GConf())

			logDir, _ := cmd.Flags().GetString("log_path")

			if logDir != "" {
				if logDir, err = homedir.Expand(logDir); err == nil {
					global.GConf().Set("log_path", logDir)
				}
			}
			auditLogDir, _ := cmd.Flags().GetString("audit_log_path")

			if auditLogDir != "" {
				if auditLogDir, err = homedir.Expand(auditLogDir); err == nil {
					global.GConf().Set("audit_log_path", auditLogDir)
				}
			}

			fmt.Println("Log path is: ", global.GConf().Get("log_path"))
			fmt.Println("audit_log path is: ", global.GConf().Get("audit_log_path"))
			fmt.Println("CryptoType is: ", crypto.CryptoType)

			return err
		},
		Run: runCommandFunc,
	}

	c.Flags().StringP("chain_id", "", "", "specify the chain id when the node is joining it without genesis file")
	c.Flags().BoolP("pprof", "", false, "start golang profile at port :6060")
	c.Flags().BoolP("statistic", "", false, "start statistic tool on specified code lines")
	c.Flags().BoolP("test", "", false, "run the node in test mode")
	// c.Flags().StringP("app", "", glb.DefaultApp, "app name")

	viper.BindPFlag("chain_id", c.Flag("chain_id"))
	viper.BindPFlag("pprof", c.Flag("pprof"))
	viper.BindPFlag("test", c.Flag("test"))
	viper.BindPFlag("statistic", c.Flag("statistic"))

	return c
}

func runCommandFunc(cmd *cobra.Command, args []string) {
	if global.GConf().GetBool("statistic") {
		utils.StartStat()
	}
	core.RunNode(global.GConf(), "", global.GConf().GetString("app_name"))
}

func setFlags(cmd *cobra.Command, conf *viper.Viper) {}
