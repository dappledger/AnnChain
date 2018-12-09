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


package commands

import (
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/urfave/cli.v1"

	cmn "github.com/dappledger/AnnChain/module/lib/go-common"
)

const SUBPATH = "/.subchain"

var (
	InitSubCommands = cli.Command{
		Name:     "initsub",
		Usage:    "init subchain config file and genesis file",
		Category: "Organization",
		Action:   initSubConf,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "configfile",
				Value: "$HOME/appconfig.toml",
			},
			cli.StringFlag{
				Name:  "genesisfile",
				Value: "$HOME/appconfig.toml",
			},
		},
	}
)

func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func initSubConf(ctx *cli.Context) error {
	if ctx.IsSet("configfile") {
		confpath := ctx.String("configfile")
		if _, err := PathExist(confpath); err != nil {
			return cli.NewExitError(err.Error(), 127)
		}

		confpath = path.Join(confpath, "appconfig.toml")

		if !cmn.FileExists(confpath) {
			cmn.MustWriteFile(confpath, []byte(configTemplate), 0644)
			fmt.Println("path of the app config file:" + confpath)
		} else {
			fmt.Println("Config file already exist!")
		}
	}

	if ctx.IsSet("genesisfile") {
		gpath := ctx.String("genesisfile")
		if _, err := PathExist(gpath); err != nil {
			return cli.NewExitError(err.Error(), 127)
		}

		gpath = path.Join(gpath, "appgenesis.json")

		if !cmn.FileExists(gpath) {
			cmn.MustWriteFile(gpath, []byte(genesisTemplate), 0644)
			fmt.Println("path of the app genesis file:" + gpath)
		} else {
			fmt.Println("Genesis file already exist!")
		}
	}

	if !ctx.IsSet("configfile") && !ctx.IsSet("genesisfile") {
		homepath, _ := homedir.Dir()
		if _, err := PathExist(homepath); err != nil {
			return cli.NewExitError(err.Error(), 127)
		}

		os.Mkdir(path.Join(homepath, SUBPATH), os.ModePerm)
		confpath := path.Join(homepath + SUBPATH, "appconfig.toml")

		if !cmn.FileExists(confpath) {
			cmn.MustWriteFile(confpath, []byte(configTemplate), 0644)
			fmt.Println("path of the app config file:" + confpath)
		} else {
			fmt.Println("Config file already exist!")
		}

		gpath := path.Join(homepath + SUBPATH, "appgenesis.json")

		if !cmn.FileExists(gpath) {
			cmn.MustWriteFile(gpath, []byte(genesisTemplate), 0644)
			fmt.Println("path of the app genesis file:" + gpath)
		} else {
			fmt.Println("Genesis file already exist!")
		}
	}

	return nil
}

var configTemplate = `#toml configuration for app
seeds = ""								  # peers to connect when the node is starting
appname = ""							  # organization's name
p2p_laddr = "tcp://0.0.0.0:46656"		  # p2p port that this node is listening
log_path = ""							  #
signbyCA = ""							  # you must require a signature from a valid CA if the blockchain is a permissioned blockchain
cosi_laddr = "tcp://0.0.0.0:46658"		  # cosi port
`

var genesisTemplate = `{
	  "app_hash": "",
	  "chain_id": "",
	  "genesis_time": "0001-01-01T00:00:00.000Z",
	  "plugins": "specialop,querycache",
	  "validators": [
		    {
			      "amount": 100,
			      "is_ca": true,
			      "name": "",
			      "pub_key": [
				        1,
				        ""
			      ]
		    }
	  ]
}
`
