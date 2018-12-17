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
| 操作系统 |          | CentOS（7以上 64位）或者Ubuntu（16.04 64位） |
|          |          |                                              |

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

`echo "GOPATH=\~/go" >> ~/.bash_profile`

变量生效

`source ~/.bash_profile`

### 1.6安装Docker及Docker-compose

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

## 第二章 工程环境配置及编译

### 2.1克隆工程

```
git clone https://github.com/dappledger/AnnChain.git ${GOPATH}/src/github.com/dappledger/AnnChain
cd ${GOPATH}/src/github.com/dappledger/AnnChain/genesis
```

### 2.2编译工程

编译成功，在./genesis/build目录中看到生成的二进制命令文件。

`make`

### 2.3编译环境变量

将编译成功后的命令目录配置到系统环境中，以便在后续操作过程的命令窗口调用。

`echo "export PATH=\$PATH:\$GOPATH/src/github.com/dappledger/AnnChain/genesis/build/genesis" >> ~/.bash_profile`

`source ~/.bash_profile`

## 第三章 链单节点部署

### 3.1克隆源码

同2.1

### 3.2编译源码

同2.2

### 3.3配置运行目录

```
echo export ANN_RUNTIME=/data/genesis >> /etc/profile  //path customize
source /etc/profile
```

### 3.4初始化链节点

`./genesis init `

初始化目录结构

```
[root@bogon .ann_runtime]# pwd    //default run dir
/root/.ann_runtime
[root@bogon .ann_runtime]# tree -L 2
.
├── addrbook.json
├── addrbook.json.bak
├── config.json
├── config.toml
├── data
│   ├── blockstore.db
│   ├── chaindata
│   ├── cs.wal
│   ├── genesisop.db
│   ├── genesisquety.db
│   ├── mempool.wal
│   ├── refuse_list.db
│   └── state.db
├── genesis.json
├── priv_validator.json
└── priv_validator.json.bak

7 directories, 9 files
```

### 3.5配置创世节点

- config.toml

  ```
  environment = "production"  //指定环境
  node_laddr = "tcp://0.0.0.0:80"  //p2p地址
  rpc_laddr = "tcp://0.0.0.0:81"   //rpc地址
  moniker = "anonymous"  //p2p匿名连接
  fast_sync = true  //快速同步数据
  db_backend = "leveldb" //链底层数据库
  seeds = "127.0.0.1:80,127.0.0.1:80"  //节点之间p2p连接
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

### 3.6启动链节点

- 命令行启动

  `nohup ./genesis node &`

- 脚本启动

  `./Genesis_service.sh start`

  ```
  Please enter the Chain-ID:test
  Please enter the P2P Listen on port:80
  Please enter the RPC Listen on port:81
  Please enter the SEEDS P2P NODE format(IP:PORT):127.0.0.1:80
  
               _                  ____            _     _
              / \    |\    |\    |    \|    |    / \    |\    |
             /   \   | \   | \   |     |    |   /   \   | \   |
            / ___ \  |  \  |  \  |     |____|  /_____\  |  \  |
           /       \ |   \ |   \ |     |    | /       \ |   \ |
          /         \|    \|    \|____/|    |/         \|    \|
  
  
          Genesis has been successfully built. 00:11:19
  
          To verify your installation run the following commands:
          For more information:
          AnnChain website: www.annchain.io/#/
          AnnChain Telegram channel @ www.annchain.io/#/news
          AnnChain resources: https://github.com/dappledger/AnnChain
          AnnChain Stack Exchange: https://...
          AnnChain wiki: https://github.com/dappledger/AnnChain/blob/master/README.md
  ```

## 第四章 Docker部署链节点

### 4.1配置Dockerfile文件

```
#Build Genesis in a stock Go builder container
FROM golang:latest as builder
MAINTAINER lvguoxin "lvguoxinlinux@163.com"
RUN apt-get update \
    && apt-get -y install net-tools \
    && apt-get -y install vim
WORKDIR $GOPATH/src/github.com/dappledger/AnnChain/genesis
ADD . $GOPATH/src/github.com/dappledger/AnnChain/genesis
RUN make
RUN ./build/genesis init

EXPOSE 46656 46657 46658

ENTRYPOINT [ "./build/genesis" ]
```

### 4.2制作docker镜像

`docker build -t annchain.io/genesis:v1.0 .`

### 4.3运行docker节点

`docker run --name node1 -d annchain.io/genesis:v1.0 node`

### 4.4查看docker运行状态

```
[root@bogon goproject]# docker ps
CONTAINER ID        IMAGE                         COMMAND                  CREATED             STATUS              PORTS               NAMES
b002176a0962        annchain.io/genesis:v1.0   "./build/genesis node -..."   8 days ago          Up 21 minutes       46656-46658/tcp     node1
```

## 第五章 自动化脚本部署链节点

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

VERSION=1.0.0
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
    git clone https://github.com/dappledger/AnnChain.git ${GOPATH}/src/github.com/dappledger/AnnChain
    cd ${GOPATH}/src/github.com/dappledger/AnnChain/genesis
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
## 第六章 链启停脚本

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
GENESIS_EX_PATH=${GOPATH}/src/gitlab.zhonganinfo.com/tech_bighealth/za-delos/build
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
