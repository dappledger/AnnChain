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
	"encoding/json"
	"fmt"

	"github.com/dappledger/AnnChain/angine/types"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/src/client/commons"
	"gopkg.in/urfave/cli.v1"
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
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	_, err := clientJSON.Call("info", []interface{}{chainID}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	res := (*tmResult).(*types.ResultInfo)
	var jsbytes []byte
	jsbytes, err = json.Marshal(res)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	fmt.Println(string(jsbytes))
	return nil
}

func numUnconfirmedTxs(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	_, err := clientJSON.Call("num_unconfirmed_txs", []interface{}{chainID}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultUnconfirmedTxs)

	fmt.Println("num of unconfirmed txs: ", res.N)
	return nil
}

func netInfo(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	_, err := clientJSON.Call("net_info", []interface{}{chainID}, tmResult)
	if err != nil {
		panic(err)
	}
	res := (*tmResult).(*types.ResultNetInfo)
	fmt.Println("listening :", res.Listening)
	for _, l := range res.Listeners {
		fmt.Println("listener :", l)
	}
	for _, p := range res.Peers {
		fmt.Println("peer address :", p.RemoteAddr,
			" pub key :", p.PubKey,
			" send status :", p.ConnectionStatus.SendMonitor.Active,
			" recieve status :", p.ConnectionStatus.RecvMonitor.Active)
	}
	return nil
}

func numArchivedBlocks(ctx *cli.Context) error {

	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	_, err := clientJSON.Call("num_archived_blocks", []interface{}{chainID}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultNumArchivedBlocks)

	fmt.Println("num of archived blocks: ", res.Num)
	return nil
}
