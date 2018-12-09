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
	//"errors"
	"fmt"
	"github.com/dappledger/AnnChain/src/client/commons"
	"time"
	//abcit "github.com/dappledger/AnnChain/ann-chain/abci/types"
	"github.com/dappledger/AnnChain/angine/types"
	angtypes "github.com/dappledger/AnnChain/angine/types"
	gcommon "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	//"github.com/dappledger/AnnChain/module/lib/go-wire"
	"gopkg.in/urfave/cli.v1"
)

type voteActions struct{}

var (
	vote         = voteActions{}
	VoteCommands = cli.Command{
		Name:     "vote",
		Usage:    "commands for vote channel request",
		Category: "Vote",
		Subcommands: []cli.Command{
			{
				Name:   "request_change_validator",
				Usage:  "make a new request to change validator",
				Action: vote.voteChangeValidator,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					cli.StringFlag{
						Name: "validator_pubkey",
					},
					anntoolFlags.power,
					anntoolFlags.isCA,
				},
			},
			{
				Name:   "sign_vote",
				Usage:  "sign a vote request, if disagree sign nil",
				Action: vote.signVote,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					cli.StringFlag{
						Name: "request_id",
					},
				},
			},
			{
				Name:   "query_votes",
				Usage:  "get all vote channel reqeusts state",
				Action: vote.queryVotes,
			},
			{
				Name:   "execute_vote",
				Usage:  "do the vote request",
				Action: vote.executeVote,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name: "request_id",
					},
				},
			},
		},
	}
)

