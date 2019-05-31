package commands

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"gopkg.in/urfave/cli.v1"

	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
	atypes "github.com/dappledger/AnnChain/gemmill/types"
	"github.com/dappledger/AnnChain/cmd/client/commons"
	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
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

	tx := ethtypes.NewTransaction(nonce, to, big.NewInt(value), gasLimit, big.NewInt(0), data)

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

	rpcResult := new(atypes.ResultBroadcastTxCommit)
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
	query[0] = atypes.QueryTxExecution
	copy(query[1:], hashBytes)

	query = append([]byte{5}, query...)

	rpcResult := new(atypes.ResultQuery)
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	if _, err = clientJSON.Call("query", []interface{}{query}, rpcResult); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	data := rpcResult.Result.Data
	payload := string(data)
	fmt.Println("payload:", payload)

	return nil
}
