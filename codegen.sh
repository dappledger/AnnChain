#!/usr/bin/env bash

gogoPath=`go list -f '{{ .Dir }}' -m github.com/gogo/protobuf`
grpcGateway=`go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway`
googLeapisPath=$grpcGateway/third_party/googleapis

rpcFile=chain/proto/grpc.proto

#generate go files from proto types
for file in `ls gemmill/protos/*/* `
do
    if [[ $file == *.proto ]]; then
       echo $file
       protoc -I .  --gofast_out=plugins=grpc,paths=source_relative:.  $file
    fi
done

#generate grpc server
echo $rpcFile
protoc -I . -I$gogoPath  -I$grpcGateway -I$googLeapisPath --gofast_out=plugins=grpc,paths=source_relative:.  \
	--swagger_out=logtostderr=true:.  --grpc-gateway_out=logtostderr=true,paths=source_relative:. $rpcFile

