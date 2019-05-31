package commands

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/cmd/client/commons"
	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	ac "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
	agtypes "github.com/dappledger/AnnChain/gemmill/types"
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
				Name:   "balance",
				Usage:  "query account's balance",
				Action: queryBalance,
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
			{
				Name:   "limittx",
				Usage:  "query rest of limited tx number if the chain enable the tx limited function ",
				Action: queryLimittx,
			},
		},
	}
)

func queryNonce(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(agtypes.ResultQuery)

	addrHex := ac.SanitizeHex(ctx.String("address"))
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

func queryBalance(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(agtypes.ResultQuery)

	addrHex := ac.SanitizeHex(ctx.String("address"))
	addr := common.Hex2Bytes(addrHex)
	query := append([]byte{2}, addr...)

	_, err := clientJSON.Call("query", []interface{}{query}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	balance := big.NewInt(0)
	rlp.DecodeBytes(rpcResult.Result.Data, balance)

	fmt.Println("query result:", balance.String())

	return nil
}

func queryReceipt(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(agtypes.ResultQuery)
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

	receiptForStorage := new(ethtypes.AnnReceiptForStorage)

	err = rlp.DecodeBytes(rpcResult.Result.Data, receiptForStorage)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	receiptJSON, _ := json.Marshal(receiptForStorage)
	fmt.Println("query result:", string(receiptJSON))

	return nil
}

func queryLimittx(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(agtypes.ResultNumLimitTx)

	_, err := clientJSON.Call("querytx", []interface{}{[]byte{9}}, rpcResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println("query result:", rpcResult.Num)
	return nil
}
