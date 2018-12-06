package app

import (
	at "github.com/dappledger/AnnChain/angine/types"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
)

type DoManageData struct {
	app *DelosApp
	op  *types.ManageDataOp
	tx  *types.Transaction
}

func (ct *DoManageData) PreCheck() at.Result {
	return at.NewResultOK([]byte{}, "")
}

func (ct *DoManageData) CheckValid(stateDup *stateDup) error {
	if ct.op.Source == types.ZERO_ADDRESS {
		return at.NewError(at.CodeType_BaseUnknownAddress, at.CodeType_BaseUnknownAddress.String())
	}
	return nil
}

func (ct *DoManageData) Apply(stateDup *stateDup) error {

	db := ct.app.dataM

	if len(ct.op.DataName) == 0 {
		return at.NewError(at.CodeType_InvalidTx, at.CodeType_InvalidTx.String())
	}

	toAdd := make(map[string]string)

	for idx, r := range ct.op.DataName {
		if err := db.QueryOneAccData(ct.op.Source, r); err != nil {
			return at.NewError(at.CodeType_BaseInvalidInput, at.CodeType_BaseInvalidInput.String())
		}
		toAdd[r] = ct.op.Data[idx]
	}

	for k, v := range toAdd {
		_, err := db.AddAccData(ct.op.Source, k, v)
		if err != nil {
			return at.NewError(at.CodeType_SaveFailed, at.CodeType_SaveFailed.String()+err.Error())
		}
		ct.SetEffects(ct.op.Source, k, v)
	}
	return nil
}

func (ct *DoManageData) SetEffects(source ethcmn.Address, dataName, data string) {
	ct.op.SetEffects(&types.ActionManageData{
		ActionBase: types.ActionBase{
			Typei:       types.OP_S_MANAGEDATA.OpInt(),
			Type:        types.OP_S_MANAGEDATA,
			FromAccount: source,
			Nonce:       ct.tx.Nonce(),
			BaseFee:     ct.tx.BaseFee(),
			Memo:        ct.tx.Data.Memo,
		},
		Name:   dataName,
		Value:  data,
		Source: source,
	}, []types.EffectObject{})
	return
}
