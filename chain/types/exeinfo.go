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

import ethcmn "github.com/dappledger/AnnChain/eth/common"

type TxExeInfo struct {
	TxHash  ethcmn.Hash
	Success bool
	Message string
	Fatal   error
}

func TxExeSuccess(txHash ethcmn.Hash) *TxExeInfo {
	return &TxExeInfo{
		TxHash:  txHash,
		Success: true,
	}
}

func TxExeFatal(txHash ethcmn.Hash, err error) *TxExeInfo {
	return &TxExeInfo{
		TxHash: txHash,
		Fatal:  err,
	}
}

func TxExeFailed(txHash ethcmn.Hash, message string) *TxExeInfo {
	return &TxExeInfo{
		TxHash:  txHash,
		Success: false,
		Message: message,
	}
}
