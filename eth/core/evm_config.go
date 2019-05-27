package core

import (
	"github.com/dappledger/AnnChain/eth/common"

	"math/big"
)

type EvmLimitType string

const (
	EvmLimitTypeTx      EvmLimitType = "tx"
	EvmLimitTypeBalance EvmLimitType = "balance"
)

var EVM_LIMIT_TYPE string

func GetEvmLimitType() EvmLimitType {
	return EvmLimitType(EVM_LIMIT_TYPE)
}

var (
	//0x068399080b34d68e0d000c55332422d06e272472 is balance administrator's address, which private key is c06e79e61571529c74772faa498c344dbe7e0e2467ad9ad130af1256a24fed7f
	BalanceAdministrator = common.HexToAddress("0x068399080b34d68e0d000c55332422d06e272472")
	//0x370cda6854d0e478dfe443e276227a1bef50cda3 is tx administrator's address, which private key is 3181e9294660054a5a7b42c4b60ca9a6e0801ac04f019c1a0bb7054b11d26a75
	TxAdministrator = common.HexToAddress("0x370cda6854d0e478dfe443e276227a1bef50cda3")
	BurningAccount = common.HexToAddress("0x0000000000000000000000000000000000000000")
)

func DefaultGenesis() Genesis {
	balance := big.NewInt(0)
	balance.SetString("100000000000000000000000000000000000000000000000000000000000",10)
	g := Genesis{Alloc: GenesisAlloc{BalanceAdministrator: {Balance: balance}}}
	return g
}