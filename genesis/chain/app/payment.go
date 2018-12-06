package app

import (
	"math/big"

	at "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/genesis/types"
)

type DoPayment struct {
	app *DelosApp
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
	if pay.op.Amount.Cmp(big0) <= 0 {
		return at.NewError(at.CodeType_LowBalance, at.CodeType_LowBalance.String())
	}
	return nil
}

func (pay *DoPayment) Apply(stateDup *stateDup) error {

	balance := stateDup.state.GetBalance(pay.op.Source)

	if new(big.Int).Sub(balance, pay.op.Amount).Cmp(big0) < 0 {
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
