package types

import (
	"math/rand"

	crypto "github.com/dappledger/AnnChain/gemmill/go-crypto"
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
		tx := make([]byte, len(SpecialTag))
		copy(tx, SpecialTag)
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
