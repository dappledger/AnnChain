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

package app

import (
	"encoding/json"
	"math/big"
	"strconv"

	at "github.com/dappledger/AnnChain/angine/types"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
)

type DoCreateContract struct {
	app *GenesisApp
	op  *types.CreateContractOp
	tx  *types.Transaction
	gas string
}

func (ca *DoCreateContract) PreCheck() at.Result {
	return at.NewResultOK([]byte{}, "")
}

func (ca *DoCreateContract) CheckValid(stateDup *stateDup) error {

	if len(ca.op.Source.Bytes()) != ethcmn.AddressLength || !stateDup.state.Exist(ca.op.Source) {
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}

	r, ok := new(big.Int).SetString(ca.op.GasLimit, 10)
	if !ok {
		panic("invalid hex in source file: " + ca.op.GasLimit)
	}

	if r.Cmp(types.MAX_GASLIMIT) > 0 {
		return at.NewError(at.CodeType_BadLimit, at.CodeType_BadLimit.String())
	}

	return nil
}

func (ca *DoCreateContract) Apply(stateDup *stateDup) error {

	var (
		jsLogs   []*json.RawMessage
		sreceipt *types.Receipt
	)
	nLimit, err := strconv.ParseInt(ca.op.GasLimit, 10, 64)
	if err != nil {
		return at.NewError(at.CodeType_BadLimit, at.CodeType_BadLimit.String())
	}
	nPrice, err := strconv.ParseInt(ca.op.Price, 10, 64)
	if err != nil {
		return at.NewError(at.CodeType_BadPrice, at.CodeType_BadPrice.String())
	}
	nAmount, err := strconv.ParseInt(ca.op.Amount, 10, 64)
	if err != nil {
		return at.NewError(at.CodeType_BadAmount, at.CodeType_BadAmount.String())
	}

	payLoad, err := ethcmn.HexToByte(ca.op.Payload)
	if err != nil {
		return at.NewError(at.CodeType_BaseInvalidInput, err.Error())
	}

	tx := NewContractCreation(stateDup.state.GetNonce(ca.op.Source), ca.op.Source, big.NewInt(nAmount), big.NewInt(nLimit), big.NewInt(nPrice), payLoad)

	stateDup.state.StartRecord(tx.Hash(), ethcmn.BytesToHash(ca.app.LoadLastBlock().AppHash), 1)

	receipt, gas, err := RunEvm(ca.app.EvmCurrentHeader, stateDup.state, tx)

	ca.gas = gas.String()

	if err != nil {
		sreceipt = &types.Receipt{
			Nonce:           ca.tx.Nonce(),
			TxHash:          ca.tx.Hash(),
			TxReceiptStatus: false,
			Message:         err.Error(),
			GasUsed:         gas,
			Height:          ca.app.EvmCurrentHeader.Number.Uint64(),
			Source:          ca.op.Source,
			GasPrice:        ca.op.Price,
			GasLimit:        ca.op.GasLimit,
			Payload:         ca.op.Payload,
			OpType:          types.OP_S_CREATECONTRACT,
		}
		stateDup.receipts = append(stateDup.receipts, sreceipt)
	} else {
		if receipt.Logs != nil {
			for _, log := range receipt.Logs {
				bytLog, err := log.MarshalJSON()
				if err != nil {
					bytLog = []byte("")
				}
				raw := json.RawMessage(bytLog)
				jsLogs = append(jsLogs, &raw)
			}
		}
		ca.op.ContractAddr = receipt.ContractAddress.Hex()
		sreceipt = &types.Receipt{
			Nonce:           ca.tx.Nonce(),
			TxHash:          ca.tx.Hash(),
			TxReceiptStatus: receipt.Status,
			Message:         "",
			ContractAddress: ca.op.ContractAddr,
			GasUsed:         gas,
			Height:          receipt.Height,
			Source:          ca.op.Source,
			GasPrice:        ca.op.Price,
			GasLimit:        ca.op.GasLimit,
			Payload:         ca.op.Payload,
			Logs:            jsLogs,
			OpType:          types.OP_S_CREATECONTRACT,
		}
		stateDup.receipts = append(stateDup.receipts, sreceipt)

	}

	ca.effects()

	return nil
}

func (ca *DoCreateContract) effects() {
	act := &types.ActionCreateContract{
		ActionBase: types.ActionBase{
			Typei:       types.OP_S_CREATECONTRACT.OpInt(),
			Type:        types.OP_S_CREATECONTRACT,
			FromAccount: ca.op.Source,
			ToAccount:   ca.tx.GetTo(),
			Nonce:       ca.tx.Nonce(),
			BaseFee:     ca.tx.BaseFee(),
			Memo:        ca.tx.Data.Memo,
			CreateAt:    ca.tx.GetCreateTime(),
		},
		ContractAddr: ca.op.ContractAddr,
		Amount:       ca.op.Amount,
		GasLimit:     ca.op.GasLimit,
		Price:        ca.op.Price,
		Gas:          ca.gas,
	}

	efts := []types.EffectObject{
		&types.EffectCreateContract{
			EffectBase: types.EffectBase{
				Typei:   types.EffectTypeCreateContract,
				Type:    types.EffectTypeCreateContract.String(),
				Account: ca.op.Source,
			},
			ContractAddr: ca.op.ContractAddr,
			Amount:       ca.op.Amount,
			GasLimit:     ca.op.GasLimit,
			Price:        ca.op.Price,
			Gas:          ca.gas,
		},
	}
	ca.op.SetEffects(act, efts)
}
