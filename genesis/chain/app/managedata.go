package app

import (
	at "github.com/dappledger/AnnChain/angine/types"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
)

type DoManageData struct {
	app *GenesisApp
	op  *types.ManageDataOp
	tx  *types.Transaction
}

type ValueCategory struct {
	value    string
	category string
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

	toAdd := make(map[string]ValueCategory)

	for idx, r := range ct.op.DataName {
		if err := db.QueryOneAccData(ct.op.Source, r); err != nil {
			return at.NewError(at.CodeType_BaseInvalidInput, err.Error())
		}
		var vc ValueCategory
		vc.value = ct.op.Data[idx]
		vc.category = ct.op.Category[idx]
		toAdd[r] = vc
	}

	for k, v := range toAdd {
		_, err := db.AddAccData(ct.op.Source, k, v.value, v.category)
		if err != nil {
			return at.NewError(at.CodeType_SaveFailed, at.CodeType_SaveFailed.String()+err.Error())
		}
		ct.SetEffects(ct.op.Source, k, v.value, v.category)
	}
	return nil
}

func (ct *DoManageData) SetEffects(source ethcmn.Address, dataName, data, category string) {
	ct.op.SetEffects(&types.ActionManageData{
		ActionBase: types.ActionBase{
			Typei:       types.OP_S_MANAGEDATA.OpInt(),
			Type:        types.OP_S_MANAGEDATA,
			FromAccount: source,
			Nonce:       ct.tx.Nonce(),
			BaseFee:     ct.tx.BaseFee(),
			Memo:        ct.tx.Data.Memo,
		},
		Name:     dataName,
		Value:    data,
		Category: category,
		Source:   source,
	}, []types.EffectObject{})
	return
}
