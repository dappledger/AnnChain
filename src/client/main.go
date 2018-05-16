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

package main

import (
	"os"

	"github.com/dappledger/AnnChain/src/client/commands"
	"github.com/dappledger/AnnChain/src/client/commons"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	commands.InitLog()
	app := cli.NewApp()
	app.Name = "anntool"
	app.Version = "0.2"

	app.Commands = []cli.Command{
		commands.SignCommand,
		commands.InitSubCommands,

		commands.EVMCommands,
		commands.AccountCommands,
		commands.QueryCommands,
		commands.TxCommands,
		commands.SpecialCommands,
		commands.OrgCommands,
		commands.EventCommands,
		commands.IkhofiCommands,

		// commands.ExamCommand,
		// commands.TransferBenchCommand,
		// commands.InitCommand,
		// commands.AnnCoinBenchCommand,

		commands.InfoCommand,
		commands.VoteCommands,
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
		cli.StringFlag{
			Name:  "target",
			Value: "",
			Usage: "specify the target chain for the following command",
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
