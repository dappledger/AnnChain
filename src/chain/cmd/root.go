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
	"go.uber.org/zap"

	agconf "github.com/dappledger/AnnChain/angine/config"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"

	//_ "github.com/dappledger/AnnChain/src/chain/app"
	"github.com/dappledger/AnnChain/src/chain/app/ikhofi"
	"github.com/dappledger/AnnChain/src/chain/node"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "ann",
	Short: "Proof-of-stake blockchain from ZhongAn Technology",
	Long: `
This is the binary of the Annchain developed by ZhongAn Technology.
The project's code name is annchain, cause we wanna mimic the structure of Human Civilization. With our annchain, you can run multiple subchains in each node simultaneously to form a very sophisticated network which will represent your role in many different organizations.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initApp)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringP("runtime", "r", "", fmt.Sprintf("angine runtime dir (default is $ANGINE_RUNTIME/%v)", agconf.DEFAULT_RUNTIME))
	RootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $CIVIL_CONFPATH/.annchain.toml)")

	viper.BindPFlag("runtime", RootCmd.PersistentFlags().Lookup("runtime"))
	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
}

func initApp() {
	node.Apps["ikhofi"] = func(l *zap.Logger, c *viper.Viper, p crypto.PrivKey) (node.Application, error) {
		return ikhofi.NewIKHOFIApp(l, ikhofi.InitIkhofiConfig(c.GetString("db_dir"), c))
	}
}
