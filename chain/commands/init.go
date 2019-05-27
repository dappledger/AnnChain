// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
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

	glb "github.com/dappledger/AnnChain/chain/commands/global"
	vrouter "github.com/dappledger/AnnChain/chain/commands/vision/routers"
	"github.com/dappledger/AnnChain/gemmill"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"

	"github.com/astaxie/beego"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
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
	c.Flags().StringVar(&appName, "app", glb.DefaultApp, "app name")
	c.Flags().StringVar(&vport, "vport", ":8080", "port of visual mode")
	return c
}

func newInitCommandFunc(cmd *cobra.Command, args []string) {
	if visual {
		vrouter.InitNode()
		beego.Run(vport)
		select {}
	}
	if !glb.CheckAppName(appName) {
		cmd.Println("appname not found")
		return
	}
	if _, err := crypto.GenPrivkeyByType(crypto.CryptoType); err != nil {
		log.Fatal(err)
	}
	defConf := glb.GenConf()
	defConf.Set("app_name", appName)

	if glb.GFlags().LogDir == "" {
		glb.GFlags().LogDir = "./"
	} else {
		var err error
		if glb.GFlags().LogDir, err = homedir.Expand(glb.GFlags().LogDir); err != nil {
			panic(err)
		}
	}

	log.Println("Log dir is: ", glb.GFlags().LogDir)
	defConf.Set("log_dir", glb.GFlags().LogDir)

	gemmill.Initialize(&gemmill.Tunes{Runtime: glb.GFlags().RuntimeDir, Conf: defConf}, chainId)
}
