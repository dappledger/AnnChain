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
	"math/big"

	at "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/genesis/types"
)

type DoPayment struct {
	app *GenesisApp
	op  *types.PaymentOp
	tx  *types.Transaction
}

func (pay *DoPayment) PreCheck() at.Result {
	return at.NewResultOK([]byte{}, "")
}

func (pay *DoPayment) CheckValid(stateDup *stateDup) error {
	if !stateDup.state.Exist(pay.op.Destination) {
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}
	if pay.op.Source.Hex() == pay.op.Destination.Hex() {
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}
	if pay.op.Amount.Cmp(big.NewInt(0)) <= 0 {
		return at.NewError(at.CodeType_LowBalance, at.CodeType_LowBalance.String())
	}
	return nil
}

func (pay *DoPayment) Apply(stateDup *stateDup) error {

	balance := stateDup.state.GetBalance(pay.op.Source)

	if new(big.Int).Sub(balance, pay.op.Amount).Cmp(big.NewInt(0)) < 0 {
		return at.NewError(at.CodeType_LowBalance, at.CodeType_LowBalance.String())
	}

	stateDup.state.SubBalance(pay.op.Source, pay.op.Amount, "payment")
	stateDup.state.AddBalance(pay.op.Destination, pay.op.Amount, "payment")

	pay.effects()

	return nil
}

func (pay *DoPayment) effects() {

	act := &types.ActionPayment{
		ActionBase: types.ActionBase{
			Typei:       types.OP_S_PAYMENT.OpInt(),
			Type:        types.OP_S_PAYMENT,
			FromAccount: pay.op.Source,
			ToAccount:   pay.op.Destination,
			Nonce:       pay.tx.Nonce(),
			BaseFee:     pay.tx.BaseFee(),
			Memo:        pay.tx.Data.Memo,
		},
		From:   pay.op.Source,
		To:     pay.op.Destination,
		Amount: pay.op.Amount.String(),
	}

	efts := []types.EffectObject{
		&types.EffectAccountCredited{
			EffectBase: types.EffectBase{
				Typei:   types.EffectTypeAccountCredited,
				Type:    types.EffectTypeAccountCredited.String(),
				Account: pay.op.Destination,
			},
			Amount: pay.op.Amount.String(),
		},
		&types.EffectAccountDebited{
			EffectBase: types.EffectBase{
				Typei:   types.EffectTypeAccountDebited,
				Type:    types.EffectTypeAccountDebited.String(),
				Account: pay.op.Source,
			},
			Amount: pay.op.Amount.String(),
		},
	}
	pay.op.SetEffects(act, efts)
}
