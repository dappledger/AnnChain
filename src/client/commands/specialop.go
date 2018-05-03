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
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/angine/types"
	anginetypes "github.com/dappledger/AnnChain/angine/types"
	gcommon "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/module/lib/go-wire"
	"github.com/dappledger/AnnChain/src/client/commons"
)

const (
	DefaultPower = 10
	DefaultIsCA  = true
)

type specialActions struct{}

var (
	specials        = specialActions{}
	SpecialCommands = cli.Command{
		Name:     "special",
		Usage:    "commands for special operations",
		Category: "Special",
		Subcommands: []cli.Command{
			{
				Name:   "change_validator",
				Usage:  "change validator set, promote node up to be a validator, change validator's voting power, change validator's CA status",
				Action: specials.ChangeValidator,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					anntoolFlags.validatorPubkey,
					anntoolFlags.isCA,
					anntoolFlags.power,
				},
			},
			{
				Name:   "remove_validator",
				Usage:  "remove a validator",
				Action: specials.ChangeValidator,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					anntoolFlags.validatorPubkey,
				},
			},
			{
				Name:   "disconnect_peer",
				Usage:  "disconnect a peer and refuse the node to join the network afterwards",
				Action: specials.DisconnectPeer,
				Flags: []cli.Flag{
					anntoolFlags.peerPubkey,
				},
			},
			{
				Name:   "add_refuse_key",
				Usage:  "add a pubkey to refuseList",
				Action: specials.AddRefuseKey,
				Flags: []cli.Flag{
					anntoolFlags.peerPubkey,
				},
			},
			{
				Name:   "delete_refuse_key",
				Usage:  "delete a pubkey from refuseList",
				Action: specials.DeleteRefuseKey,
				Flags: []cli.Flag{
					anntoolFlags.peerPubkey,
				},
			},
			{
				Name:   "list_refuse_key",
				Usage:  "list all refuse key",
				Action: listRefuseKey,
			},
		},
	}
)

func (act specialActions) newCmd(t string, msg interface{}) *types.SpecialOPCmd {
	c := &types.SpecialOPCmd{}
	c.CmdType = t
	c.Time = time.Now()
	if err := c.LoadMsg(msg); err != nil {
		fmt.Println(err)
		return nil
	}

	return c
}

func (act specialActions) constructJSONRPCPayload(c *types.SpecialOPCmd) ([]byte, error) {
	return types.TagSpecialOPTx(wire.BinaryBytes(c)), nil
}

func (act specialActions) jsonRPC(chainID string, p []byte) (*anginetypes.RPCResult, error) {
	clt := rpcclient.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(anginetypes.RPCResult)
	_, err := clt.Call("request_special_op", []interface{}{chainID, p}, tmResult)
	if err != nil {
		return nil, err
	}
	return tmResult, nil
}

func (act specialActions) makeValidatorAttrWithPk(pk string) (*anginetypes.ValidatorAttr, error) {
	pubkeyByte32, err := types.StringTo32byte(pk)
	if err != nil {
		return nil, err
	}
	validator := &anginetypes.ValidatorAttr{
		PubKey: pubkeyByte32[:],
		Power:  DefaultPower,
		IsCA:   DefaultIsCA,
	}
	return validator, nil
}

func (act specialActions) ChangeValidator(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chain is required", 127)
	}
	if !ctx.IsSet("validator_pubkey") {
		return cli.NewExitError("missing 2validator's pubkey", 127)
	}
	privKey, err := requirePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	chainID, validatorPub := ctx.GlobalString("target"), gcommon.SanitizeHex(ctx.String("validator_pubkey"))
	power := ctx.Uint64("power")
	isCA := ctx.Bool("isCA")
	vAttr, err := act.makeValidatorAttrWithPk(validatorPub)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	vAttr.IsCA = isCA
	vAttr.Power = power

	scmd := act.newCmd(types.SpecialOP_ChangeValidator, vAttr)
	pubkey := privKey.PubKey().(*crypto.PubKeyEd25519)
	scmd.PubKey = pubkey[:]

	signMessage, _ := json.Marshal(scmd)
	signature := privKey.Sign(signMessage).(*crypto.SignatureEd25519)
	scmd.Signature = signature[:]

	cmdBytes, _ := json.Marshal(scmd)

	res, err := act.jsonRPC(chainID, types.TagSpecialOPTx(cmdBytes))
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println((*res).(*anginetypes.ResultRequestSpecialOP))

	return nil
}

func (act specialActions) DisconnectPeer(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chain is required", 127)
	}
	chainID := ctx.GlobalString("target")
	peerPubByte32, err := keyTo32byte(ctx)
	if err != nil {
		return err
	}
	scmd := act.newCmd(types.SpecialOP_Disconnect, peerPubByte32[:])
	payload, err := act.constructJSONRPCPayload(scmd)
	if err != nil {
		return err
	}
	res, err := act.jsonRPC(chainID, payload)
	if err != nil {
		return err
	}

	fmt.Println((*res).(*anginetypes.ResultRequestSpecialOP))

	return nil
}

func (act specialActions) AddRefuseKey(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("missing chainid", 127)
	}
	chainID := ctx.GlobalString("target")

	peerPubByte32, err := keyTo32byte(ctx)
	if err != nil {
		return err
	}
	scmd := act.newCmd(types.SpecialOP_AddRefuseKey, peerPubByte32[:])
	payload, err := act.constructJSONRPCPayload(scmd)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	res, err := act.jsonRPC(chainID, payload)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println((*res).(*anginetypes.ResultRequestSpecialOP))
	return nil
}

func (act specialActions) DeleteRefuseKey(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("missing chainid", 127)
	}
	chainID := ctx.GlobalString("target")

	peerPubByte32, err := keyTo32byte(ctx)
	if err != nil {
		return err
	}
	scmd := act.newCmd(types.SpecialOP_DeleteRefuseKey, peerPubByte32[:])
	payload, err := act.constructJSONRPCPayload(scmd)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	res, err := act.jsonRPC(chainID, payload)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println((*res).(*anginetypes.ResultRequestSpecialOP))
	return nil
}

func listRefuseKey(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("missing chainid", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := rpcclient.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(anginetypes.RPCResult)
	_, err := clientJSON.Call("list_refuse_key", []interface{}{chainID}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	res := (*tmResult).(*anginetypes.ResultRefuseList)
	fmt.Println("refuseList: ", res.Result)
	return nil
}

func keyTo32byte(ctx *cli.Context) (peerByte32 [32]byte, err error) {
	if !ctx.IsSet("peer_pubkey") {
		err = cli.NewExitError("missing peer's pubkey", 127)
		return
	}
	peerPub := gcommon.SanitizeHex(ctx.String("peer_pubkey"))
	peerByte32, err = types.StringTo32byte(peerPub)
	return
}
