.PHONY: Annchain.genesis all
all: genesis

genesis:
	go build -ldflags "-X github.com/dappledger/AnnChain/genesis/chain/version.commitVer=`git rev-parse HEAD`" -o ./build/genesis ./chain
