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
	"fmt"
	"math/big"
	"strconv"

	at "github.com/dappledger/AnnChain/angine/types"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
)

type DoExcuteContract struct {
	app *GenesisApp
	op  *types.ExcuteContractOp
	tx  *types.Transaction
}

func (ca *DoExcuteContract) PreCheck() at.Result {
	return at.NewResultOK([]byte{}, "")
}

func (ca *DoExcuteContract) CheckValid(stateDup *stateDup) error {

	if ca.op.Source == types.ZERO_ADDRESS || ca.op.To == types.ZERO_ADDRESS {
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}

	r, ok := new(big.Int).SetString(ca.op.GasLimit, 10)

	if !ok || r.Cmp(types.MAX_GASLIMIT) > 0 {
		return at.NewError(at.CodeType_BadLimit, at.CodeType_BadLimit.String())
	}

	return nil
}

func (ca *DoExcuteContract) Apply(stateDup *stateDup) error {
	var (
		jsLogs   []*json.RawMessage
		sreceipt *types.Receipt
		err      error
	)

	if !stateDup.state.Exist(ca.op.Source) || !stateDup.state.Exist(ca.op.To) {
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}
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

	hashBytes := stateDup.state.GetCodeHash(ca.op.To)
	if len(hashBytes) != ethcmn.HashLength || ethcmn.EmptyHash(hashBytes) {
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}

	payLoad, err := ethcmn.HexToByte(ca.op.Payload)
	if err != nil {
		return at.NewError(at.CodeType_BaseInvalidInput, err.Error())
	}

	tx := NewContractTransaction(stateDup.state.GetNonce(ca.op.Source), ca.op.Source, ca.op.To, big.NewInt(nAmount), big.NewInt(nLimit), big.NewInt(nPrice), payLoad)

	stateDup.state.StartRecord(tx.Hash(), ethcmn.BytesToHash(ca.app.LoadLastBlock().AppHash), 1)

	receipt, _, err := RunEvm(ca.app.EvmCurrentHeader, stateDup.state, tx)
	if err != nil {
		sreceipt = &types.Receipt{
			Nonce:           ca.tx.Nonce(),
			TxHash:          ca.tx.Hash(),
			TxReceiptStatus: false,
			Message:         err.Error(),
			ContractAddress: ca.op.To.Hex(),
			Height:          ca.app.EvmCurrentHeader.Number.Uint64(),
			Source:          ca.op.Source,
			GasPrice:        ca.op.Price,
			GasLimit:        ca.op.GasLimit,
			Payload:         ca.op.Payload,
			OpType:          types.OP_S_EXECUTECONTRACT,
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

		sreceipt = &types.Receipt{
			Nonce:           ca.tx.Nonce(),
			TxHash:          ca.tx.Hash(),
			TxReceiptStatus: receipt.Status,
			Message:         "",
			Res:             ethcmn.Bytes2Hex(receipt.Ret),
			ContractAddress: ca.op.To.Hex(),
			GasUsed:         receipt.GasUsed,
			Height:          receipt.Height,
			Source:          ca.op.Source,
			GasPrice:        ca.op.Price,
			GasLimit:        ca.op.GasLimit,
			Logs:            jsLogs,
			Payload:         ca.op.Payload,
			OpType:          types.OP_S_EXECUTECONTRACT,
		}
		ca.op.Gas = fmt.Sprintf("%v", sreceipt.GasUsed)
		stateDup.receipts = append(stateDup.receipts, sreceipt)
	}

	ca.effects()

	return nil
}

func (ca *DoExcuteContract) effects() {
	act := &types.ActionExcuteContract{
		ActionBase: types.ActionBase{
			Typei:       types.OP_S_EXECUTECONTRACT.OpInt(),
			Type:        types.OP_S_EXECUTECONTRACT,
			FromAccount: ca.op.Source,
			ToAccount:   ca.op.To,
			Nonce:       ca.tx.Nonce(),
			BaseFee:     ca.tx.BaseFee(),
			Memo:        ca.tx.Data.Memo,
			CreateAt:    ca.tx.GetCreateTime(),
		},
		ContractAddr: ca.op.To.Hex(),
		Amount:       ca.op.Amount,
		GasLimit:     ca.op.GasLimit,
		Price:        ca.op.Price,
		Gas:          ca.op.Gas,
	}

	efts := []types.EffectObject{
		&types.EffectExcuteContract{
			EffectBase: types.EffectBase{
				Typei:   types.EffectTypeExecuteContract,
				Type:    types.EffectTypeExecuteContract.String(),
				Account: ca.op.Source,
			},
			ContractAddr: ca.op.To.Hex(),
			Amount:       ca.op.Amount,
			GasLimit:     ca.op.GasLimit,
			Price:        ca.op.Price,
			Gas:          ca.op.Gas,
		},
	}
	ca.op.SetEffects(act, efts)
}
