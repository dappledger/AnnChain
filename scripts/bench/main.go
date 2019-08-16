package main

import (
	"crypto/ecdsa"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/dappledger/AnnChain/chain/types"
	"github.com/dappledger/AnnChain/cmd/client/commands"
	"github.com/dappledger/AnnChain/eth/accounts/abi"
	"github.com/dappledger/AnnChain/eth/common"
	etypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	gcrypto "github.com/dappledger/AnnChain/gemmill/go-crypto"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/rpc/client"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
)

var version = "0.9.0"

var (
	ethSigner = etypes.HomesteadSigner{}
	gasLimit  = uint64(90000000000)
	txSize    = 250

	requestNum int
	procNum    int
	method     string
	argument   int64

	bytecode = "6060604052341561000f57600080fd5b60d38061001d6000396000f3006060604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806360fe47b114604e5780636d4ce63c14606e575b600080fd5b3415605857600080fd5b606c60048080359060200190919050506094565b005b3415607857600080fd5b607e609e565b6040518082815260200191505060405180910390f35b8060008190555050565b600080549050905600a165627a7a723058202f196d73b019bfadd056125155d3a0728f0f73dad0eb44ac6dfc4eb3c6a0a3ec0029"
)

func main() {
	// server := flag.String("server", "localhost:46657", "rpc server address")
	actionType := flag.String("action", "", "action type: basic, create, read, execute")
	flag.IntVar(&procNum, "proc_num", 1, "total number of processes")
	flag.IntVar(&requestNum, "req_num", 1, "total number of requests")
	flag.StringVar(&method, "method", "set", "method")
	// flag.StringVar(&arguments, "arguments", "", "arguments")
	flag.Int64Var(&argument, "argument", 0, "argument")

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	endpoints := strings.Split(flag.Arg(0), ",")
	client := rpcclient.NewClientJSONRPC(endpoints[0])

	switch *actionType {
	case "basic":
		basic(client)
	case "create":
		createContract(client, nil, nil)
	case "read":
		readContract(client)
	case "call":
		// args := []interface{}{big.NewInt(188)}
		// argList := strings.Split(arguments, ",")
		// args := make([]interface{}, len(argList))
		// for i, arg := range argList {
		// 	args[i] = arg
		// }
		// args := []interface{}{big.NewInt(argument)}
		args := []interface{}{big.NewInt(1), big.NewInt(argument)}

		callContract(client, method, args)
	case "basic_goroutine":
		basic_goroutine(client)
	case "create_goroutine":
		create_goroutine(client)
	case "read_goroutine":
		read_goroutine(client)
	case "call_goroutine":
		call_goroutine(client)
	default:
		panic("unsupport action: " + *actionType)
	}
}

func basic(client *rpcclient.ClientJSONRPC) {
	timeStart := time.Now().UnixNano() / 1000000
	fmt.Printf("START: %v\n", timeStart)
	for i := 0; i < requestNum; i++ {
		// randomTx := generateTx(1, i, "genesis")

		startTime := time.Now().UnixNano()

		tx := new(types.Transaction)
		tx.Data.TimeStamp = uint64(time.Now().UnixNano())
		pk := gcrypto.GenPrivKeyEd25519()
		tx.Data.Caller = common.BytesToAddress(pk.PubKey().Bytes())
		tx.Data.CryptoType = gcrypto.CryptoTypeZhongAn
		tx.Data.PublicKey = pk.PubKey().Bytes()

		sig := pk.Sign(tx.SigHash().Bytes())
		tx.Data.Signature = sig.Bytes()
		rlpedTx, err := tx.EncodeRLP()
		panicErr(err)

		rpcResult := new(gtypes.RPCResult)
		_, err = client.Call("broadcast_tx_commit", []interface{}{rlpedTx}, rpcResult)
		if err != nil {
			fmt.Println("client.Call failed with error: ", err)
		}
		panicErr(err)
		fmt.Printf("request %v spent %v ms\n", i, (time.Now().UnixNano()-startTime)/1000000)
	}
	fmt.Printf("END: %v\n", time.Now().UnixNano()/1000000)
}

