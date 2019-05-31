package commands

import (
	"fmt"

	"gopkg.in/urfave/cli.v1"
)

var (
	VERSION         string
	VersionCommands = cli.Command{
		Name:   "version",
		Action: ShowVersion,
		Usage:  "show version of rtool",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "version",
				Usage: "show version of rtool",
				Value: "0",
			},
		},
	}
)

func ShowVersion(ctx *cli.Context) {
	fmt.Println("version:", VERSION)
}
