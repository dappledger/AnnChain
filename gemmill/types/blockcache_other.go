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

package types

import (
	"math/rand"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
)

// =============for test code

func TxsLenForTest() int {
	return rand.Intn(100) + 100
}

func TxsNumForTest() int {
	return rand.Intn(10000)
}

func GenTxsForTest(numtx, numExTx int) ([]Tx, []Tx) {
	txs := make([]Tx, 0, numtx)
	extxs := make([]Tx, 0, numExTx)
	for i := 0; i < numtx; i++ {
		txs = append(txs, Tx(crypto.CRandBytes(TxsLenForTest())))
	}
	for i := 0; i < numExTx; i++ {
		tx := make([]byte, len(AdminTag))
		copy(tx, AdminTag)
		tx = append(tx, crypto.CRandBytes(TxsLenForTest())...)
		extxs = append(extxs, tx)
	}
	return txs, extxs
}

func GenCommitForTest(height int64, chainID string) *Commit {
	return &Commit{}
}

func randomTo2Nums(total int) (int, int) {
	randn := rand.Intn(total)
	return randn, total - randn
}

func GenBlockForTest(height int64, ntxs, nexTxs int) *Block {
	chainID := "test-chain"
	txs, extxs := GenTxsForTest(ntxs, nexTxs)
	block, _ := MakeBlock(
		height,
		chainID,
		txs,
		extxs,
		GenCommitForTest(height, chainID),
		nil,
		GenBlockID(),
		crypto.CRandBytes(20),
		crypto.CRandBytes(20),
		crypto.CRandBytes(20),
		20)
	return block
}

func GenBlockID() BlockID {
	return BlockID{
		Hash: crypto.CRandBytes(20),
		PartsHeader: PartSetHeader{
			Total: 20,
			Hash:  crypto.CRandBytes(20),
		},
	}
}

func GenBlocksForTest(num int) []*Block {
	blocks := make([]*Block, num)
	for i := 0; i < num; i++ {
		txs, exTxs := randomTo2Nums(TxsNumForTest())
		blocks[i] = GenBlockForTest(int64(i+1), txs, exTxs)
	}
	return blocks
}
