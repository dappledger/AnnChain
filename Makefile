.PHONY: ann api all
all: genesis 

#
ver=github.com/dappledger/AnnChain/chain/types.commitVer=`git rev-parse --short=8 HEAD`
tag=github.com/dappledger/AnnChain/chain/types.Version=`git describe --exact-match --tags $(git log -n1 --pretty='%h')`
gVer=github.com/dappledger/AnnChain/chain/types.gversion=`go run version/gemmill.go`
balance=github.com/dappledger/AnnChain/chain/types.lversion=balance
evmBalance=github.com/dappledger/AnnChain/eth/core.EVM_LIMIT_TYPE=balance
tx=github.com/dappledger/AnnChain/chain/types.lversion=tx
evmTx=github.com/dappledger/AnnChain/eth/core.EVM_LIMIT_TYPE=tx

#
rtoolver=github.com/dappledger/AnnChain/cmd/client/commands.VERSION=`git rev-parse --short=8 HEAD`

genesis:
	go build -ldflags "-X $(ver) -X $(tag)"  -o ./build/genesis ./cmd/genesis
limit_balance_genesis:
	go build -ldflags "-X $(ver) -X $(tag) -X $(gVer) -X $(balance) -X $(evmBalance)"  -o ./build/genesis ./cmd/genesis
limit_tx_genesis:
	go build -ldflags "-X $(ver) -X $(tag) -X $(gVer) -X $(tx) -X $(evmTx)"  -o ./build/genesis ./cmd/genesis




clean:
	rm ./genesis*

test:
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
