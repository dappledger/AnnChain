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
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/cmd/client/commons"
	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
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
				Name:   "receipt",
				Usage:  "",
				Action: queryReceipt,
				Flags: []cli.Flag{
					anntoolFlags.hash,
				},
			},
		},
	}
)

func queryNonce(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(gtypes.ResultQuery)

	addrHex := gcmn.SanitizeHex(ctx.String("address"))
	addr := common.Hex2Bytes(addrHex)
	query := append([]byte{1}, addr...)

	_, err := clientJSON.Call("query", []interface{}{query}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	nonce := new(uint64)
	rlp.DecodeBytes(rpcResult.Result.Data, nonce)

	fmt.Println("query result:", *nonce)

	return nil
}

func queryReceipt(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(gtypes.ResultQuery)
	hash := ctx.String("hash")
	if strings.Index(hash, "0x") == 0 {
		hash = hash[2:]
	}

	hashBytes := common.Hex2Bytes(hash)
	query := append([]byte{3}, hashBytes...)
	_, err := clientJSON.Call("query", []interface{}{query}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	receiptForStorage := new(types.ReceiptForStorage)

	err = rlp.DecodeBytes(rpcResult.Result.Data, receiptForStorage)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	receiptJSON, _ := json.Marshal(receiptForStorage)
	fmt.Println("query result:", string(receiptJSON))

	return nil
}
