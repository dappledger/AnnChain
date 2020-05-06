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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	rtypes "github.com/dappledger/AnnChain/chain/types"
	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/core/types"
	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/cmd/client/commons"
	"github.com/dappledger/AnnChain/eth/rlp"
	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
)

var (
	KvPutCommands = cli.Command{
		Name:     "put",
		Usage:    "operations for put key value",
		Category: "put",
		Action:   putKeyValue,
		Flags: []cli.Flag{
			anntoolFlags.privateKey,
		},
	}
	KvGetCommands = cli.Command{
		Name:     "get",
		Usage:    "operations for get key value, if",
		Category: "get",
		Action:   getKeyValue,
		Flags: []cli.Flag{
			anntoolFlags.pageNum,
			anntoolFlags.pageSize,
		},
	}
)

func getKeyValue(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(gtypes.ResultQuery)
	keyStr := ctx.Args().First()
	pageNum := ctx.Uint("page_num")
	if pageNum != 0 {
		return queryKeyUpdateHistory(ctx)
	}

	query := append([]byte{rtypes.QueryType_Key}, []byte(keyStr)...)

	_, err := clientJSON.Call("query", []interface{}{query}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	fmt.Println("query result:", string(rpcResult.Result.Data))
	return nil
}

func putKeyValue(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(gtypes.ResultBroadcastTxCommit)
	if ctx.NArg() < 2 {
		return cli.NewExitError(fmt.Errorf("need key and value"), 127)
	}
	keyStr := ctx.Args().First()
	valueStr := ctx.Args().Get(1)
	privkey := ctx.String("priv_key")
	if privkey == "" {
		return cli.NewExitError("privkey is required", 127)
	}

	if strings.Index(privkey, "0x") == 0 || strings.Index(privkey, "0X") == 0 {
		privkey = privkey[2:]
	}

	privBytes := common.Hex2Bytes(privkey)

	addr, err := getAddrBytes(privBytes)
	if err != nil {
		return err
	}

	nonce, _ := getNonce(addr)
	kvBytes, err := rlp.EncodeToBytes(&rtypes.KV{Key: []byte(keyStr), Value: []byte(valueStr)})
	if err != nil {
		return err
	}

	txdata := append(rtypes.KVTxType, kvBytes...)
	tx := types.NewTransaction(nonce, common.Address{}, big.NewInt(0), gasLimit, big.NewInt(0), txdata)
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

	_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{b}, rpcResult)
	if err != nil {
		return err
	}

	hash := rpcResult.TxHash
	fmt.Println("tx result:", hash)

	return nil
}

func queryKeyUpdateHistory(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(gtypes.ResultQuery)
	keyStr := ctx.Args().First()
	//ValueHistoryResult
	pageNum := ctx.Uint("page_num")
	if pageNum == 0 {
		pageNum = 1
	}
	pageSize := ctx.Uint("page_size")
	if pageSize == 0 {
		pageSize = 10
	}
	query := append([]byte{rtypes.QueryType_Key_Update_History}, putUint32(uint32(pageNum))...)
	query = append(query, putUint32(uint32(pageSize))...)
	query = append(query, []byte(keyStr)...)

	_, err := clientJSON.Call("query", []interface{}{query}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	response := &gtypes.ValueHistoryResult{}
	err = rlp.DecodeBytes(rpcResult.Result.Data, response)
	if err != nil {
		fmt.Println(rpcResult.Result)
		return cli.NewExitError(err.Error(), 127)
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println("query result:", string(responseJSON))

	return nil
}

func putUint32(i uint32) []byte {
	index := make([]byte, 4)
	binary.BigEndian.PutUint32(index, i)
	return index
}
