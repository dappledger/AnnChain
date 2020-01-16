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
	"log"

	"github.com/astaxie/beego"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/dappledger/AnnChain/chain/commands/global"
	"github.com/dappledger/AnnChain/chain/commands/vision/routers"
	"github.com/dappledger/AnnChain/gemmill"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
)

var (
	chainId string
	visual  bool
	appName string
	vport   string
)

func NewInitCommand() *cobra.Command {
	c := &cobra.Command{
		Use:       "init",
		Short:     "init runtime dir",
		Args:      cobra.OnlyValidArgs,
		ValidArgs: []string{"chainid", "crypto"},
		Run:       newInitCommandFunc,
	}

	c.Flags().StringVar(&chainId, "chainid", "", "manually specify chainId")
	c.Flags().BoolVar(&visual, "visual", false, "whether init node visually")
	c.Flags().StringVar(&appName, "app", global.DefaultApp, "app name")
	c.Flags().StringVar(&vport, "vport", ":8080", "port of visual mode")
	return c
}

func newInitCommandFunc(cmd *cobra.Command, args []string) {
	if visual {
		routers.InitNode()
		beego.Run(vport)
		select {}
	}
	if !global.CheckAppName(appName) {
		cmd.Println("appname not found")
		return
	}
	if _, err := crypto.GenPrivkeyByType(crypto.CryptoType); err != nil {
		log.Fatal(err)
	}
	defConf := global.GenConf()
	defConf.Set("app_name", appName)

	if global.GFlags().LogDir == "" {
		global.GFlags().LogDir = "./"
	} else {
		var err error
		if global.GFlags().LogDir, err = homedir.Expand(global.GFlags().LogDir); err != nil {
			panic(err)
		}
	}
	if global.GFlags().AuditLogDir == "" {
		global.GFlags().AuditLogDir = "./"
	} else {
		var err error
		if global.GFlags().AuditLogDir, err = homedir.Expand(global.GFlags().AuditLogDir); err != nil {
			panic(err)
		}
	}

	log.Println("Log dir is: ", global.GFlags().LogDir)
	defConf.Set("log_dir", global.GFlags().LogDir)

	log.Println("audit_log_path is: ", global.GFlags().AuditLogDir)
	defConf.Set("audit_log_path", global.GFlags().AuditLogDir)

	gemmill.Initialize(&gemmill.Tunes{Runtime: global.GFlags().RuntimeDir, Conf: defConf}, chainId)
}
