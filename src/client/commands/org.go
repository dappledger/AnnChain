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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	client "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	civil "github.com/dappledger/AnnChain/src/chain/node"
	"github.com/dappledger/AnnChain/src/client/commons"
	cvtools "github.com/dappledger/AnnChain/src/tools"
)

type orgActions struct{}

var (
	oa = orgActions{}

	OrgCommands = cli.Command{
		Name:     "organization",
		Usage:    "commands for organization control",
		Category: "Organization",
		Subcommands: []cli.Command{
			{
				Name:   "create",
				Usage:  "create a organization",
				Action: oa.Create,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					cli.StringFlag{
						Name:  "genesisfile",
						Usage: "point the genesis file for the sharding",
						Value: "",
					},
					cli.StringFlag{
						Name:  "configfile",
						Usage: "custom configs for the sharding",
						Value: "",
					},
				},
			},
			{
				Name:   "join",
				Usage:  "join a organization",
				Action: oa.Join,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					cli.StringFlag{
						Name:  "orgid",
						Usage: "specify the organization's chain id",
						Value: "",
					},
					cli.StringFlag{
						Name:  "genesisfile",
						Usage: "point the genesis file for the sharding",
						Value: "",
					},
					cli.StringFlag{
						Name:  "configfile",
						Usage: "custom configs for the sharding",
						Value: "",
					},
				},
			},
			{
				Name:   "leave",
				Usage:  "",
				Action: oa.Leave,
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					cli.StringFlag{
						Name:  "orgid",
						Usage: "specify the organization's chain id",
						Value: "",
					},
				},
			},
		},
	}
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func fileData(str string) ([]byte, error) {
	path, _ := PathExists(str)
	if !path {
		fstr := strings.Replace(str, "\\\r\n", "\r\n", -1)
		fstr = strings.Replace(fstr, "\\\"", "\"", -1)
		return []byte(fstr), nil
	}
	return ioutil.ReadFile(str)
}

func (o *orgActions) Create(ctx *cli.Context) error {
	tx := civil.OrgTx{
		Act:  civil.OrgCreate,
		Time: time.Now(),
	}

	if err := o.joinOrCreate(ctx, &tx); err != nil {
		return err
	}

	return nil
}

func (o *orgActions) Join(ctx *cli.Context) error {
	tx := civil.OrgTx{
		Act:  civil.OrgJoin,
		Time: time.Now(),
	}

	if err := o.joinOrCreate(ctx, &tx); err != nil {
		return err
	}

	return nil
}

func (o *orgActions) joinOrCreate(ctx *cli.Context, tx *civil.OrgTx) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chain is required", 127)
	}
	chainID := ctx.GlobalString("target")

	privKeyPtr, err := requirePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	pubkey := privKeyPtr.PubKey().(*crypto.PubKeyEd25519)
	tx.PubKey = pubkey[:]

	var gdata, cdata []byte
	if tx.Act == civil.OrgCreate && !ctx.IsSet("genesisfile") {
		return cli.NewExitError("genesis is required", 127)
	}
	if ctx.IsSet("genesisfile") {
		gpath := ctx.String("genesisfile")
		if gdata, err = fileData(gpath); err != nil {
			return cli.NewExitError(err.Error(), 127)
		}
		genesis := types.GenesisDocFromJSON(gdata)
		tx.ChainID, tx.Genesis = genesis.ChainID, *genesis
	} else {
		if !ctx.IsSet("orgid") {
			return cli.NewExitError("please tell me which chain you wanna join", 127)
		}
		tx.ChainID = ctx.String("orgid")
	}

	if !ctx.IsSet("configfile") {
		return cli.NewExitError("config is required", 127)
	}
	cpath, conf := ctx.String("configfile"), make(map[string]interface{})
	if cdata, err = fileData(cpath); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	if err = toml.Unmarshal(cdata, &conf); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	if _, ok := conf["p2p_laddr"]; !ok {
		return cli.NewExitError("p2p_laddr is missing, give a available port for p2p gossip", 127)
	}
	if _, ok := conf["appname"]; !ok {
		return cli.NewExitError("app name should be in the conf", 127)
	} else if tx.App, ok = conf["appname"].(string); !ok {
		return cli.NewExitError("app name should be a string", 127)
	}
	civil.CheckConfNeedInApp(tx.App, conf)
	tx.Config = conf

	tx.Signature, _ = cvtools.TxSign(tx, privKeyPtr)
	txBytes, err := cvtools.TxToBytes(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	txBytes = types.WrapTx(civil.OrgTag, txBytes)

	clnt := client.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	_, err = clnt.Call("broadcast_tx_"+commons.CallMode, []interface{}{chainID, txBytes}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	hash, _ := cvtools.TxHash(tx)

	fmt.Println("send ok: ", hex.EncodeToString(hash))

	return nil
}

func (o *orgActions) Leave(ctx *cli.Context) error {
	var err error
	tx := &civil.OrgTx{
		Act:  civil.OrgLeave,
		Time: time.Now(),
	}

	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chain is required", 127)
	}
	if !ctx.IsSet("orgid") {
		return cli.NewExitError("orgid is required", 127)
	}
	chainID, orgID := ctx.GlobalString("target"), ctx.String("orgid")

	privKeyPtr, err := requirePrivKey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	pubkey := privKeyPtr.PubKey().(*crypto.PubKeyEd25519)
	tx.PubKey, tx.ChainID = pubkey[:], orgID

	tx.Signature, _ = cvtools.TxSign(tx, privKeyPtr)

	txBytes, err := json.Marshal(tx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	txBytes = types.WrapTx(civil.OrgTag, txBytes)

	clnt := client.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	_, err = clnt.Call("broadcast_tx_"+commons.CallMode, []interface{}{chainID, txBytes}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println("send ok")

	return nil
}

func requirePubKey(ctx *cli.Context) (*crypto.PubKeyEd25519, error) {
	if !ctx.IsSet("pubkey") {
		return nil, fmt.Errorf("pubkey is required")
	}
	pub := ctx.String("pubkey")
	pubBytes, err := hex.DecodeString(pub)
	if err != nil {
		return nil, err
	}
	pubkey := crypto.PubKeyEd25519{}
	copy(pubkey[:], pubBytes)
	return &pubkey, nil
}

func requirePrivKey(ctx *cli.Context) (*crypto.PrivKeyEd25519, error) {
	if !ctx.IsSet("privkey") {
		return nil, fmt.Errorf("privkey is required")
	}
	priv := ctx.String("privkey")
	privBytes, err := hex.DecodeString(priv)
	if err != nil {
		return nil, err
	}
	privKey := crypto.PrivKeyEd25519{}
	copy(privKey[:], privBytes)
	return &privKey, nil
}
