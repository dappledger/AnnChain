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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/chain/types"
	"github.com/dappledger/AnnChain/cmd/client/commons"
	"github.com/dappledger/AnnChain/eth/accounts/abi"
	"github.com/dappledger/AnnChain/eth/core"
	etypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	gcommon "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/rpc/client"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
)

type adminActions struct{}

var (
	admins        = adminActions{}
	AdminCommands = cli.Command{
		Name:     "admin",
		Usage:    "commands for admin operations",
		Category: "Admin",
		Subcommands: []cli.Command{
			{
				Name:   "add_peer",
				Usage:  "add node<peer> into validator_set;",
				Action: AddPeer,
				Flags: []cli.Flag{
					anntoolFlags.cType,
					anntoolFlags.verbose,
					anntoolFlags.nPrivs,
				},
			},
			{
				Name:   "remove_node",
				Usage:  "remove node from validator_set",
				Action: RemoveNode,
				Flags: []cli.Flag{
					anntoolFlags.validatorPubkey,
					anntoolFlags.verbose,
					anntoolFlags.nPrivs,
				},
			},
			{
				Name:   "change_node",
				Usage:  "change node of validator_set",
				Action: UpdateNode,
				Flags: []cli.Flag{
					anntoolFlags.validatorPubkey,
					anntoolFlags.power,
					anntoolFlags.cType,
					anntoolFlags.verbose,
					anntoolFlags.nPrivs,
				},
			},
		},
	}
)

func Power() string {
	return anntoolFlags.power.GetName()
}

func CType() string {
	return anntoolFlags.cType.GetName()
}

func Verbose() string {
	return anntoolFlags.verbose.GetName()
}

func PubKey() string {
	return anntoolFlags.validatorPubkey.GetName()
}

func PrivKey() string {
	return anntoolFlags.validatorPrivKey.GetName()
}

func NPrivs() string {
	return anntoolFlags.nPrivs.GetName()
}

//-------------------------------------------------------------------
func AddPeer(ctx *cli.Context) error {
	crypto.NodeInit(ctx.String("crypto_type"))
	fmt.Println("Input Privkey of addnode  for user:")
	key, err := readNodePrivKey(ctx)
	if err != nil {
		return cli.NewExitError("read  validator's privKey:"+err.Error(), 127)
	}
	if ctx.IsSet(Verbose()) && ctx.Bool(Verbose()) {
		fmt.Printf("fetch privkey of addr(%x)\n", key.PubKey().Address())
	}
	fmt.Printf("\nNow fetch CA-Node;")
	au, err := NewAdminOPUser(ctx)
	if err != nil {
		return cli.NewExitError("NewAdminOPUser :"+err.Error(), 127)
	}
	err = au.MakeAddPeerMsg(key)
	if err != nil {
		return cli.NewExitError("MakeAddPeerMsg :"+err.Error(), 127)
	}
	hash, err := adminContractCall(au)
	if err != nil {
		return cli.NewExitError("adminContractCall :"+err.Error(), 127)
	}
	fmt.Println("hash=", hash)
	return nil
}

func RemoveNode(ctx *cli.Context) error {
	if !ctx.IsSet(PubKey()) {
		return cli.NewExitError("missing 2validator's pubkey", 127)
	}
	au, err := NewAdminOPUser(ctx)
	if err != nil {
		return cli.NewExitError("NewAdminOPUser :"+err.Error(), 127)
	}
	validatorPub := gcommon.SanitizeHex(ctx.String(PubKey()))
	err = au.MakeRemoveNodeMsg(validatorPub)
	if err != nil {
		return cli.NewExitError("MakeAddPeerMsg :"+err.Error(), 127)
	}
	hash, err := adminContractCall(au)
	if err != nil {
		return cli.NewExitError("adminContractCall :"+err.Error(), 127)
	}
	fmt.Println("hash=", hash)
	return nil
}

func UpdateNode(ctx *cli.Context) error {
	if !ctx.IsSet(PubKey()) {
		return cli.NewExitError("missing 2validator's pubkey", 127)
	}

	if !ctx.IsSet(Power()) {
		fmt.Println("'power' was not set;set default value '0'!")
	}
	au, err := NewAdminOPUser(ctx)
	if err != nil {
		return cli.NewExitError("NewAdminOPUser :"+err.Error(), 127)
	}
	validatorPub := gcommon.SanitizeHex(ctx.String(PubKey()))
	power := ctx.Int64(Power())
	if power < 0 {
		power = 0
	}
	err = au.MakeUpdateNodeMsg(validatorPub, power)
	if err != nil {
		return cli.NewExitError("MakeAddPeerMsg :"+err.Error(), 127)
	}
	hash, err := adminContractCall(au)
	if err != nil {
		return cli.NewExitError("MakeAddPeerMsg :"+err.Error(), 127)
	}
	fmt.Println("hash=", hash)
	return nil
}

