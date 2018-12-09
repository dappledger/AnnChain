package flags

import (
	"os"

	"gopkg.in/urfave/cli.v1"
)

var (
	GlobalFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "datadir",
			Usage: "Data directory for the databases and keystore",
			Value: os.Getenv("HOME") + "/.annchain",
		},
	}
)
