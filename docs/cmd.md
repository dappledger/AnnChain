# AnnChain Command Tool

This document is about AnnChain usage. It does not involve AnnChain installation, node deployment and start-up. 
This note uses gtool to deploy and execute smart contracts, and Go/Java SDK to perform contract-related operations.

## Create Account


```
./build/gtool account create

privkey: C579D84396CC7D425AFD5ED700140ECA3A0EF9D7E6FB007C4C09CBDE0359D6AF
address: 771403C283A3F46CDA462F7AEFF5DFD28B00F106
```

## Create Contract


Nodes need to be started before performing intelligent contract-related operations.
The default node has been started for the following actions

##### Command

```
gtool --backend <validator's IP:RPC Port> evm create --abif <abi filepath> --callf <input json filepath> --nonce <account nonce>
```

In addition to the above parameters, you need to input the private key of the validator node in the running chain when prompting "Privkey for user".

##### Result

```
contract address 
tx result 
```

##### Demo

```
cd AnnChain
./build/gtool --backend "tcp://127.0.0.1:46657" evm create --abif ./scripts/examples/evm/sample.abi --callf ./scripts/examples/evm/sample.json --nonce 0
Privkey for user : 
C579D84396CC7D425AFD5ED700140ECA3A0EF9D7E6FB007C4C09CBDE0359D6AF
contract address: 0xAe119075bd77dE2d8e32629bdb439D967A1EcFe6									
tx result: 0x3121cda109485a5478cb5ff227f8699dd6fa76a69869cc42a12b1b32b9c4b885
```

sample.abi 

```abi
[
        {
                "constant": false,
                "inputs": [
                        {
                                "name": "Id",
                                "type": "uint256"
                        },
                        {
                                "name": "Amount",
                                "type": "uint256"
                        }
                ],
                "name": "createCheckInfos",
                "outputs": [],
                "payable": false,
                "stateMutability": "nonpayable",
                "type": "function"
        },
        {
                "constant": true,
                "inputs": [
                        {
                                "name": "Id",
                                "type": "uint256"
                        }
                ],
                "name": "getPremiumInfos",
                "outputs": [
                        {
                                "name": "",
                                "type": "uint256"
                        }
                ],
                "payable": false,
                "stateMutability": "view",
                "type": "function"
        },
        {
                "anonymous": false,
                "inputs": [
                        {
                                "indexed": false,
                                "name": "Id",
                                "type": "uint256"
                        },
                        {
                                "indexed": false,
                                "name": "Amount",
                                "type": "uint256"
                        }
                ],
                "name": "InputLog",
                "type": "event"
        }
]
```

sample.json  

```json
{
  "bytecode" : "6060604052341561000f57600080fd5b6101818061001e6000396000f30060606040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063a6226f2114610051578063b051c1e01461007d575b600080fd5b341561005c57600080fd5b61007b60048080359060200190919080359060200190919050506100b4565b005b341561008857600080fd5b61009e6004808035906020019091905050610136565b6040518082815260200191505060405180910390f35b60007fb45ab3e8c50935ce2fa51d37817fd16e7358a3087fd93a9ac7fbddb22a926c358383604051808381526020018281526020019250505060405180910390a1828160000181905550818160010181905550806000808581526020019081526020016000206000820154816000015560018201548160010155905050505050565b60008060008381526020019081526020016000206001015490509190505600a165627a7a723058207eaf119132cfc4008c97339b874c4c16d20d27a72875e55a6a22a29fee30876d0029",																										
  "params" :[]																					 
}
```

| parameter     | description         |
| -------- | ------------ |
| bytecode | Contract bytecode |
| params   | Input parameter of calling function      |

## Execute Contract

##### Command

```
gtool --backend <validator's IP:RPC Port> evm call --abif <abi filepath> --callf <input json filepath> --nonce <account nonce>
```

In addition to the above parameters, you need to input the private key of the validator node in the running chain when prompting "Privkey for user".


##### Result

```
tx result 
```

##### Demo

```
./build/gtool --backend "tcp://127.0.0.1:46657" evm call --abif ./scripts/examples/evm/sample.abi --callf ./scripts/examples/evm/sample_execute.json --nonce 1
Privkey for user : 
C579D84396CC7D425AFD5ED700140ECA3A0EF9D7E6FB007C4C09CBDE0359D6AF
tx result: 0x2b41d9c05a7be5b85586c53b5a2d3cacc1ded323a18f1c62c51bc2aea0953b55
```

sample_execute.json  
```json
{
  "contract" : "0xAe119075bd77dE2d8e32629bdb439D967A1EcFe6",		
  "function" : "createCheckInfos",															
  "params":[																									
    1, 100
  ]
}
```

| parameter     | description     |
| -------- | -------- |
| contract | Contract address |
| function | Which contract's function do you want to call |
| params   | Input parameter of calling function |

## Read Contract

##### Command

```
gtool --backend <validator's IP:RPC Port> evm read --abif <abi filepath> --callf <input json filepath> 
```

In addition to the above parameters, you need to input the private key of the validator node in the running chain when prompting "Privkey for user".

##### Result

```
parse result {type , value}
```

##### Demo

```
./build/gtool --backend "tcp://127.0.0.1:46657" evm read --abif ./scripts/examples/evm/sample.abi --callf ./scripts/examples/evm/sample_read.json
Privkey for user : 
C579D84396CC7D425AFD5ED700140ECA3A0EF9D7E6FB007C4C09CBDE0359D6AF
parse result: *big.Int 100
```

sample_read.json  
```json
{
  "contract" : "0xAe119075bd77dE2d8e32629bdb439D967A1EcFe6",	
  "function" : "getPremiumInfos",															
  "params":[																									
    1
  ]
}
```

| parameter     | description     |
| -------- | -------- |
| contract | Contract address |
| function | Which contract's function do you want to call |
| params   | Input parameter of calling function |

## Query Nonce

##### Command

```
gtool --backend <validator's IP:RPC Port>  query nonce --address <account address>
```

##### Result

```
query result nonce
```

##### Demo

```
./build/gtool --backend "tcp://127.0.0.1:46657" query nonce --address 771403c283a3f46cda462f7aeff5dfd28b00f106
query result: 2
```

## Query Receipt

##### Commadn

```
gtool --backend <validator's IP:RPC Port> query receipt --hash <tx hash>
```

##### Result

```
query result receipt
```

##### Demo

```
./build/gtool --backend "tcp://127.0.0.1:46657" query receipt --hash 0x2b41d9c05a7be5b85586c53b5a2d3cacc1ded323a18f1c62c51bc2aea0953b55
query result: {"root":null,"status":1,"cumulativeGasUsed":21656,"logsBloom":"0x00000000000000000000000000800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000020000000000000000000000000000000000000000000000000000000000000","logs":[{"address":"0xae119075bd77de2d8e32629bdb439d967a1ecfe6","topics":["0xb45ab3e8c50935ce2fa51d37817fd16e7358a3087fd93a9ac7fbddb22a926c35"],"data":"0x00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000064","blockNumber":"0x64e","transactionHash":"0x2b41d9c05a7be5b85586c53b5a2d3cacc1ded323a18f1c62c51bc2aea0953b55","transactionIndex":"0x0","blockHash":"0x000000000000000000000000ec83a146ca731fdffe4bef69ad260d7389732e87","logIndex":"0x0","removed":false}],"transactionHash":"0x2b41d9c05a7be5b85586c53b5a2d3cacc1ded323a18f1c62c51bc2aea0953b55","contractAddress":"0x0000000000000000000000000000000000000000","gasUsed":21656}
```
