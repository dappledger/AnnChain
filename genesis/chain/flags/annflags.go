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

package flags

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	AnnFlags = []cli.Flag{
		cli.StringFlag{
			Name: "privkey",
		},
		cli.BoolFlag{
			Name: "reinit",
		},
		cli.StringFlag{
			Name: "issuer_account",
		},
		cli.StringFlag{
			Name: "terminal_account",
		},
		cli.StringFlag{
			Name:  "moniker",
			Usage: "node name",
		},
		cli.StringFlag{
			Name:  "node_laddr",
			Usage: "node listen address(0.0.0.0:0 means any interface any port)",
		},
		cli.StringFlag{
			Name:  "seeds",
			Usage: "comma separated host:port seed nodes",
		},
		cli.BoolFlag{
			Name:  "pprof",
			Usage: "start pprof server at :6060",
		},
		cli.StringFlag{
			Name:  "rpc_laddr",
			Usage: "RPC listen address. Port required",
		},
		cli.StringFlag{
			Name:  "log_level",
			Usage: "log level",
		},
		cli.BoolTFlag{
			Name:  "pex",
			Usage: "enable peer-exchange (dev feature)",
		},
		cli.BoolTFlag{
			Name:  "fast_sync",
			Usage: "fast blockchain syncing",
		},
		cli.BoolFlag{
			Name:  "skip_upnp",
			Usage: "skip UPNP configuration",
		},
		cli.IntFlag{
			Name:  "connection_reset_wait",
			Usage: "set sleep time when 'connection reset by peer' occurs",
			Value: 300,
		},
		cli.BoolFlag{
			Name:  "addrbook_strict",
			Usage: "disable to book unroutables, ex. RFC1918,RFC3927,RFC4862,RFC4193,RFC4843,Local",
		},
		cli.BoolFlag{
			Name:  "mempool_recheck",
			Usage: "Recheck mempool good txs if any txs were committed in the block.",
		},
		cli.BoolFlag{
			Name:  "test",
			Usage: "initial test accounts",
		},
		cli.BoolFlag{
			Name:  "init_official",
			Usage: "initial official account, give it an great amount of money",
		},
	}
)