func basic_goroutine(client *rpcclient.ClientJSONRPC) {
	timeStart := time.Now().UnixNano() / 1000000
	// fmt.Printf("START: %v\n", timeStart)

	// txs := make(chan []byte, requestNum)
	// res := make(chan struct {
	// 	int64
	// 	error
	// }, requestNum)
	// for i := 0; i < requestNum; i++ {
	// 	tx := new(types.Transaction)
	// 	tx.Data.TimeStamp = uint64(time.Now().UnixNano())
	// 	pk := gcrypto.GenPrivKeyEd25519()
	// 	tx.Data.Caller = common.BytesToAddress(pk.PubKey().Bytes())
	// 	tx.Data.CryptoType = gcrypto.CryptoTypeED25519
	// 	tx.Data.PublicKey = pk.PubKey().Bytes()

	// 	sig := pk.Sign(tx.SigHash().Bytes())
	// 	tx.Data.Signature = sig.Bytes()
	// 	rlpedTx, err := tx.EncodeRLP()
	// 	panicErr(err)
	// 	txs <- rlpedTx
	// }

	// for i := 0; i < procNum; i++ {
	// 	go func(workerId int) {
	// 		fmt.Println("Goroutine index: ", workerId)
	// 		for tx := range txs {
	// 			startTime := time.Now().UnixNano()
	// 			rpcResult := new(gtypes.RPCResult)
	// 			_, err := client.Call("broadcast_tx_commit", []interface{}{tx}, rpcResult)
	// 			if err != nil {
	// 				// fmt.Println("client.Call failed with error: ", err)
	// 				panicErr(err)
	// 				res <- struct {
	// 					int64
	// 					error
	// 				}{(time.Now().UnixNano() - startTime) / 1000000, err}
	// 				continue
	// 			}
	// 			res <- struct {
	// 				int64
	// 				error
	// 			}{(time.Now().UnixNano() - startTime) / 1000000, err}
	// 			// fmt.Printf("request %v spent %v ms\n", i, (time.Now().UnixNano()-startTime)/1000000)
	// 		}
	// 	}(i)
	// }

	// close(txs)

	// for i := 0; i < requestNum; i++ {
	// 	<-res
	// }

	var wg sync.WaitGroup
	for i := 0; i < requestNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			startTime := time.Now().UnixNano()

			fmt.Println("Goroutine index: ", idx)
			tx := new(types.Transaction)
			tx.Data.TimeStamp = uint64(time.Now().UnixNano())
			pk := gcrypto.GenPrivKeyEd25519()
			tx.Data.Caller = common.BytesToAddress(pk.PubKey().Bytes())
			tx.Data.CryptoType = gcrypto.CryptoTypeZhongAn
			tx.Data.PublicKey = pk.PubKey().Bytes()

			sig := pk.Sign(tx.SigHash().Bytes())
			tx.Data.Signature = sig.Bytes()
			rlpedTx, err := tx.EncodeRLP()
			panicErr(err)

			rpcResult := new(gtypes.RPCResult)
			_, err = client.Call("broadcast_tx_commit", []interface{}{rlpedTx}, rpcResult)
			if err != nil {
				fmt.Println("client.Call failed with error: ", err)
			}
			panicErr(err)
			fmt.Printf("request %v spent %v ms\n", idx, (time.Now().UnixNano()-startTime)/1000000)
		}(i)
	}
	wg.Wait()

	timeSpent := time.Now().UnixNano()/1000000 - timeStart
	fmt.Printf("[Proc Number.     ]: %v\n", procNum)
	fmt.Printf("[Request Number   ]: %v\n", requestNum)
	fmt.Printf("[Total Time Spent ]: %v ms\n", timeSpent)
	fmt.Printf("[Estimated TPS 	  ]: %v\n", float64(requestNum*1000)/float64(timeSpent))
}

