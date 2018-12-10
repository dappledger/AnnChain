## Genesis Docker部署节点

本文档能够让初学者快速体验Genesis链，仅需要在安装了Docker的机器上clone了Genesis链的代码后运行本文档中的命令，即可启动Genesis节点。

推荐使用Docker version 1.13.1, build 8633870/1.13.1以上的版本，安装方法参照[官方文档](https://docs.docker.com/]%E6%88%96%E6%9C%AC%E6%96%87%E6%A1%A3%E9%99%84%E5%BD%95%E3%80%82)

#### 配置Dockerfile文件

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
注意dockerfile位置：$GOPATH/src/github.com/dappledger/AnnChain/genesis
```

#### 制作Docker镜像

```
docker build -t annchain.io/genesis:v1.0 .
```

#### 运行Docker节点

```
docker run --name node1 -d annchain.io/genesis:v1.0 node
```

#### 查看Docker运行状态

```
[root@bogon goproject]# docker ps
CONTAINER ID        IMAGE                         COMMAND                  CREATED             STATUS              PORTS               NAMES
b002176a0962        annchain.io/genesis:v1.0   "./build/genesis node -..."   8 days ago          Up 21 minutes       46656-46658/tcp     node1
```

