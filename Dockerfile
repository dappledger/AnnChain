# compile environment;
FROM golang:1.12-alpine as builder
#install libs
#you shold replace when you in china.
RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g" /etc/apk/repositories
RUN apk add build-base make git
#copy files;
ADD . /AnnChain
WORKDIR /AnnChain
RUN GO111MODULE="on" GOPROXY="https://goproxy.cn" make genesis

# package environment;
FROM alpine:latest
#you shold replace when you in china.
RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g" /etc/apk/repositories
RUN apk add libc6-compat
WORKDIR /genesis
COPY --from=builder /AnnChain/build/genesis /bin/
ENTRYPOINT ["genesis"]

