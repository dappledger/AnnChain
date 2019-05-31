package main

import (
	"os"

	"github.com/dappledger/AnnChain/cmd/client/commands"
	"github.com/dappledger/AnnChain/cmd/client/commons"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	commands.InitLog()

	app := cli.NewApp()
	app.Name = "anntool"
	app.Version = "0.6"

	app.Commands = []cli.Command{
		commands.SignCommand,

		commands.EVMCommands,
		commands.AccountCommands,
		commands.QueryCommands,
		commands.TxCommands,
		commands.SpecialCommands,
		commands.InfoCommand,

		commands.VersionCommands,

		commands.SetupCommand,
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
