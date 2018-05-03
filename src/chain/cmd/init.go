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
	"path"

	"github.com/astaxie/beego"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/dappledger/AnnChain/angine"
	civconf "github.com/dappledger/AnnChain/src/chain/config"
	vrouter "github.com/dappledger/AnnChain/src/vision/routers"

	cmn "github.com/dappledger/AnnChain/module/lib/go-common"
)

const (
	CONFPATH = "confpath"
)

// abcCmd represents the abc command
var initCmd = &cobra.Command{
	Use:       "init",
	Short:     "create config, genesis, priv files in the runtime directory",
	Long:      `create config, genesis, priv files in the runtime directory`,
	Args:      cobra.OnlyValidArgs,
	ValidArgs: []string{"chainid", "professional"},
	Run: func(cmd *cobra.Command, args []string) {
		runtime := viper.GetString("runtime")
		if cmd.Flag("visual").Value.String() == "true" {
			vrouter.InitNode(runtime)
			beego.Run(viper.GetString("vport"))
			select {}
			return
		}
		chainID := cmd.Flag("chain_id").Value.String()
		angine.Initialize(&angine.Tunes{
			Runtime: runtime,
		}, chainID)
	},
}

func init() {

	initCmd.Flags().String("chainid", "", "name of the chain")
	initCmd.Flags().Bool("visual", false, "whether init node visually")
	initCmd.Flags().String("vport", ":8080", "port under visual mode")
	RootCmd.AddCommand(initCmd)

	viper.BindPFlag("vport", initCmd.Flag("vport"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// abcCmd.PersistentFlags().String("foo", "", "A help for foo")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// abcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initCivilConfig(pathStr string) {
	if pathStr == "" {
		pathStr = CivilPath()
	}

	var dirpath, filepath string
	if path.Ext(pathStr) == "" {
		dirpath = pathStr
		filepath = path.Join(pathStr, ".annchain.toml")
	} else {
		dirpath = path.Dir(pathStr)
		filepath = pathStr
	}

	if err := cmn.EnsureDir(dirpath, 0700); err != nil {
		cmn.PanicSanity(err)
	}
	if !cmn.FileExists(filepath) {
		cmn.MustWriteFile(filepath, []byte(civconf.Template), 0644)
		fmt.Println("path of the annchain config file: " + filepath)
	}
}

var civilPath string

func CivilPath() string {
	if len(civilPath) == 0 {
		civilPath = viper.GetString(CONFPATH)
		if len(civilPath) == 0 {
			civilPath, _ = homedir.Dir()
		}
	}
	return civilPath
}
