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

package types

import (
	"fmt"
	"math/big"

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/eth/rlp"
)

type OperatorItfc interface {
	Type() TYPE_OP
	Bytes() []byte // no error ret
	GetOperationEffects() (ActionObject, []EffectObject)
	GetSourceAccount() ethcmn.Address
}

type BaseOp struct {
	Source ethcmn.Address
	To     ethcmn.Address
	act    ActionObject
	efts   []EffectObject
}

func (bp *BaseOp) GetOperationEffects() (ActionObject, []EffectObject) {
	return bp.act, bp.efts
}

func (bp *BaseOp) SetEffects(act ActionObject, effects []EffectObject) {
	bp.act = act
	bp.efts = effects
}

func (bp *BaseOp) GetSourceAccount() ethcmn.Address {
	return bp.Source
}

type CreateAccountOp struct {
	BaseOp
	TargetAddress ethcmn.Address
	StartBalance  *big.Int
}

func (cop *CreateAccountOp) Type() TYPE_OP {
	return OP_S_CREATEACCOUNT.OpInt()
}

func (cop *CreateAccountOp) Bytes() []byte {
	bys, err := rlp.EncodeToBytes(cop)
	if err != nil {
		fmt.Println(err)
	}
	return bys
}

type PaymentOp struct {
	BaseOp
	Destination ethcmn.Address
	Amount      *big.Int
}

func (pop *PaymentOp) Type() TYPE_OP {
	return OP_S_PAYMENT.OpInt()
}

func (pop *PaymentOp) Bytes() []byte {
	bys, err := rlp.EncodeToBytes(pop)
	if err != nil {
		fmt.Println(err)
	}
	return bys
}

type ManageDataOp struct {
	BaseOp
	DataName []string
	Data     []string
	Category []string
}

func (mop *ManageDataOp) Type() TYPE_OP {
	return OP_S_MANAGEDATA.OpInt()
}

func (mop *ManageDataOp) Bytes() []byte {
	bys, err := rlp.EncodeToBytes(mop)
	if err != nil {
		fmt.Println(err)
	}
	return bys

}

type CreateContractOp struct {
	BaseOp
	ContractAddr string
	Payload      string
	Price        string
	Amount       string
	GasLimit     string
	Params       []byte
}

func (ccontract *CreateContractOp) Type() TYPE_OP {
	return OP_S_CREATECONTRACT.OpInt()
}

func (ccontract *CreateContractOp) Bytes() []byte {
	bys, err := rlp.EncodeToBytes(ccontract)
	if err != nil {
		fmt.Println(err)
	}
	return bys
}

type ExcuteContractOp struct {
	BaseOp
	Payload  string
	Price    string
	Amount   string
	GasLimit string
	Gas      string
}

func (econtract *ExcuteContractOp) Type() TYPE_OP {
	return OP_S_CREATECONTRACT.OpInt()
}

func (econtract *ExcuteContractOp) Bytes() []byte {
	bys, err := rlp.EncodeToBytes(econtract)
	if err != nil {
		fmt.Println(err)
	}
	return bys
}
