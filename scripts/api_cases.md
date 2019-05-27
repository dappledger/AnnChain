
### 1. create

API: `/v1/accounts/create`
Method: `GET`

Result:

```json
{
	"isSuccess": true,
	"result": {
	"privkey": "1a0feb4332e9dad4ca3b823f8e29c500114e2417e393dd1d7cde3058fb2ffed9",
	"address": "0xe9351b7ae49250db4f5cc68e8118bb7e770cf439"
	}
}
```

### 2. nonces

API: `/v1/nonces/:address`
Method: `GET`

Command:

```shell
curl 127.0.0.1:8889/v1/nonces/0xe9351b7ae49250db4f5cc68e8118bb7e770cf439
```

Result:

```json
{
  "isSuccess": true,
  "result": "0"
}
```


### 3. receipt

API: `/v1/receipt/:tx`
Method: `GET`

Command:

```shell
curl 127.0.0.1:8889/v1/receipt/0x6ba71f04b8b92f2c3d621b5f1078e3899712690f077db9e9b32363edc33e8e78
```

Result:

```json
{
	"isSuccess": true,
	"result": {
		"PostState": "",
		"CumulativeGasUsed": 2543328,
		"Bloom": "0x00000000000000008800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000002000000000000",
		"Logs": [
			{
				"address": "0xec66c781c74ef6b9c85ad534969bc1c30aa58ecd",
				"topics": [
					"0xe6d5659a0d290aa0da2625afddd1e9298b982e9c98a94bf9421a3d0cae6fa038"
				],
				"data": "0x000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000002bc00000000000000000000000000000000000000000000000000000000000000c8",
				"blockNumber": "0x28d138",
				"transactionIndex": "0x0",
				"transactionHash": "0x6ba71f04b8b92f2c3d621b5f1078e3899712690f077db9e9b32363edc33e8e78",
				"blockHash": "0x00000000000000000000000089cd831e34dfa8d246e7107e60df57350c192cd6",
				"logIndex": "0x0",
				"removed": false
			}
		],
		"TxHash": "0x6ba71f04b8b92f2c3d621b5f1078e3899712690f077db9e9b32363edc33e8e78",
		"ContractAddress": "0x0000000000000000000000000000000000000000",
		"GasUsed": 21848
	}
}
```


### 4. contract

1) create

API: `/v1/contract/create`
Method: `POST`

```json
{
 	"privkey" : "1a0feb4332e9dad4ca3b823f8e29c500114e2417e393dd1d7cde3058fb2ffed9",
 	"code" : "6060604052341561000f57600080fd5b6101d08061001e6000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680630aefd2a114610051578063af1d210014610088575b600080fd5b341561005c57600080fd5b61007260048080359060200190919050506100bd565b6040518082815260200191505060405180910390f35b341561009357600080fd5b6100bb60048080359060200190919080359060200190919080359060200190919050506100dc565b005b6000806000838152602001908152602001600020600301549050919050565b60007fe6d5659a0d290aa0da2625afddd1e9298b982e9c98a94bf9421a3d0cae6fa03884848460405180848152602001838152602001828152602001935050505060405180910390a18183101561013257600080fd5b838160020181905550828160000181905550818160010181905550806001015481600001540381600301819055508060008086815260200190815260200160002060008201548160000155600182015481600101556002820154816002015560038201548160030155905050505050505600a165627a7a72305820552c2fdd443c6883ec802909dbd6db10f12338d31b099ee8fa489517f9d1866d0029",
 	"abiDefinition" :"[{\"constant\":true,\"inputs\":[{\"name\":\"companyId\",\"type\":\"uint256\"}],\"name\":\"getNetprofitInfos\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"companyId\",\"type\":\"uint256\"},{\"name\":\"totalProfit\",\"type\":\"uint256\"},{\"name\":\"incomeTax\",\"type\":\"uint256\"}],\"name\":\"createNetprofitInfos\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"nameId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"totalProfit\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"incomeTax\",\"type\":\"uint256\"}],\"name\":\"InputLog\",\"type\":\"event\"}]", 
 	"params" :[]
 }

```

Result:

```json
{
	"isSuccess": true,
	"result": {
	"contract": "0xec66c781c74ef6b9c85ad534969bc1c30aa58ecd",
	"tx": "0xd72be22d8ca7f1778ec4c2a33023a220d191f604b397604d966ad57514c566ec"
	}
}
```

2) call

API: `/v1/contract/call`
Method: `POST`

```json
{
	"contract" : "0xec66c781c74ef6b9c85ad534969bc1c30aa58ecd",
  	"privkey" : "1a0feb4332e9dad4ca3b823f8e29c500114e2417e393dd1d7cde3058fb2ffed9",
  	"abiDefinition" : "[{\"constant\":true,\"inputs\":[{\"name\":\"companyId\",\"type\":\"uint256\"}],\"name\":\"getNetprofitInfos\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"companyId\",\"type\":\"uint256\"},{\"name\":\"totalProfit\",\"type\":\"uint256\"},{\"name\":\"incomeTax\",\"type\":\"uint256\"}],\"name\":\"createNetprofitInfos\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"nameId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"totalProfit\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"incomeTax\",\"type\":\"uint256\"}],\"name\":\"InputLog\",\"type\":\"event\"}]",
  	"method" : "createNetprofitInfos",
  	"params":[
		1, 700, 200
  ]
}
```

Result:

```json
{
	"isSuccess": true,
	"result": "0x6ba71f04b8b92f2c3d621b5f1078e3899712690f077db9e9b32363edc33e8e78"
}
```

3) read

API: `/v1/contract/read`
Method: `POST`

```json
{
	"contract" : "0xec66c781c74ef6b9c85ad534969bc1c30aa58ecd",
    "privkey" : "1a0feb4332e9dad4ca3b823f8e29c500114e2417e393dd1d7cde3058fb2ffed9",
    "abiDefinition" : "[{\"constant\":true,\"inputs\":[{\"name\":\"companyId\",\"type\":\"uint256\"}],\"name\":\"getNetprofitInfos\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"companyId\",\"type\":\"uint256\"},{\"name\":\"totalProfit\",\"type\":\"uint256\"},{\"name\":\"incomeTax\",\"type\":\"uint256\"}],\"name\":\"createNetprofitInfos\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"nameId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"totalProfit\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"incomeTax\",\"type\":\"uint256\"}],\"name\":\"InputLog\",\"type\":\"event\"}]",
  	"method" : "getNetprofitInfos",
 	"params":[
  		1
  	]
}
```

Result:

```json
{
	"isSuccess": true,
	"result": 500
}
```





