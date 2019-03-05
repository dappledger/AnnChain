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

package main

import (
	"encoding/hex"
	"fmt"
	_ "net/http/pprof"
	"os"
	"path"
	"strings"

	acfg "github.com/dappledger/AnnChain/angine/config"
	"github.com/dappledger/AnnChain/ann-module/lib/ed25519"
	"github.com/dappledger/AnnChain/ann-module/lib/go-config"
	"github.com/dappledger/AnnChain/genesis/chain/app"
	_ "github.com/dappledger/AnnChain/genesis/chain/app"
	dcfg "github.com/dappledger/AnnChain/genesis/chain/config"
	"github.com/dappledger/AnnChain/genesis/chain/flags"
	"github.com/dappledger/AnnChain/genesis/chain/log"
	"github.com/dappledger/AnnChain/genesis/chain/node"
	"github.com/dappledger/AnnChain/genesis/chain/start"
	"github.com/dappledger/AnnChain/genesis/chain/version"
	"github.com/dappledger/AnnChain/genesis/eth/core/state"
	"github.com/dappledger/AnnChain/genesis/tools"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"gopkg.in/urfave/cli.v1"
)

const (
	DataDirFlag = "datadir"
)

var (
	annConf *config.MapConfig

	NodeCommand = cli.Command{
		Name:     "node",
		Usage:    "start a node",
		Category: "Blockchain Commands",
		Action:   createNode,
	}
	ShowValidatorCommand = cli.Command{
		Name:     "show_validator",
		Action:   showVar,
		Usage:    "show validators",
		Category: "Blockchain Commands",
	}
	GenValidatorCommand = cli.Command{
		Name:     "gen_validator",
		Action:   genVar,
		Usage:    "generate validators",
		Category: "Blockchain Commands",
	}
	ProbeUpnpCommand = cli.Command{
		Name:     "probe_upnp",
		Action:   probeupnp,
		Usage:    "probe the upnp",
		Category: "Blockchain Commands",
	}
	VersionCommand = cli.Command{
		Name:     "version",
		Action:   prtversion,
		Usage:    "print the version",
		Category: "Blockchain Commands",
	}
	InitCommand = cli.Command{
		Name:     "init",
		Action:   geninit,
		Usage:    "init annchain",
		Category: "Blockchain Commands",
	}
	RsetCommand = cli.Command{
		Name:     "reset",
		Action:   genreset,
		Usage:    "Reset Privalidator,clean the data and shards",
		Category: "Blockchain Commands",
	}
	SignCommand = cli.Command{
		Name:     "sign",
		Action:   signCA,
		Usage:    "sign CA info",
		Category: "CA Commands",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "sec",
				Value: "0",
			},
			cli.StringFlag{
				Name:  "pub",
				Value: "0",
			},
		},
	}

	logger *zap.Logger
)

func main() {

	app := cli.NewApp()
	app.Name = "Genesis"
	app.HideVersion = true
	app.Version = version.GetVersion()
	app.Copyright = "COPYRIGHT GOES TO ZHONGAN TECHNOLOGY"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "admin",
			Email: ""}}

	app.Flags = flags.GlobalFlags
	NodeCommand.Flags = append(NodeCommand.Flags, flags.AnnFlags...)

	app.Commands = []cli.Command{
		InitCommand,
		NodeCommand,
		VersionCommand,
		SignCommand,
		RsetCommand,
	}

	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalIsSet(DataDirFlag) {
			annConf = acfg.GetConfig(tools.ExpandPath(ctx.GlobalString(DataDirFlag)))
		} else {
			annConf = acfg.GetConfig("")
		}
		env := annConf.GetString("environment")
		logMode := annConf.GetString("log_mode")
		logpath := ctx.String("log_path")
		if logpath == "" {
			var err error
			logpath, err = os.Getwd()
			if annConf.IsSet("log_path") {
				logpath = path.Join(logpath, annConf.GetString("log_path"))
				os.Mkdir(logpath, os.ModePerm)
			} else {
				if err != nil {
					cli.NewExitError(err.Error(), -1)
				}
			}
		}
		logger = log.Initialize(logMode, env, path.Join(logpath, "node.output.log"), path.Join(logpath, "node.err.log"))
		state.Init(logger)

		return nil
	}

	app.After = func(ctx *cli.Context) error {
		return nil
	}

	defer func() {
		if logger != nil {
			logger.Sync()
		}
	}()

	if err := app.Run(os.Args); err != nil {

	}
}

