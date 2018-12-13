package app

import (
	"encoding/json"
	"math/big"

	"github.com/dappledger/AnnChain/genesis/eth/rlp"

	at "github.com/dappledger/AnnChain/angine/types"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
)

func makeResultData(i interface{}) at.Result {
	jdata, err := json.Marshal(i)
	if err != nil {
		return at.NewError(at.CodeType_InternalError, err.Error())
	}
	return at.NewResultOK(jdata, "")
}

func makeRpcResultData(i interface{}) at.NewRPCResult {
	return at.NewRpcResultOK(i, "")
}

func wrapEffectResultData(effects []types.EffectData) at.Result {
	if len(effects) == 0 {
		return makeResultData(effects)
	}
	mres := make([]map[string]interface{}, len(effects))
	for i, v := range effects {
		mres[i] = make(map[string]interface{})
		err := json.Unmarshal([]byte(v.JSONData), &mres[i])
		if err != nil {
			return at.NewError(at.CodeType_InternalError, err.Error())
		}
		mres[i]["id"] = v.EffectID
	}

	return makeResultData(mres)
}

func wrapActionResultData(actions []types.ActionData) at.NewRPCResult {
	if len(actions) == 0 {
		return at.NewRpcResultOK(actions, "")
	}
	mres := make([]map[string]interface{}, len(actions))
	for i, v := range actions {
		mres[i] = make(map[string]interface{})
		err := json.Unmarshal([]byte(v.JSONData), &mres[i])
		if err != nil {
			return at.NewRpcError(at.CodeType_InternalError, err.Error())
		}
		mres[i]["id"] = v.ActionID
	}

	return at.NewRpcResultOK(mres, "")
}

func (app *GenesisApp) queryDoContract(bs []byte) at.NewRPCResult {

	var err error

	dupState := app.state.DeepCopy()

	tx := &types.Transaction{}

	err = rlp.DecodeBytes(bs, tx)
	if err != nil {
		return at.NewRpcError(at.CodeType_WrongRLP, err.Error())
	}

	if tx.SignString() != "" {
		if err = tx.CheckSig(); err != nil {
			return at.NewRpcError(at.CodeType_BaseInvalidSignature, err.Error())
		}
	}

	if hashBytes := dupState.GetCodeHash(tx.GetTo()); len(hashBytes) != ethcmn.HashLength || ethcmn.EmptyHash(hashBytes) {
		return at.NewRpcError(at.CodeType_BaseUnknownAddress, "contract address not exist")
	}

	var queryPayload types.QueryContract

	if err = json.Unmarshal(tx.GetOperation(), &queryPayload); err != nil {
		return at.NewRpcError(at.CodeType_InvalidTx, err.Error())
	}

	payload, err := ethcmn.HexToByte(queryPayload.Payload)
	if err != nil {
		return at.NewRpcError(at.CodeType_InvalidTx, err.Error())
	}

	etx := NewContractTransaction(dupState.GetNonce(tx.GetFrom()), tx.GetFrom(), tx.GetTo(), big.NewInt(0), ethcmn.MaxBig, big.NewInt(0), payload)

	res, _, err := QueryContractExcute(dupState, etx)
	if err != nil {
		return at.NewRpcError(at.CodeType_InvalidTx, err.Error())
	}

	if res == nil {
		return at.NewRpcError(at.CodeType_NullData, "No Data!")
	}

	return at.NewRpcResultOK(ethcmn.Bytes2Hex(res), "")
}

func (app *GenesisApp) queryAllLedgers(cursor, limit uint64, order string) at.NewRPCResult {
	res, err := app.dataM.QueryAllLedgerHeaderData(cursor, limit, order)
	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, err.Error())
	}

	return at.NewRpcResultOK(res, "")
}

func (app *GenesisApp) queryLedger(seq *big.Int) at.NewRPCResult {
	res, err := app.dataM.QueryLedgerHeaderData(seq)
	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, err.Error())
	}

	if res == nil {
		return at.NewRpcError(at.CodeType_NullData, "No Data!")
	}

	return at.NewRpcResultOK(res, "")
}

