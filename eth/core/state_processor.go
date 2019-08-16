// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/core/state"
	"github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/core/vm"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/params"
	"github.com/dappledger/AnnChain/eth/rlp"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"

	"math/big"
)

var Big1 = big.NewInt(1)

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, 0, err
	}
	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc, author)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)

	// Apply the transaction to the current state (included in the env)
	_, gas, failed, err := ApplyMessage(vmenv, msg, gp)
	if err != nil {
		return nil, 0, err
	}
	// Update the state with pending changes
	var root []byte
	// if config.IsByzantium(header.Number) {
	// 	statedb.Finalise(true)
	// } else {
	// 	root = statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
	// }
	//Edit by zhongan
	statedb.Finalise(true)
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)

	// Edit by zhongan
	txBytes, _ := rlp.EncodeToBytes(tx)
	receipt.TxHash = common.BytesToHash(gtypes.Tx(txBytes).Hash())

	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	// Edit by zhongan
	receipt.Logs = statedb.GetLogs(receipt.TxHash)
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, gas, err
}