func (act voteActions) voteChangeValidator(ctx *cli.Context) error {
	validatorPub := ctx.String("validator_pubkey")
	if validatorPub == "" {
		return cli.NewExitError("missing validator's pubkey", 127)
	}
	chainID := ctx.GlobalString("target")
	validatorPub = gcommon.SanitizeHex(validatorPub)
	privkey, err := requirePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	power := ctx.Uint64("power")
	isCA := ctx.Bool("isCA")
	pubyte32, err := angtypes.StringTo32byte(validatorPub)
	if err != nil {
		return err
	}

	//construct special op cmd as vote channel MSG
	validator := &angtypes.ValidatorAttr{
		PubKey: pubyte32[:],
		Power:  DefaultPower,
		IsCA:   DefaultIsCA,
	}
	validator.IsCA = isCA
	validator.Power = power
	spopcmd := act.newSPOPcmd(angtypes.SpecialOP_ChangeValidator, validator)
	pubkey := privkey.PubKey().(*crypto.PubKeyEd25519)
	spopcmd.PubKey = pubkey[:]

	signMessage, _ := json.Marshal(spopcmd)
	signature := privkey.Sign(signMessage).(*crypto.SignatureEd25519)
	spopcmd.Signature = signature[:]

	spopBytes, _ := json.Marshal(spopcmd)

	//construct vote channel cmd
	var votecmd types.VoteChannelCmd
	votecmd.Msg = types.TagSpecialOPTx(spopBytes)
	votecmd.CmdCode = types.VoteChannel
	votecmd.SubCmd = types.VoteChannel_NewRequest
	votecmd.Votetype = types.SpecialOP_ChangeValidator

	cmdBytes, _ := json.Marshal(votecmd)
	payload := types.TagVoteChannelTx(cmdBytes)
	client := rpcclient.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(angtypes.RPCResult)
	fmt.Println("get params :", chainID, power, isCA, privkey.KeyString())
	_, err = client.Call("request_vote_channel", []interface{}{chainID, payload}, tmResult)
	if err != nil {
		fmt.Println("exit after call request ...", err)
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println("result :", *tmResult)
	//fmt.Println((*tmResult).(*angtypes.ResultRequestVoteChannel))
	return nil
}

func (act voteActions) signVote(ctx *cli.Context) error {
	chainID := ctx.GlobalString("target")
	reqid := ctx.String("request_id")
	if reqid == "" {
		return cli.NewExitError("missing request id", 127)
	}

	privkey, err := requirePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	//signpub := ctx.String("sign_pubkey")
	//signpub = gcommon.SanitizeHex(signpub)
	//signb32, err := types.StringTo32byte(signpub)
	//if err != nil {
	//	return err
	//}

	bytereqid, _ := types.StringToAnybyte(reqid)
	reqsig := crypto.SignatureEd25519{}
	copy(reqsig[:], bytereqid[:])
	//reqsig, err := crypto.SignatureFromBytes(bytereqid)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//}

	var votecmd types.VoteChannelCmd
	votecmd.Id = reqsig.Bytes()
	votecmd.CmdCode = types.VoteChannel
	votecmd.SubCmd = types.VoteChannel_Sign
	sig := privkey.Sign(votecmd.Id)

	//signpubkey := crypto.PubKeyEd25519{}
	//copy(signpubkey[:], sig[:])
	votecmd.Txmsg = sig.Bytes()

	cmdbytes, _ := json.Marshal(votecmd)
	payload := types.TagVoteChannelTx(cmdbytes)
	client := rpcclient.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(angtypes.RPCResult)
	_, err = client.Call("request_vote_channel", []interface{}{chainID, payload}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println((*tmResult).(*angtypes.ResultRequestVoteChannel))
	return nil
}

func (act voteActions) queryVotes(ctx *cli.Context) error {
	chainID := ctx.GlobalString("target")
	var votecmd types.VoteChannelCmd
	votecmd.CmdCode = types.VoteChannel
	votecmd.SubCmd = types.VoteChannel_query

	cmdbytes, _ := json.Marshal(votecmd)
	payload := types.TagVoteChannelTx(cmdbytes)
	client := rpcclient.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(angtypes.RPCResult)
	_, err := client.Call("request_vote_channel", []interface{}{chainID, payload}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	var allReq []types.VoteChannelCmd
	if err = json.Unmarshal((*tmResult).(*angtypes.ResultRequestVoteChannel).Data, &allReq); err != nil {
		fmt.Println(err)
		return err
	}
	for _, val := range allReq {
		sigb := crypto.SignatureEd25519{}
		copy(sigb[:], val.Id[:])
		pub := crypto.PubKeyEd25519{}
		copy(pub[:], val.Sender[:])
		fmt.Println("ID: ", fmt.Sprintf("%X", sigb), "vote type :", val.Votetype, "req sender: ", pub.KeyString())
	}
	return nil
}

func (act voteActions) executeVote(ctx *cli.Context) error {
	chainID := ctx.GlobalString("target")
	reqid := ctx.String("request_id")
	if reqid == "" {
		return cli.NewExitError("missing request id", 127)
	}

	bytereqid, _ := types.StringToAnybyte(reqid)
	var votecmd types.VoteChannelCmd
	reqsig := crypto.SignatureEd25519{}
	copy(reqsig[:], bytereqid[:])
	//reqsig, err := crypto.SignatureFromBytes(bytereqid)
	//if err != nil {
	//	return err
	//}
	votecmd.Id = reqsig.Bytes()
	votecmd.CmdCode = types.VoteChannel
	votecmd.SubCmd = types.VoteChannel_Exec

	cmdbytes, _ := json.Marshal(votecmd)
	payload := types.TagVoteChannelTx(cmdbytes)
	client := rpcclient.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(angtypes.RPCResult)
	_, err := client.Call("request_vote_channel", []interface{}{chainID, payload}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println((*tmResult).(*angtypes.ResultRequestVoteChannel))
	return nil
}

func (act voteActions) newSPOPcmd(t string, msg interface{}) *types.SpecialOPCmd {
	c := &types.SpecialOPCmd{}
	c.CmdType = t
	c.Time = time.Now()
	if err := c.LoadMsg(msg); err != nil {
		fmt.Println(err)
		return nil
	}

	return c
}
