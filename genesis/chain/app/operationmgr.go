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

	at "github.com/dappledger/AnnChain/angine/types"
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/types"
	"go.uber.org/zap"
)

type OperationManager struct {
	DB  types.OperationDBItfc
	app *GenesisApp
}

func (m *OperationManager) Init(db types.OperationDBItfc, app *GenesisApp) {
	m.DB = db
	m.app = app
}

func (m *OperationManager) PreCheck(tx *types.Transaction) at.Result {
	_, err := m.NewOperator(tx)
	if err != nil {
		return at.NewError(at.CodeType_BaseInvalidInput, err.Error())
	}
	return at.NewResultOK([]byte{}, "")
}

func (m *OperationManager) ExecTx(stateDup *stateDup, tx *types.Transaction) error {

	operator, err := m.NewOperator(tx)
	if err != nil {
		return err
	}

	if err = operator.CheckValid(stateDup); err != nil {
		logger.Debug("[operation],checkvalid err", zap.String("type:", tx.GetOpName()), zap.String("err:", err.Error()))
		return err
	}

	if err = operator.Apply(stateDup); err != nil {
		logger.Debug("[operation],apply err", zap.String("type:", tx.GetOpName()), zap.String("err:", err.Error()))
		return err
	}

	return nil
}

type DoOperatorItfc interface {
	PreCheck() at.Result
	CheckValid(*stateDup) error
	Apply(*stateDup) error
}

func (m *OperationManager) NewOperator(tx *types.Transaction) (DoOperatorItfc, error) {

	switch tx.GetOpName() {

	case types.OP_S_CREATEACCOUNT.OpStr():

		opCreateAccount := new(types.CreateAccount)

		if err := json.Unmarshal(tx.GetOperation(), opCreateAccount); err != nil {
			return nil, err
		}
		startBalance, err := big.NewInt(0).SetString(opCreateAccount.StartingBalance, 10)
		if err != nil {
			return nil, err
		}

		createAccountOp := &types.CreateAccountOp{
			TargetAddress: tx.GetTo(),
			StartBalance:  startBalance,
		}
		createAccountOp.BaseOp.Source = tx.GetFrom()
		createAccountOp.BaseOp.To = tx.GetTo()

		tx.SetOperatorItfc(createAccountOp)

		return &DoCreateAccount{
			tx:  tx,
			app: m.app,
			op:  createAccountOp,
		}, nil

	case types.OP_S_PAYMENT.OpStr():

		opPayment := new(types.Payment)

		if err := json.Unmarshal(tx.GetOperation(), opPayment); err != nil {
			return nil, err
		}

		amount, err := big.NewInt(0).SetString(opPayment.Amount, 10)
		if err != nil {
			return nil, err
		}

		paymentOp := &types.PaymentOp{
			Destination: tx.GetTo(),
			Amount:      amount,
		}
		paymentOp.BaseOp.Source = tx.GetFrom()

		tx.SetOperatorItfc(paymentOp)

		return &DoPayment{
			tx:  tx,
			app: m.app,
			op:  paymentOp,
		}, nil

	case types.OP_S_MANAGEDATA.OpStr():

		var opManageData types.ManageData

		if err := json.Unmarshal(tx.GetOperation(), &opManageData.Mparis); err != nil {
			return nil, err
		}

		manageData := new(types.ManageDataOp)

		for k, v := range opManageData.Mparis {
			manageData.DataName = append(manageData.DataName, k)
			manageData.Data = append(manageData.Data, v.Value)
			manageData.Category = append(manageData.Category, v.Category)
		}

		manageData.BaseOp.Source = tx.GetFrom()

		tx.SetOperatorItfc(manageData)

		return &DoManageData{
			app: m.app,
			op:  manageData,
			tx:  tx,
		}, nil

	case types.OP_S_CREATECONTRACT.OpStr():

		opCreateContract := new(types.CreateContract)

		if err := json.Unmarshal(tx.GetOperation(), opCreateContract); err != nil {
			return nil, err
		}

		createContractOp := &types.CreateContractOp{
			Price:    opCreateContract.Price,
			GasLimit: opCreateContract.GasLimit,
			Amount:   opCreateContract.Amount,
			Payload:  opCreateContract.Payload,
		}

		createContractOp.BaseOp.Source = tx.GetFrom()

		tx.SetOperatorItfc(createContractOp)

		return &DoCreateContract{
			tx:  tx,
			app: m.app,
			op:  createContractOp,
		}, nil

	case types.OP_S_EXECUTECONTRACT.OpStr():

		opExcuteContract := new(types.ExcuteContract)

		if err := json.Unmarshal(tx.GetOperation(), opExcuteContract); err != nil {
			return nil, err
		}

		excuteContractOp := &types.ExcuteContractOp{
			Price:    opExcuteContract.Price,
			GasLimit: opExcuteContract.GasLimit,
			Amount:   opExcuteContract.Amount,
			Payload:  opExcuteContract.Payload,
		}

		excuteContractOp.BaseOp.Source = tx.GetFrom()
		excuteContractOp.BaseOp.To = tx.GetTo()

		tx.SetOperatorItfc(excuteContractOp)

		return &DoExcuteContract{
			tx:  tx,
			app: m.app,
			op:  excuteContractOp,
		}, nil
	}
	return nil, fmt.Errorf("operation not exist")
}

func (m *OperationManager) CheckSignature(source *ethcmn.Address, sigaccs []ethcmn.Address) error {

	return nil
}
