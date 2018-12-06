## 一、Genesis介绍

Genesis作为Annchain第一代公链系统，采用DPOS+PBFT混合共识，合约模块化，开发工具化的区块链基础设施。Genesis主要的目标是要能够以区块链技术为基础，承载众安科技的众多产品，快速开发部署区块链的应用模式。Genesis采用三层架构:协议层、扩展层和应用层，分别用以存储不可篡改的原始数据、实现各种功能（例如，智能合约）和运行各种应用（例如银行的移动应用）。每个区块链节点可以部署多个子链，子链承载不同的应用。目前，我们支持的应用有以太坊合约引擎和JVM合约引擎。
注：Annchain.Genesis的开源协议为GPL3.0，详情参考[LICENSES](https://github.com/dappledger/AnnChain/blob/master/LICENSE)

### 1.1本教程你能学到什么

- [ ] 了解annChain.Genesis基础概念
- [ ] 了解annChain.Genesis命令
- [ ] 如何配置工程开发环境
- [ ] 如何配置创世节点
- [ ] 了解节点的部署方式

### 1.2面向读者

- [ ] 大学生
- [ ] 区块链的技术爱好者
- [ ] 互联网工程师
- [ ] 区块链工程师
- [ ] 运维工程师

### 1.3涉及技术

- [ ] 版本管理：Git工具
- [ ] 编程语言：Golang
- [ ] 编程工具：Vscode/JetBrains GoLand/Atom/LiteIDE
- [ ] 操作系统：CentOS/Ubuntu/Mac/Windos
- [ ] 运维工具：Docker/Docker-compose/Docker-machine

## 二、前期环境准备

本章节主要介绍在部署annChain.Genesis链节点之前，预先安装一些需要涉及到的环境配置和工具：

- [ ] 版本管理工具Git；
- [ ] Golang环境配置；
- [ ] curl下载工具；
- [ ] Docker容器和Docker-compose工具；

### 2.1安装Git工具

Windows：

[download](https://git-scm.com/download/win)

Mac：

`brew install git`

CentOS:

`yum install -y git`

Ubuntu:

`apt-get install git`

### 2.2安装Goland

[download](https://plugins.jetbrains.com/go)

### 2.3安装并配置Golang环境

安装Golang

`brew install go  //Mac`

`apt-get install go //Ubuntu`

`yum install -y go //CentOS`

配置环境变量

`echo "GOPATH=\~/go" >> ~/.bash_profile`

变量生效

`source ~/.bash_profile`

### 2.4安装Docker及Docker-compose

Windows:

[download](https://docs.docker.com/docker-for-windows/install/)

Ubuntu:

```
#wget -qO- https://get.docker.com/ | sh
#sudo service docker start
#sudo curl -L "https://github.com/docker/compose/releases/download/1.22.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
```

Mac:

`brew cask install docker`

`brew install docker-compose`

## 三、工程环境配置及编译

### 3.1克隆工程

1. 创建工程

   `mkdir -p $GOPATH/src/github.com/dappledger/AnnChain/angine`

   `mkdir -p $GOPATH/src/github.com/dappledger/AnnChain/ann-module`

   `mkdir -p $GOPATH/src/github.com/dappledger/AnnChain/go-sdk`

   `mkdir -p $GOPATH/src/github.com/dappledger/AnnChain/genesis`

2. 克隆工程

   ```
   git clone http://github.com/dappledger/AnnChain/ann-module.git ${GOPATH}/src/github.com/dappledger/AnnChain/ann-module
   cd ${GOPATH}/src/github.com/dappledger/AnnChain/ann-module
   
   git clone http://github.com/dappledger/AnnChain/angine.git ${GOPATH}/src/github.com/dappledger/AnnChain/angine
   cd ${GOPATH}/src/github.com/dappledger/AnnChain/angine
   
   git clone http://github.com/dappledger/AnnChain/go-sdk.git ${GOPATH}/src/github.com/dappledger/AnnChain/go-sdk
   
   git clone http://github.com/dappledger/AnnChain/genesis.git ${GOPATH}/src/github.com/dappledger/AnnChain/genesis
   cd ${GOPATH}/src/github.com/dappledger/AnnChain/genesis
   ```

3. 切换版本

   注意：除了go-sdk是master主分支，其余的三个工程包括：angine、ann-module、genesis都是genesis分支。

   `git checkout -b genesis origin/genesis `

### 3.2编译工程

编译成功，在/genesis/build目录中看到生成的二进制命令文件。

`make`

### 3.3编译环境变量

将编译成功后的命令目录配置到系统环境中，以便在后续操作过程的命令窗口调用。

`echo "export PATH=\$PATH:\$GOPATH/src/github.com/dappledger/AnnChain/build/bin" >> ~/.bash_profile`

`source ~/.bash_profile`

## 四、链单节点部署

### 4.1克隆源码

同3.1

### 4.2编译源码

同3.2

### 4.3初始化链节点

`./genesis init `

初始化目录结构

```
[root@iZuf6do02x7n4nr9lv00ybZ annchain-data]# tree -L 2
.
├── addrbook.json
├── addrbook.json.bak
├── config.toml
├── data
│   ├── archive
│   ├── archive.db
│   ├── blockstore.db
│   ├── cs.wal
│   ├── eventcodebase.db
│   ├── eventstate.db
│   ├── eventwarehouse.db
│   ├── mempool.wal
│   ├── metropolis.db
│   ├── orgStatus.db
│   ├── query_cache
│   ├── refuse_list.db
│   ├── state.db
│   └── votechannel.db
├── genesis.json
├── priv_validator.json
├── priv_validator.json.bak
└── shards
    └── annchain-evm
17 directories, 6 files
```

### 4.4配置创世节点

- config.toml

  ```
  environment = "production"
  node_laddr = "tcp://0.0.0.0:80"
  rpc_laddr = "tcp://0.0.0.0:81"
  moniker = "anonymous"
  fast_sync = true
  db_backend = "leveldb"
  seeds = "127.0.0.1:80,127.0.0.1:80"
  signbyCA = ""
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

### 4.5启动链节点

- 命令行启动

  `nohup ./genesis node &`

- 脚本启动

  `./Genesis_service.sh start`

## 五、Docker部署链节点

### 5.1配置Dockerfile文件

```
FROM golang:latest as builder
MAINTAINER lvguoxin "lvguoxin@zhongan.io"
RUN apt-get update \
    && apt-get -y install net-tools \
    && apt-get -y install vim
WORKDIR $GOPATH/src/github.com/dappledger/AnnChain/genesis
ADD . $GOPATH/src/github.com/dappledger/AnnChain/genesis
RUN make
RUN ./build/genesis init

EXPOSE 46656 46657 46658

ENTRYPOINT [ "./build/genesis" ]
注意dockerfile位置：$GOPATH/src/github.com/dappledger/AnnChain/genesis
```

### 5.2制作docker镜像

`docker build -t annchain.io/genesis:v1.0 .`

### 5.3运行docker节点

`docker run --name node1 -d annchain.io/genesis:v1.0 node`

### 5.4查看docker运行状态

```
[root@bogon goproject]# docker ps
CONTAINER ID        IMAGE                         COMMAND                  CREATED             STATUS              PORTS               NAMES
b002176a0962        annchain.io/genesis:v1.0   "./build/genesis node -..."   8 days ago          Up 21 minutes       46656-46658/tcp     node1
```

## 六、自动化脚本部署链节点

`./Genesis_install.sh`

```
#!/bin/bash

##########################################################################
# This is the annChain.Genesis automated install script for Linux and Mac OS.
# This file was downloaded from https://github.com/dappledger/AnnChain
#
# Copyright (c) 2018, Respective Authors all rights reserved.
#
# After December 1, 2018 this software is available under the following terms:
# 
# The MIT License
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
#
# https://github.com/dappledger/AnnChain/blob/master/README.md
##########################################################################

VERSION=0.1
TIME_BEGIN=$( date -u +%s )

#install golang
function InstallGo()
{
    source /etc/os-release
    case $ID in
    debian|ubuntu|devuan)
        sudo apt-get install curl
        ;;
    centos|fedora|rhel)
        yumdnf="yum"
        if test "$(echo "$VERSION_ID >= 22" | bc)" -ne 0;then
            yumdnf="dnf"
        fi
        $yumdnf install -y curl
        ;;
    *)
        exit 1
        ;;
    esac

    echo -e "Please enter the genesis requirements above go 1.9:\c"
    read -e go_version

    curl -SL -o go.tar.gz https://www.golangtc.com/static/go/${go_version}/go${go_version}.linux-amd64.tar.gz && \
    tar -C  /usr/local/ -zxf go.tar.gz && \
    rm -rf go.tar.gz
    
    #GOPATH
    echo -e "Please enter the GOPATH dir:\c"
    read -e GOPATHDIR
    mkdir -p ${GOPATHDIR}
#config env
cat <<EOF >> /etc/profile
export GOROOT=/usr/local/go
export GOPATH=${GOPATHDIR}
export PATH=\$GOROOT/bin:\$PATH
EOF
    source /etc/profile
    go env
    go version
}

#install git 
function InstallGit()
{
    source /etc/os-release
    case $ID in
    debian|ubuntu|devuan)
        sudo apt-get install git
        ;;
    centos|fedora|rhel)
        yumdnf="yum"
        if test "$(echo "$VERSION_ID >= 22" | bc)" -ne 0;then
            yumdnf="dnf"
        fi
        $yumdnf install -y git
        ;;
    *)
        exit 1
        ;;
    esac 
}