func (app *GenesisApp) queryPaymentsData(q types.ActionsQuery) at.NewRPCResult {
	res, err := app.dataM.QueryPaymentData(q)
	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, err.Error())
	}

	if res == nil {
		return at.NewRpcError(at.CodeType_NullData, "No Data!")
	}

	return wrapActionResultData(res)
}

func (app *GenesisApp) queryEffectsData(q types.EffectsQuery) at.Result {
	res, err := app.dataM.QueryEffectData(q)
	if err != nil {
		return at.NewError(at.CodeType_InternalError, err.Error())
	}

	return wrapEffectResultData(res)
}

func (app *GenesisApp) queryActionsData(q types.ActionsQuery) at.NewRPCResult {
	res, err := app.dataM.QueryActionData(q)
	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, err.Error())
	}

	if res == nil {
		return at.NewRpcError(at.CodeType_NullData, "No Data!")
	}

	return wrapActionResultData(res)
}

func (app *GenesisApp) queryAllTxs(cursor, limit uint64, order string) at.NewRPCResult {
	res, err := app.dataM.QueryAllTxs(cursor, limit, order)
	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, err.Error())
	}

	if res == nil {
		return at.NewRpcError(at.CodeType_NullData, "No Data!")
	}

	return at.NewRpcResultOK(res, "")
}

func (app *GenesisApp) queryTx(txhash ethcmn.Hash) at.Result {
	res, err := app.dataM.QuerySingleTx(&txhash)
	if err != nil {
		return at.NewError(at.CodeType_InternalError, err.Error())
	}

	if res == nil {
		return at.NewError(at.CodeType_InternalError, "Not exist")
	}

	data := []types.TransactionData{*res}

	return makeResultData(data)
}

func (app *GenesisApp) queryAccountTxs(addr ethcmn.Address, cursor, limit uint64, order string) at.NewRPCResult {
	res, err := app.dataM.QueryAccountTxs(&addr, cursor, limit, order)
	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, err.Error())
	}

	if res == nil {
		return at.NewRpcError(at.CodeType_NullData, "No Data!")
	}

	return at.NewRpcResultOK(res, "")
}

func (app *GenesisApp) queryHeightTxs(height string, cursor, limit uint64, order string) at.NewRPCResult {
	res, err := app.dataM.QueryHeightTxs(height, cursor, limit, order)
	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, err.Error())
	}

	if res == nil {
		return at.NewRpcError(at.CodeType_NullData, "No Data!")
	}

	return at.NewRpcResultOK(res, "")
}

func (app *GenesisApp) queryAccountManagedata(addr ethcmn.Address, category string, name string, cursor, limit uint64, order string) at.NewRPCResult {
	res, err := app.dataM.QueryAccountManagedata(addr, category, name, cursor, limit, order)

	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, err.Error())
	}

	if len(res) == 0 {
		return at.NewRpcError(at.CodeType_NullData, "No Data!")
	}

	return at.NewRpcResultOK(res, "")
}

func (app *GenesisApp) queryAccountSingleManageData(addr ethcmn.Address, keys string) at.NewRPCResult {
	res, err := app.dataM.QuerySingleManageData(addr, keys)

	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, err.Error())
	}

	if len(res) == 0 {
		return at.NewRpcError(at.CodeType_NullData, "No Data!")
	}
	return at.NewRpcResultOK(res, "")
}

func (app *GenesisApp) queryAccountCategoryManageData(addr ethcmn.Address, category string) at.NewRPCResult {
	res, err := app.dataM.QueryCategoryManageData(addr, category)

	if err != nil {
		return at.NewRpcError(at.CodeType_InternalError, err.Error())
	}

	if len(res) == 0 {
		return at.NewRpcError(at.CodeType_NullData, "No Data!")
	}
	return at.NewRpcResultOK(res, "")
}
