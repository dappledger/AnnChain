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
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	ac "github.com/dappledger/AnnChain/module/lib/go-common"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	civil "github.com/dappledger/AnnChain/src/chain/node"
	"github.com/dappledger/AnnChain/src/client/commons"
)

var (
	QueryCommands = cli.Command{
		Name:     "query",
		Usage:    "operations for query state",
		Category: "Query",
		Subcommands: []cli.Command{
			{
				Name:   "nonce",
				Usage:  "query account's nonce",
				Action: queryNonce,
				Flags: []cli.Flag{
					anntoolFlags.addr,
				},
			},
			{
				Name:   "balance",
				Usage:  "query account's balance",
				Action: queryBalance,
				Flags: []cli.Flag{
					anntoolFlags.addr,
				},
			},
			{
				Name:   "receipt",
				Usage:  "",
				Action: queryReceipt,
				Flags: []cli.Flag{
					anntoolFlags.hash,
				},
			},
			{
				Name:   "events",
				Usage:  "query events on the node",
				Action: queryEvents,
				Flags:  []cli.Flag{},
			},
			{
				Name:   "event_code",
				Usage:  "",
				Action: queryEventCode,
				Flags: []cli.Flag{
					anntoolFlags.codeHash,
				},
			},
			{
				Name:   "apps",
				Usage:  "query apps on the node",
				Action: queryNodeApps,
				Flags:  []cli.Flag{},
			},
		},
	}
)

func queryNonce(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	addrHex := ac.SanitizeHex(ctx.String("address"))
	addr := common.Hex2Bytes(addrHex)
	query := append([]byte{1}, addr...)

	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)
	nonce := new(uint64)
	rlp.DecodeBytes(res.Result.Data, nonce)

	fmt.Println("query result:", *nonce)

	return nil
}

func queryBalance(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	addrHex := ac.SanitizeHex(ctx.String("address"))
	addr := common.Hex2Bytes(addrHex)
	query := append([]byte{2}, addr...)

	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)
	balance := big.NewInt(0)
	rlp.DecodeBytes(res.Result.Data, balance)

	fmt.Println("query result:", balance.String())

	return nil
}

func queryReceipt(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	hashHex := ac.SanitizeHex(ctx.String("hash"))
	hash := common.Hex2Bytes(hashHex)
	query := append([]byte{3}, hash...)
	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	receiptForStorage := new(ethtypes.ReceiptForStorage)

	err = rlp.DecodeBytes(res.Result.Data, receiptForStorage)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	receiptJSON, _ := json.Marshal(receiptForStorage)
	fmt.Println("query result:", string(receiptJSON))

	return nil
}

func queryTx(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")

	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	hashHex := ac.SanitizeHex(ctx.String("hash"))
	fmt.Println(hashHex)
	hash := common.Hex2Bytes(hashHex)
	query := append([]byte{3}, hash...)
	fmt.Println(len(query))
	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	receiptForStorage := new(ethtypes.ReceiptForStorage)

	err = rlp.DecodeBytes(res.Result.Data, receiptForStorage)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println("query result:", receiptForStorage.ContractAddress.Hex())

	return nil
}

func queryEvents(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")

	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	query := []byte{civil.QueryEvents}
	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	if res.Result.IsOK() {
		buffers := bytes.NewBuffer(res.Result.Data)
		dec := gob.NewDecoder(buffers)
		keys := make([]string, 0)
		if err := dec.Decode(&keys); err != nil {
			return cli.NewExitError("fail to decode result: "+err.Error(), 127)
		}

		fmt.Println(keys)
		return nil
	}

	return cli.NewExitError(res.Result.Log, 127)
}

func queryEventCode(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	if !ctx.IsSet("code_hash") {
		return cli.NewExitError("query code_hash is missing", 127)
	}
	chainID := ctx.GlobalString("target")

	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	code_hash, err := hex.DecodeString(ctx.String("code_hash"))
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	_, err = clientJSON.Call("event_code", []interface{}{chainID, code_hash}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	if res.Result.IsOK() {
		fmt.Println(string(res.Result.Data))
		return nil
	}

	return cli.NewExitError(res.Result.Log, 127)
}

func queryNodeApps(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	_, err := clientJSON.Call("organizations", nil, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	res := (*tmResult).(*types.ResultOrgs)
	fmt.Println(*res)
	return nil
}