func adminContractCall(au *AdminOPUser) (string, error) {
	var abiJson abi.ABI
	abiJson, err := abi.JSON(strings.NewReader(core.AdminABI))
	if err != nil {
		return "", err
	}
	args := []interface{}{au.AdminTxData()}
	var calldata []byte
	calldata, err = abiJson.Pack(core.AdminMethod, args...)
	if err != nil {
		return "", err
	}
	waptx := etypes.NewTransaction(au.Nonce, core.AdminTo, big.NewInt(0), gasLimit, big.NewInt(0), calldata)
	if !bytes.Equal(calldata, waptx.Data()) {
		return "", nil
	}
	signer, sig, err := SignTx(au.Priv, waptx)
	if err != nil {
		return "", err
	}
	sigTx, err := waptx.WithSignature(signer, sig)
	if err != nil {
		return "", err
	}
	b, err := rlp.EncodeToBytes(sigTx)
	if err != nil {
		return "", err
	}

	clt := rpcclient.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(gtypes.ResultBroadcastTxCommit)
	_, err = clt.Call("broadcast_tx_commit", []interface{}{b}, rpcResult)
	if err != nil {
		return "", err
	}
	if rpcResult.Code != 0 {
		return "", fmt.Errorf(rpcResult.Log)
	}
	hash := rpcResult.TxHash
	return hash, nil
}

type AdminOPUser struct {
	//account info;
	Priv  []byte
	Addr  []byte
	Nonce uint64
	//node info;
	privs []crypto.PrivKey
	//msg info;
	adminOpData []byte
}

func NewAdminOPUser(ctx *cli.Context) (*AdminOPUser, error) {
	privKeys, err := requireNodePrivKeys(ctx)
	if err != nil {
		return nil, err
	}
	priv, addr, err := createAccount()
	if err != nil {
		return nil, err
	}
	nonce, err := getNonce(addr)
	if err != nil {
		return nil, err
	}
	user := &AdminOPUser{
		Priv:        priv,
		Addr:        addr,
		Nonce:       nonce,
		privs:       privKeys,
		adminOpData: nil,
	}
	return user, nil
}

func getNonce(addrBytes []byte) (uint64, error) {
	query := append([]byte{types.QueryType_Nonce}, addrBytes...)
	clientJSON := rpcclient.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(gtypes.ResultQuery)
	_, err := clientJSON.Call("query", []interface{}{query}, rpcResult)
	if err != nil {
		return 0, err
	}
	nonce := new(uint64)
	rlp.DecodeBytes(rpcResult.Result.Data, nonce)
	return *nonce, nil
}

func (au *AdminOPUser) AdminTxData() []byte {
	return gtypes.TagAdminOPTx(au.adminOpData)
}

func (au *AdminOPUser) MakeAddPeerMsg(privKey crypto.PrivKey) error {
	pubKey := privKey.PubKey()
	vAttr, err := au.makeValidatorAttr(pubKey.KeyString(), 0, gtypes.ValidatorCmdAddPeer)
	if err != nil {
		return err
	}
	scmd, err := au.ContractsAdminCmd(vAttr)
	if err != nil {
		return err
	}
	sig := privKey.Sign(scmd.Msg)
	scmd.SelfSign = crypto.GetNodeSigBytes(sig)
	au.adminOpData, err = json.Marshal(scmd)
	return err
}

func (au *AdminOPUser) MakeUpdateNodeMsg(pub string, power int64) error {
	vAttr, err := au.makeValidatorAttr(pub, power, gtypes.ValidatorCmdUpdateNode)
	if err != nil {
		return err
	}
	scmd, err := au.ContractsAdminCmd(vAttr)
	au.adminOpData, err = json.Marshal(scmd)
	return err
}

func (au *AdminOPUser) MakeRemoveNodeMsg(pub string) error {
	vAttr, err := au.makeValidatorAttr(pub, 0, gtypes.ValidatorCmdRemoveNode)
	if err != nil {
		return err
	}
	scmd, err := au.ContractsAdminCmd(vAttr)
	au.adminOpData, err = json.Marshal(scmd)
	return err
}

func (au *AdminOPUser) makeValidatorAttr(nodepub string, power int64, cmd gtypes.ValidatorCmd) (*gtypes.ValidatorAttr, error) {
	if strings.HasPrefix(nodepub, "0x") || strings.HasPrefix(nodepub, "0X") {
		nodepub = nodepub[2:]
	}
	pub, err := hex.DecodeString(nodepub)
	if err != nil {
		return nil, err
	}
	return &gtypes.ValidatorAttr{
		PubKey: pub,
		Cmd:    cmd,
		Power:  power,
		Nonce:  au.Nonce,
		Addr:   au.Addr,
	}, nil
}

func (au *AdminOPUser) ContractsAdminCmd(vAttr *gtypes.ValidatorAttr) (*gtypes.AdminOPCmd, error) {
	vdata, err := json.Marshal(vAttr)
	if err != nil {
		return nil, err
	}
	scmd := &gtypes.AdminOPCmd{}
	scmd.CmdType = gtypes.AdminOpChangeValidator
	scmd.Time = time.Now()
	scmd.Msg = vdata
	for _, pk := range au.privs {
		pubk := crypto.GetNodePubkeyBytes(pk.PubKey())
		sigbytes := crypto.GetNodeSigBytes(pk.Sign(vdata))
		scmd.SInfos = append(scmd.SInfos, gtypes.SigInfo{pubk, sigbytes})
	}
	return scmd, nil
}
