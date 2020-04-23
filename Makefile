.PHONY: ann api all
all: genesis gtool 
test: test-eth test-gemmill

#
ver=github.com/dappledger/AnnChain/chain/types.commitVer=`git rev-parse --short=8 HEAD`
tag=github.com/dappledger/AnnChain/chain/types.Version=`git describe --tags --always $(git log -n1 --pretty='%h')`
#
gtoolver=github.com/dappledger/AnnChain/cmd/client/commands.VERSION=`git rev-parse --short=8 HEAD`

genesis:
	go build -ldflags "-X $(ver) -X $(tag)"  -o ./build/genesis ./cmd/genesis

gtool:
	go build -ldflags "-X $(ver) -X $(tag)" -o ./build/gtool ./cmd/client

clean:
	rm -rf build

test-eth:
	go test -v ./chain/app/evm
	go test -v ./eth/rlp
	go test -v ./eth/accounts/abi
	go test -v ./eth/common
	go test -v ./eth/crypto
	go test -v ./eth/ethdb
	go test -v ./eth/event
	go test -v ./eth/metrics
	go test -v ./eth/params
	go test -v ./eth/trie
	go test -v ./eth/core/vm
	go test -v ./eth/core/state
	go test -v ./eth/core/types
	go test -v ./eth/core/rawdb
	go test -v ./eth/core

test-gemmill:
	go test -v ./gemmill/blockchain
	go test -v ./gemmill/config
	go test -v ./gemmill/consensus
	go test -v ./gemmill/ed25519
	go test -v ./gemmill/ed25519/extra25519
	go test -v ./gemmill/go-crypto
	go test -v ./gemmill/go-utils
	go test -v ./gemmill/go-wire
	go test -v ./gemmill/go-wire/expr
	go test -v ./gemmill/modules/go-autofile
	go test -v ./gemmill/modules/go-clist
	go test -v ./gemmill/modules/go-common
	go test -v ./gemmill/modules/go-db
	go test -v ./gemmill/modules/go-events
	# 'go test flowrate' failed maybe your machine upgrade required.
	go test -v ./gemmill/modules/go-flowrate/flowrate
	go test -v ./gemmill/modules/go-log
	go test -v ./gemmill/modules/go-merkle
	go test -v ./gemmill/p2p
	go test -v ./gemmill/refuse_list
	go test -v ./gemmill/types
	go test -v ./gemmill/utils

image:
	docker build -t genesis:latest -f Dockerfile .
	docker tag genesis:latest annchain/genesis:latest

# docker build and run
fastrun:image
	docker-compose -f docker-compose.yaml up

clean_fastrun:
	docker-compose -f docker-compose.yaml stop
	docker-compose -f docker-compose.yaml rm