func createContract(client *rpcclient.ClientJSONRPC, pkEcdsa *ecdsa.PrivateKey, wg *sync.WaitGroup) {
	bytecode := "6060604052341561000f57600080fd5b6101818061001e6000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063a6226f2114610051578063b051c1e01461007d575b600080fd5b341561005c57600080fd5b61007b60048080359060200190919080359060200190919050506100b4565b005b341561008857600080fd5b61009e6004808035906020019091905050610136565b6040518082815260200191505060405180910390f35b60007fb45ab3e8c50935ce2fa51d37817fd16e7358a3087fd93a9ac7fbddb22a926c358383604051808381526020018281526020019250505060405180910390a1828160000181905550818160010181905550806000808581526020019081526020016000206000820154816000015560018201548160010155905050505050565b60008060008381526020019081526020016000206001015490509190505600a165627a7a723058207eaf119132cfc4008c97339b874c4c16d20d27a72875e55a6a22a29fee30876d0029"
	// timeStart := time.Now().UnixNano() / 1000000
	// fmt.Printf("START: %v\n", timeStart)

	if wg != nil {
		defer wg.Done()
	}

	if pkEcdsa == nil {
		var err error
		// pkEcdsa, err = crypto.GenerateKey()
		pkEcdsa, err = crypto.ToECDSA(common.Hex2Bytes(defaultPrivKey))
		panicErr(err)
	}

	for i := 0; i < requestNum; i++ {
		// pkEcdsa = crypto.ToECDSA(common.Hex2Bytes(defaultPrivKey))
		caller := crypto.PubkeyToAddress(pkEcdsa.PublicKey)
		nonce, err := getNonce(client, caller.Hex())
		fmt.Printf("nonce: %v\n", nonce)
		panicErr(err)
		data := common.Hex2Bytes(bytecode)
		tx := etypes.NewContractCreation(nonce, big.NewInt(0), gasLimit, big.NewInt(0), data)

		sig, _ := crypto.Sign(ethSigner.Hash(tx).Bytes(), pkEcdsa)
		sigTx, _ := tx.WithSignature(ethSigner, sig)
		rlpSignedTx, err := rlp.EncodeToBytes(sigTx)
		panicErr(err)

		startTime := time.Now().UnixNano()

		rpcResult := new(gtypes.RPCResult)
		_, err = client.Call("broadcast_tx_commit", []interface{}{rlpSignedTx}, rpcResult)
		panicErr(err)
		fmt.Printf("request %v spent %v ms\n", i, (time.Now().UnixNano()-startTime)/1000000)

		// res := (*rpcResult).(*types.ResultBroadcastTx)
		// if res.Code != types.CodeType_OK {
		// 	fmt.Println(res.Code, string(res.Data))
		// 	panic(res.Data)
		// }

		// fmt.Println(res.Code, string(res.Data))
		contractAddr := crypto.CreateAddress(caller, nonce)
		fmt.Println("contract address:", contractAddr.Hex())
		time.Sleep(time.Millisecond * 1)

		nonce, err = getNonce(client, caller.Hex())
		fmt.Println("nonce: ", nonce)
	}

	// fmt.Printf("END: %v\n", time.Now().UnixNano()/1000000)

	// timeSpent := (time.Now().UnixNano()/1000000 - timeStart) // millisecond
	// fmt.Printf("END(%v requests complete!): %v ms\n", requestNum, timeSpent)
	// fmt.Printf("TPS: %v\n", int64(requestNum)*1000/timeSpent)

	// ============================== read request from file =========================
	// rd := bufio.NewReaderSize(os.Stdin, 1024*1024*124)
	// for {
	//  line, _, err := rd.ReadLine()
	//  if nil != err || io.EOF == err {
	//      break
	//  }
	//  js, err := simplejson.NewJson([]byte{line})
	//  if err != nil {
	//      panic(err)
	//  }

	// _, err := client.Call("broadcast_tx_commit", []interface{}{rlpSignedTx}, rpcResult)

	//  fmt.Printf("%v %v spent %v ms\n", time.Now().UnixNano()/1000000, string(line), (time.Now().UnixNano()-s)/1000000)
	// }
	// ======================================================================
}

