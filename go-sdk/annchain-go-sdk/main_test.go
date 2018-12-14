package annchain

import (
	"encoding/hex"
	"testing"
	"time"
)

var abis string = `[
	{
		"constant": false,
		"inputs": [
			{
				"name": "min",
				"type": "uint256"
			},
			{
				"name": "max",
				"type": "uint256"
			}
		],
		"name": "GetRand",
		"outputs": [
			{
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "a",
				"type": "int256"
			},
			{
				"name": "b",
				"type": "int256"
			}
		],
		"name": "Add",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "GetResult",
		"outputs": [
			{
				"name": "",
				"type": "int256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "constructor"
	}
]`

var bytcode string = "608060405234801561001057600080fd5b50610224806100206000396000f3006080604052600436106100565763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166342bd1fd6811461005b578063830fb67c146100885780639a7d9af1146100a5575b600080fd5b34801561006757600080fd5b506100766004356024356100ba565b60408051918252519081900360200190f35b34801561009457600080fd5b506100a3600435602435610174565b005b3480156100b157600080fd5b5061007661017a565b6000808215156100cd576000915061016d565b6100dd428063ffffffff61018016565b604051602001808281526020019150506040516020818303038152906040526040518082805190602001908083835b6020831061012b5780518252601f19909201916020918201910161010c565b5181516020939093036101000a600019018019909116921691909117905260405192018290039091209350505084840390508181151561016757fe5b06840191505b5092915050565b01600055565b60005490565b818101828110156101f257604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601360248201527f536166654d61746820616464206661696c656400000000000000000000000000604482015290519081900360640190fd5b929150505600a165627a7a723058200888f8586d6f2d60d479c6ae0040be6f361480ff68add913ca005bb614be474c0029"

var superAddr string = "0x65188459a1dc65984a0c7d4a397ed3986ed0c853"
var superPriv string = "7cb4880c2d4863f88134fd01a250ef6633cc5e01aeba4c862bedbf883a148ba8"

var testAddr string = "0xc409aaf73698fdb5995c4d85f6033d5e90d2f2bd"
var testPriv string = "64d18eb7061dff419581c1af98201b76c7ab6db538b1cf65123c470ccc6d5929"

var (
	newAddr, newPriv string
	nonce            uint64
	hash             string

	contractAddress string
	contractReceipt string
	contractHash    string

	height uint64
)

func TestGen(t *testing.T) {

	newPriv, newAddr = GenerateKey()

	t.Log(newPriv, newAddr, len(newAddr))

}

func GetNonce(source string) uint64 {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	nonce, _, _ := client.QueryNonce(source)

	return nonce
}

//**************************Account TEST *************************************

func TestCreateAccount(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	chash, code, err := client.CreateAccount(GetNonce(superAddr), superPriv, "100", "memo", superAddr, testAddr, "1000")

	hash = chash

	t.Log(hash, code, err)

	time.Sleep(time.Second)
}
func TestQueryAccount(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryAccount(superAddr)

	t.Log(result, code, err)
}

//****************************************************************************

//**************************ManageData TEST***********************************

func TestManageData(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	var datas map[string]ManageDataValueParam

	datas = make(map[string]ManageDataValueParam)

	datas["8"] = ManageDataValueParam{Value: "zhaoyang", Category: "B"}
	datas["9"] = ManageDataValueParam{Value: "fanhongyue", Category: "B"}

	result, code, err := client.ManageData(GetNonce(superAddr), superPriv, "100", "memo", superAddr, datas)

	t.Log(result, code, err)

	hash = result

	time.Sleep(time.Second)
}

func TestQueryAccountManageDatas(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryAccountManageDatas(superAddr, "asc", 10, 0)

	t.Log(result, code, err)
}

func TestQueryAccountManageData(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryAccountManageData(superAddr, "3")

	t.Log(result, code, err)
}

func TestQueryAccountCategoryManageData(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryAccountCategoryManageData(superAddr, "B")

	t.Log(result, code, err)
}

//*********************************************************************************************

//************************************Ledger TEST *********************************************

func TestQueryTransactions(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryTransactions("desc", 10, 0)

	t.Log(result, code, err)
}

func TestQueryTransaction(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryTransaction(hash)

	t.Log(result, code, err)

	time.Sleep(time.Second)
}

func TestQueryAccountTransactions(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryAccountTransactions(superAddr, "asc", 10, 0)

	t.Log(result, code, err)
}

func TestQueryLedgerTransactions(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryLedgerTransactions(23, "asc", 10, 0)

	t.Log(result, code, err)
}

//*********************************************************************************************

//************************************Contract TEST********************************************

func TestCreateContract(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	param, err := NewCreateContractParam("1", "1000000000", "0", bytcode, abis, []interface{}{})
	if err != nil {
		t.Log(err)
		return
	}

	ccontractHash, code, err := client.CreateContract(GetNonce(superAddr), superPriv, "100", "", superAddr, param)

	contractHash = ccontractHash

	t.Log(contractHash, code, err)

	time.Sleep(time.Second)
}

func TestQueryCreateReceipt(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryReceipt(contractHash)

	contractAddress = result.ContractAddress

	t.Log(result, code, err)
}

func TestQueryContractExist(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryContractExist(contractAddress)

	t.Log(result, code, err)
}

func TestExecuteContract(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	param, err := NewExecuteContractParam("1", "10000000", "0", "GetRand", abis, []interface{}{1, 200})
	if err != nil {
		t.Log(err)
		return
	}

	result, code, err := client.ExcuteContract(GetNonce(superAddr), superPriv, "100", "", superAddr, contractAddress, param)

	contractHash = result

	t.Log(result, code, err)

	time.Sleep(time.Second)
}

func TestQueryExecuteReceipt(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryReceipt(contractHash)

	bResult, _ := hex.DecodeString(result.Result)

	outs, err := unpackResultToArray("GetRand", abis, bResult)

	t.Log(result, outs, code, err)
}

func TestQueryContract(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryContract(superPriv, superAddr, contractAddress, "GetResult", abis, []interface{}{})

	t.Log(result, code, err)
}

//*************************************************************************************************

//*************************************Payment TEST************************************************

func TestPayment(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.Payment(GetNonce(superAddr), superPriv, "0", "memo", superAddr, testAddr, "100")

	t.Log(result, code, err)

	hash = result

	time.Sleep(time.Second)
}

func TestQueryPayments(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryPayments("asc", 100, 0)

	t.Log(result, code, err)
}

func TestQueryAccountPayments(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryAccountPayments(superAddr, "asc", 100, 0)

	t.Log(result, code, err)
}

func TestQueryPayment(t *testing.T) {

	client := NewAnnChainClient("tcp://127.0.0.1:46657")

	result, code, err := client.QueryPayment(hash)

	t.Log(result, code, err)
}

//*************************************************************************************************
