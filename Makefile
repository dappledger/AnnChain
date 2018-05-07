.PHONY: ann anntool test all
all: ann anntool test

ann:
	go build -ldflags "-X github.com/dappledger/AnnChain/src/chain/version.commitVer=`git rev-parse HEAD`" -o ./build/ann ./src/chain
anntool:
	go build -ldflags "-X github.com/dappledger/AnnChain/src/client/main.version=`git rev-parse HEAD`" -o ./build/anntool ./src/client
test:
	 # todo go test all case
	#go test ./ann/dbs/refuse_list
	#go test ./ann/mempool
	#go test ./ann/consensus
	#go test ./ann/blockchain
	go test ./src/tools/state
