package commands

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"gopkg.in/urfave/cli.v1"

	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
	agtypes "github.com/dappledger/AnnChain/gemmill/types"
	"github.com/dappledger/AnnChain/cmd/client/commons"
	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
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

	nonceMap  = make(map[common.Address]uint64)
	ethSigner = ethtypes.HomesteadSigner{}
)

func newPrivkey(n int64) *ecdsa.PrivateKey {
	hex := "a8971729fbc199fb3459529cebcd8704791fc699d88ac89284f23ff8e7f00000"
	keyInt := big.NewInt(0)
	keyInt.SetBytes(common.Hex2Bytes(hex))
	keyInt.Add(keyInt, big.NewInt(n))
	privekey, err := crypto.ToECDSA(keyInt.Bytes())
	if err != nil {
		panic(err)
	}
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
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
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

		tx := ethtypes.NewTransaction(nonce, to, big.NewInt(1), gasLimit, big.NewInt(0), []byte{})
		sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), privkey)
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
		rpcResult := new(agtypes.ResultBroadcastTx)
		_, err = clientJSON.Call("broadcast_tx_async", []interface{}{b}, rpcResult)
		if err != nil {
			panic(err)
		}
		nonceMap[meAddress] = nonceMap[meAddress] + 1
		fmt.Println("tx result:", sigTx.Hash().Hex(), meAddress.Hex())
	}

}

func benchAnnCoin(ctx *cli.Context) error {
	pks := "d6e2a2a9b0f8be93ee0773087fa68bcdfa84621c9c4fc2740d1d640a54d754df"
	privekey, err := crypto.ToECDSA(common.Hex2Bytes(pks))
	if err != nil {
		panic(err)
	}
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	fromAddr := crypto.PubkeyToAddress(privekey.PublicKey) //common.StringToAddress("680008fb232b293cbfee5f1c9c82dde51b03495f")
	toAddr := common.HexToAddress("9cef2ef1197ff8bd475307aac3e27261df88059d")
	nonce := getNonce(fromAddr)
	for i := 0; i < 500; i++ {
		transferAnnCoin(fromAddr, toAddr, privekey, clientJSON, nonce)
		nonce++
	}

	return nil
}

func transferAnnCoin(fromAddr common.Address, toAddr common.Address, privkey *ecdsa.PrivateKey, clientJSON *cl.ClientJSONRPC, nonce uint64) {
	tx := ethtypes.NewTransaction(nonce, toAddr, big.NewInt(1), gasLimit, big.NewInt(0), []byte{})
	sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), privkey)
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
	rpcResult := new(agtypes.ResultBroadcastTx)
	_, err = clientJSON.Call("broadcast_tx_async", []interface{}{b}, rpcResult)
	if err != nil {
		panic(err)
	}
	fmt.Println("tx result:")
}

func getNonce(address common.Address) uint64 {
	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)
	rpcResult := new(agtypes.ResultQuery)

	query := append([]byte{1}, address.Bytes()...)

	_, err := clientJSON.Call("abci_query", []interface{}{query}, rpcResult)
	if err != nil {
		panic(err)
	}

	nonce := new(uint64)
	rlp.DecodeBytes(rpcResult.Result.Data, nonce)

	fmt.Println("query result:", *nonce)

	return *nonce
}
