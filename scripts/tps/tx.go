// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
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

package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	cl "github.com/dappledger/AnnChain/gemmill/rpc/client"
	"github.com/dappledger/AnnChain/gemmill/types"
	atypes "github.com/dappledger/AnnChain/gemmill/types"
)

var (
	// this signer appears to be a must in evm 1.5.9
	ethSigner = ethtypes.HomesteadSigner{}
)

func sendTx(privkey *ecdsa.PrivateKey, nonce uint64, to common.Address, data []byte, commit bool) error {
	tx := ethtypes.NewTransaction(nonce, to, big.NewInt(0), 90000, big.NewInt(0), data)

	if privkey != nil {
		sig, err := crypto.Sign(ethSigner.Hash(tx).Bytes(), privkey)
		panicErr(err)
		sigTx, err := tx.WithSignature(ethSigner, sig)
		panicErr(err)
		b, err := rlp.EncodeToBytes(sigTx)
		panicErr(err)

		clientJSON := cl.NewClientJSONRPC(rpcTarget)
		if !commit {
			rpcResult := new(types.ResultBroadcastTx)
			_, err = clientJSON.Call("broadcast_tx_async", []interface{}{b}, rpcResult)
			panicErr(err)

			hash := atypes.Tx(b).Hash()
			fmt.Println("tx result:", hex.EncodeToString(hash))

			return nil
		}

		rpcResult := new(types.ResultBroadcastTxCommit)
		_, err = clientJSON.Call("broadcast_tx_commit", []interface{}{b}, rpcResult)
		panicErr(err)

		hash := atypes.Tx(b).Hash()
		fmt.Println("tx result:", hex.EncodeToString(hash))

		return nil
	}

	return nil
}
