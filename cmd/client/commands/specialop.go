package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"

	gcrypto "github.com/dappledger/AnnChain/gemmill/go-crypto"
	gcommon "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/rpc/client"
	"github.com/dappledger/AnnChain/gemmill/types"
	"github.com/dappledger/AnnChain/cmd/client/commons"
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
					anntoolFlags.validatorPubkey,
					anntoolFlags.isCA,
					anntoolFlags.power,
					anntoolFlags.cType,
				},
			},
			{
				Name:   "remove_validator",
				Usage:  "remove a validator",
				Action: specials.ChangeValidator,
				Flags: []cli.Flag{
					anntoolFlags.validatorPubkey,
					anntoolFlags.cType,
				},
			},
			{
				Name:   "disconnect_peer",
				Usage:  "disconnect a peer and refuse the node to join the network afterwards",
				Action: specials.DisconnectPeer,
				Flags: []cli.Flag{
					anntoolFlags.validatorPubkey,
					anntoolFlags.cType,
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
					anntoolFlags.cType,
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
	//return agtypes.TagSpecialOPTx(wire.BinaryBytes(c)), nil
	bytedata, _ := json.Marshal(c)
	return types.TagSpecialOPTx(bytedata), nil
}

func (act specialActions) jsonRPC(p []byte) (*types.ResultRequestSpecialOP, error) {
	clt := rpcclient.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(types.ResultRequestSpecialOP)
	_, err := clt.Call("request_special_op", []interface{}{p}, rpcResult)
	if err != nil {
		return nil, err
	}
	return rpcResult, nil
}

func (act specialActions) makeValidatorAttrWithPk(pk string) (*types.ValidatorAttr, error) {
	if strings.HasPrefix(pk, "0x") || strings.HasPrefix(pk, "0X") {
		pk = pk[2:]
	}
	pub, err := hex.DecodeString(pk)
	if err != nil {
		return nil, err
	}
	if len(pub) != gcrypto.NodePubkeyLen() {
		return nil, fmt.Errorf("pubkey format error:need %d's bytes;but %d", gcrypto.NodePubkeyLen(), len(pub))
	}
	validator := &types.ValidatorAttr{
		PubKey: pub,
		Power:  DefaultPower,
		IsCA:   DefaultIsCA,
	}
	return validator, nil
}

func (act specialActions) ChangeValidator(ctx *cli.Context) error {
	if !ctx.IsSet("validator_pubkey") {
		return cli.NewExitError("missing 2validator's pubkey", 127)
	}
	privKey, err := requireNodePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	validatorPub := gcommon.SanitizeHex(ctx.String("validator_pubkey"))
	power := ctx.Uint64("power")
	isCA := ctx.Bool("isCA")
	vAttr, err := act.makeValidatorAttrWithPk(validatorPub)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	vAttr.IsCA = isCA
	vAttr.Power = power

	scmd := act.newCmd(types.SpecialOP_ChangeValidator, vAttr)
	scmd.PubKey, _ = hex.DecodeString(privKey.PubKey().KeyString())

	signMessage, _ := json.Marshal(scmd)
	scmd.Signature, _ = hex.DecodeString(privKey.Sign(signMessage).KeyString())

	cmdBytes, _ := json.Marshal(scmd)

	res, err := act.jsonRPC(types.TagSpecialOPTx(cmdBytes))
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println(*res)

	return nil
}

func (act specialActions) DisconnectPeer(ctx *cli.Context) error {
	if !ctx.IsSet("validator_pubkey") {
		return cli.NewExitError("missing 2validator's pubkey", 127)
	}
	privKey, err := requireNodePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	validatorPub := gcommon.SanitizeHex(ctx.String("validator_pubkey"))
	msg, err := hex.DecodeString(validatorPub)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	scmd := act.newCmd(types.SpecialOP_Disconnect, nil)
	scmd.Msg = msg

	if err = fillSigInCmd(scmd, privKey); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	cmdBytes, err := json.Marshal(scmd)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res, err := act.jsonRPC(types.TagSpecialOPTx(cmdBytes))
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println(*res)

	return nil
}

func (act specialActions) AddRefuseKey(ctx *cli.Context) error {

	peerPubByte32, err := keyTo32byte(ctx)
	if err != nil {
		return err
	}
	scmd := act.newCmd(types.SpecialOP_AddRefuseKey, peerPubByte32[:])
	payload, err := act.constructJSONRPCPayload(scmd)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	res, err := act.jsonRPC(payload)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println(*res)
	return nil
}

func fillSigInCmd(scmd *types.SpecialOPCmd, privKey gcrypto.PrivKey) error {
	var err error
	scmd.PubKey, err = hex.DecodeString(privKey.PubKey().KeyString())
	if err != nil {
		return err
	}
	signMessage, err := json.Marshal(scmd)
	if err != nil {
		return err
	}
	scmd.Signature, err = hex.DecodeString(privKey.Sign(signMessage).KeyString())
	return err
}

func (act specialActions) DeleteRefuseKey(ctx *cli.Context) error {
	privKey, err := requireNodePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	peerPubByte32, err := keyTo32byte(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	scmd := act.newCmd(types.SpecialOP_DeleteRefuseKey, peerPubByte32[:])

	if err = fillSigInCmd(scmd, privKey); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	payload, err := act.constructJSONRPCPayload(scmd)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	res, err := act.jsonRPC(payload)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println(*res)
	return nil
}

func listRefuseKey(ctx *cli.Context) error {
	clientJSON := rpcclient.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(types.RPCResult)
	_, err := clientJSON.Call("list_refuse_key", []interface{}{}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	res := (*rpcResult).(*types.ResultRefuseList)
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
