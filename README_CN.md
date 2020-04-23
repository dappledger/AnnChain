# AnnChain/[English](README.md)

![banner](docs/img/ann.png)

<br/>

AnnChain 是众安科技的新一代联盟链的核心引擎，具有高安全、高性能、高可用特性。旨在为企业提供一个高集成的区块链系统。联盟的特性，使它非常适合商业组织之间的商业合作，以及安全高频金融事务,等要求很高的场景。已经在几十个实际的业务场景中得到了应用。

[![version](https://img.shields.io/github/v/tag/dappledger/AnnChain.svg?sort=semver)](https://github.com/dappledger/AnnChain/releases/latest)
[![API Reference](https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667)](https://godoc.org/github.com/dappledger/AnnChain)
[![Go version](https://img.shields.io/badge/go-1.12.0-blue.svg)](https://github.com/moovweb/gvm)
[![Go Report Card](https://goreportcard.com/badge/github.com/dappledger/AnnChain)](https://goreportcard.com/report/github.com/dappledger/AnnChain)
[![Travis](https://travis-ci.org/dappledger/AnnChain.svg?branch=master)](https://travis-ci.org/dappledger/AnnChain)
[![license](https://img.shields.io/github/license/dappledger/AnnChain.svg)](https://github.com/dappledger/AnnChain/blob/master/LICENSE)

| Branch | Tests                                                                                                                                                | Coverage                                                                                                                             |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| master | [![TravisCI](https://travis-ci.org/dappledger/AnnChain.svg?branch=master)](https://travis-ci.org/dappledger/AnnChain)  | [![codecov](https://codecov.io/gh/dappledger/AnnChain/branch/master/graph/badge.svg)](https://codecov.io/gh/dappledger/AnnChain) |



## Releases
自动构建的稳定 [发布版](https://github.com/dappledger/AnnChain/releases).



## 环境要求

| 要求 | 描述              |
| ----------- | ------------------ |
| Go 版本  | Go1.12.0 或更高 |


## 源码编译

``` shell
export GOPATH=$HOME/.gopkgs

git clone https://github.com/dappledger/AnnChain.git

cd AnnChain

./get_pkgs.sh

make
```

## 支持的共识

AnnChain 支持 bft 共识和 raft 共识，bft为默认共识。如果需要使用raft,可按如下操作。

##### 第一步, 在config.toml文件设设置共识为raft :

``` shell
consensus = "raft"
```

##### 然后, 在运行目录增加 raft 节点配置 文件 raft-cluster.json以四个节点为例):

``` shell
{
    "advertise": "ann7939-validator8fc99df2-2.default.svc.cluster.local:23000",
    "local": {
        "bind": "0.0.0.0:23000",
        "pub_key": [
            1,
            "35EC28D113DB8D057140F903BAB049770CABAD4C2838509602552511C3F2D2E3"
        ],
        "rpc": "ann7939-validator8fc99df2-2.default.svc.cluster.local:47000"
    },
    "peers": [
        {
            "bind": "ann7939-validator480649ca-0.default.svc.cluster.local:23000",
            "pub_key": [
                1,
                "7B788FD0A5A1504C438B2D6B5602717C07F5E82D25175B4065B75C46017B770D"
            ],
            "rpc": "ann7939-validator480649ca-0.default.svc.cluster.local:47000"
        },
        {
            "bind": "ann7939-validatorb14a47dc-1.default.svc.cluster.local:23000",
            "pub_key": [
                1,
                "1FE0A5560BB9376348CB8F218BDA2011280606571DB20B841FA9F7560143796D"
            ],
            "rpc": "ann7939-validatorb14a47dc-1.default.svc.cluster.local:47000"
        },
        {
            "bind": "ann7939-validator8fc99df2-2.default.svc.cluster.local:23000",
            "pub_key": [
                1,
                "35EC28D113DB8D057140F903BAB049770CABAD4C2838509602552511C3F2D2E3"
            ],
            "rpc": "ann7939-validator8fc99df2-2.default.svc.cluster.local:47000"
        },
        {
            "bind": "ann7939-validatore78bd527-3.default.svc.cluster.local:23000",
            "pub_key": [
                1,
                "3C521E9D3D942654FA1E6C52E7B3A4EDE059E047FB4DF4F00F04C092149002EA"
            ],
            "rpc": "10.103.237.176:47000"
        }
    ]
}
```

* advertise: advertise 广播地址用于节点之间相连

* local.bind:  raft 协议的本地绑定端口

* local.pub_key: 节点的公钥 类似与 pbft 公钥

* local.rpc: 节点的 rpc 绑定地址

* peers: 其他节点的绑定地址公钥信息，包括自己


## 快速入手

#### 单节点

``` shell
./build.sh genesis

./build/genesis init

./build/genesis run
```

#### 使用docker-compose的本地集群

``` shell
# docker build image and docker-compose run
make fastrun

# remove cluster
make clean_fastrun
```

## 用法

[命令行工具](docs/cmd_CN.md)
<br/>
[Golang SDK](https://github.com/dappledger/AnnChain-go-sdk)
<br/>
[Java SDK](https://github.com/dappledger/ann-java-sdk)


## 应用

- [浏览器](https://github.com/dappledger/ann-explorer)



## 贡献

如果您有任何问题，请[提交](https://github.com/dappledger/AnnChain/issues).
<br/>
如果喜欢贡献代码, 请先fork, 修复问题, 提交代码， 再发送[合并请求](https://github.com/dappledger/AnnChain/pulls) 供项目维护者审阅代码和合并。






