package types

import (
	"math/rand"

	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	crypto "github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

// =============for test code

func TxsLenForTest() int {
	return rand.Intn(100) + 100
}

func TxsNumForTest() int {
	return rand.Intn(10000)
}

func GenTxsForTest(numtx, numExTx int) []Tx {
	txs := make([]Tx, 0, numtx+numExTx)
	for i := 0; i < numtx; i++ {
		txs = append(txs, Tx(crypto.CRandBytes(TxsLenForTest())))
	}
	for i := 0; i < numExTx; i++ {
		tx := make([]byte, len(SpecialTag))
		copy(tx, SpecialTag)
		tx = append(tx, crypto.CRandBytes(TxsLenForTest())...)
		txs = append(txs, tx)
	}
	return txs
}

func GenCommitCacheForTest(height def.INT, chainID string) *CommitCache {
	return NewCommitCache(&pbtypes.Commit{})
}

func randomTo2Nums(total int) (int, int) {
	randn := rand.Intn(total)
	return randn, total - randn
}

func GenBlockForTest(height def.INT, txs, exTxs int) *BlockCache {
	maker := crypto.CRandBytes(32)
	chainID := "test-chain"
	blockCache, _ := MakeBlock(
		maker,
		height,
		chainID,
		GenTxsForTest(txs, exTxs),
		GenCommitCacheForTest(height, chainID),
		GenBlockID(),
		crypto.CRandBytes(20),
		crypto.CRandBytes(20),
		crypto.CRandBytes(20),
		20,
		100)
	return blockCache
}

func GenBlockID() pbtypes.BlockID {
	return pbtypes.BlockID{
		Hash: crypto.CRandBytes(20),
		PartsHeader: &pbtypes.PartSetHeader{
			Total: 20,
			Hash:  crypto.CRandBytes(20),
		},
	}
}

func GenBlocksForTest(num int) []*BlockCache {
	blocks := make([]*BlockCache, num)
	for i := 0; i < num; i++ {
		txs, exTxs := randomTo2Nums(TxsNumForTest())
		blocks[i] = GenBlockForTest(def.INT(i+1), txs, exTxs)
	}
	return blocks
}

func GenBlocksPbForTest(num int) []*pbtypes.Block {
	blocks := GenBlocksForTest(num)
	pbb := make([]*pbtypes.Block, num)
	for i := range blocks {
		pbb[i] = blocks[i].Block
	}
	return pbb
}
