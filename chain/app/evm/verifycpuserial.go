// Copyright © 2017 ZhongAn Technology
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

package evm

import (
	etypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
)

func exeWithCPUSerialVeirfy(txs gtypes.Txs, beginExec BeginExecFunc) {
	for i, raw := range txs {
		txbs := gtypes.Tx(txs[i])
		exec, end := beginExec()
		err := execTx(txbs, exec, i, raw)
		end(raw, err)
	}
}

func execTx(atx gtypes.Tx, exec ExecFunc, index int, raw gtypes.Tx) error {
	var tx *etypes.Transaction
	if len(atx) > 0 {
		tx = new(etypes.Transaction)
		if err := rlp.DecodeBytes(atx, tx); err != nil {
			return err
		}
		if err := exec(index, raw, tx); err != nil {
			return err
		}
	}
	return nil
}