#Determine if the command was executed successfully
function ExIsOk()
{
    if [ $? -ne 0 ];then
        echo "Command execution failed"
        exit 1
    fi
}

#build Genesis
function BuildGenesis()
{
    git clone http://github.com/dappledger/AnnChain/ann-module.git ${GOPATH}/src/github.com/dappledger/AnnChain/ann-module
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/ann-module
    git checkout -b genesis origin/genesis >/dev/null 2>&1
    git clone http://github.com/dappledger/AnnChain/angine.git ${GOPATH}/src/github.com/dappledger/AnnChain/angine
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/angine
    git checkout -b genesis origin/genesis >/dev/null 2>&1
    git clone http://github.com/dappledger/AnnChain/go-sdk.git ${GOPATH}/src/github.com/dappledger/AnnChain/go-sdk
    git clone http://github.com/dappledger/AnnChain/genesis.git ${GOPATH}/src/github.com/dappledger/AnnChain/genesis
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/genesis
    git checkout -b genesis origin/genesis >/dev/null 2>&1
    printf "\\n\\tBeginning build version: %s\\n" "${VERSION}"
    make
}

#init Genesis
function InitGenesis()
{
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/genesis/build
    ./genesis init >/dev/null 2>&1 &
}

#Configuration config.toml
function ConfigToml()
{
    GENESIS_PATH=$HOME/.ann_runtime
    echo -e "Please enter the Chain-ID:\c"
    read -e CHAINID
    sed -i 's/"annchain.*"/"'${CHAINID}'"/g' ${GENESIS_PATH}/genesis.json

    echo -e "Please enter the P2P Listen on port:\c"
    read -e NODE_LADDR
    sed -i "s/46656/${NODE_LADDR}/g" ${GENESIS_PATH}/config.toml

    echo -e "Please enter the RPC Listen on port:\c"
    read -e RPC_LADDR
    sed -i "s/46657/${RPC_LADDR}/g" ${GENESIS_PATH}/config.toml

    echo -e "Please enter the SEEDS P2P NODE format(IP:PORT):\c"
    read -e SEEDS
    sed -i 's/seeds = ""/seeds = "'${SEEDS}'"/g' ${GENESIS_PATH}/config.toml
    
}

