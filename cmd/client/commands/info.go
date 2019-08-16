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
	"encoding/json"
	"fmt"

	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/cmd/client/commons"
	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
	"github.com/dappledger/AnnChain/gemmill/types"
)

var (
	InfoCommand = cli.Command{
		Name:  "info",
		Usage: "get annchain info",
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "last_block",
				Action: lastBlockInfo,
			},
			cli.Command{
				Name:   "num_unconfirmed_txs",
				Action: numUnconfirmedTxs,
			},
			cli.Command{
				Name:   "net",
				Action: netInfo,
			},
			cli.Command{
				Name:   "num_archived_blocks",
				Action: numArchivedBlocks,
			},
		},
	}
)

func lastBlockInfo(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(types.ResultInfo)
	_, err := clientJSON.Call("info", nil, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	var jsbytes []byte
	jsbytes, err = json.Marshal(rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	fmt.Println(string(jsbytes))
	return nil
}

func numUnconfirmedTxs(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(types.ResultUnconfirmedTxs)
	_, err := clientJSON.Call("num_unconfirmed_txs", nil, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println("num of unconfirmed txs: ", rpcResult.N)
	return nil
}

func netInfo(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(types.ResultNetInfo)
	_, err := clientJSON.Call("net_info", nil, rpcResult)
	if err != nil {
		panic(err)
	}
	fmt.Println("listening :", rpcResult.Listening)
	for _, l := range rpcResult.Listeners {
		fmt.Println("listener :", l)
	}
	for _, p := range rpcResult.Peers {
		fmt.Println("peer address :", p.RemoteAddr,
			" pub key :", p.PubKey,
			" send status :", p.ConnectionStatus.SendMonitor.Active,
			" recieve status :", p.ConnectionStatus.RecvMonitor.Active)
	}
	return nil
}

func numArchivedBlocks(ctx *cli.Context) error {

	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(types.ResultNumArchivedBlocks)
	_, err := clientJSON.Call("num_archived_blocks", []interface{}{}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println("num of archived blocks: ", rpcResult.Num)
	return nil
}