func readContract(client *rpcclient.ClientJSONRPC) {
	contractAddr := "0x670958483e7281a77832642d4856e309276b48fc"
	contractAbi := "[{\"constant\":false,\"inputs\":[{\"name\":\"Id\",\"type\":\"uint256\"},{\"name\":\"Amount\",\"type\":\"uint256\"}],\"name\":\"createCheckInfos\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"Id\",\"type\":\"uint256\"}],\"name\":\"getPremiumInfos\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"Id\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"Amount\",\"type\":\"uint256\"}],\"name\":\"InputLog\",\"type\":\"event\"}]"

	abiJson, err := abi.JSON(strings.NewReader(contractAbi))
	panicErr(err)
	data, err := abiJson.Pack("getPremiumInfos", []interface{}{big.NewInt(1)}...)
	panicErr(err)

	privkey := gcmn.SanitizeHex(defaultPrivKey)
	pkEcdsa, err := crypto.HexToECDSA(privkey)
	panicErr(err)
	caller := crypto.PubkeyToAddress(pkEcdsa.PublicKey)
	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)
	fmt.Println("Nonce: ", nonce)

	tx := etypes.NewTransaction(nonce, common.HexToAddress(contractAddr), big.NewInt(0), gasLimit, big.NewInt(0), data)
	sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), pkEcdsa)
	panicErr(err)
	sigTx, err := tx.WithSignature(ethSigner, sig)
	panicErr(err)
	rlpSignedTx, err := rlp.EncodeToBytes(sigTx)
	panicErr(err)

	timeStart := time.Now().UnixNano() / 1000000
	fmt.Printf("START: %v\n", timeStart)
	query := append([]byte{types.QueryType_Contract}, rlpSignedTx...)

	for i := 0; i < requestNum; i++ {
		// startTime := time.Now().UnixNano()
		rpcResult := new(gtypes.RPCResult)
		_, err = client.Call("query", []interface{}{query}, rpcResult)
		panicErr(err)

		res := (*rpcResult).(*gtypes.ResultQuery)
		// fmt.Println("query result:", common.Bytes2Hex(res.Result.Data))
		parseResult, _ := commands.UnpackResult("get", abiJson, string(res.Result.Data))
		fmt.Println("parse result:", reflect.TypeOf(parseResult), parseResult)
		// fmt.Printf("request %v spent %v ms\n", i, (time.Now().UnixNano()-startTime)/1000000)
	}
	fmt.Printf("END: %v\n", time.Now().UnixNano()/1000000)
	// timeSpent := (time.Now().UnixNano()/1000000 - timeStart) // millisecond
	// fmt.Printf("END(%v requests complete!): %v ms\n", requestNum, timeSpent)
	// fmt.Printf("TPS: %v\n", int64(requestNum)*1000/timeSpent)
}