#Run Genesis
function RunGenesis()
{
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/genesis/build
    nohup ./genesis node >/dev/null 2>&1  &
}

#Check that the Genesis process is successful
function CheckGenesisIsOK()
{
    txtbld=$(tput bold)
    bldred=${txtbld}$(tput setaf 2)
    txtrst=$(tput sgr0)
    TIME_END=$(( $(date -u +%s) - ${TIME_BEGIN} ))
    SERVICE_STATUS=`ps -ef | grep genesis | grep -v grep|wc -l`
    if [ $SERVICE_STATUS -gt 0 ];then
	#printf "\t#####################################################\v"
	printf "${bldred}\n"
        printf "\t     _                  ____            _     _\n"     
	printf "\t    / \    |\    |\    |    \|    |    / \    |\    |\n"
	printf "\t   /   \   | \   | \   |     |    |   /   \   | \   |\n"
	printf "\t  / ___ \  |  \  |  \  |     |____|  /_____\  |  \  |\n"
	printf "\t /       \ |   \ |   \ |     |    | /       \ |   \ |\n"
	printf "\t/         \|    \|    \|____/|    |/         \|    \|\n${txtrst}\n"
	#printf "\t#####################################################\v"
	printf "\\n\\tGenesis has been successfully built. %02d:%02d:%02d\\n\\n" $(($TIME_END/3600)) $(($TIME_END%3600/60)) $(($TIME_END%60))
	printf "\\tTo verify your installation run the following commands:\\n"
	printf "\\tFor more information:\\n"
	printf "\\tAnnChain website: www.annchain.io/#/\\n"
	printf "\\tAnnChain Telegram channel @ www.annchain.io/#/news\\n"
	printf "\\tAnnChain resources: https://github.com/dappledger/AnnChain\\n"
	printf "\\tAnnChain Stack Exchange: https://...\\n"
	printf "\\tAnnChain wiki: https://github.com/dappledger/AnnChain/blob/master/README.md\\n\\n\\n"
				
    else 
	echo "service not ok"
    fi
}

