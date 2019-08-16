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

package core

import (
	"math/big"

	"github.com/dappledger/AnnChain/eth/common"
)

var (
	AdminTo   = common.HexToAddress("0x02000000") //contract addr;
	AdminCode = common.Hex2Bytes("60806040526004361061003b576000357c010000000000000000000000000000000000000000000000000000000090048063ba9c716e14610040575b600080fd5b34801561004c57600080fd5b506101066004803603602081101561006357600080fd5b810190808035906020019064010000000081111561008057600080fd5b82018360208201111561009257600080fd5b803590602001918460018302840111640100000000831117156100b457600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050610108565b005b60603382604051602001808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166c0100000000000000000000000002815260140182805190602001908083835b6020831015156101855780518252602082019150602081019050602083039250610160565b6001836020036101000a0380198251168184511680821785525050505050509050019250505060405160208183030381529060405290506000602082510190506000808284600060fe600019f18015156101de57600080fd5b5050505056fea165627a7a72305820de66ac7f0ed7f9f6d566ddabc3faa2fb4abe3b21b05fd4bbb8b178cac78e7d680029")
)

const (
	AdminMethod = "changenode"
	AdminABI    = `[
		{
			"constant": false,
			"inputs": [
				{
					"name": "txdata",
					"type": "bytes"
				}
			],
			"name": "changenode",
			"outputs": [],
			"payable": false,
			"stateMutability": "nonpayable",
			"type": "function"
		}
	]`
)

/*
//admin.sol
pragma solidity ^0.5.1;

contract Admin {
    function changenode(bytes memory txdata) public {

        bytes memory data = abi.encodePacked(msg.sender, txdata);
        uint size = data.length + 32;

        assembly {
            let success := call(not(0), 0xfe, 0, data, size, 0, 0)
            if iszero(success){
                revert(0, 0)
            }
        }
    }
}
*/

func DefaultGenesis() Genesis {
	//set allocator;
	galloc := GenesisAlloc{
		AdminTo: {
			Code:    AdminCode,
			Balance: big.NewInt(0),
		},
	}

	g := Genesis{Alloc: galloc}
	return g
}