func callContract(client *rpcclient.ClientJSONRPC, method string, args []interface{}) {
	contractAddr := "0x670958483e7281a77832642d4856e309276b48fc"
	contractAbi := "[{\"constant\":false,\"inputs\":[{\"name\":\"Id\",\"type\":\"uint256\"},{\"name\":\"Amount\",\"type\":\"uint256\"}],\"name\":\"createCheckInfos\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"Id\",\"type\":\"uint256\"}],\"name\":\"getPremiumInfos\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"Id\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"Amount\",\"type\":\"uint256\"}],\"name\":\"InputLog\",\"type\":\"event\"}]"

	abiJson, err := abi.JSON(strings.NewReader(contractAbi))
	panicErr(err)
	// args := []interface{}{big.NewInt(188)}
	data, err := abiJson.Pack(method, args...) // contract function to be called
	panicErr(err)

	fmt.Println("data:", hex.EncodeToString(data))

	// pkEcdsa, err := crypto.GenerateKey()
	// panicErr(err)

	privkey := gcmn.SanitizeHex(defaultPrivKey)
	pkEcdsa, err := crypto.HexToECDSA(privkey)
	panicErr(err)
	caller := crypto.PubkeyToAddress(pkEcdsa.PublicKey)
	nonce, err := getNonce(client, caller.Hex())
	panicErr(err)
	fmt.Println("Nonce: ", nonce)

	// timeStart := time.Now().UnixNano() / 1000000
	// fmt.Printf("START: %v\n", timeStart)
	for i := 0; i < requestNum; i++ {
		tx := etypes.NewTransaction(nonce+uint64(i), common.HexToAddress(contractAddr), big.NewInt(0), gasLimit, big.NewInt(0), data)

		sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), pkEcdsa)
		panicErr(err)
		sigTx, err := tx.WithSignature(ethSigner, sig)
		panicErr(err)
		rlpSignedTx, err := rlp.EncodeToBytes(sigTx)
		panicErr(err)

		// startTime := time.Now().UnixNano()
		rpcResult := new(gtypes.RPCResult)
		_, err = client.Call("broadcast_tx_commit", []interface{}{rlpSignedTx}, rpcResult)
		panicErr(err)
		// fmt.Printf("request %v spent %v ms\n", i, (time.Now().UnixNano()-startTime)/1000000)

		// res := (*rpcResult).(*gtypes.ResultBroadcastTx)
		// if res.Code != gtypes.CodeType_OK {
		// 	fmt.Println(res.Code, string(res.Data), res.Log)
		// 	return errors.New(string(res.Data))
		// }
		nonce, err = getNonce(client, caller.Hex())
		fmt.Println("nonce: ", nonce)
	}
	// timeSpent := (time.Now().UnixNano()/1000000 - timeStart) // millisecond
	// fmt.Printf("END(%v requests complete!): %v ms\n", requestNum, timeSpent)
	// fmt.Printf("TPS: %v\n", int64(requestNum)*1000/timeSpent)
}

func create_goroutine(client *rpcclient.ClientJSONRPC) {
	timeStart := time.Now().UnixNano() / 1000000

	var wg sync.WaitGroup
	for i := 0; i < procNum; i++ {
		pkEcdsa, err := crypto.GenerateKey()
		panicErr(err)

		wg.Add(1)
		go createContract(client, pkEcdsa, &wg)
		// go func(idx int, privKey *ecdsa.PrivateKey) {
		// 	startTime := time.Now().UnixNano()

		// 	defer wg.Done()
		// 	// pk := crypto.ToECDSA(common.Hex2Bytes(privKey))
		// 	caller := crypto.PubkeyToAddress(pkEcdsa.PublicKey)
		// 	nonce, err := getNonce(client, caller.Hex())
		// 	panicErr(err)

		// 	data := common.Hex2Bytes(bytecode)
		// 	tx := etypes.NewContractCreation(nonce, big.NewInt(0), gasLimit, big.NewInt(0), data)

		// 	sig, _ := crypto.Sign(tx.SigHash(ethSigner).Bytes(), pkEcdsa)
		// 	sigTx, _ := tx.WithSignature(ethSigner, sig)
		// 	rlpSignedTx, err := rlp.EncodeToBytes(sigTx)
		// 	panicErr(err)

		// 	rpcResult := new(gtypes.RPCResult)
		// 	_, err = client.Call("broadcast_tx_commit", []interface{}{rlpSignedTx}, rpcResult)
		// 	panicErr(err)
		// 	fmt.Printf("request %v spent %v ms\n", i, (time.Now().UnixNano()-startTime)/1000000)

		// 	// res := (*rpcResult).(*types.ResultBroadcastTx)
		// 	// if res.Code != types.CodeType_OK {
		// 	// 	fmt.Println(res.Code, string(res.Data))
		// 	// 	panic(res.Data)
		// 	// }

		// 	// fmt.Println(res.Code, string(res.Data))
		// 	contractAddr := crypto.CreateAddress(caller, nonce)
		// 	fmt.Println("contract address:", contractAddr.Hex())
		// }(i, pkEcdsa)
	}

	wg.Wait()
	timeSpent := time.Now().UnixNano()/1000000 - timeStart
	fmt.Printf("[Proc Number.     ]: %v\n", procNum)
	fmt.Printf("[Request Number   ]: %v\n", requestNum)
	fmt.Printf("[Total Time Spent ]: %v ms\n", timeSpent)
	fmt.Printf("[Estimated TPS 	  ]: %v\n", float64(requestNum*1000)/float64(timeSpent))
}

