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