func createNode(ctx *cli.Context) {
	parseFlags(annConf, ctx)
	dcfg.LoadDefaultConfig(annConf, ctx.GlobalString(DataDirFlag))
	initApp := app.NewGenesisApp(annConf, logger)
	node.RunNode(logger, annConf, initApp)
}

func showVar(ctx *cli.Context) {
	parseFlags(annConf, ctx)
	start.Show_validator(logger, annConf)
}

func genVar(ctx *cli.Context) {
	parseFlags(annConf, ctx)
	start.Gen_validator(logger, annConf)
}

func probeupnp(ctx *cli.Context) {
	parseFlags(annConf, ctx)
	start.Probe_upnp(logger)
}

func prtversion(ctx *cli.Context) {
	fmt.Println(version.GetCommitVersion())
}

func geninit(ctx *cli.Context) {
	parseFlags(annConf, ctx)
	start.Initfiles(annConf)
}

func genreset(ctx *cli.Context) {
	parseFlags(annConf, ctx)
	start.Reset_all(logger, annConf)
}

func parseFlags(conf config.Config, ctx *cli.Context) {
	conf.Set("context", ctx)
	fillNodeConfigs(conf, ctx)
	fillP2PConfigs(conf, ctx)
}

func fillP2PConfigs(conf config.Config, ctx *cli.Context) {
	if ctx.IsSet("connection_reset_wait") {
		t := ctx.Int("connection_reset_wait")
		m := conf.GetMap("p2p")
		m["connection_reset_wait"] = t
		conf.Set("p2p", m)
	}
}

func fillNodeConfigs(conf config.Config, ctx *cli.Context) {
	toBeMerged := []string{"moniker", "node_laddr", "seeds", "rpc_laddr", "log_level"}
	for _, name := range toBeMerged {
		if ctx.IsSet(name) {
			conf.Set(name, ctx.String(name))
		}
	}
	conf.Set("pprof", ctx.Bool("pprof"))
	conf.Set("test", ctx.Bool("test"))
	conf.Set("fast_sync", ctx.BoolT("fast_sync"))
	conf.Set("skip_upnp", ctx.Bool("skip_upnp"))
	conf.Set("pex_reactor", ctx.BoolT("pex"))
	conf.Set("addrbook_strict", ctx.Bool("addrbook_strict"))
	conf.Set("mempool_recheck", ctx.Bool("mempool_recheck"))
}

func signCA(ctx *cli.Context) error {
	sec := ctx.String("sec")
	pub := ctx.String("pub")

	if sec == "" || pub == "" {
		fmt.Println("sec or pub is null. exit")
		return nil
	}

	if len(sec) != 128 {
		fmt.Println("Invalid sec:", sec)
		return nil
	}

	skBs, err := hex.DecodeString(sec)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	var sk [ed25519.PrivateKeySize]byte

	copy(sk[:], skBs)

	for _, p := range strings.Split(pub, ",") {
		p = strings.Trim(p, " ")
		if len(p) != 64 {
			fmt.Println("Invalid pub:", p)
			return nil
		}
		pubBytes, _ := hex.DecodeString(p)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}
		sig := ed25519.Sign(&sk, pubBytes)
		ss := hex.EncodeToString(sig[:])
		fmt.Printf("%s : %s\n", p, ss)
	}

	return nil
}
