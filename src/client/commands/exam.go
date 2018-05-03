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
	"fmt"
	"math/big"
	"sync"

	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/angine/types"
	cl "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/src/client/commons"
	"gopkg.in/urfave/cli.v1"
)

var (
	ExamCommand = cli.Command{
		Name:   "exam",
		Action: exam,
		Flags: []cli.Flag{
			cli.Int64Flag{
				Name:  "start",
				Value: 0,
			},
		},
	}
	InitCommand = cli.Command{
		Name:   "initial",
		Action: initial,
		Flags: []cli.Flag{
			cli.Int64Flag{
				Name:  "start",
				Value: 0,
			},
		},
	}

	nonceFor911 uint64 = 0
)

func initial2(ctx *cli.Context) error {
	start := ctx.Int64("start")

	for i := start; i < 1000; i++ {
		privkey := newPrivkey(i)
		fmt.Println(crypto.PubkeyToAddress(privkey.PublicKey).Hex())
	}

	return nil
}

func initial(ctx *cli.Context) error {

	nonceFor911 = getNonce(common.HexToAddress("0x7752b42608a0f1943c19fc5802cb027e60b4c911"))
	start := ctx.Int64("start")
	var i int64
	for i = start; i < 1000+start; i++ {
		privkey := newPrivkey(i)
		addr := crypto.PubkeyToAddress(privkey.PublicKey)
		transferToAddress(addr, nonceFor911)
		nonceFor911 += 1
		fmt.Println(addr.Hex())
	}

	return nil
}

func exam(ctx *cli.Context) error {

	hex := "a8971729fbc199fb3459529cebcd8704791fc699d88ac89284f23ff8e7f00000"
	keyInt := big.NewInt(0)
	keyInt.SetBytes(common.Hex2Bytes(hex))
	var i int64
	var wg sync.WaitGroup
	start := ctx.Int64("start")
	for i = start; i < start+1000; i++ {
		privkey := crypto.ToECDSA(keyInt.Bytes())
		addr := crypto.PubkeyToAddress(privkey.PublicKey)
		transferToAddress(addr, uint64(i))
		keyInt.Add(keyInt, big.NewInt(1))
		fmt.Println(addr.Hex())

	}
	wg.Wait()

	return nil
}

func transferTo911(serial int64) {

	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	addr := common.HexToAddress("0x7752b42608a0f1943c19fc5802cb027e60b4c911")
	tx := ethtypes.NewTransaction(0, addr, big.NewInt(1), big.NewInt(90000), big.NewInt(0), []byte{})

	hex := "a8971729fbc199fb3459529cebcd8704791fc699d88ac89284f23ff8e7f00000"

	keyInt := big.NewInt(0)
	keyInt.SetBytes(common.Hex2Bytes(hex))

	keyInt.Add(keyInt, big.NewInt(serial))
	privkey := crypto.ToECDSA(keyInt.Bytes())

	fmt.Println("from address:", crypto.PubkeyToAddress(privkey.PublicKey))

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
	res := (*tmResult).(*types.ResultBroadcastTx)
	fmt.Println(common.Bytes2Hex(res.Data))
}

func transferToAddress(addr common.Address, nonce uint64) {

	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)

	tx := ethtypes.NewTransaction(nonce, addr, big.NewInt(10000), big.NewInt(90000), big.NewInt(0), []byte{})

	key, err := crypto.HexToECDSA("a8971729fbc199fb3459529cebcd8704791fc699d88ac89284f23ff8e7fca7d6")
	if err != nil {
		panic(err)
	}
	sig, err := crypto.Sign(tx.SigHash(ethSigner).Bytes(), key)
	sigTx, err := tx.WithSignature(ethSigner, sig)

	b, err := rlp.EncodeToBytes(sigTx)
	if err != nil {
		panic(err)
	}

	tmResult := new(types.RPCResult)
	_, err = clientJSON.Call("broadcast_tx_sync", []interface{}{b}, tmResult)
	if err != nil {
		panic(err)
	}
}
