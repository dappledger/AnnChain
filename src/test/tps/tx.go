package main

import (
	"fmt"
	"math/big"

	"time"

	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/angine/types"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
)

var (
	// this signer appears to be a must in evm 1.5.9
	ethSigner = ethtypes.HomesteadSigner{}
)

func sendTx(privkey, toAddr string, value int64) error {

	nonce := uint64(time.Now().UnixNano())
	to := common.HexToAddress(toAddr)

	tx := ethtypes.NewTransaction(nonce, to, big.NewInt(value), big.NewInt(90000), big.NewInt(0), []byte{})

	key, err := crypto.HexToECDSA(privkey)
	panicErr(err)
	sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
	panicErr(err)
	sigTx, err := tx.WithSignature(ethSigner, sig)
	panicErr(err)
	b, err := rlp.EncodeToBytes(sigTx)
	panicErr(err)

	tmResult := new(types.RPCResult)
	clientJSON := cl.NewClientJSONRPC(logger, rpcTarget)
	_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{defaultChainID, b}, tmResult)
	panicErr(err)

	res := (*tmResult).(*types.ResultBroadcastTxCommit)
	fmt.Println("******************result", res)

	return nil
}
