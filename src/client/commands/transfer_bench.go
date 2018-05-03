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
	"fmt"
	"math/big"

	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/angine/types"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/src/client/commons"
)

var (
	TransferBenchCommand = cli.Command{
		Name:   "bench",
		Action: transferBench,
		Flags: []cli.Flag{
			cli.Int64Flag{
				Name:  "start",
				Value: 0,
			},
			cli.Int64Flag{
				Name:  "times",
				Value: 1,
			},
		},
	}

	AnnCoinBenchCommand = cli.Command{
		Name:   "benchcoin",
		Action: benchAnnCoin,
		Flags:  []cli.Flag{},
	}

	nonceMap = make(map[common.Address]uint64)
)

func newPrivkey(n int64) *ecdsa.PrivateKey {
	hex := "a8971729fbc199fb3459529cebcd8704791fc699d88ac89284f23ff8e7f00000"
	keyInt := big.NewInt(0)
	keyInt.SetBytes(common.Hex2Bytes(hex))
	keyInt.Add(keyInt, big.NewInt(n))
	privekey := crypto.ToECDSA(keyInt.Bytes())
	return privekey
}

func transferBench(ctx *cli.Context) error {

	start := ctx.Int64("start")
	times := ctx.Int64("times")
	//hex := "a8971729fbc199fb3459529cebcd8704791fc699d88ac89284f23ff8e7f00000"
	//keyInt := big.NewInt(0)
	//keyInt.SetBytes(common.Hex2Bytes(hex))
	var i int64 = 0
	for ; i < 1000; i++ {
		transComplement(start, i, 1000, times)
		//
		//go func(j int64) {
		//	transComplement(j, 1000)
		//}(i % 1000)
		//if i % 100 == 99 {
		//	time.Sleep(50 * time.Millisecond)
		//}
	}
	return nil
}

func transComplement(start int64, serial int64, n int64, times int64) {
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	privkey := newPrivkey(start + serial)
	toPrivkey := newPrivkey(start + (1000 - serial))
	meAddress := crypto.PubkeyToAddress(privkey.PublicKey)
	fmt.Println("address:", meAddress.Hex())
	to := crypto.PubkeyToAddress(toPrivkey.PublicKey)

	var i int64
	for i = 0; i < times; i++ {
		nonce := nonceMap[meAddress]
		if nonce == 0 {
			nonce = getNonce(meAddress)
			nonceMap[meAddress] = nonce
		}

		tx := ethtypes.NewTransaction(nonce, to, big.NewInt(1), big.NewInt(90000), big.NewInt(0), []byte{})
		sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), privkey)
		if err != nil {
			panic(err)
		}
		sigTx, err := tx.WithSignature(ethSigner, sig)
		if err != nil {
			panic(err)
		}
		b, err := rlp.EncodeToBytes(sigTx)
		if err != nil {
			panic(err)
		}
		tmResult := new(types.RPCResult)
		_, err = clientJSON.Call("broadcast_tx_async", []interface{}{b}, tmResult)
		if err != nil {
			panic(err)
		}
		nonceMap[meAddress] = nonceMap[meAddress] + 1
		fmt.Println("tx result:", sigTx.Hash().Hex(), meAddress.Hex())
	}

}

func benchAnnCoin(ctx *cli.Context) error {
	pks := "d6e2a2a9b0f8be93ee0773087fa68bcdfa84621c9c4fc2740d1d640a54d754df"
	privekey := crypto.ToECDSA(common.Hex2Bytes(pks))
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	fromAddr := crypto.PubkeyToAddress(privekey.PublicKey) //common.StringToAddress("680008fb232b293cbfee5f1c9c82dde51b03495f")
	toAddr := common.StringToAddress("9cef2ef1197ff8bd475307aac3e27261df88059d")
	nonce := getNonce(fromAddr)
	for i := 0; i < 500; i++ {
		transferAnnCoin(fromAddr, toAddr, privekey, clientJSON, nonce)
		nonce++
	}

	return nil
}

func transferAnnCoin(fromAddr common.Address, toAddr common.Address, privkey *ecdsa.PrivateKey, clientJSON *cl.ClientJSONRPC, nonce uint64) {
	tx := ethtypes.NewTransaction(nonce, toAddr, big.NewInt(1), big.NewInt(90000), big.NewInt(0), []byte{})
	sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), privkey)
	if err != nil {
		panic(err)
	}
	sigTx, err := tx.WithSignature(ethSigner, sig)
	if err != nil {
		panic(err)
	}
	b, err := rlp.EncodeToBytes(sigTx)
	if err != nil {
		panic(err)
	}
	tmResult := new(types.RPCResult)
	_, err = clientJSON.Call("broadcast_tx_sync", []interface{}{b}, tmResult)
	if err != nil {
		panic(err)
	}
	fmt.Println("tx result:")
}

func getNonce(address common.Address) uint64 {
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	query := append([]byte{1}, address.Bytes()...)

	_, err := clientJSON.Call("abci_query", []interface{}{query}, tmResult)
	if err != nil {
		panic(err)
	}

	res := (*tmResult).(*types.ResultQuery)
	nonce := new(uint64)
	rlp.DecodeBytes(res.Result.Data, nonce)

	fmt.Println("query result:", *nonce)

	return *nonce
}
