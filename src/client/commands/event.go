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
	"time"

	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	client "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/src/chain/node"
	"github.com/dappledger/AnnChain/src/client/commons"
	cvtools "github.com/dappledger/AnnChain/src/tools"
	"gopkg.in/urfave/cli.v1"
)

type eventActions struct{}

var (
	ea = eventActions{}

	EventCommands = cli.Command{
		Name:     "event",
		Usage:    "trigger the inter-chain event system",
		Category: "Event",
		Subcommands: []cli.Command{
			{
				Name:   "upload-code",
				Usage:  "upload a segment of lisp/lua to process event data for a particular organization",
				Action: ea.UploadCode,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					cli.StringFlag{
						Name:  "code",
						Usage: "code segment",
						Value: "",
					},
					cli.StringFlag{
						Name:  "ownerid",
						Usage: "chain id of the code owner",
						Value: "",
					},
				},
			},
			{
				Name:   "request",
				Usage:  "",
				Action: ea.Request,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					cli.StringFlag{
						Name:  "listener",
						Usage: "listener chainID",
						Value: "",
					},
					cli.StringFlag{
						Name:  "source",
						Usage: "event source chainID",
						Value: "",
					},
					cli.StringFlag{
						Name:  "listener_hash",
						Usage: "code hash on the listener chain",
						Value: "",
					},
					cli.StringFlag{
						Name:  "source_hash",
						Usage: "code hash on the source chain",
						Value: "",
					},
				},
			},
			{
				Name:   "unsubscribe",
				Usage:  "",
				Action: ea.Unsubscribe,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					cli.StringFlag{
						Name:  "listener",
						Usage: "listener chainID",
						Value: "",
					},
					cli.StringFlag{
						Name:  "source",
						Usage: "event source chainID",
						Value: "",
					},
				},
			},
		},
	}
)

func (ea *eventActions) UploadCode(ctx *cli.Context) (err error) {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chain is required", 127)
	}
	if !ctx.IsSet("code") {
		return cli.NewExitError("code is required", 127)
	}
	if !ctx.IsSet("ownerid") {
		return cli.NewExitError("ownerid is required", 127)
	}

	targetChainID, code, ownerid := ctx.GlobalString("target"), ctx.String("code"), ctx.String("ownerid")
	privkeyPtr, err := requirePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	pubkey := (*privkeyPtr).PubKey().(*crypto.PubKeyEd25519)

	tx := &node.EventUploadCodeTx{
		Code:  code,
		Owner: ownerid,
		Time:  time.Now(),
	}
	tx.PubKey = pubkey[:]
	if tx.Signature, err = cvtools.TxSign(tx, privkeyPtr); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txBytes, err := cvtools.TxToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txBytes = agtypes.WrapTx(node.EventUploadCodeTag, txBytes)

	clnt := client.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(agtypes.RPCResult)
	if _, err = clnt.Call("broadcast_tx_sync", []interface{}{targetChainID, txBytes}, tmResult); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	r := (*tmResult).(*agtypes.ResultBroadcastTx)
	if r.Code == pbtypes.CodeType_OK {
		codehash, _ := cvtools.HashKeccak([]byte(code))
		fmt.Println(hex.EncodeToString(codehash))
		return nil
	}

	return nil
}

func (ea *eventActions) Request(ctx *cli.Context) (err error) {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chain is required", 127)
	}
	if !ctx.IsSet("listener") {
		return cli.NewExitError("listener is required", 127)
	}
	if !ctx.IsSet("source") {
		return cli.NewExitError("source is required", 127)
	}
	if !ctx.IsSet("listener_hash") {
		return cli.NewExitError("listener_hash is required", 127)
	}
	if !ctx.IsSet("source_hash") {
		return cli.NewExitError("source_hash is required", 127)
	}

	chainID, listener, source := ctx.GlobalString("target"), ctx.String("listener"), ctx.String("source")
	sourceStr, listenerStr := ctx.String("source_hash"), ctx.String("listener_hash")

	sourceHash, err := hex.DecodeString(sourceStr)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	listenerHash, err := hex.DecodeString(listenerStr)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	privkeyPtr, err := requirePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	pubkey := (*privkeyPtr).PubKey().(*crypto.PubKeyEd25519)

	tx := &node.EventRequestTx{
		Source:       source,
		Listener:     listener,
		SourceHash:   sourceHash,
		ListenerHash: listenerHash,
		Time:         time.Now(),
	}
	tx.PubKey = pubkey[:]
	if tx.Signature, err = cvtools.TxSign(tx, privkeyPtr); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txBytes, err := cvtools.TxToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txBytes = agtypes.WrapTx(node.EventRequestTag, txBytes)

	clnt := client.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(agtypes.RPCResult)
	_, err = clnt.Call("broadcast_tx_commit", []interface{}{chainID, txBytes}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	hash, _ := cvtools.TxHash(tx)
	fmt.Println("tx result:", hex.EncodeToString(hash))

	return nil
}

func (ea *eventActions) Unsubscribe(ctx *cli.Context) (err error) {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chain is required", 127)
	}
	if !ctx.IsSet("listener") {
		return cli.NewExitError("listener is required", 127)
	}
	if !ctx.IsSet("source") {
		return cli.NewExitError("source is required", 127)
	}

	chainID, listener, source := ctx.GlobalString("target"), ctx.String("listener"), ctx.String("source")
	privkeyPtr, err := requirePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	pubkey := (*privkeyPtr).PubKey().(*crypto.PubKeyEd25519)
	tx := &node.EventUnsubscribeTx{
		Source:   source,
		Listener: listener,
		Proof:    []byte{},
		Time:     time.Now(),
	}
	tx.PubKey = pubkey[:]
	if tx.Signature, err = cvtools.TxSign(tx, privkeyPtr); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txBytes, err := cvtools.TxToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txBytes = agtypes.WrapTx(node.EventUnsubscribeTag, txBytes)

	clnt := client.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(agtypes.RPCResult)
	_, err = clnt.Call("broadcast_tx_commit", []interface{}{chainID, txBytes}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	fmt.Println("send ok")

	return nil
}
