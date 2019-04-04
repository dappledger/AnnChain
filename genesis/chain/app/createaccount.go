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

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"

	at "github.com/dappledger/AnnChain/angine/types"
)

type DoCreateAccount struct {
	app *GenesisApp
	op  *types.CreateAccountOp
	tx  *types.Transaction
}

func (ca *DoCreateAccount) PreCheck() at.Result {

	if ca.op.Source == types.ZERO_ADDRESS || ca.op.TargetAddress == types.ZERO_ADDRESS {
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}
	if ca.op.StartBalance == nil || ca.op.StartBalance.Cmp(types.BIG_MINBLC) < 0 {
		return at.NewError(at.CodeType_LowBalance, at.CodeType_LowBalance.String())
	}
	return at.NewResultOK([]byte{}, "")
}

func (ca *DoCreateAccount) CheckValid(stateDup *stateDup) error {

	var isInitAddr bool = false

	for _, initAddr := range ca.app.Init_Accounts {
		addr := ethcmn.HexToAddress(initAddr.Address)
		if addr.Hex() == ca.op.Source.Hex() {
			isInitAddr = true
			break
		}
	}
	if !isInitAddr {
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}
	if stateDup.state.Exist(ca.op.TargetAddress) {
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}
	if stateDup.state.GetBalance(ca.op.Source).Cmp(ca.op.StartBalance) < 0 {
		return at.NewError(at.CodeType_LowBalance, at.CodeType_LowBalance.String())
	}

	return nil
}

func (ca *DoCreateAccount) Apply(stateDup *stateDup) error {

	stateDup.state.CreateAccount(ca.op.TargetAddress)

	stateDup.state.SubBalance(ca.op.Source, ca.op.StartBalance, "createaccount")

	stateDup.state.SetBalance(ca.op.TargetAddress, ca.op.StartBalance, "createaccount")

	ca.effects()

	return nil
}

func (ca *DoCreateAccount) effects() {

	act := &types.ActionCreateAccount{

		ActionBase: types.ActionBase{
			Typei:       types.OP_S_CREATEACCOUNT.OpInt(),
			Type:        types.OP_S_CREATEACCOUNT,
			FromAccount: ca.op.Source,
			ToAccount:   ca.op.TargetAddress,
			Nonce:       ca.tx.Nonce(),
			BaseFee:     ca.tx.BaseFee(),
			Memo:        ca.tx.Data.Memo,
		},

		StartingBalance: ca.op.StartBalance,
	}

	efts := []types.EffectObject{

		&types.EffectAccountCreated{
			EffectBase: types.EffectBase{
				Typei:   types.EffectTypeAccountCreated,
				Type:    types.EffectTypeAccountCreated.String(),
				Account: ca.op.TargetAddress,
			},
			StartingBalance: ca.op.StartBalance,
		},

		&types.EffectAccountCredited{
			EffectBase: types.EffectBase{
				Typei:   types.EffectTypeAccountCredited,
				Type:    types.EffectTypeAccountCredited.String(),
				Account: ca.op.TargetAddress,
			},
			Amount: ca.op.StartBalance.String(),
		},

		&types.EffectAccountDebited{
			EffectBase: types.EffectBase{
				Typei:   types.EffectTypeAccountDebited,
				Type:    types.EffectTypeAccountDebited.String(),
				Account: ca.op.Source,
			},
			Amount: ca.op.StartBalance.String(),
		},
	}
	ca.op.SetEffects(act, efts)
}
