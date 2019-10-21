# AnnChain Command Tool

This document is about AnnChain's usage. It does not involve AnnChain installation, node deployment and start-up. 
This note uses gtool to deploy and execute smart contracts, also can use the Go/Java SDK to perform contract-related operations.

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
Privkey for user: //account's private key
```

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
Privkey for user: //account's private key
```


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
Privkey for user: //account's private key
```

##### Result

```
result：value
```

##### Demo

```
./build/gtool --backend "tcp://127.0.0.1:46657" evm read --abif ./scripts/examples/evm/sample.abi --callf ./scripts/examples/evm/sample_read.json
Privkey for user : 
C579D84396CC7D425AFD5ED700140ECA3A0EF9D7E6FB007C4C09CBDE0359D6AF
result: 100
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

## New Node Synchronous Chain's Data

If a new node wants to join the chain, it needs to synchronize the chain data in the first. The details are as follows:

#### Initialize a new node

##### Command

```
./build/genesis init
```

##### Result

```
Log dir is:  ./
Initialized chain_id: genesis-SyaIbH genesis_file: /root/.genesis/genesis.json priv_validator: /root/.genesis/priv_validator.json
Check the files generated, make sure everything is OK.
```

The node's information can be found from `priv_validator.json`

- The node's private key is `C25E861FC9083455FB1D47CDB1DCBC49597370C4A7E014C07D6D8E7BF9849F95FB192BF3F6D8B2DD2FA8CAE2F5E9B6B64597CD88BC7778820646B68A7E9D02F9`
- The node's public key is `FB192BF3F6D8B2DD2FA8CAE2F5E9B6B64597CD88BC7778820646B68A7E9D02F9`

#### Update `config.toml`

- `seeds` 

  Filling in the  `p2p_laddr` of the nodes which are validators in the chain in order to get the current state of the chain.

  ##### Demo

  ```
  seeds = "0.0.0.0:46000,0.0.0.0:46001,0.0.0.0:46002,0.0.0.0:46003"
  ```

- `signbyca`

  If `auth_by_ca`is true, CA node's signature must be filled in `signbyca`. The details are as follows:

  ##### Command

  ```
  ./build/gtool sign --pub <new node's public key>
  Node Privkey for user://CA node's private key
  ```

  ##### Result

  ```
  <new node's public key>:<CA node's signature>
  ```

  ##### Demo

  ```
  ./build/gtool sign --pub FB192BF3F6D8B2DD2FA8CAE2F5E9B6B64597CD88BC7778820646B68A7E9D02F9
  Node Privkey for user:
  2948184E586A8079538D6B033388CA094507D1339157AD397F687B17D327C237A05B3A182A3006024DD632823EA37F5B3742286AC759767DBD7B422C60175810
  FB192BF3F6D8B2DD2FA8CAE2F5E9B6B64597CD88BC7778820646B68A7E9D02F9 : ca4b244dbeadd0418434a42d8f4cc19570f0a5f091213460a2597601cd8f25d25daca5254dcd1b7599da3bacb3df9c73b91c34e3ac9ebd58fddf32fa18398805
  ```

  ```
  signbyca = "ca4b244dbeadd0418434a42d8f4cc19570f0a5f091213460a2597601cd8f25d25daca5254dcd1b7599da3bacb3df9c73b91c34e3ac9ebd58fddf32fa18398805"	
  ```

#### Update `genesis.json`

`genesis.json` should be replaced by CA nodes' `genesis.json` .Then run the node:

```
./build/genesis run
```

## Add Validator Node

If the new node wants to vote, It must be a validator node. The details are as follows:

##### Command

```
./build/gtool admin add_peer --nPrivs <the number of CA nodes which needed to validate the behavior>
Input Privkey of addnode  for user: //new node's private key
Now fetch CA-Node;need n private keys; please input n' keys: //CA nodes' private keys，n is the number of CA nodes which needed to validate the behavior
```

##### Result

```
hash //tx hash
```

##### Demo

```
./build/gtool admin add_peer --nPrivs 1
Input Privkey of addnode  for user:
C25E861FC9083455FB1D47CDB1DCBC49597370C4A7E014C07D6D8E7BF9849F95FB192BF3F6D8B2DD2FA8CAE2F5E9B6B64597CD88BC7778820646B68A7E9D02F9
Now fetch CA-Node;need 1 private keys; please input 1' keys:
2948184E586A8079538D6B033388CA094507D1339157AD397F687B17D327C237A05B3A182A3006024DD632823EA37F5B3742286AC759767DBD7B422C60175810
hash= 0x288f30b4e5904b2cddf3d157bb7a4820229c947bf0ee00c51019f136071d8e19
```

## Update Validator Node's Voting Power

If the new node wants to update voting power, It must be a validator node in the first. The initial power is 0.

##### Command

```
./build/gtool admin change_node --validator_pubkey <new node's public key> --power <new voting power> --nPrivs <the number of CA nodes which needed to validate the behavior>
need n private keys; please input n' keys: //CA nodes' private keys，n is the number of CA nodes which needed to validate the behavior
```

##### Result

```
hash //tx hash
```

##### Demo

```
./build/gtool admin change_node --validator_pubkey FB192BF3F6D8B2DD2FA8CAE2F5E9B6B64597CD88BC7778820646B68A7E9D02F9 --power 20 --nPrivs 1
need 1 private keys; please input 1' keys:
2948184E586A8079538D6B033388CA094507D1339157AD397F687B17D327C237A05B3A182A3006024DD632823EA37F5B3742286AC759767DBD7B422C60175810
hash= 0x110ee98edba177ede906e5d8175d9f787bfed61f3bb841537327d6f8128c6dbe
```

## Remove Node

```
./build/gtool admin remove_node --validator_pubkey <the public key of the node which needed to be removed> --nPrivs <the number of CA nodes which needed to validate the behavior>
need n private keys; please input n' keys //CA nodes' private keys，n is the number of CA nodes which needed to validate the behavior
```

##### Result

```
hash //tx hash
```

##### Demo

```
./build/gtool admin remove_node --validator_pubkey FB192BF3F6D8B2DD2FA8CAE2F5E9B6B64597CD88BC7778820646B68A7E9D02F9 --nPrivs 1
need 1 private keys; please input 1' keys:
2948184E586A8079538D6B033388CA094507D1339157AD397F687B17D327C237A05B3A182A3006024DD632823EA37F5B3742286AC759767DBD7B422C60175810
hash= 0x05248b13cda3b0feb40a9ec47d27e43c5a4648278e1473125c33089ad24d1d3d
```

