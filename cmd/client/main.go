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

package main

import (
	"os"

	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/chain/types"
	"github.com/dappledger/AnnChain/cmd/client/commands"
	"github.com/dappledger/AnnChain/cmd/client/commons"
)

func main() {
	commands.InitLog()

	app := cli.NewApp()
	app.Name = "anntool"
	app.Version = types.GetVersion()

	app.Commands = []cli.Command{
		commands.SignCommand,

		commands.EVMCommands,
		commands.AccountCommands,
		commands.QueryCommands,
		commands.TxCommands,
		commands.AdminCommands,
		commands.InfoCommand,

		commands.VersionCommands,
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "callmode",
			Usage:       "rpc call mode: sync or commit",
			Value:       "sync",
			Destination: &commons.CallMode,
		},
		cli.StringFlag{
			Name:        "backend",
			Value:       "tcp://localhost:46657",
			Destination: &commons.QueryServer,
			Usage:       "rpc address of the node",
		},
	}

	app.Before = func(ctx *cli.Context) error {
		if commons.CallMode == "sync" || commons.CallMode == "commit" {
			return nil
		}

		return cli.NewExitError("invalid sync mode", 127)
	}

	_ = app.Run(os.Args)
}
