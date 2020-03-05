#!/usr/bin/env bash
export GO111MODULE=on
#install protoc
#Protocol Compiler Installation: https://github.com/protocolbuffers/protobuf/releases
#install tools
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go get -u github.com/gogo/protobuf/protoc-gen-gofast
go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc

gogoPath=`go list -f '{{ .Dir }}' -m github.com/gogo/protobuf`
grpcGateway=`go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway`
googLeapisPath=$grpcGateway/third_party/googleapis

rpcFile=chain/proto/grpc.proto

generate go files from proto types
for file in `ls gemmill/protos/*/*.proto `
do
   echo $file
   protoc -I .  --gofast_out=plugins=grpc,paths=source_relative:.  $file
done

#generate grpc server with swagger docs
echo $rpcFile
protoc -I . -I$gogoPath  -I$grpcGateway -I$googLeapisPath --gofast_out=plugins=grpc,paths=source_relative:.  \
	--swagger_out=logtostderr=true:.  --grpc-gateway_out=logtostderr=true,paths=source_relative:. $rpcFile

mv chain/proto/grpc.swagger.json docs/

#generate grpc markdown docs
protoc -I . -I$gogoPath  -I$grpcGateway -I$googLeapisPath --doc_out=./docs  --doc_opt=docs/resource/markdown.tmpl,api.md chain/proto/grpc.proto `ls gemmill/protos/*/*.proto`

