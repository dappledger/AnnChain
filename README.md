# AnnChain

![banner](docs/img/ann.png)

<br/>

AnnChain is the core engine of the new generation alliance chain of Zhongan Science and Technology. It has the characteristics of high security, high performance and high availability. It aims to provide a tightly integrated block chain system for enterprises. It is very suitable for business cooperation among commercial organizations with alliance nature, and also for high-frequency financial transactions and security. A demanding scene. Dozens of actual business scenarios have been applied.

[![version](https://img.shields.io/github/v/tag/dappledger/AnnChain.svg?sort=semver)](https://github.com/dappledger/AnnChain/releases/latest)
[![API Reference](https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667)](https://godoc.org/github.com/dappledger/AnnChain)
[![Go version](https://img.shields.io/badge/go-1.12.0-blue.svg)](https://github.com/moovweb/gvm)
[![Go Report Card](https://goreportcard.com/badge/github.com/dappledger/AnnChain)](https://goreportcard.com/report/github.com/dappledger/AnnChain)
[![Travis](https://travis-ci.org/dappledger/AnnChain.svg?branch=master)](https://travis-ci.org/dappledger/AnnChain)
[![license](https://img.shields.io/github/license/dappledger/AnnChain.svg)](https://github.com/dappledger/AnnChain/blob/master/LICENSE)

| Branch | Tests                                                                                                                                                | Coverage                                                                                                                             |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| master | [![CircleCI](https://circleci.com/gh/dappledger/AnnChain/tree/master.svg?style=shield)](https://circleci.com/gh/dappledger/AnnChain/tree/master)  | [![codecov](https://codecov.io/gh/dappledger/AnnChain/branch/master/graph/badge.svg)](https://codecov.io/gh/dappledger/AnnChain) |



## Releases

Automated builds are available for stable [releases](https://github.com/dappledger/AnnChain/releases).



## Requirements

| Requirement | Notes              |
| ----------- | ------------------ |
| Go version  | Go1.12.0 or higher |


## Building the source 

``` shell
export GOPATH=$HOME/.gopkgs

git clone https://github.com/dappledger/AnnChain.git

cd AnnChain

./get_pkgs.sh

make
```

## Quick Start

#### Single node

``` shell
./build/genesis init

./build/genesis run
```

#### Local cluster using docker-compose

``` shell
## start cluster
➜  docker-compose up

## remove cluster
➜  docker-compose down
```

## Usage

[Command Tool](https://github.com/dappledger/AnnChain/tree/master/docs/cmd.md)
[Golang SDK](https://github.com/dappledger/AnnChain-go-sdk)
[Java SDK](https://github.com/dappledger/ann-java-sdk)


## Applications

- [Explorer](https://github.com/dappledger/ann-explorer)



## Contributing

If you have any questions,please [report](https://github.com/dappledger/AnnChain/issues).
<br/>
If you'd like to contribute code, please fork, fix, commit and send a [pull request](https://github.com/dappledger/AnnChain/pulls) for the maintainers to review and merge into the main code base






