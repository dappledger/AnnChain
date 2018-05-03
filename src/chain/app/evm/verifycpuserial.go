package evm

import (
	"bytes"

	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/angine/types"
)

func exeWithCPUSerialVeirfy(signer ethtypes.Signer, txs [][]byte, quit chan struct{},
	whenExec func(index int, raw []byte, tx *ethtypes.Transaction, abi []byte), whenError func(raw []byte, err error)) error {
	for i, raw := range txs {
		var abi, txBytes []byte
		txType := raw[:4]
		switch {
		case bytes.Equal(txType, EVMTxTag):
			txBytes = raw[4:]
		case bytes.Equal(txType, EVMCreateContractTxTag):
			if txCreate, err := DecodeCreateContract(types.UnwrapTx(raw)); err != nil {
				whenError(raw, err)
				continue
			} else {
				txBytes = txCreate.EthTx
				abi = txCreate.EthAbi
			}
		}
		var tx *ethtypes.Transaction
		if len(txBytes) > 0 {
			tx = new(ethtypes.Transaction)
			if err := rlp.DecodeBytes(txBytes, tx); err != nil {
				whenError(raw, err)
				continue
			}
		}
		whenExec(i, raw, tx, abi)
	}

	return nil
}
