.PHONY: ann anntool vanntool test all
all: ann anntool vanntool test

ann:
	go build -ldflags "-X github.com/dappledger/AnnChain/src/chain/version.commitVer=`git rev-parse HEAD`" -o ./build/ann ./src/chain
anntool:
	go build -ldflags "-X github.com/dappledger/AnnChain/src/client/main.version=`git rev-parse HEAD`" -o ./build/anntool ./src/client
vanntool:
	go build  -o ./build/vanntool ./vanntool
test:
	 # todo go test all case
	#go test ./ann/dbs/refuse_list
	#go test ./ann/mempool
	#go test ./ann/consensus
	#go test ./ann/blockchain
	go test ./src/tools/state
proto:
	protoc --proto_path=$(GOPATH)/src --proto_path=src/chain/app/remote --go_out=plugins=grpc:src/chain/app/remote src/chain/app/remote/*.proto
	protoc --proto_path=$(GOPATH)/src --proto_path=src/example/types --go_out=plugins=grpc:src/example/types src/example/types/*.proto
	#protoc --proto_path=$(GOPATH)/src --proto_path=src/chain/node/protos --gofast_out=plugins=grpc:src/chain/node/protos src/chain/node/protos/*.proto
	protoc --proto_path=src/types --go_out=src/types src/types/*.proto