func read_goroutine(client *rpcclient.ClientJSONRPC) {
	timeStart := time.Now().UnixNano() / 1000000
	var wg sync.WaitGroup
	for i := 0; i < requestNum; i++ {
		abiJson, err := abi.JSON(strings.NewReader(defaultAbis))
		panicErr(err)
		data, err := abiJson.Pack("get", []interface{}{}...)
		panicErr(err)

		privkey := gcmn.SanitizeHex(defaultPrivKey)
		pkEcdsa, err := crypto.HexToECDSA(privkey)
		panicErr(err)
		caller := crypto.PubkeyToAddress(pkEcdsa.PublicKey)
		nonce, err := getNonce(client, caller.Hex())
		panicErr(err)

		wg.Add(1)
		go func(idx int, nonce uint64) {
			defer wg.Done()
			startTime := time.Now().UnixNano()

			tx := etypes.NewTransaction(nonce, common.HexToAddress(defaultContractAddr), big.NewInt(0), gasLimit, big.NewInt(0), data)
			sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), pkEcdsa)
			panicErr(err)
			sigTx, err := tx.WithSignature(ethSigner, sig)
			panicErr(err)
			rlpSignedTx, err := rlp.EncodeToBytes(sigTx)
			panicErr(err)

			query := append([]byte{types.QueryType_Contract}, rlpSignedTx...)

			rpcResult := new(gtypes.RPCResult)
			_, err = client.Call("query", []interface{}{query}, rpcResult)
			panicErr(err)

			// res := (*rpcResult).(*gtypes.ResultQuery)
			// fmt.Println("query result:", common.Bytes2Hex(res.Result.Data))
			// parseResult, _ := unpackResult(callfunc, aabbii, string(res.Result.Data))
			// fmt.Println("parse result:", reflect.TypeOf(parseResult), parseResult)
			fmt.Printf("request %v spent %v ms\n", i, (time.Now().UnixNano()-startTime)/1000000)
		}(i, nonce)
	}

	wg.Wait()
	timeSpent := time.Now().UnixNano()/1000000 - timeStart
	fmt.Printf("[Proc Number.     ]: %v\n", procNum)
	fmt.Printf("[Request Number   ]: %v\n", requestNum)
	fmt.Printf("[Total Time Spent ]: %v ms\n", timeSpent)
	fmt.Printf("[Estimated TPS 	  ]: %v\n", float64(requestNum*1000)/float64(timeSpent))
}

