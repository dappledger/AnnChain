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
