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
	"crypto/ecdsa"
	_ "encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/golang/protobuf/proto"
	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/angine/types"
	ac "github.com/dappledger/AnnChain/module/lib/go-common"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/src/chain/app/ikhofi"
	"github.com/dappledger/AnnChain/src/client/commons"
)

var (
	ikhofiSigner = ikhofi.DawnSigner{}

	//ContractCommands defines a more git-like subcommand system
	IkhofiCommands = cli.Command{
		Name:     "ikhofi",
		Usage:    "operations for ikhofi",
		Category: "Contract",
		Subcommands: []cli.Command{
			{
				Name:   "execute",
				Usage:  "execute a contract",
				Action: execute,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "method",
						Value: "",
					},
					cli.StringFlag{
						Name:  "contractid",
						Value: "",
					},
					anntoolFlags.privkey,
					anntoolFlags.bytecode,
				},
			}, {
				Name:   "query",
				Usage:  "query a contract",
				Action: query,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "method",
						Value: "",
					},
					cli.StringFlag{
						Name:  "contractid",
						Value: "",
					},
					anntoolFlags.privkey,
				},
			},
		},
	}
)

type ContractParam struct {
	ChainID    string
	ContractID string
	MethodName string
	Args       []string
	Privkey    *ecdsa.PrivateKey
	ByteCode   []byte
}

type QueryResult struct {
	Value string `json:"value"`
	Msg   string `json:"msg"`
}

func substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

func getMethod(str string) (method string, args []string) {
	// get method from method strings
	index := strings.Index(str, "(")
	method = substr(str, 0, index)

	// get argument list from method strings
	argStr := substr(str, index+1, len(str)-index-2)
	args = []string{}
	if argStr != "" {
		args = strings.Split(argStr, ",")
	}
	for i := 0; i < len(args); i++ {
		arg := strings.Trim(args[i], "' ")
		args[i] = string(arg)
	}
	return method, args
}

func execute(ctx *cli.Context) error {

	param, err := ParseParam(ctx.String("method"), ctx.String("contractid"), ctx.GlobalString("target"), ctx.String("privkey"))
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	if param.MethodName == ikhofi.SystemDeployMethod || param.MethodName == ikhofi.SystemUpgradeMethod {
		if err = checkFileContent(param.Args); err != nil {
			return err
		}
		if ctx.IsSet("bytecode") {
			param.ByteCode, err = fileData(ctx.String("bytecode"))
			if err != nil {
				return cli.NewExitError(err.Error(), 127)
			}
		} else {
			bytez, err := getFileContent(param.Args)
			if err != nil {
				return cli.NewExitError(err.Error(), 127)
			}
			param.ByteCode = bytez
		}
		fmt.Println("contract address is:", param.Args[0])
		fmt.Println("contract file path is:", param.Args[1])
	}
	err = contractExecute(&param)
	if err != nil {
		return cli.NewExitError(err.Error(), 110)
	}

	return nil
}

func query(ctx *cli.Context) error {

	param, err := ParseParam(ctx.String("method"), ctx.String("contractid"), ctx.GlobalString("target"), ctx.String("privkey"))
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	result, err := ContractQuery(&param)
	if err != nil {
		return cli.NewExitError(err.Error(), 110)
	}
	if result.Value != "" {
		fmt.Println("query result value:", result.Value)
		fmt.Println("query result message:", result.Msg)
	}

	return nil
}

func checkFileContent(args []string) error {
	if len(args) != 2 {
		return errors.New("Invalid method deploy or upgrade`s argument")
	}

	if args[0] == ikhofi.SystemContractId {
		return errors.New("Unable to deploy or upgrade 'system' contract")
	} //todo check exist contract_id
	return nil
}

// checkFileContent() should be called first
func getFileContent(args []string) (bytez []byte, err error) {
	bytez, err = ioutil.ReadFile(args[1])
	return
}

func ContractExecute(param *ContractParam) (txhash []byte, err error) {

	addr := crypto.PubkeyToAddress(param.Privkey.PublicKey)
	// make NewTransaction
	tx := ikhofi.NewTransaction(addr, param.ContractID, param.MethodName, param.Args, param.ByteCode)

	// sign tx using privkey
	sig, errS := crypto.Sign(tx.SigHash(ikhofiSigner).Bytes(), param.Privkey)
	if errS != nil {
		err = errS
		return
	}
	signedTx, errT := tx.WithSignature(ikhofiSigner, sig)
	if errT != nil {
		err = errT
		return
	}
	// encode tx to bytes
	txpb := signedTx.Transaction2Pb()
	b, errM := proto.Marshal(txpb)
	if errM != nil {
		err = errM
		return
	}

	// call rpc service
	tmResult := new(types.RPCResult)
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	_, err = clientJSON.Call("broadcast_tx_sync", []interface{}{param.ChainID, b}, tmResult)
	return signedTx.Hash[:], err
}

func contractExecute(param *ContractParam) (err error) {
	var txHash []byte
	txHash, err = ContractExecute(param)
	fmt.Printf("txHash:%x\n", txHash)
	return
}

func ContractQuery(param *ContractParam) (result QueryResult, err error) {

	addr := crypto.PubkeyToAddress(param.Privkey.PublicKey)
	// new a transaction struct
	tx := ikhofi.NewTransaction(addr, param.ContractID, param.MethodName, param.Args, nil)

	// sign tx using privkey
	sig, _ := crypto.Sign(tx.SigHash(ikhofiSigner).Bytes(), param.Privkey)
	signedTx, _ := tx.WithSignature(ikhofiSigner, sig)

	// encode tx to bytes
	txpb := signedTx.Transaction2Pb()
	query, errM := proto.Marshal(txpb)
	if errM != nil {
		err = errM
		return
	}

	// call rpc service
	tmResult := new(types.RPCResult)
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	_, err = clientJSON.Call("query", []interface{}{param.ChainID, query}, tmResult)
	if err != nil {
		return
	}

	res := (*tmResult).(*types.ResultQuery)
	queryRes := &ikhofi.Result{}
	err = proto.Unmarshal(res.Result.Data, queryRes)
	if err != nil {
		return
	}
	fmt.Println("query result code:", queryRes.Code)
	result.Value = queryRes.Value
	result.Msg = queryRes.Msg

	return
}

func ParseParam(method, id, chainID, privkey string) (param ContractParam, err error) {

	if method == "" {
		err = errors.New("Required parameter method")
		return
	}
	if id == "" {
		err = errors.New("Required parameter id")
		return
	}
	if chainID == "" {
		err = errors.New("Required parameter chain_id")
		return
	}
	if privkey == "" {
		err = errors.New("Required parameter privkey")
		return
	}
	privkey = ac.SanitizeHex(privkey)
	ecdsaKey, errC := crypto.HexToECDSA(privkey)
	if errC != nil {
		err = errC
		return
	}

	methodName, args := getMethod(method)
	if id == ikhofi.SystemContractId {
		if !(methodName == ikhofi.SystemDeployMethod ||
			methodName == ikhofi.SystemUpgradeMethod ||
			methodName == ikhofi.SystemQueryContractIdExits ||
			methodName == ikhofi.SystemQueryEventFilterById ||
			methodName == ikhofi.SystemQueryEventFilterByIdAndType) {
			err = errors.New(fmt.Sprintf("Invalid system contract method: %s", methodName))
			return
		}
	}
	param = ContractParam{
		ChainID:    chainID,
		ContractID: id,
		MethodName: methodName,
		Args:       args,
		Privkey:    ecdsaKey,
	}
	return
}