func call_goroutine(client *rpcclient.ClientJSONRPC) {
	timeStart := time.Now().UnixNano() / 1000000

	contractAddr := "0xb24e3d7537b4389e6c730799fc64701e95457706"
	contractAbi := "[{\"constant\":false,\"inputs\":[{\"name\":\"x\",\"type\":\"uint256\"}],\"name\":\"set\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

	var wg sync.WaitGroup
	for i := 0; i < requestNum; i++ {
		pkEcdsa, err := crypto.GenerateKey()
		panicErr(err)

		wg.Add(1)
		go func(idx int, privKey *ecdsa.PrivateKey, contractAddr string) {
			defer wg.Done()
			startTime := time.Now().UnixNano()

			abiJson, err := abi.JSON(strings.NewReader(contractAbi))
			panicErr(err)
			args := []interface{}{big.NewInt(188)}
			data, err := abiJson.Pack("set", args...) // contract function to be called
			panicErr(err)

			// privkey := gcmn.SanitizeHex(privKey)
			// pkEcdsa, err := crypto.HexToECDSA(privkey)
			// panicErr(err)
			caller := crypto.PubkeyToAddress(pkEcdsa.PublicKey)
			nonce, err := getNonce(client, caller.Hex())
			panicErr(err)

			tx := etypes.NewTransaction(nonce+uint64(i), common.HexToAddress(contractAddr), big.NewInt(0), gasLimit, big.NewInt(0), data)
			sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), pkEcdsa)
			panicErr(err)
			sigTx, err := tx.WithSignature(ethSigner, sig)
			panicErr(err)
			rlpSignedTx, err := rlp.EncodeToBytes(sigTx)
			panicErr(err)

			rpcResult := new(gtypes.RPCResult)
			_, err = client.Call("broadcast_tx_commit", []interface{}{rlpSignedTx}, rpcResult)
			panicErr(err)
			fmt.Printf("request %v spent %v ms\n", i, (time.Now().UnixNano()-startTime)/1000000)

			// res := (*rpcResult).(*gtypes.ResultBroadcastTx)
			// if res.Code != gtypes.CodeType_OK {
			// 	fmt.Println(res.Code, string(res.Data), res.Log)
			// 	return errors.New(string(res.Data))
			// }
		}(i, pkEcdsa, contractAddr)
	}

	wg.Wait()
	timeSpent := time.Now().UnixNano()/1000000 - timeStart
	fmt.Printf("[Proc Number.     ]: %v\n", procNum)
	fmt.Printf("[Request Number   ]: %v\n", requestNum)
	fmt.Printf("[Total Time Spent ]: %v ms\n", timeSpent)
	fmt.Printf("[Estimated TPS 	  ]: %v\n", float64(requestNum*1000)/float64(timeSpent))
}

func generateTx(connIndex int, txNumber int, chainId string) []byte {
	tx := make([]byte, txSize)

	binary.PutUvarint(tx[:8], uint64(connIndex))
	binary.PutUvarint(tx[8:16], uint64(txNumber))

	chainHash := md5.Sum([]byte(chainId))
	copy(tx[16:32], chainHash[:16])
	binary.PutUvarint(tx[32:40], uint64(time.Now().Unix()))

	// 40-* random data
	if _, err := rand.Read(tx[40:]); err != nil {
		panic(errors.Wrap(err, "failed to read random bytes"))
	}

	return tx
}

func getNonce(client *rpcclient.ClientJSONRPC, fromAddress string) (uint64, error) {
	rpcResult := new(gtypes.RPCResult)

	addr := common.Hex2Bytes(gcmn.SanitizeHex(fromAddress))
	query := append([]byte{types.QueryType_Nonce}, addr...)

	if client == nil {
		// panic("client is nil")
		return 0, fmt.Errorf("client is nil")
	}
	_, err := client.Call("query", []interface{}{query}, rpcResult)
	if err != nil {
		// panic(err)
		return 0, err
	}

	res := (*rpcResult).(*gtypes.ResultQuery)
	if res.Result.IsErr() {
		// fmt.Println(res.Result.Code, res.Result.Log)
		// panic(res.Result.Error())
		return 0, fmt.Errorf(res.Result.Error())
	}
	nonce := new(uint64)
	err = rlp.DecodeBytes(res.Result.Data, nonce)
	if err != nil {
		return 0, err
	}
	// panicErr(err)

	return *nonce, nil
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
