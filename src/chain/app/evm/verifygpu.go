package evm

import (
	"bytes"

	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/angine/types"
)

type gpuTx struct {
	index    int
	rawbytes types.Tx
	tx       *ethtypes.Transaction
	abi      []byte // only used in contract creation tx
	err      error
}

func exeWithGPUVeirfy(signer ethtypes.Signer, txs [][]byte, quit chan struct{},
	whenExec func(index int, raw []byte, tx *ethtypes.Transaction, abi []byte), whenError func(bs []byte, err error)) error {

	// // make queue
	// gpuTxQ := make([]gpuTx, len(txs))
	// makeGPUTxQueue(txs, gpuTxQ)

	// // pick txs for validation
	// trackTree := make(map[*ethtypes.Transaction]*gpuTx, len(txs))
	// batchTxForGPU := make([]*ethtypes.Transaction, 0, len(txs))
	// for i, v := range gpuTxQ {
	// 	if v.err != nil {
	// 		whenError(v.rawbytes, v.err)
	// 	} else {
	// 		batchTxForGPU = append(batchTxForGPU, v.tx)
	// 		trackTree[v.tx] = &gpuTxQ[i]
	// 	}
	// }

	// // do batch validation
	// ethtypes.BatchSender(EthSigner, batchTxForGPU)()

	// // execute valid txs
	// for _, v := range batchTxForGPU {
	// 	gtx := trackTree[v]
	// 	if _, err := ethtypes.LoadFrom(v); err != nil {
	// 		whenError(gtx.rawbytes, fmt.Errorf("error found in sigature validation"))
	// 	} else {
	// 		whenExec(gtx.index, gtx.rawbytes, gtx.tx, gtx.abi)
	// 	}
	// }

	return nil
}

func makeGPUTxQueue(txs [][]byte, gpuTxQ []gpuTx) {
	for i, raw := range txs {
		gpuTxQ[i].rawbytes = raw
		gpuTxQ[i].index = i

		var txBytes []byte
		txType := raw[:4]
		switch {
		case bytes.Equal(txType, EVMTxTag):
			txBytes = raw[4:]
		case bytes.Equal(txType, EVMCreateContractTxTag):
			if txCreate, err := DecodeCreateContract(types.UnwrapTx(raw)); err != nil {
				gpuTxQ[i].err = err
			} else {
				txBytes = txCreate.EthTx
				gpuTxQ[i].abi = txCreate.EthAbi
			}
		}
		if len(txBytes) > 0 {
			gpuTxQ[i].tx = new(ethtypes.Transaction)
			if err := rlp.DecodeBytes(txBytes, gpuTxQ[i].tx); err != nil {
				gpuTxQ[i].err = err
			}
		}
	}
}