if ! [ -x "$(command -v git)" ];then
    InstallGit
else
    echo "Git has been installed!!!!"
fi

if ! [ -x "$(command -v go)" ];then
    InstallGo
else 
    echo "Golang has been installed !!!"
fi

BuildGenesis
InitGenesis
ConfigToml
RunGenesis
CheckGenesisIsOK
```

`./Genesis_service.sh`

```
#!/bin/sh
# chkconfig: 2345 64 36
# description: AnnChain.genesis startup
# Author:lvguoxin
# Time:2018-10-30 15:39:34
# Name:annchainService.sh
# Version:V1.0
# Description:This is a test script.

[ -f /etc/init.d/functions ] && source /etc/init.d/functions
GENESIS_EX_PATH=${GOPATH}/src/github.com/dappledger/AnnChain/genesis/build
RETURN_VALUE=0


#log failure output func
function LogFailureMsg()
{
    echo "Genesis service ERROR!$@"
}

#log success output func
function LogSuccessMsg()
{
    echo "Genesis service SUCCESS!$@"
}

#Genesis start service
function start()
{
    cd ${GENESIS_EX_PATH}
    echo "start Genesis service"
    nohup ./genesis node >/dev/null 2>&1  &
    if [ $? -ne 0 ];then
        LogFailureMsg
    else
        LogSuccessMsg
    fi
}

#Genesis stop service
function stop()
{   
    echo "Stop Genesis service"
    kill `ps -ef | grep genesis | grep -v grep|awk '{print $2}'`  >/dev/null 2>&1 &
    if [ $? -ne 0 ];then
        LogFailureMsg
    else
        LogSuccessMsg
    fi
}


case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    *)
	echo "Usage:$0{start|stop|restart}"
        exit 1
esac
exit $RETURN_VALUE
```

