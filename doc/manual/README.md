# annChain.Genesis区块链操作手册

## 前言

以下代码及操作命令是在CentOS7/Ubuntu操作系统上为例。

## 第一章 部署annChain.Genesis环境

本章节主要介绍annChain.Genesis区块链环境的部署。包括机器配置，部署软件环境和编译源码。

### 1.1机器配置

|   参数   | 最低配置 |                     推荐                     |
| :------: | :------: | :------------------------------------------: |
|   CPU    |  1.5GHz  |                    2.4GHz                    |
|   内存   |   1GB    |                     4GB                      |
|   核数   |   2核    |                     4核                      |
|   带宽   |   1Mb    |                     5Mb                      |
|   磁盘   |   100G   |                     500G                     |
| 操作系统 |          | CentOS（7以上 64位）或者Ubuntu（16.04 64位） |

### 1.2软件工具

- [ ] 版本管理工具Git；
- [ ] Golang环境配置；
- [ ] curl下载工具；
- [ ] Docker容器和Docker-compose工具；

### 1.3安装Git工具

Windows：

[download](https://git-scm.com/download/win)

Mac：

`brew install git`

CentOS:

`yum install -y git`

Ubuntu:

`apt-get install git`

### 1.4安装Goland

[download](https://plugins.jetbrains.com/go)

### 1.5安装并配置Golang环境

安装Golang

`brew install go  //Mac`

`apt-get install go //Ubuntu`

`yum install -y go //CentOS`

配置环境变量

`echo "export GOPATH=/root/goproject" >> ~/.bash_profile`

`echo "export GOROOT=/usr/local/go">> ~/.bash_profile`

`echo "export PATH=$GOROOT/bin:$PATH" >> ~/.bash_profile`

变量生效

`source ~/.bash_profile`

## 第二章 工程环境配置及编译

### 2.1克隆工程

```
git clone https://github.com/dappledger/AnnChain.git 
```

### 2.2编译工程

编译成功，在./AnnChain/bin目录中看到生成的二进制命令文件。

`./build.sh`

### 2.3编译环境变量

将编译成功后的命令目录配置到系统环境中，以便在后续操作过程的命令窗口调用。

`echo "export PATH=\$PATH:\$GOPATH/src/github.com/dappledger/AnnChain/bin/genesis" >> ~/.bash_profile`

`source ~/.bash_profile`

## 第三章 链单节点部署

### 3.1克隆源码

同2.1

### 3.2编译源码

同2.2

### 3.3初始化链节点

`./genesis init `

初始化目录结构

```
[root@bogon .genesis]# pwd
/root/.genesis
[root@bogon .genesis]# tree -L 2
.
├── config.toml
├── data
│   ├── archive
│   ├── archive.db
│   ├── blockstore.db
│   ├── chaindata
│   ├── cs.wal
│   ├── evm.db
│   ├── mempool.wal
│   ├── query_cache
│   ├── refuse_list.db
│   ├── state.db
│   └── votechannel.db
├── genesis.json
├── priv_validator.json
└── priv_validator.json.bak

12 directories, 4 files
```

### 3.4配置创世节点

- config.toml

  ```
  app_name = "evm"
  auth_by_ca = true
  block_size = 2000
  db_backend = "leveldb"
  environment = "production"
  fast_sync = true
  log_path = ""
  moniker = "anonymous"
  non_validator_auth_by_ca = false
  non_validator_node_auth = false
  p2p_laddr = "tcp://x.x.x.x:46656"
  rpc_laddr = "tcp://0.0.0.0:46657"
  seeds = "x.x.x.x:46656,x.x.x.x:46656,x.x.x.x:46656,x.x.x.x:46656,x.x.x.x:46656,x.x.x.x:46656,x.x.x.x:46656"
  signbyca = "8B27A3BDAF3FD47E4C143303BD030D6D02EECC6925C539031D8A5B157D4FC07B9E77D7E63EA795ECADDB39D7EA4511BE712210EB6D91D087A22717EFA1D2FA00"
  skip_upnp = true
  threshold_blocks = 0
  timeout_commit = 2000
  timeout_precommit = 2000
  timeout_precommit_delta = 1000
  timeout_prevote = 2000
  timeout_prevote_delta = 1000
  timeout_propose = 3000
  timeout_propose_delta = 1000
  tracerouter_msg_ttl = 5
  ```

- priv_validator.json

  ```
  {
  	//节点公钥
          "pub_key": [
  	//公钥加密算法，代表ED25519，暂不支持修改
                  1,
                  "D0425EECB2B0A2080C164FD7665CC6DA7B9F9ECE676B1DD27B6492FF599C85BA"
          ],
  	//共识状态，无需修改
          "last_height": 455838,
  	//共识状态，无需修改
          "last_round": 0,
  	//共识状态，无需修改
          "last_step": 3,
  	//共识状态，无需修改
          "last_signature": [
                  1,
                  "A0297E209C88D1B018AED59B4450698B943BF7AD960278F0DEE6AF1DB8A1DE97C194452B850EAC47AE6C2B52040E4E5EDE8D3C77B6133B8DBAD22168D947FA0F
  "
          ],
  	//共识状态，无需修改
          "last_signbytes": "0A0E68656C6C6F2D616E6E636861696E124E0A14352F610F1E501AD05612CF6D530A9388F760C555189EE91B280232300A149124A24CE66BB4A02B
  EDB26AAB8B46D4B794FF771218080112143BE53F7892B1D4D82F8595E15D75E730EB3643C9",
  	//节点私钥
          "priv_key": [
                  1,
                  "84E43AD6CA3C5F71FE7C321D8D1D553A087802EC596E37727DAC513F6FB0F302D0425EECB2B0A2080C164FD7665CC6DA7B9F9ECE676B1DD27B6492FF599C85BA
  "
          ]
  }
  ```

- genesis.json

  ```
  {
  	//创世时间
          "genesis_time": "0001-01-01T00:00:00Z",  
  	//链ID        
  	"chain_id": "hello-annchain",
  	//验证节点数组        
  	"validators": [
                  {
  		      //节点公钥
                          "pub_key": [
                                  1,
                              "D0425EECB2B0A2080C164FD7665CC6DA7B9F9ECE676B1DD27B6492FF599C85BA"
                          ],
  		      //权重
                          "amount": 100,
  		      //无用
                          "name": "",
                           //是否是CA节点，auth_by_ca=true时有效
                          "is_ca": true
                  }
          ],
          //自定义起始的state状态
          "app_hash": "",
          //插件
          "plugins": "specialop,querycache"
  }
  ```

### 3.5启动链节点

- 命令行启动

  `nohup ./genesis run & //更多参数请查看 ./genesis -h` 

