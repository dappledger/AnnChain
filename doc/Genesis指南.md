## AnnChain.Genesis链指南

### AnnChain网络概述

AnnChain.Genesis是底层为链表式结构的区块链系统，致力于快速创建一条高性能、可扩展、可自由配置的区块链基础设施，让创建去中心化应用更加简单快速，并在技术社区中提供了多种实际应用场景。

### 创建账户

你在Genesis网络上执行任何一种操作的前提是创建账户，区块链中的所有资产、资源都将与账户密切联系。每一个账户都是一对公钥与私钥，Genesis使用私钥为每一笔交易加密，保证交易的安全性，公钥通常可以安全共享，其他人通过公钥地址来识别你，通过验证私钥加密来验证你是否为这笔交易授权。你永远都不应该与任何人共享私钥，一旦私钥泄露那么你账户下的所有资源都将变的不安全。

公私钥使用secp256k1非对称加密算法生成，Golang SDK中生成公私钥对代码展示：

```
func GenerateKey() (privekey, address string) {

	privkey, err := crypto.GenerateKey()
	if err != nil {
		return "", ""
	}

	privekey = ethcmn.Bytes2Hex(crypto.FromECDSA(privkey))

	address = crypto.PubkeyToAddress(privkey.PublicKey).Hex()
}
```

该函数调用核心算法生成两个字符串，分别是私钥、公钥。此时生成的账户为离线状态，其信息并没有在区块链上，因此需要进一步做创建账户的激活动作才能在区块链上作为可操作用户。

在SDK中展现为执行如下操作：

```
var superAddr string = "0x65188459a1dc65984a0c7d4a397ed3986ed0c853"
var superPriv string = "7cb4880c2d4863f88134fd01a250ef6633cc5e01aeba4c862bedbf883a148ba8"

client := NewAnnChainClient("tcp://127.0.0.1:46657") //链接到区块链rpc地址

client.CreateAccount(GetNonce(superAddr), superPriv, "100", "memo", superAddr, account, "1000") //组织创建账户交易
```

Genesis链中的资产在链启动时一次性发放到超级账户中，代码中默认的超级账户的公私钥对如上代码所示，开发者可在本地安装后测试使用。

### 创建发送交易

针对账户在区块链上的每一种操作都可视为一笔交易。

1. **构造交易**

   Genesis链以transaction结构体的方式封装交易具体内容，结构体使用rlp编码为二进制流发送到区块链网络，比如下面一笔转账交易：

   ```
   {
   	"basefee": "100",
   	"nonce": "2",
   	"memo": "payment",
   	"optype": "payment",
   	"form": "1743fe87b946d8aa4f5fd3094386c4cafc98c31c",
   	"to": "1743fe87b946d8aa4f5fd3094386c4cafc98c31c",
   	"operation":
   				{
   					"amount": "100"
   				},
   	"signature": ""
   }
   ```

   结构体按照顺序（顺序不可乱）、字段类型填写相应的参数，然后将basefee~operation之间的字段经过rlp编码生成二进制数据，然后对该二进制流使用keccack sha-3散列算法计算其hash值，再使用用户的私钥对该hash加签，加签后的数据进行16进制编码，放到字符串字段signature中，这样一笔转账交易构造完成。

2. **发送交易**

   Genesis链使用Json-RPC协议进行数据交互，通过构造交易生成的数据流按照 Json-Rpc协议预定义的调用方式发送到区块链网络，例如上面的转账交易可以通过调用curl命令，调试接口：

   ```
   Curl -d '{"jsonrpc":"2.0","method":"payment","params":[""],"id":"1"}' -H "Content-Type:application/json" -X POST "http://localhost:46657"
   ```

   详情请查看API接口文档。

### 智能合约

在区块链技术领域，智能合约是指基于预定事件触发、不可篡改、自动执行的计算机程序。

Genesis链提供基于Solidity 语言开发的图灵完备的智能合约系统，为保证合约的可扩展、可移植性，我们沿用了以太坊的EVM系统，对于合约编程可参考 [solidity官方文档](https://solidity.readthedocs.io/en/v0.5.2/)。

如果对以太坊的EVM熟练的开发者，那么对我们Genesis链提供的智能合约系统将同样的驾轻就熟，针对智能合约的相关操作我们提供了一下方法：

创建合约

执行合约

查询合约接口

查询合约是否存在

查询合约bytecode

查询合约codehash

查询合约票据

具体函数使用可参考API文档，通过上面提供的合约方法，为开发者提供方便、完整的合约开发使用功能。

### 数据实体