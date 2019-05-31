package commands

import (
	"fmt"
	"math/big"
	"sync"

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
		privkey, err := crypto.ToECDSA(keyInt.Bytes())
		if err != nil {
			return err
		}
		addr := crypto.PubkeyToAddress(privkey.PublicKey)
		transferToAddress(addr, uint64(i))
		keyInt.Add(keyInt, big.NewInt(1))
		fmt.Println(addr.Hex())

	}
	wg.Wait()

	return nil
}

func transferToAddress(addr common.Address, nonce uint64) {

	clientJSON := cl.NewClientJSONRPC(commons.QueryServer)

	tx := ethtypes.NewTransaction(nonce, addr, big.NewInt(10000), gasLimit, big.NewInt(0), []byte{})

	key := "a8971729fbc199fb3459529cebcd8704791fc699d88ac89284f23ff8e7fca7d6"

	privBytes := common.Hex2Bytes(key)

	signer, sig, err := SignTx(privBytes, tx)
	if err != nil {
		panic(err)
	}
	sigTx, err := tx.WithSignature(signer, sig)
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
}
