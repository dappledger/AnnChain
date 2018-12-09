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
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/angine/types"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/src/client/commons"
)

var (
	TxCommands = cli.Command{
		Name:     "tx",
		Usage:    "operations for transaction",
		Category: "Transaction",
		Subcommands: []cli.Command{
			{
				Name:   "send",
				Usage:  "send a transaction",
				Action: sendTx,
				Flags: []cli.Flag{
					anntoolFlags.addr,
					anntoolFlags.pwd,
					anntoolFlags.payload,
					anntoolFlags.privkey,
					anntoolFlags.nonce,
					anntoolFlags.to,
					anntoolFlags.value,
				},
			},
			{
				Name:   "query",
				Usage:  "query transaction execution info",
				Action: txQuery,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name: "txHash",
					},
				},
			},
		},
	}
)

func txQuery(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is required", 127)
	}
	if !ctx.IsSet("txHash") {
		return cli.NewExitError("txHash is required", 127)
	}
	chainID, txhash := ctx.GlobalString("target"), ctx.String("txHash")
	hashBytes, err := hex.DecodeString(txhash)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	payload := make([]byte, len(hashBytes)+1)
	payload[0] = types.QueryTxExecution
	copy(payload[1:], hashBytes)

	tmResult := new(types.RPCResult)
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	if _, err = clientJSON.Call("query", []interface{}{chainID, payload}, tmResult); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	res := (*tmResult).(*types.ResultQuery)

	info := &types.TxExecutionResult{}
	info.FromBytes(res.Result.Data)

	fmt.Printf("%+v\n", info)
	return nil
}

func sendTx(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is required", 127)
	}
	chainID := ctx.GlobalString("target")
	privkey := ctx.String("privkey")
	address := ctx.String("address")
	passwd := ctx.String("passwd")

	nonce := ctx.Uint64("nonce")
	to := common.HexToAddress(ctx.String("to"))
	value := ctx.Int64("value")
	payload := ctx.String("payload")
	data := common.Hex2Bytes(payload)
	if len(data) == 0 {
		data = []byte(payload)
	}

	if privkey == "" && (address == "" || passwd == "") {
		panic("should provide privkey or address-passwd pair.")
	}

	tx := ethtypes.NewTransaction(nonce, to, big.NewInt(value), big.NewInt(90000), big.NewInt(0), data)

	if privkey != "" {
		key, err := crypto.HexToECDSA(privkey)
		if err != nil {
			panic(err)
		}
		sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
		sigTx, err := tx.WithSignature(ethSigner, sig)

		b, err := rlp.EncodeToBytes(sigTx)
		if err != nil {
			panic(err)
		}

		tmResult := new(types.RPCResult)
		clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
		_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{chainID, b}, tmResult)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}

		fmt.Println("tx result:", sigTx.Hash().Hex())
	}

	return nil
}
