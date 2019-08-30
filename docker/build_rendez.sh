#!/bin/sh

rm Dockerfile
echo 'FROM alpine:latest
RUN apk add --no-cache libc6-compat
COPY ./genesis /bin/

WORKDIR /genesis

ENTRYPOINT ["genesis"]' > Dockerfile

docker build -t genesis:latest -f Dockerfile ../build
docker tag genesis:latest annchain/genesis:latest
rm Dockerfile