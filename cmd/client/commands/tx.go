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
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/cmd/client/commons"
	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
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
					anntoolFlags.payload,
					anntoolFlags.nonce,
					anntoolFlags.to,
					anntoolFlags.value,
				},
			},
			{
				Name:   "data",
				Usage:  "query transaction execution info",
				Action: txData,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name: "txHash",
					},
				},
			},
		},
	}
)

func sendTx(ctx *cli.Context) error {
	nonce := ctx.Uint64("nonce")
	to := common.HexToAddress(ctx.String("to"))
	value := ctx.Int64("value")
	payload := ctx.String("payload")

	data := []byte(payload)

	tx := types.NewTransaction(nonce, to, big.NewInt(value), gasLimit, big.NewInt(0), data)

	key, err := requireAccPrivky(ctx)
	if err != nil {
		return err
	}

	privBytes := common.Hex2Bytes(key)

	signer, sig, err := SignTx(privBytes, tx)
	if err != nil {
		return err
	}
	sigTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		return err
	}

	b, err := rlp.EncodeToBytes(sigTx)
	if err != nil {
		return err
	}

	rpcResult := new(gtypes.ResultBroadcastTxCommit)
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{b}, rpcResult)
	if err != nil {
		return err
	}

	hash := rpcResult.TxHash
	fmt.Println("tx result:", hash)

	return nil
}

func txData(ctx *cli.Context) error {
	if !ctx.IsSet("txHash") {
		return cli.NewExitError("txHash is required", 127)
	}
	txhash := ctx.String("txHash")
	if strings.Index(txhash, "0x") == 0 {
		txhash = txhash[2:]
	}

	hashBytes, err := hex.DecodeString(txhash)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	query := make([]byte, len(hashBytes)+1)
	query[0] = gtypes.QueryTxExecution
	copy(query[1:], hashBytes)

	query = append([]byte{5}, query...)

	rpcResult := new(gtypes.ResultQuery)
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	if _, err = clientJSON.Call("query", []interface{}{query}, rpcResult); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	data := rpcResult.Result.Data
	payload := string(data)
	fmt.Println("payload:", payload)

	return nil
}
